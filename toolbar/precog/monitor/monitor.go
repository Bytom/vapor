package monitor

import (
	"io/ioutil"
	// "os"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	// dbm "github.com/vapor/database/leveldb"

	cfg "github.com/vapor/config"
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
	dirPath *os.File
}

func NewMonitor(cfg *config.Config, db *gorm.DB) *monitor {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		log.Fatal(err)
	}

	return &monitor{
		cfg:     cfg,
		db:      db,
		dirPath: dirPath,
	}
}

func (m *monitor) Run() {
	defer os.RemoveAll(m.dirPath)

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
	// TODO: 包一下？  common cfg 之类的？
	mCfg := &cfg.Config{
		BaseConfig: cfg.DefaultBaseConfig(),
		P2P:        cfg.DefaultP2PConfig(),
		Federation: cfg.DefaultFederationConfig(),
	}
	// TODO: fix
	mCfg.DBPath = m.dirPath

	// swPrivKey, err := signlib.NewPrivKey()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// testDB := dbm.NewDB("testdb", "leveldb", dirPath)
	// TODO: clean up
	// log.Println("Federation.Xpubs", mCfg.Federation.Xpubs)
	sw, err := p2p.NewSwitch(mCfg)
	if err != nil {
		log.Fatal(err)
	}

	sw.Start()
	defer sw.Stop()
}

func (m *monitor) makeSwitch() {

}

func (m *monitor) monitorRountine() error {
	// TODO: dail nodes, get lantency & best_height
	// TODO: decide check_height("best best_height" - "confirmations")
	// TODO: get blockhash by check_height, get latency
	// TODO: update lantency, active_time and status
	return nil
}
