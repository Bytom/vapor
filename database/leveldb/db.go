package leveldb

import . "github.com/tendermint/tmlibs/common"

type DB interface {
	Get([]byte) []byte
	Set([]byte, []byte)
	SetSync([]byte, []byte)
	Delete([]byte)
	DeleteSync([]byte)
	Close()
	NewBatch() Batch
	Iterator() Iterator
	IteratorPrefix([]byte) Iterator
	IteratorPrefixWithStart(Prefix, start []byte, isReverse bool) Iterator

	// For debugging
	Print()
	Stats() map[string]string
}

type Batch interface {
	Set(key, value []byte)
	Delete(key []byte)
	Write()
}

type Iterator interface {
	Next() bool

	Key() []byte
	Value() []byte
	Seek([]byte) bool

	Release()
	Error() error
}

//-----------------------------------------------------------------------------

const (
	utxoPrefix  byte = iota //UTXOPrefix is StandardUTXOKey prefix
	sutxoPrefix             //SUTXOPrefix is ContractUTXOKey prefix
	contractPrefix
	contractIndexPrefix
	accountPrefix // AccountPrefix is account ID prefix
	accountAliasPrefix
	accountIndexPrefix
	txPrefix            //TxPrefix is wallet database transactions prefix
	txIndexPrefix       //TxIndexPrefix is wallet database tx index prefix
	unconfirmedTxPrefix //UnconfirmedTxPrefix is txpool unconfirmed transactions prefix
	globalTxIndexPrefix //GlobalTxIndexPrefix is wallet database global tx index prefix
	walletKey
	miningAddressKey
	coinbaseAbKey
	recoveryKey //recoveryKey key for db store recovery info.
)

// leveldb key prefix
var (
	colon               byte = 0x3a
	store                    = []byte("store:")
	accountStore             = []byte("AS:")
	walletStore              = []byte("WS:")
	UTXOPrefix               = append(accountStore, utxoPrefix, colon)
	SUTXOPrefix              = append(walletStore, sutxoPrefix, colon)
	ContractPrefix           = append(accountStore, contractPrefix, colon)
	ContractIndexPrefix      = append(accountStore, contractIndexPrefix, colon)
	AccountPrefix            = append(accountStore, accountPrefix, colon) // AccountPrefix is account ID prefix
	AccountAliasPrefix       = append(walletStore, accountAliasPrefix, colon)
	AccountIndexPrefix       = append(accountStore, accountIndexPrefix, colon)
	TxPrefix                 = append(walletStore, txPrefix, colon)            //TxPrefix is wallet database transactions prefix
	TxIndexPrefix            = append(walletStore, txIndexPrefix, colon)       //TxIndexPrefix is wallet database tx index prefix
	UnconfirmedTxPrefix      = append(walletStore, unconfirmedTxPrefix, colon) //UnconfirmedTxPrefix is txpool unconfirmed transactions prefix
	GlobalTxIndexPrefix      = append(walletStore, globalTxIndexPrefix, colon) //GlobalTxIndexPrefix is wallet database global tx index prefix
	WalletKey                = append(walletStore, walletKey)
	MiningAddressKey         = append(walletStore, miningAddressKey)
	CoinbaseAbKey            = append(walletStore, coinbaseAbKey)
	RecoveryKey              = append(walletStore, recoveryKey)
)

const (
	LevelDBBackendStr   = "leveldb" // legacy, defaults to goleveldb.
	CLevelDBBackendStr  = "cleveldb"
	GoLevelDBBackendStr = "goleveldb"
	MemDBBackendStr     = "memdb"
)

type dbCreator func(name string, dir string) (DB, error)

var backends = map[string]dbCreator{}

func registerDBCreator(backend string, creator dbCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

func NewDB(name string, backend string, dir string) DB {
	db, err := backends[backend](name, dir)
	if err != nil {
		PanicSanity(Fmt("Error initializing DB: %v", err))
	}
	return db
}
