package monitor

import (
	"net"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vapor/p2p"
	"github.com/vapor/toolbar/precog/database/orm"
)

func (m *monitor) connectionRoutine() {
	// TODO: fix
	// ticker := time.NewTicker(time.Duration(m.cfg.CheckFreqMinutes) * time.Minute)
	ticker := time.NewTicker(time.Duration(m.cfg.CheckFreqMinutes) * time.Second)
	for ; true; <-ticker.C {
		m.Lock()

		if err := m.dialNodes(); err != nil {
			log.WithFields(log.Fields{"err": err}).Error("dialNodes")
		}
	}
}

func (m *monitor) dialNodes() error {
	log.Info("Start to reconnect to nodes...")
	var nodes []*orm.Node
	if err := m.db.Model(&orm.Node{}).Find(&nodes).Error; err != nil {
		return err
	}

	addresses := make([]*p2p.NetAddress, 0)
	for i := 0; i < len(nodes); i++ {
		address := p2p.NewNetAddressIPPort(net.ParseIP(nodes[i].IP), nodes[i].Port)
		addresses = append(addresses, address)
	}

	// connected peers will be skipped in switch.DialPeers()
	m.sw.DialPeers(addresses)
	log.Info("DialPeers done.")
	m.processDialResults()
	m.checkStatus()
	return nil
}

func (m *monitor) checkStatus() {
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
}
