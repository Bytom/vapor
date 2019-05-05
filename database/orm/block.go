package orm

import (
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

type BlockStoreState struct {
	StoreKey string `gorm:"primary_key"`
	Height   uint64
	Hash     string
}

type BlockHeader struct {
	ID                     uint   `gorm:"AUTO_INCREMENT"`
	BlockHash              string `sql:"index"`
	Height                 uint64 `sql:"index"`
	Version                uint64
	PreviousBlockHash      string
	Timestamp              uint64
	TransactionsMerkleRoot string
	TransactionStatusHash  string
}

func stringToHash(str string) (*bc.Hash, error) {
	hash := &bc.Hash{}
	if err := hash.UnmarshalText([]byte(str)); err != nil {
		return nil, err
	}
	return hash, nil
}

func (bh *BlockHeader) ToTypesBlockHeader() (*types.BlockHeader, error) {
	previousBlockHash, err := stringToHash(bh.PreviousBlockHash)
	if err != nil {
		return nil, err
	}

	transactionsMerkleRoot, err := stringToHash(bh.TransactionsMerkleRoot)
	if err != nil {
		return nil, err
	}
	transactionStatusHash, err := stringToHash(bh.TransactionStatusHash)
	if err != nil {
		return nil, err
	}

	return &types.BlockHeader{
		Version:           bh.Version,
		Height:            bh.Height,
		PreviousBlockHash: *previousBlockHash,
		Timestamp:         bh.Timestamp,
		BlockCommitment: types.BlockCommitment{
			TransactionsMerkleRoot: *transactionsMerkleRoot,
			TransactionStatusHash:  *transactionStatusHash,
		},
	}, nil
}
