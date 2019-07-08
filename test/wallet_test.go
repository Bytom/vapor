package test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/vapor/account"
	"github.com/vapor/asset"
	"github.com/vapor/blockchain/pseudohsm"
	"github.com/vapor/blockchain/signers"
	"github.com/vapor/blockchain/txbuilder"
	"github.com/vapor/config"
	"github.com/vapor/consensus"
	"github.com/vapor/crypto/ed25519/chainkd"
	"github.com/vapor/database"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/event"
	"github.com/vapor/protocol"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	wt "github.com/vapor/wallet"
)

func TestWalletUpdate(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	config.CommonConfig = config.DefaultConfig()
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer func() {
		testDB.Close()
		os.RemoveAll("temp")
	}()

	store := database.NewStore(testDB)
	walletStore := database.NewWalletStore(testDB)
	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)

	chain, err := protocol.NewChain(store, txPool, dispatcher)
	if err != nil {
		t.Fatal(err)
	}

	accountStore := database.NewAccountStore(testDB)
	accountManager := account.NewManager(accountStore, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("test_pub1", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub1.XPub}, 1, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	controlProg.KeyIndex = 1

	reg := asset.NewRegistry(testDB, chain)
	// reg := asset.NewRegistry(testDB, nil)
	asset := bc.AssetID{V0: 5}

	utxos := []*account.UTXO{}
	btmUtxo := mockUTXO(controlProg, consensus.BTMAssetID)
	utxos = append(utxos, btmUtxo)
	OtherUtxo := mockUTXO(controlProg, &asset)
	utxos = append(utxos, OtherUtxo)

	_, txData, err := mockTxData(utxos, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	tx := types.NewTx(*txData)
	block := mockSingleBlock(tx)
	txStatus := bc.NewTransactionStatus()
	txStatus.SetStatus(0, false)
	txStatus.SetStatus(1, false)
	store.SaveBlock(block, txStatus)

	w := newMockWallet(walletStore, accountManager, reg, chain, dispatcher, true)
	err = w.Wallet.AttachBlock(block)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := w.Wallet.GetTransactionByTxID(tx.ID.String()); err != nil {
		t.Fatal(err)
	}

	wants, err := w.Wallet.GetTransactions(testAccount.ID, "", 1, false)
	if len(wants) != 1 {
		t.Fatal(err)
	}

	if wants[0].ID != tx.ID {
		t.Fatal("account txID mismatch")
	}

	for position, tx := range block.Transactions {
		get := testDB.Get(database.CalcGlobalTxIndexKey(tx.ID.String()))
		bh := block.BlockHeader.Hash()
		expect := database.CalcGlobalTxIndex(&bh, uint64(position))
		if !reflect.DeepEqual(get, expect) {
			t.Fatalf("position#%d: compare retrieved globalTxIdx err", position)
		}
	}
}

// func TestRescanWallet(t *testing.T) {
// 	rescan := func() {
// 		// prepare wallet & db.
// 		dirPath, err := ioutil.TempDir(".", "")
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		defer os.RemoveAll(dirPath)

// 		config.CommonConfig = config.DefaultConfig()
// 		testDB := dbm.NewDB("testdb", "leveldb", "temp")
// 		walletStore := database.NewWalletStore(testDB)
// 		defer func() {
// 			testDB.Close()
// 			os.RemoveAll("temp")
// 		}()

// 		store := database.NewStore(testDB)
// 		dispatcher := event.NewDispatcher()
// 		txPool := protocol.NewTxPool(store, dispatcher)
// 		chain, err := protocol.NewChain(store, txPool, dispatcher)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		statusInfo := wt.StatusInfo{
// 			Version:  uint(1),
// 			WorkHash: bc.Hash{V0: 0xff},
// 		}
// 		if err := walletStore.SetWalletInfo(&statusInfo); err != nil {
// 			t.Fatal(err)
// 		}
// 		walletInfo, err := walletStore.GetWalletInfo()
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		accountStore := database.NewAccountStore(testDB)
// 		accountManager := account.NewManager(accountStore, chain)
// 		// w := newMockWallet(walletStore, accountManager, nil, chain, dispatcher, false)
// 		w, err := wt.NewWallet(walletStore, accountManager, nil, nil, chain, dispatcher, false)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		var hash [32]byte
// 		w.Status.WorkHash = bc.NewHash(hash)
// 		w.Status = *walletInfo

// 		// rescan wallet.
// 		if err := w.LoadWalletInfo(); err != nil {
// 			t.Fatal(err)
// 		}

// 		block := config.GenesisBlock()
// 		if w.Status.WorkHash != block.Hash() {
// 			t.Fatal("reattach from genesis block")
// 		}
// 	}

// 	for i := 0; i < 100; i++ {
// 		fmt.Printf("case i: %v\n", i)
// 		rescan()
// 	}
// }

func TestMemPoolTxQueryLoop(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatal(err)
	}
	config.CommonConfig = config.DefaultConfig()
	testDB := dbm.NewDB("testdb", "leveldb", dirPath)
	defer func() {
		testDB.Close()
		os.RemoveAll(dirPath)
	}()

	store := database.NewStore(testDB)
	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)

	chain, err := protocol.NewChain(store, txPool, dispatcher)
	if err != nil {
		t.Fatal(err)
	}

	accountStore := database.NewAccountStore(testDB)
	accountManager := account.NewManager(accountStore, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("test_pub1", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub1.XPub}, 1, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	controlProg.KeyIndex = 1

	reg := asset.NewRegistry(testDB, chain)
	asset := bc.AssetID{V0: 5}

	utxos := []*account.UTXO{}
	btmUtxo := mockUTXO(controlProg, consensus.BTMAssetID)
	utxos = append(utxos, btmUtxo)
	OtherUtxo := mockUTXO(controlProg, &asset)
	utxos = append(utxos, OtherUtxo)

	_, txData, err := mockTxData(utxos, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	tx := types.NewTx(*txData)
	txStatus := bc.NewTransactionStatus()
	txStatus.SetStatus(0, false)
	walletStore := database.NewWalletStore(testDB)
	w := newMockWallet(walletStore, accountManager, reg, chain, dispatcher, false)
	go w.Wallet.MemPoolTxQueryLoop()
	w.Wallet.EventDispatcher.Post(protocol.TxMsgEvent{TxMsg: &protocol.TxPoolMsg{TxDesc: &protocol.TxDesc{Tx: tx}, MsgType: protocol.MsgNewTx}})
	time.Sleep(time.Millisecond * 10)
	if _, err := w.Wallet.GetUnconfirmedTxByTxID(tx.ID.String()); err != nil {
		t.Fatal("dispatch new tx msg error:", err)
	}
	w.Wallet.EventDispatcher.Post(protocol.TxMsgEvent{TxMsg: &protocol.TxPoolMsg{TxDesc: &protocol.TxDesc{Tx: tx}, MsgType: protocol.MsgRemoveTx}})
	time.Sleep(time.Millisecond * 10)
	txs, err := w.Wallet.GetUnconfirmedTxs(testAccount.ID)
	if err != nil {
		t.Fatal("get unconfirmed tx error:", err)
	}

	if len(txs) != 0 {
		t.Fatal("dispatch remove tx msg error")
	}

	w.Wallet.EventDispatcher.Post(protocol.TxMsgEvent{TxMsg: &protocol.TxPoolMsg{TxDesc: &protocol.TxDesc{Tx: tx}, MsgType: 2}})
}

func mockUTXO(controlProg *account.CtrlProgram, assetID *bc.AssetID) *account.UTXO {
	utxo := &account.UTXO{}
	utxo.OutputID = bc.Hash{V0: 1}
	utxo.SourceID = bc.Hash{V0: 2}
	utxo.AssetID = *assetID
	utxo.Amount = 1000000000
	utxo.SourcePos = 0
	utxo.ControlProgram = controlProg.ControlProgram
	utxo.AccountID = controlProg.AccountID
	utxo.Address = controlProg.Address
	utxo.ControlProgramIndex = controlProg.KeyIndex
	return utxo
}

func mockTxData(utxos []*account.UTXO, testAccount *account.Account) (*txbuilder.Template, *types.TxData, error) {
	tplBuilder := txbuilder.NewBuilder(time.Now())

	for _, utxo := range utxos {
		txInput, sigInst, err := account.UtxoToInputs(testAccount.Signer, utxo)
		if err != nil {
			return nil, nil, err
		}
		tplBuilder.AddInput(txInput, sigInst)

		out := &types.TxOutput{}
		if utxo.AssetID == *consensus.BTMAssetID {
			out = types.NewIntraChainOutput(utxo.AssetID, 100, utxo.ControlProgram)
		} else {
			out = types.NewIntraChainOutput(utxo.AssetID, utxo.Amount, utxo.ControlProgram)
		}
		tplBuilder.AddOutput(out)
	}

	return tplBuilder.Build()
}

type mockWallet struct {
	Wallet *wt.Wallet
}

func newMockWallet(store wt.WalletStore, account *account.Manager, asset *asset.Registry, chain *protocol.Chain, dispatcher *event.Dispatcher, txIndexFlag bool) *mockWallet {
	w, err := wt.NewWallet(store, account, asset, nil, chain, dispatcher, txIndexFlag)
	if err != nil {
		panic(err)
	}
	var hash [32]byte
	w.Status.WorkHash = bc.NewHash(hash)
	return &mockWallet{w}
}

func mockSingleBlock(tx *types.Tx) *types.Block {
	return &types.Block{
		BlockHeader: types.BlockHeader{
			Version: 1,
			Height:  1,
		},
		Transactions: []*types.Tx{config.GenesisTx(), tx},
	}
}
