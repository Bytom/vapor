package chainmgr

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/errors"
	"github.com/vapor/netsync/peers"
	"github.com/vapor/p2p/security"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

const (
	maxNumOfParallelFetchBlocks = 7
	blockProcessChSize          = 1024
	blocksProcessChSize         = 128
	headersProcessChSize        = 1024
	maxNumOfFastSyncPeers       = 128
)

var (
	requireBlockTimeout   = 20 * time.Second
	requireHeadersTimeout = 30 * time.Second
	requireBlocksTimeout  = 50 * time.Second

	errRequestBlocksTimeout = errors.New("request blocks timeout")
	errRequestTimeout       = errors.New("request timeout")
	errPeerDropped          = errors.New("Peer dropped")
	errSendMsg              = errors.New("send message error")
)

type MsgFetcher interface {
	resetParameter()
	addSyncPeer(peerID string)
	requireBlock(peerID string, height uint64) (*types.Block, error)
	parallelFetchBlocks(work []*fetchBlocksWork, downloadedBlockCh chan struct{}, downloadResult chan bool, ProcessResult chan bool, wg *sync.WaitGroup)
	parallelFetchHeaders(peers []*peers.Peer, locator []*bc.Hash, stopHash *bc.Hash, skip uint64) (map[string][]*types.BlockHeader, error)
}

type fetchBlocksWork struct {
	startHeader, stopHeader *types.BlockHeader
}

type fetchBlocksResult struct {
	startHeight, stopHeight uint64
	err                     error
}

type msgFetcher struct {
	storage          Storage
	syncPeers        *fastSyncPeers
	peers            *peers.PeerSet
	blockProcessCh   chan *blockMsg
	blocksProcessCh  chan *blocksMsg
	headersProcessCh chan *headersMsg
	blocksMsgChanMap map[string]chan []*types.Block
	mux              sync.RWMutex
}

func newMsgFetcher(storage Storage, peers *peers.PeerSet) *msgFetcher {
	return &msgFetcher{
		storage:          storage,
		syncPeers:        newFastSyncPeers(),
		peers:            peers,
		blockProcessCh:   make(chan *blockMsg, blockProcessChSize),
		blocksProcessCh:  make(chan *blocksMsg, blocksProcessChSize),
		headersProcessCh: make(chan *headersMsg, headersProcessChSize),
		blocksMsgChanMap: make(map[string]chan []*types.Block),
	}
}

func (mf *msgFetcher) addSyncPeer(peerID string) {
	mf.syncPeers.add(peerID)
}

func (mf *msgFetcher) processBlock(peerID string, block *types.Block) {
	mf.blockProcessCh <- &blockMsg{block: block, peerID: peerID}
}

func (mf *msgFetcher) processBlocks(peerID string, blocks []*types.Block) {
	mf.blocksProcessCh <- &blocksMsg{blocks: blocks, peerID: peerID}
	mf.mux.RLock()
	blocksMsgChan, ok := mf.blocksMsgChanMap[peerID]
	mf.mux.RUnlock()
	if !ok {
		log.WithFields(log.Fields{"module": logModule}).Warn("msg from unsolicited peer")
		return
	}

	blocksMsgChan <- blocks
}

func (mf *msgFetcher) processHeaders(peerID string, headers []*types.BlockHeader) {
	mf.headersProcessCh <- &headersMsg{headers: headers, peerID: peerID}
}

func (mf *msgFetcher) fetchBlocks(work *fetchBlocksWork, peerID string) ([]*types.Block, error) {
	defer mf.syncPeers.setIdle(peerID)
	startHash := work.startHeader.Hash()
	stopHash := work.stopHeader.Hash()
	blocks, err := mf.requireBlocks(peerID, []*bc.Hash{&startHash}, &stopHash)
	if err != nil {
		mf.peers.ErrorHandler(peerID, security.LevelConnException, err)
		return nil, err
	}

	if err := mf.verifyBlocksMsg(blocks, work.startHeader, work.stopHeader); err != nil {
		mf.peers.ErrorHandler(peerID, security.LevelConnException, err)
		return nil, err
	}

	return blocks, nil
}

func (mf *msgFetcher) fetchBlocksProcess(work *fetchBlocksWork, peerCh chan string, downloadedBlockCh chan struct{}, closeCh chan struct{}) error {
	for {
		select {
		case peerID := <-peerCh:
			for {
				blocks, err := mf.fetchBlocks(work, peerID)
				if err != nil {
					log.WithFields(log.Fields{"module": logModule, "startHeight": work.startHeader.Height, "stopHeight": work.stopHeader.Height, "error": err}).Info("failed on fetch blocks")
					break
				}

				if err := mf.storage.writeBlocks(peerID, blocks); err != nil {
					log.WithFields(log.Fields{"module": logModule, "error": err}).Info("write block error")
				}

				// send to block process pool
				downloadedBlockCh <- struct{}{}

				// work completed
				if blocks[len(blocks)-1].Height >= work.stopHeader.Height-1 {
					return nil
				}

				//unfinished work, continue
				work.startHeader = &blocks[len(blocks)-1].BlockHeader
			}
		case <-closeCh:
			return nil
		}
	}
}

func (mf *msgFetcher) fetchBlocksWorker(workCh chan *fetchBlocksWork, peerCh chan string, resultCh chan *fetchBlocksResult, closeCh chan struct{}, downloadedBlockCh chan struct{}, wg *sync.WaitGroup) {
	for {
		select {
		case work := <-workCh:
			err := mf.fetchBlocksProcess(work, peerCh, downloadedBlockCh, closeCh)
			resultCh <- &fetchBlocksResult{startHeight: work.startHeader.Height, stopHeight: work.stopHeader.Height, err: err}
		case <-closeCh:
			wg.Done()
			return
		}
	}
}

