package monitor

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	// dbm "github.com/vapor/database/leveldb"

	vaporCfg "github.com/vapor/config"
	"github.com/vapor/p2p"
	// conn "github.com/vapor/p2p/connection"
	// "github.com/vapor/p2p/signlib"
	// "github.com/vapor/consensus"
	"github.com/vapor/toolbar/precog/config"
	"github.com/vapor/toolbar/precog/database/orm"
)

type monitor struct {
	cfg     *config.Config
	db      *gorm.DB
	nodeCfg *vaporCfg.Config
}

func NewMonitor(cfg *config.Config, db *gorm.DB) *monitor {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		log.Fatal(err)
	}

	nodeCfg := &vaporCfg.Config{
		BaseConfig: vaporCfg.DefaultBaseConfig(),
		P2P:        vaporCfg.DefaultP2PConfig(),
		Federation: vaporCfg.DefaultFederationConfig(),
	}
	nodeCfg.DBPath = dirPath

	return &monitor{
		cfg:     cfg,
		db:      db,
		nodeCfg: nodeCfg,
	}
}

func (m *monitor) Run() {
	defer os.RemoveAll(m.nodeCfg.DBPath)

	m.updateBootstrapNodes()
	go m.discovery()
	ticker := time.NewTicker(time.Duration(m.cfg.CheckFreqSeconds) * time.Second)
	for ; true; <-ticker.C {
		// TODO: lock?
		m.monitorRountine()
	}
}

// create or update: https://github.com/jinzhu/gorm/issues/1307
func (m *monitor) updateBootstrapNodes() {
	for _, node := range m.cfg.Nodes {
		ormNode := &orm.Node{
			PublicKey: node.PublicKey.String(),
			Alias:     node.Alias,
			Host:      node.Host,
			Port:      node.Port,
		}

		if err := m.db.Where(&orm.Node{PublicKey: ormNode.PublicKey}).
			Assign(&orm.Node{
				Alias: node.Alias,
				Host:  node.Host,
				Port:  node.Port,
			}).FirstOrCreate(ormNode).Error; err != nil {
			log.Error(err)
			continue
		}
	}
}

// TODO:
// implement logic first, and then refactor
// /home/gavin/work/go/src/github.com/vapor/
// p2p/test_util.go
// p2p/switch_test.go
func (m *monitor) discovery() {
	sw, err := m.makeSwitch()
	if err != nil {
		log.Fatal(err)
	}

	sw.Start()
	defer sw.Stop()
}

func (m *monitor) calcNetID() (*p2p.Switch, error) {
	var data []byte
	var h [32]byte
	data = append(data, m.nodeCfg.GenesisBlock().Hash().Bytes()...)
	magic := make([]byte, 8)
	magicNumber := uint64(0x054c5638)
	binary.BigEndian.PutUint64(magic, magicNumber)
	data = append(data, magic[:]...)
	sha3pool.Sum256(h[:], data)
	return binary.BigEndian.Uint64(h[:8])
}

func (m *monitor) makeSwitch() (*p2p.Switch, error) {
	// TODO: 包一下？  common cfg 之类的？

	var err error
	var l Listener
	var listenAddr string
	var discv *dht.Network
	var lanDiscv *mdns.LANDiscover

	// swPrivKey, err := signlib.NewPrivKey()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// TODO: whatz that for
	// testDB := dbm.NewDB("testdb", "leveldb", dirPath)
	// TODO: clean up
	// log.Println("Federation.Xpubs", mCfg.Federation.Xpubs)
	sw, err := p2p.NewSwitch(mCfg)
	if err != nil {
		return nil, err
	}

	return sw, nil
}

func (m *monitor) monitorRountine() error {
	// TODO: dail nodes, get lantency & best_height
	// TODO: decide check_height("best best_height" - "confirmations")
	// TODO: get blockhash by check_height, get latency
	// TODO: update lantency, active_time and status
	return nil
}
