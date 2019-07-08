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
	UTXOPrefix               = append(store, utxoPrefix, colon)
	SUTXOPrefix              = append(store, sutxoPrefix, colon)
	ContractPrefix           = []byte{contractPrefix, contractPrefix, colon}
	ContractIndexPrefix      = []byte{contractIndexPrefix, colon}
	AccountPrefix            = []byte{accountPrefix, colon} // AccountPrefix is account ID prefix
	AccountAliasPrefix       = []byte{accountAliasPrefix, colon}
	AccountIndexPrefix       = []byte{accountIndexPrefix, colon}
	TxPrefix                 = []byte{txPrefix, colon}            //TxPrefix is wallet database transactions prefix
	TxIndexPrefix            = []byte{txIndexPrefix, colon}       //TxIndexPrefix is wallet database tx index prefix
	UnconfirmedTxPrefix      = []byte{unconfirmedTxPrefix, colon} //UnconfirmedTxPrefix is txpool unconfirmed transactions prefix
	GlobalTxIndexPrefix      = []byte{globalTxIndexPrefix, colon} //GlobalTxIndexPrefix is wallet database global tx index prefix
	WalletKey                = []byte{walletKey}
	MiningAddressKey         = []byte{miningAddressKey}
	CoinbaseAbKey            = []byte{coinbaseAbKey}
	RecoveryKey              = []byte{recoveryKey}
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