func (mf *msgFetcher) parallelFetchBlocks(works []*fetchBlocksWork, downloadedBlockCh chan struct{}, downloadStop chan bool, ProcessStop chan bool, wg *sync.WaitGroup) {
	workSize := len(works)
	workCh := make(chan *fetchBlocksWork, workSize)
	peerCh := make(chan string, maxNumOfFastSyncPeers)
	resultCh := make(chan *fetchBlocksResult, workSize)
	closeCh := make(chan struct{})

	var workWg sync.WaitGroup
	for i := 0; i <= maxNumOfParallelFetchBlocks && i < workSize; i++ {
		workWg.Add(1)
		go mf.fetchBlocksWorker(workCh, peerCh, resultCh, closeCh, downloadedBlockCh, &workWg)
	}

	for _, work := range works {
		workCh <- work
	}

	syncPeers := mf.syncPeers.selectIdlePeers()
	for i := 0; i < len(syncPeers) && i < maxNumOfFastSyncPeers; i++ {
		peerCh <- syncPeers[i]
	}

	defer func() {
		close(closeCh)
		workWg.Wait()
		close(workCh)
		close(resultCh)
		downloadStop <- true
		wg.Done()
	}()

	//collect fetch results
	for resultCount := 0; resultCount < workSize && mf.syncPeers.size() > 0; resultCount++ {
		select {
		case result := <-resultCh:
			if result.err != nil {
				log.WithFields(log.Fields{"module": logModule, "startHeight": result.startHeight, "stopHeight": result.stopHeight, "err": result.err}).Error("failed on fetch blocks")
				return
			}

			peer, err := mf.syncPeers.selectIdlePeer()
			if err != nil {
				log.WithFields(log.Fields{"module": logModule, "err": result.err}).Warn("failed on find fast sync peer")
				break
			}
			peerCh <- peer
		case <-ProcessStop:
			return
		}
	}
}

func (mf *msgFetcher) parallelFetchHeaders(peers []*peers.Peer, locator []*bc.Hash, stopHash *bc.Hash, skip uint64) (map[string][]*types.BlockHeader, error) {
	result := make(map[string][]*types.BlockHeader)
	response := make(map[string]bool)
	for _, peer := range peers {
		if ok := peer.GetHeaders(locator, stopHash, skip); !ok {
			continue
		}
		result[peer.ID()] = nil
	}

	timeout := time.NewTimer(requireHeadersTimeout)
	defer timeout.Stop()
	for {
		select {
		case msg := <-mf.headersProcessCh:
			if _, ok := result[msg.peerID]; ok {
				result[msg.peerID] = append(result[msg.peerID], msg.headers[:]...)
				response[msg.peerID] = true
				if len(response) == len(result) {
					return result, nil
				}
			}
		case <-timeout.C:
			log.WithFields(log.Fields{"module": logModule, "err": errRequestTimeout}).Warn("failed on parallel fetch headers")
			return result, errors.Wrap(errRequestTimeout, "parallelFetchHeaders")
		}
	}
}

func (mf *msgFetcher) requireBlock(peerID string, height uint64) (*types.Block, error) {
	peer := mf.peers.GetPeer(peerID)
	if peer == nil {
		return nil, errPeerDropped
	}

	if ok := peer.GetBlockByHeight(height); !ok {
		return nil, errSendMsg
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

func (mf *msgFetcher) requireBlocks(peerID string, locator []*bc.Hash, stopHash *bc.Hash) ([]*types.Block, error) {
	peer := mf.peers.GetPeer(peerID)
	if peer == nil {
		mf.syncPeers.delete(peerID)
		return nil, errPeerDropped
	}

	receiveCh := make(chan []*types.Block, 1)
	mf.mux.Lock()
	mf.blocksMsgChanMap[peerID] = receiveCh
	mf.mux.Unlock()

	if ok := peer.GetBlocks(locator, stopHash); !ok {
		return nil, errSendMsg
	}

	timeout := time.NewTimer(requireBlocksTimeout)
	defer timeout.Stop()
	select {
	case blocks := <-receiveCh:
		return blocks, nil
	case <-timeout.C:
		return nil, errRequestBlocksTimeout
	}
}

func (mf *msgFetcher) resetParameter() {
	mf.blocksMsgChanMap = make(map[string]chan []*types.Block)
	mf.syncPeers = newFastSyncPeers()
	mf.storage.resetParameter()
	//empty chan
	for {
		select {
		case <-mf.blocksProcessCh:
		case <-mf.headersProcessCh:
		default:
			return
		}
	}
}

func (mf *msgFetcher) verifyBlocksMsg(blocks []*types.Block, startHeader, stopHeader *types.BlockHeader) error {
	// null blocks
	if len(blocks) == 0 {
		return errors.New("null blocks msg")
	}

	// blocks more than request
	if uint64(len(blocks)) > stopHeader.Height-startHeader.Height+1 {
		return errors.New("exceed length blocks msg")
	}

	// verify start block
	if blocks[0].Hash() != startHeader.Hash() {
		return errors.New("get mismatch blocks msg")
	}

	// verify blocks continuity
	for i := 0; i < len(blocks)-1; i++ {
		if blocks[i].Hash() != blocks[i+1].PreviousBlockHash {
			return errors.New("get discontinuous blocks msg")
		}
	}

	return nil
}
