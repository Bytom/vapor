package monitor

import (
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/vapor/toolbar/precog/config"
	"github.com/vapor/toolbar/precog/database/orm"
)

type monitor struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewMonitor(cfg *config.Config, db *gorm.DB) *monitor {
	return &monitor{
		cfg: cfg,
		db:  db,
	}
}

func (m *monitor) Run() {
	if err := m.updateBootstrapNodes(); err != nil {
		log.Error(err)
	}

	go m.discovery()
	ticker := time.NewTicker(time.Duration(m.cfg.CheckFreqSeconds) * time.Second)
	for ; true; <-ticker.C {
		// TODO: lock?
		m.monitorRountine()
	}
}

// create or update: https://github.com/jinzhu/gorm/issues/1307
func (m *monitor) updateBootstrapNodes() error {
	for _, node := range m.cfg.Nodes {
		ormNode := &orm.Node{
			PubKey: node.PubKey.String(),
			Alias:  node.Alias,
			Host:   node.Host,
			Port:   node.Port,
		}
		if err := m.db.Where(&orm.Node{PubKey: ormNode.PubKey}).
			Assign(&orm.Node{
				Alias: node.Alias,
				Host:  node.Host,
				Port:  node.Port,
			}).FirstOrCreate(ormNode).Error; err != nil {
			return err
		}
	}

	return nil
}

// TODO:
func (m *monitor) discovery() {
}

func (m *monitor) monitorRountine() error {
	// TODO: dail, get lantency & best_height
	// TODO: decide check_height("best best_height" - "confirmations")
	// TODO: get blockhash by check_height, get latency
	// TODO: update lantency, active_time and status
	return nil
}
