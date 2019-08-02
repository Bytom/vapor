package monitor

import (
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/vapor/toolbar/precog/config"
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
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Duration(m.cfg.CheckFreqSeconds) * time.Second)
	for ; true; <-ticker.C {
		// TODO: lock?
		m.monitorRountine()
	}
}

func (m *monitor) updateBootstrapNodes() error {
	// TODO: updated existed nodes
	// TODO: add new nodes
	return nil
}

func (m *monitor) monitorRountine() error {
	// TODO: dail
	// TODO: get blockhash by height
	// TODO: update
	return nil
}
