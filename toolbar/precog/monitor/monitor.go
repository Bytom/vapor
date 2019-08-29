package monitor

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	vaporCfg "github.com/vapor/config"
	"github.com/vapor/crypto/ed25519/chainkd"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/errors"
	"github.com/vapor/event"
	"github.com/vapor/netsync/chainmgr"
	"github.com/vapor/netsync/consensusmgr"
	"github.com/vapor/netsync/peers"
	"github.com/vapor/p2p"
	"github.com/vapor/p2p/discover/dht"
	"github.com/vapor/p2p/discover/mdns"
	"github.com/vapor/p2p/signlib"
	"github.com/vapor/test/mock"
	"github.com/vapor/toolbar/precog/config"
)

type monitor struct {
	*sync.RWMutex
	cfg     *config.Config
	db      *gorm.DB
	nodeCfg *vaporCfg.Config
	sw      *p2p.Switch
	privKey chainkd.XPrv
	chain   *mock.Chain
	txPool  *mock.Mempool
	// discvMap maps a node's public key to the node itself
	discvMap       map[string]*dht.Node
	dialCh         chan struct{}
	checkStatusCh  chan struct{}
	bestHeightSeen uint64
	peers          *peers.PeerSet
}

func NewMonitor(cfg *config.Config, db *gorm.DB) *monitor {
	dbPath, err := makePath()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("makePath")
	}

	nodeCfg := &vaporCfg.Config{
		BaseConfig: vaporCfg.DefaultBaseConfig(),
		P2P:        vaporCfg.DefaultP2PConfig(),
		Federation: vaporCfg.DefaultFederationConfig(),
	}
	nodeCfg.DBPath = dbPath
	nodeCfg.ChainID = "mainnet"
	privKey, err := signlib.NewPrivKey()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("NewPrivKey")
	}

	chain, txPool, err := mockChainAndPool()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("mockChainAndPool")
	}

	return &monitor{
		RWMutex:        &sync.RWMutex{},
		cfg:            cfg,
		db:             db,
		nodeCfg:        nodeCfg,
		privKey:        privKey.(chainkd.XPrv),
		chain:          chain,
		txPool:         txPool,
		discvMap:       make(map[string]*dht.Node),
		dialCh:         make(chan struct{}, 1),
		checkStatusCh:  make(chan struct{}, 1),
		bestHeightSeen: uint64(0),
	}
}

func makePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	dataPath := usr.HomeDir + "/.vapor/precog"
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return "", err
	}

	return dataPath, nil
}

func (m *monitor) Run() {
	var seeds []string
	for _, node := range m.cfg.Nodes {
		seeds = append(seeds, fmt.Sprintf("%s:%d", node.IP, node.Port))
		if err := m.upSertNode(&node); err != nil {
			log.WithFields(log.Fields{"node": node, "err": err}).Error("upSertNode")
		}
	}
	m.nodeCfg.P2P.Seeds = strings.Join(seeds, ",")
	if err := m.makeSwitch(); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("makeSwitch")
	}

	m.dialCh <- struct{}{}
	go m.discoveryRoutine()
	go m.connectNodesRoutine()
	go m.checkStatusRoutine()
}

func (m *monitor) makeSwitch() error {
	l, listenAddr := p2p.GetListener(m.nodeCfg.P2P)
	discv, err := dht.NewDiscover(m.nodeCfg, m.privKey, l.ExternalAddress().Port, m.cfg.NetworkID)
	if err != nil {
		return err
	}

	// no need for lanDiscv, but passing &mdns.LANDiscover{} will cause NilPointer
	lanDiscv := mdns.NewLANDiscover(mdns.NewProtocol(), int(l.ExternalAddress().Port))
	m.sw, err = p2p.NewSwitch(m.nodeCfg, discv, lanDiscv, l, m.privKey, listenAddr, m.cfg.NetworkID)
	if err != nil {
		return err
	}

	m.peers = peers.NewPeerSet(m.sw)
	if err := m.prepareReactors(m.peers); err != nil {
		return errors.Wrap(err, "prepareReactors")
	}

	return nil
}

func (m *monitor) prepareReactors(peers *peers.PeerSet) error {
	dispatcher := event.NewDispatcher()
	// add ConsensusReactor for consensusChannel
	_ = consensusmgr.NewManager(m.sw, m.chain, peers, dispatcher)
	fastSyncDB := dbm.NewDB("fastsync", m.nodeCfg.DBBackend, m.nodeCfg.DBDir())
	// add ProtocolReactor to handle msgs
	if _, err := chainmgr.NewManager(m.nodeCfg, m.sw, m.chain, m.txPool, dispatcher, peers, fastSyncDB); err != nil {
		return err
	}

	for label, reactor := range m.sw.GetReactors() {
		log.WithFields(log.Fields{"label": label, "reactor": reactor}).Debug("start reactor")
		if _, err := reactor.Start(); err != nil {
			return nil
		}
	}

	m.sw.GetSecurity().RegisterFilter(m.sw.GetNodeInfo())
	m.sw.GetSecurity().RegisterFilter(m.sw.GetPeers())
	return m.sw.GetSecurity().Start()
}

func (m *monitor) checkStatusRoutine() {
	for range m.checkStatusCh {
		for _, peer := range m.sw.GetPeers().List() {
			peer.Start()
			m.peers.AddPeer(peer)
		}
		log.WithFields(log.Fields{"num": len(m.sw.GetPeers().List()), "peers": m.sw.GetPeers().List()}).Info("connected peers")

		for _, peer := range m.sw.GetPeers().List() {
			p := m.peers.GetPeer(peer.ID())
			if p == nil {
				continue
			}

			if err := p.SendStatus(m.chain.BestBlockHeader(), m.chain.LastIrreversibleHeader()); err != nil {
				log.WithFields(log.Fields{"peer": p, "err": err}).Error("SendStatus")
				m.peers.RemovePeer(p.ID())
			}
		}

		for _, peerInfo := range m.peers.GetPeerInfos() {
			if peerInfo.Height > m.bestHeightSeen {
				m.bestHeightSeen = peerInfo.Height
			}
		}
		log.WithFields(log.Fields{"bestHeight": m.bestHeightSeen}).Info("peersInfo")
		m.processPeerInfos(m.peers.GetPeerInfos())

		for _, peer := range m.sw.GetPeers().List() {
			p := m.peers.GetPeer(peer.ID())
			if p == nil {
				continue
			}

			m.peers.RemovePeer(p.ID())
		}
		log.Info("Disonnect all peers.")

		m.Unlock()
		m.dialCh <- struct{}{}
	}
}
