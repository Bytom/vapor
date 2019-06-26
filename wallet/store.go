package wallet

import (
	"github.com/vapor/blockchain/query"
	"github.com/vapor/common"
	"github.com/vapor/protocol/bc"
)

// WalletStorer interface contains wallet storage functions.
type WalletStorer interface {
	InitBatch()
	CommitBatch()
	GetAssetDefinition(*bc.AssetID) []byte
	SetAssetDefinition(*bc.AssetID, []byte)
	GetRawProgram(common.Hash) []byte
	GetAccountByAccountID(string) []byte
	DeleteTransactions(uint64)
	SetTransaction(uint64, uint32, string, []byte)
	DeleteUnconfirmedTransaction(string)
	SetGlobalTransactionIndex(string, *bc.Hash, uint64)
	GetStandardUTXO(bc.Hash) []byte
	GetTransaction(string) ([]byte, error)
	GetGlobalTransaction(string) []byte
	GetTransactions() ([]*query.AnnotatedTx, error)
	GetUnconfirmedTransactions() ([]*query.AnnotatedTx, error)
	GetUnconfirmedTransaction(string) []byte
	SetUnconfirmedTransaction(string, []byte)
	DeleteStardardUTXO(bc.Hash)
	DeleteContractUTXO(bc.Hash)
	SetStandardUTXO(bc.Hash, []byte)
	SetContractUTXO(bc.Hash, []byte)
	GetWalletInfo() []byte
	SetWalletInfo([]byte)
	DeleteWalletTransactions()
	DeleteWalletUTXOs()
	GetAccountUTXOs(key string) [][]byte
	SetRecoveryStatus([]byte, []byte)
	DeleteRecoveryStatus([]byte)
	GetRecoveryStatus([]byte) []byte
}
