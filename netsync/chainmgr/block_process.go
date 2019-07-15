package chainmgr

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/errors"
	"github.com/vapor/netsync/peers"
	"github.com/vapor/p2p/security"
)

var errOrphanBlock = errors.New("fast sync inserting orphan block")

type BlockProcessor interface {
	process(chan struct{}, chan struct{}, uint64, *sync.WaitGroup)
}

type blockProcessor struct {
	chain   Chain
	storage Storage
	peers   *peers.PeerSet
}

func newBlockProcessor(chain Chain, storage Storage, peers *peers.PeerSet) *blockProcessor {
	return &blockProcessor{
		chain:   chain,
		peers:   peers,
		storage: storage,
	}
}

func (bp *blockProcessor) insert(blockStorage *blockStorage) error {
	isOrphan, err := bp.chain.ProcessBlock(blockStorage.block)
	if isOrphan {
		bp.peers.ProcessIllegal(blockStorage.peerID, security.LevelMsgIllegal, errOrphanBlock.Error())
		return errOrphanBlock
	}

	if err != nil {
		bp.peers.ProcessIllegal(blockStorage.peerID, security.LevelMsgIllegal, err.Error())
	}
	return err
}

func (bp *blockProcessor) process(downloadNotifyCh chan struct{}, ProcessStop chan struct{}, startHeight uint64, wg *sync.WaitGroup) {
	defer func() {
		close(ProcessStop)
		wg.Done()
	}()

	nextHeight := startHeight + 1
	for {
		for {
			block, err := bp.storage.readBlock(nextHeight)
			if err != nil {
				break
			}

			if err := bp.insert(block); err != nil {
				log.WithFields(log.Fields{"module": logModule, "err": err}).Error("failed on process block")
				return
			}

			bp.storage.deleteBlock(nextHeight)
			nextHeight++
		}

		if _, ok := <-downloadNotifyCh; !ok {
			return
		}
	}
}
