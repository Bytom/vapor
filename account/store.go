package account

import (
	"github.com/vapor/common"
	"github.com/vapor/crypto/ed25519/chainkd"
	"github.com/vapor/protocol/bc"
)

// AccountStorer interface contains account storage functions.
type AccountStorer interface {
	InitBatch()
	CommitBatch()
	SetAccount(string, string, []byte)
	SetAccountIndex([]chainkd.XPub, uint64)
	GetAccountByAccountAlias(string) []byte
	GetAccountByAccountID(string) []byte
	GetAccountIndex([]chainkd.XPub) []byte
	DeleteAccountByAccountAlias(string)
	DeleteAccountByAccountID(string)
	DeleteRawProgram(common.Hash)
	DeleteBip44ContractIndex(string)
	DeleteContractIndex(string)
	GetContractIndex(string) []byte
	GetAccountUTXOs(string) [][]byte
	DeleteStandardUTXO(bc.Hash)
	GetCoinbaseArbitrary() []byte
	SetCoinbaseArbitrary([]byte)
	GetMiningAddress() []byte
	GetFirstAccount() ([]byte, error)
	SetMiningAddress([]byte)
	GetBip44ContractIndex(string, bool) []byte
	GetRawProgram(common.Hash) []byte
	GetAccounts(string) [][]byte
	GetControlPrograms() ([][]byte, error)
	SetRawProgram(common.Hash, []byte)
	SetContractIndex(string, uint64)
	SetBip44ContractIndex(string, bool, uint64)
	GetUTXOs() [][]byte
	GetStandardUTXO(bc.Hash) []byte
	GetContractUTXO(bc.Hash) []byte
	SetStandardUTXO(bc.Hash, []byte)
}
