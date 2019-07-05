package chainmgr

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/karalabe/cookiejar.v2/collections/prque"

	"github.com/vapor/errors"
	"github.com/vapor/netsync/peers"
	"github.com/vapor/p2p/security"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

const (
	maxParallelTasksNum  = 7
	blockProcessChSize   = 1024
	blocksProcessChSize  = 128
	headersProcessChSize = 1024
)

var (
	requireBlockTimeout         = 20 * time.Second
	requireHeadersTimeout       = 30 * time.Second
	requireBlocksTimeout        = 50 * time.Second
	parallelFetchHeadersTimeout = 50 * time.Second
	parallelFetchBlocksTimeout  = 200 * time.Second
)

type MsgFetcher interface {
	resetParameter()
	requireBlock(peerID string, height uint64) (*types.Block, error)
	parallelFetchBlocks(syncPeers []string, taskQueue *prque.Prque, downloadedBlockCh chan *downloadedBlock, downloadResult chan bool, ProcessResult chan bool, wg *sync.WaitGroup, num int)
	parallelFetchHeaders(peers []*peers.Peer, locator []*bc.Hash, stopHash *bc.Hash, skip uint64) (map[string][]*types.BlockHeader, error)
}

type msgFetcher struct {
	storage          Storage
	peers            *peers.PeerSet
	blockProcessCh   chan *blockMsg
	blocksProcessCh  chan *blocksMsg
	headersProcessCh chan *headersMsg
}

func newMsgFetcher(storage Storage, peers *peers.PeerSet) *msgFetcher {
	return &msgFetcher{
		storage:          storage,
		peers:            peers,
		blockProcessCh:   make(chan *blockMsg, blockProcessChSize),
		blocksProcessCh:  make(chan *blocksMsg, blocksProcessChSize),
		headersProcessCh: make(chan *headersMsg, headersProcessChSize),
	}
}

func (mf *msgFetcher) processBlock(peerID string, block *types.Block) {
	mf.blockProcessCh <- &blockMsg{block: block, peerID: peerID}
}

func (mf *msgFetcher) processBlocks(peerID string, blocks []*types.Block) {
	mf.blocksProcessCh <- &blocksMsg{blocks: blocks, peerID: peerID}
}

func (mf *msgFetcher) processHeaders(peerID string, headers []*types.BlockHeader) {
	mf.headersProcessCh <- &headersMsg{headers: headers, peerID: peerID}
}

func (mf *msgFetcher) requireBlock(peerID string, height uint64) (*types.Block, error) {
	peer := mf.peers.GetPeer(peerID)
	if peer == nil {
		return nil, errPeerDropped
	}

	if ok := peer.GetBlockByHeight(height); !ok {
		return nil, errPeerDropped
	}

	timeout := time.NewTimer(requireBlockTimeout)
	defer timeout.Stop()

	for {
		select {
		case msg := <-mf.blockProcessCh:
			if msg.peerID != peerID {
				continue
			}
			if msg.block.Height != height {
				continue
			}
			return msg.block, nil
		case <-timeout.C:
			return nil, errors.Wrap(errRequestTimeout, "requireBlock")
		}
	}
}

func (mf *msgFetcher) requireBlocks(peerID string, locator []*bc.Hash, stopHash *bc.Hash) error {
	peer := mf.peers.GetPeer(peerID)
	if peer == nil {
		return errPeerDropped
	}

	if ok := peer.GetBlocks(locator, stopHash); !ok {
		return errPeerDropped
	}

	return nil
}

func (mf *msgFetcher) requireHeaders(peerID string, locator []*bc.Hash, stopHash *bc.Hash, skip uint64) error {
	peer := mf.peers.GetPeer(peerID)
	if peer == nil {
		return errPeerDropped
	}

	if ok := peer.GetHeaders(locator, stopHash, skip); !ok {
		return errPeerDropped
	}

	return nil
}

