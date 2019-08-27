package monitor

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/vapor/crypto/ed25519/chainkd"
	"github.com/vapor/netsync/peers"
	"github.com/vapor/p2p"
	"github.com/vapor/toolbar/precog/common"
	"github.com/vapor/toolbar/precog/config"
	"github.com/vapor/toolbar/precog/database/orm"
)

// create or update: https://github.com/jinzhu/gorm/issues/1307
func (m *monitor) upSertNode(node *config.Node) error {
	if node.XPub != nil {
		node.PublicKey = fmt.Sprintf("%v", node.XPub.PublicKey().String())
	}

	ormNode := &orm.Node{PublicKey: node.PublicKey}
	if err := m.db.Where(&orm.Node{PublicKey: node.PublicKey}).First(ormNode).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if node.Alias != "" {
		ormNode.Alias = node.Alias
	}
	if node.XPub != nil {
		ormNode.Xpub = node.XPub.String()
	}
	ormNode.Host = node.Host
	ormNode.Port = node.Port
	return m.db.Where(&orm.Node{PublicKey: ormNode.PublicKey}).
		Assign(&orm.Node{
			Xpub:  ormNode.Xpub,
			Alias: ormNode.Alias,
			Host:  ormNode.Host,
			Port:  ormNode.Port,
		}).FirstOrCreate(ormNode).Error
}

func (m *monitor) processDialResults() error {
	for _, peer := range m.sw.GetPeers().List() {
		if err := m.processDialResult(peer); err != nil {
			log.Error(err)
		}
	}
	return nil
}

// TODO: add start time here
func (m *monitor) processDialResult(peer *p2p.Peer) error {
	xPub := &chainkd.XPub{}
	if err := xPub.UnmarshalText([]byte(peer.Key)); err != nil {
		return err
	}

	ormNodeLiveness := &orm.NodeLiveness{}
	if err := m.db.Model(&orm.NodeLiveness{}).
		Joins("join nodes on nodes.id = node_livenesses.node_id").
		Where("nodes.public_key = ?", xPub.PublicKey().String()).First(ormNodeLiveness).Error; err != nil {
		return err
	}

	return nil
}

func (m *monitor) processPeerInfos(peerInfos []*peers.PeerInfo) error {
	for _, peerInfo := range peerInfos {
		dbTx := m.db.Begin()
		if err := m.processPeerInfo(dbTx, peerInfo); err != nil {
			log.Error(err)
			dbTx.Rollback()
		} else {
			dbTx.Commit()
		}
	}

	return nil
}

func (m *monitor) processPeerInfo(dbTx *gorm.DB, peerInfo *peers.PeerInfo) error {
	xPub := &chainkd.XPub{}
	if err := xPub.UnmarshalText([]byte(peerInfo.ID)); err != nil {
		return err
	}

	ormNode := &orm.Node{}
	if err := dbTx.Model(&orm.Node{}).Where(&orm.Node{PublicKey: xPub.PublicKey().String()}).First(ormNode).Error; err != nil {
		return err
	}

	log.Debugf("peerInfo.Ping: %v", peerInfo.Ping)
	ping, err := time.ParseDuration(peerInfo.Ping)
	if err != nil {
		log.Debugf("Parse ping time err: %v", err)
	}

	// TODO: preload?
	ormNodeLiveness := &orm.NodeLiveness{
		NodeID:        ormNode.ID,
		BestHeight:    ormNode.BestHeight,
		AvgLantencyMS: sql.NullInt64{Int64: ping.Nanoseconds() / 1000, Valid: true},
		// PingTimes     uint64
		// PongTimes     uint64
	}
	if err := dbTx.Model(&orm.NodeLiveness{}).Where("node_id = ? AND status != ?", ormNode.ID, common.NodeOfflineStatus).
		UpdateColumn(&orm.NodeLiveness{
			BestHeight:    ormNodeLiveness.BestHeight,
			AvgLantencyMS: ormNodeLiveness.AvgLantencyMS,
		}).FirstOrCreate(ormNodeLiveness).Error; err != nil {
		return err
	}

	if err := dbTx.Model(&orm.Node{}).Where(&orm.Node{PublicKey: xPub.PublicKey().String()}).
		UpdateColumn(&orm.Node{
			Alias:      peerInfo.Moniker,
			Xpub:       peerInfo.ID,
			BestHeight: peerInfo.Height,
			// LatestDailyUptimeMinutes uint64
		}).First(ormNode).Error; err != nil {
		return err
	}

	return nil
}
