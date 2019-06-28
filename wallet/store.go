package wallet

import (
	acc "github.com/vapor/account"
	"github.com/vapor/asset"
	"github.com/vapor/blockchain/query"
	"github.com/vapor/common"
	"github.com/vapor/protocol/bc"
)

// WalletStorer interface contains wallet storage functions.
type WalletStorer interface {
	InitBatch()
	CommitBatch()
	GetAssetDefinition(*bc.AssetID) (*asset.Asset, error)
	SetAssetDefinition(*bc.AssetID, []byte)
	GetControlProgram(common.Hash) (*acc.CtrlProgram, error)
	GetAccountByAccountID(string) (*acc.Account, error)
	DeleteTransactions(uint64)
	SetTransaction(uint64, *query.AnnotatedTx) error
	DeleteUnconfirmedTransaction(string)
	SetGlobalTransactionIndex(string, *bc.Hash, uint64)
	GetStandardUTXO(bc.Hash) (*acc.UTXO, error)
	GetTransaction(string) (*query.AnnotatedTx, error)
	GetGlobalTransactionIndex(string) []byte
	GetTransactions() ([]*query.AnnotatedTx, error)
	GetUnconfirmedTransactions() ([]*query.AnnotatedTx, error)
	GetUnconfirmedTransaction(string) (*query.AnnotatedTx, error)
	SetUnconfirmedTransaction(string, *query.AnnotatedTx) error
	DeleteStardardUTXO(bc.Hash)
	DeleteContractUTXO(bc.Hash)
	SetStandardUTXO(bc.Hash, *acc.UTXO) error
	SetContractUTXO(bc.Hash, *acc.UTXO) error
	GetWalletInfo() []byte // need move database.NewWalletStore in wallet package
	SetWalletInfo([]byte)  // need move database.NewWalletStore in wallet package
	DeleteWalletTransactions()
	DeleteWalletUTXOs()
	GetAccountUTXOs(string) ([]*acc.UTXO, error)
	SetRecoveryStatus([]byte, []byte) // recoveryManager.state isn't exported outside
	DeleteRecoveryStatus([]byte)
	GetRecoveryStatus([]byte) []byte // recoveryManager.state isn't exported outside
}
