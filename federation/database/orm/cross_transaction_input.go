package orm

import (
	"database/sql"

	"github.com/vapor/federation/types"
)

type CrossTransactionInput struct {
	ID            uint64 `gorm:"primary_key"`
	MainchainTxID uint64
	SidechainTxID sql.NullInt64
	SourcePos     uint64
	AssetID       uint64
	AssetAmount   uint64
	Script        string
	CreatedAt     types.Timestamp
	UpdatedAt     types.Timestamp

	MainchainTransaction *CrossTransaction `gorm:"foreignkey:MainchainTxID"`
	SidechainTransaction *CrossTransaction `gorm:"foreignkey:SidechainTxID"`
	Asset                *Asset            `gorm:"foreignkey:AssetID"`
}