func (mf *msgFetcher) parallelFetchBlocks(syncPeers []string, taskQueue *prque.Prque, downloadedBlockCh chan *downloadedBlock, downloadComplete chan bool, ProcessComplete chan bool, wg *sync.WaitGroup, num int) {
	defer wg.Done()

	timeout := time.NewTimer(requireBlocksTimeout)
	defer timeout.Stop()
	fetchBlocksTimeout := time.NewTimer(parallelFetchBlocksTimeout)
	defer fetchBlocksTimeout.Stop()

	tasks := newBlocksTasks()
	timeoutQueue := newTimeoutQueue(requireBlocksTimeout)
	for {
		if taskQueue.Size() == 0 && tasks.size() == 0 {
			downloadComplete <- true
			return
		}
		// schedule task
		for !taskQueue.Empty() && tasks.size() <= maxParallelTasksNum {
			task := taskQueue.PopItem().(*blocksTask)
			peerID, err := mf.peers.SelectPeer(syncPeers, task.stopHeader.Height+fastSyncPivotGap)
			if err != nil {
				if tasks.size() == 0 {
					downloadComplete <- true
					return
				}
				taskQueue.Push(task, -float32(task.index))
				log.WithFields(log.Fields{"module": logModule, "err": err}).Debug("failed on select valid peer")
				break
			}

			startHash := task.startHeader.Hash()
			stopHash := task.stopHeader.Hash()
			if err := mf.requireBlocks(peerID, []*bc.Hash{&startHash}, &stopHash); err != nil {
				taskQueue.Push(task, -float32(task.index))
				log.WithFields(log.Fields{"module": logModule, "err": err}).Error("failed on send require blocks msg")
				continue
			}

			tasks.addTask(peerID, task)
			timeoutQueue.addTimer(peerID)
			if d := timeoutQueue.getNextTimeoutDuration(); d != nil {
				timeout.Reset(*d)
			}
		}

		select {
		case msg := <-mf.blocksProcessCh:
			mf.peers.SetIdle(msg.peerID)
			//check message from the requested peer.
			task, ok := tasks.getTask(msg.peerID)
			if !ok {
				mf.peers.ErrorHandler(msg.peerID, security.LevelMsgIllegal, errors.New("get unsolicited blocks msg"))
				break
			}

			//reset timeout
			timeoutQueue.delTimer(msg.peerID)
			if d := timeoutQueue.getNextTimeoutDuration(); d != nil {
				timeout.Reset(*d)
			}

			if err := mf.verifyBlocksMsg(msg, task.startHeader, task.stopHeader); err != nil {
				mf.peers.ErrorHandler(msg.peerID, security.LevelMsgIllegal, err)
				taskQueue.Push(task, -float32(task.index))
				break
			}

			if err := mf.storage.WriteBlocks(msg.peerID, msg.blocks); err != nil {
				log.WithFields(log.Fields{"module": logModule, "error": err}).Info("write block error")
				downloadComplete <- true
				return
			}

			downloadedBlockCh <- &downloadedBlock{startHeight: msg.blocks[0].Height, stopHeight: msg.blocks[len(msg.blocks)-1].Height}
			tasks.delTask(msg.peerID)
			//unfinished task, continue
			if msg.blocks[len(msg.blocks)-1].Height < task.stopHeader.Height-1 {
				log.WithFields(log.Fields{"module": logModule, "task": task.index}).Info("task unfinished")
				task.startHeader = &msg.blocks[len(msg.blocks)-1].BlockHeader
				taskQueue.Push(task, -float32(task.index))
			}
		case <-timeout.C:
			peerID := timeoutQueue.getFirstTimeoutID()
			if peerID == nil {
				break
			}

			task, ok := tasks.getTask(*peerID)
			if !ok {
				break
			}
			timeoutQueue.delTimer(*peerID)
			//reset timeout
			if d := timeoutQueue.getNextTimeoutDuration(); d != nil {
				timeout.Reset(*d)
			}
			log.WithFields(log.Fields{"module": logModule, "peerID": peerID, "error": errRequestTimeout}).Info("failed on fetch blocks")
			mf.peers.ErrorHandler(*peerID, security.LevelConnException, errors.New("require blocks timeout"))
			taskQueue.Push(task, -float32(task.index))
			tasks.delTask(*peerID)
		case <-fetchBlocksTimeout.C:
			downloadComplete <- true
			return
		case <-ProcessComplete:
			return
		}
	}
}

