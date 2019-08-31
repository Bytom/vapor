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

func (m *monitor) upsertNode(node *config.Node) error {
	ormNode := &orm.Node{
		IP:   node.IP,
		Port: node.Port,
	}
	if err := m.db.Where(ormNode).First(ormNode).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	ormNode.PublicKey = node.PublicKey
	if node.XPub != nil {
		ormNode.Xpub = node.XPub.String()
		ormNode.PublicKey = fmt.Sprintf("%v", node.XPub.PublicKey().String())
	}
	return m.db.Save(ormNode).Error
}

func (m *monitor) processDialResults(peerList []*p2p.Peer) error {
	var ormNodes []*orm.Node
	if err := m.db.Model(&orm.Node{}).Find(&ormNodes).Error; err != nil {
		return err
	}

	addressMap := make(map[string]*orm.Node, len(ormNodes))
	for _, ormNode := range ormNodes {
		addressMap[fmt.Sprintf("%s:%d", ormNode.IP, ormNode.Port)] = ormNode
	}

	connMap := make(map[string]bool, len(ormNodes))
	// connected peers
	for _, peer := range peerList {
		connMap[peer.ListenAddr] = true
		if err := m.processConnectedPeer(addressMap[peer.ListenAddr]); err != nil {
			log.WithFields(log.Fields{"peer listenAddr": peer.ListenAddr, "err": err}).Error("processConnectedPeer")
		}
	}

	// offline peers
	for _, ormNode := range ormNodes {
		if _, ok := connMap[peer.ListenAddr]; ok {
			continue
		}

		if err := m.processOfflinePeer(ormNode); err != nil {
			log.WithFields(log.Fields{"peer publicKey": ormNode.PublicKey, "err": err}).Error("processOfflinePeer")
		}
	}

	return nil
}

func (m *monitor) processConnectedPeer(ormNode *orm.Node) error {
	ormNodeLiveness := &orm.NodeLiveness{NodeID: ormNode.ID}
	err := m.db.Preload("Node").Where(ormNodeLiveness).Last(ormNodeLiveness).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	ormNodeLiveness.PongTimes += 1
	if ormNode.Status == common.NodeOfflineStatus {
		ormNode.Status = common.NodeUnknownStatus
	}
	ormNodeLiveness.Node = ormNode
	return m.db.Save(ormNodeLiveness).Error
}

func (m *monitor) processOfflinePeer(ormNode *orm.Node) error {
	ormNode.Status = common.NodeOfflineStatus
	return m.db.Save(ormNode).Error
}

func (m *monitor) processPeerInfos(peerInfos []*peers.PeerInfo) {
	for _, peerInfo := range peerInfos {
		dbTx := m.db.Begin()
		if err := m.processPeerInfo(dbTx, peerInfo); err != nil {
			log.WithFields(log.Fields{"peerInfo": peerInfo, "err": err}).Error("processPeerInfo")
			dbTx.Rollback()
		} else {
			dbTx.Commit()
		}
	}
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

	if ormNode.Status == common.NodeOfflineStatus {
		return fmt.Errorf("node %s status error", ormNode.PublicKey)
	}

	log.WithFields(log.Fields{"ping": peerInfo.Ping}).Debug("peerInfo")
	ping, err := time.ParseDuration(peerInfo.Ping)
	if err != nil {
		return err
	}

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	var ormNodeLivenesses []*orm.NodeLiveness
	if err := dbTx.Preload("Node").Model(&orm.NodeLiveness{}).
		Where("node_id = ? AND updated_at >= ?", ormNode.ID, yesterday).
		Order(fmt.Sprintf("created_at %s", "DESC")).
		Find(&ormNodeLivenesses).Error; err != nil {
		return err
	}

	// update latest liveness
	latestLiveness := ormNodeLivenesses[0]
	lantencyMS := ping.Nanoseconds() / 1000
	if lantencyMS != 0 {
		ormNode.AvgLantencyMS = sql.NullInt64{
			Int64: (ormNode.AvgLantencyMS.Int64*int64(latestLiveness.PongTimes) + lantencyMS) / int64(latestLiveness.PongTimes+1),
			Valid: true,
		}
	}
	latestLiveness.PongTimes += 1
	if peerInfo.Height != 0 {
		latestLiveness.BestHeight = peerInfo.Height
	}
	if err := dbTx.Save(latestLiveness).Error; err != nil {
		return err
	}

	// calc LatestDailyUptimeMinutes
	total := 0 * time.Minute
	ormNodeLivenesses[0].UpdatedAt = now
	for _, ormNodeLiveness := range ormNodeLivenesses {
		if ormNodeLiveness.CreatedAt.Before(yesterday) {
			ormNodeLiveness.CreatedAt = yesterday
		}

		total += ormNodeLiveness.UpdatedAt.Sub(ormNodeLiveness.CreatedAt)
	}

	return dbTx.Model(&orm.Node{}).Where(&orm.Node{PublicKey: xPub.PublicKey().String()}).
		UpdateColumn(&orm.Node{
			Alias:                    peerInfo.Moniker,
			Xpub:                     peerInfo.ID,
			BestHeight:               peerInfo.Height,
			LatestDailyUptimeMinutes: uint64(total.Minutes()),
		}).First(ormNode).Error
}
