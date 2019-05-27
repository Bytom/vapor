package protocol

import (
	"errors"

	"github.com/vapor/database/storage"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/protocol/state"
)

var (
	ErrNotFoundVoteResult = errors.New("can't find the vote result by given sequence")
)

// Store provides storage interface for blockchain data
type Store interface {
	BlockExist(*bc.Hash) bool

	GetBlock(*bc.Hash) (*types.Block, error)
	GetStoreStatus() *BlockStoreState
	GetTransactionStatus(*bc.Hash) (*bc.TransactionStatus, error)
	GetTransactionsUtxo(*state.UtxoViewpoint, []*bc.Tx) error
	GetUtxo(*bc.Hash) (*storage.UtxoEntry, error)
	GetVoteResult(uint64) (*state.VoteResult, error)

	LoadBlockIndex(uint64) (*state.BlockIndex, error)
	SaveBlock(*types.Block, *bc.TransactionStatus) error
	SaveChainStatus(*state.BlockNode, *state.BlockNode, *state.UtxoViewpoint, map[bc.Hash]bool, map[uint64]*state.VoteResult) error
	SaveChainNodeStatus(*state.BlockNode, *state.BlockNode) error
}

// BlockStoreState represents the core's db status
type BlockStoreState struct {
	Height             uint64
	Hash               *bc.Hash
	IrreversibleHeight uint64
	IrreversibleHash   *bc.Hash
}
