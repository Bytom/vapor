package blockproposer

import (
	"encoding/hex"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/account"
	"github.com/vapor/config"
	"github.com/vapor/consensus"
	"github.com/vapor/event"
	"github.com/vapor/proposal"
	"github.com/vapor/protocol"
)

const (
	logModule = "blockproposer"
)

// BlockProposer propose several block in specified time range
type BlockProposer struct {
	sync.Mutex
	chain           *protocol.Chain
	accountManager  *account.Manager
	txPool          *protocol.TxPool
	started         bool
	quit            chan struct{}
	eventDispatcher *event.Dispatcher
}

// generateBlocks is a worker that is controlled by the proposeWorkerController.
// It is self contained in that it creates block templates and attempts to solve
// them while detecting when it is performing stale work and reacting
// accordingly by generating a new block template.  When a block is verified, it
// is submitted.
//
// It must be run as a goroutine.
func (b *BlockProposer) generateBlocks() {
	xpub := config.CommonConfig.PrivateKey().XPub()
	xpubStr := hex.EncodeToString(xpub[:])
	ticker := time.NewTicker(consensus.BlockTimeInterval * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-b.quit:
			return
		case <-ticker.C:
		}

		bestBlockHeader := b.chain.BestBlockHeader()
		bestBlockHash := bestBlockHeader.Hash()
		now := uint64(time.Now().UnixNano() / 1e6)
		base := max(now, bestBlockHeader.Timestamp)
		minTimeToNextBlock := consensus.BlockTimeInterval - now%consensus.BlockTimeInterval
		nextBlockTime := base + minTimeToNextBlock
		if nextBlockTime-now < consensus.BlockTimeInterval/10 {
			nextBlockTime += consensus.BlockTimeInterval
		}
		// nextBlockTime := uint64(time.Now().UnixNano() / 1e6)
		// if minNextBlockTime := bestBlockHeader.Timestamp + consensus.BlockTimeInterval; nextBlockTime < minNextBlockTime {
		// 	nextBlockTime = minNextBlockTime
		// }

		if isBlocker, err := b.chain.IsBlocker(&bestBlockHash, xpubStr, nextBlockTime); !isBlocker {
			log.WithFields(log.Fields{"module": logModule, "error": err, "pubKey": xpubStr}).Debug("fail on check is next blocker")
			continue
		}

		block, err := proposal.NewBlockTemplate(b.chain, b.txPool, b.accountManager, nextBlockTime)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "error": err}).Error("failed on create NewBlockTemplate")
			continue
		}

		isOrphan, err := b.chain.ProcessBlock(block)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "height": block.BlockHeader.Height, "error": err}).Error("proposer fail on ProcessBlock")
			continue
		}

		log.WithFields(log.Fields{"module": logModule, "height": block.BlockHeader.Height, "isOrphan": isOrphan, "tx": len(block.Transactions)}).Info("proposer processed block")
		// Broadcast the block and announce chain insertion event
		if err = b.eventDispatcher.Post(event.NewProposedBlockEvent{Block: *block}); err != nil {
			log.WithFields(log.Fields{"module": logModule, "height": block.BlockHeader.Height, "error": err}).Error("proposer fail on post block")
		}
	}
}

func max(x, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

// Start begins the block propose process as well as the speed monitor used to
// track hashing metrics.  Calling this function when the block proposer has
// already been started will have no effect.
//
// This function is safe for concurrent access.
func (b *BlockProposer) Start() {
	b.Lock()
	defer b.Unlock()

	// Nothing to do if the miner is already running
	if b.started {
		return
	}

	b.quit = make(chan struct{})
	go b.generateBlocks()

	b.started = true
	log.Infof("block proposer started")
}

// Stop gracefully stops the proposal process by signalling all workers, and the
// speed monitor to quit.  Calling this function when the block proposer has not
// already been started will have no effect.
//
// This function is safe for concurrent access.
func (b *BlockProposer) Stop() {
	b.Lock()
	defer b.Unlock()

	// Nothing to do if the miner is not currently running
	if !b.started {
		return
	}

	close(b.quit)
	b.started = false
	log.Info("block proposer stopped")
}

// IsProposing returns whether the block proposer has been started.
//
// This function is safe for concurrent access.
func (b *BlockProposer) IsProposing() bool {
	b.Lock()
	defer b.Unlock()

	return b.started
}

// NewBlockProposer returns a new instance of a block proposer for the provided configuration.
// Use Start to begin the proposal process.  See the documentation for BlockProposer
// type for more details.
func NewBlockProposer(c *protocol.Chain, accountManager *account.Manager, txPool *protocol.TxPool, dispatcher *event.Dispatcher) *BlockProposer {
	return &BlockProposer{
		chain:           c,
		accountManager:  accountManager,
		txPool:          txPool,
		eventDispatcher: dispatcher,
	}
}