func (mf *msgFetcher) parallelFetchHeaders(peers []*peers.Peer, locator []*bc.Hash, stopHash *bc.Hash, skip uint64) (map[string][]*types.BlockHeader, error) {
	timeout := time.NewTimer(requireHeadersTimeout)
	defer timeout.Stop()
	fetchHeadersTimeout := time.NewTimer(parallelFetchBlocksTimeout)
	defer fetchHeadersTimeout.Stop()

	result := make(map[string][]*types.BlockHeader)
	tasks := newHeadersTasks()
	for _, peer := range peers {
		tasks.addTask(peer.ID())
	}

	timeoutQueue := newTimeoutQueue(requireHeadersTimeout)
	for {
		for len(tasks.getPeers(unstart)) > 0 && len(tasks.getPeers(process)) < maxParallelTasksNum {
			unstartedPeers := tasks.getPeers(unstart)
			if len(unstartedPeers) > 0 {
				requestPeerID := unstartedPeers[0]
				if err := mf.requireHeaders(requestPeerID, locator, stopHash, skip); err != nil {
					tasks.addRequestNum(requestPeerID)
					log.WithFields(log.Fields{"module": logModule, "err": err}).Error("failed on send require headers msg")
					continue
				}

				tasks.setStatus(requestPeerID, process)
				timeoutQueue.addTimer(requestPeerID)
				if d := timeoutQueue.getNextTimeoutDuration(); d != nil {
					timeout.Reset(*d)
				}
			}
		}

		if len(tasks.getPeers(complete)) == len(peers) {
			return result, nil
		}

		select {
		case msg := <-mf.headersProcessCh:
			ok := tasks.isRequestedPeer(msg.peerID)
			if !ok {
				mf.peers.ErrorHandler(msg.peerID, security.LevelMsgIllegal, errors.New("get unsolicited blocks msg"))
				break
			}

			//reset timeout
			timeoutQueue.delTimer(msg.peerID)
			if d := timeoutQueue.getNextTimeoutDuration(); d != nil {
				timeout.Reset(*d)
			}
			tasks.setStatus(msg.peerID, complete)
			result[msg.peerID] = append(result[msg.peerID], msg.headers[:]...)
		case <-timeout.C:
			peerID := timeoutQueue.getFirstTimeoutID()
			if peerID == nil {
				break
			}
			ok := tasks.isRequestedPeer(*peerID)
			//task, ok := tasks[*peerID]
			if !ok {
				break
			}
			tasks.addRequestNum(*peerID)
			timeoutQueue.delTimer(*peerID)
			//reset timeout
			if d := timeoutQueue.getNextTimeoutDuration(); d != nil {
				timeout.Reset(*d)
			}
			log.WithFields(log.Fields{"module": logModule, "peerID": peerID, "error": errRequestTimeout}).Info("failed on fetch headers")
			mf.peers.ErrorHandler(*peerID, security.LevelConnException, errors.New("require blocks timeout"))
		case <-fetchHeadersTimeout.C:
			return nil, errors.New("parallel fetch headers timeout")
		}
	}
}

func (mf *msgFetcher) resetParameter() {
	for len(mf.blocksProcessCh) > 0 {
		<-mf.blocksProcessCh
	}

	for len(mf.headersProcessCh) > 0 {
		<-mf.headersProcessCh
	}
	mf.storage.ResetParameter()
}

func (mf *msgFetcher) verifyBlocksMsg(msg *blocksMsg, startHeader, stopHeader *types.BlockHeader) error {
	// null blocks
	if len(msg.blocks) == 0 {
		return errors.New("null blocks msg")
	}

	// blocks more than request
	if uint64(len(msg.blocks)) > stopHeader.Height-startHeader.Height+1 {
		return errors.New("exceed length blocks msg")
	}

	// verify start block
	if msg.blocks[0].Hash() != startHeader.Hash() {
		return errors.New("get mismatch blocks msg")
	}

	// verify blocks continuity
	for i := 0; i < len(msg.blocks)-1; i++ {
		if msg.blocks[i].Hash() != msg.blocks[i+1].PreviousBlockHash {
			return errors.New("get discontinuous blocks msg")
		}
	}

	return nil
}
