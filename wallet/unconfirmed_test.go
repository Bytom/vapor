package wallet

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/vapor/account"
	"github.com/vapor/asset"
	"github.com/vapor/blockchain/pseudohsm"
	"github.com/vapor/blockchain/query"
	"github.com/vapor/blockchain/signers"
	"github.com/vapor/consensus"
	"github.com/vapor/crypto/ed25519/chainkd"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/event"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/testutil"
)

func TestWalletUnconfirmedTxs(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	testStore := NewStore(testDB)
	defer os.RemoveAll("temp")

	accountManager := account.NewManager(testDB, nil)
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

	reg := asset.NewRegistry(testDB, nil)
	asset := bc.AssetID{V0: 5}

	dispatcher := event.NewDispatcher()
	w := mockWallet(testStore, accountManager, reg, nil, dispatcher, false)
	utxos := []*account.UTXO{}
	btmUtxo := mockUTXO(controlProg, consensus.BTMAssetID)
	utxos = append(utxos, btmUtxo)

	OtherUtxo := mockUTXO(controlProg, &asset)
	utxos = append(utxos, OtherUtxo)
	_, txData, err := mockTxData(utxos, testAccount)
	if err != nil {
		t.Fatal(err)
	}
	testTx := types.NewTx(*txData)
	w.saveUnconfirmedTx(testTx)

	txs := AnnotatedTxs([]*types.Tx{testTx}, w)
	wantTx := txs[0]
	gotTx, err := w.GetUnconfirmedTxByTxID(testTx.ID.String())
	if !testutil.DeepEqual(gotTx.ID, wantTx.ID) {
		t.Errorf(`transaction got=%#v; want=%#v`, gotTx.ID, wantTx.ID)
	}

	wantTxs := AnnotatedTxs([]*types.Tx{testTx}, w)
	gotTxs, err := w.GetUnconfirmedTxs("")
	for i, want := range wantTxs {
		if !testutil.DeepEqual(gotTxs[i].ID, want.ID) {
			t.Errorf(`the NO %d transaction, tx got=%#v; want=%#v`, i, gotTxs[i].ID.String(), want.ID.String())
		}

		for j, input := range want.Inputs {
			if !testutil.DeepEqual(gotTxs[i].Inputs[j].AccountID, input.AccountID) {
				t.Errorf(`the NO %d transaction input, accountID got=%#v; want=%#v`, j, gotTxs[i].Inputs[j].AccountID, input.AccountID)
			}

			if !testutil.DeepEqual(gotTxs[i].Inputs[j].AssetID, input.AssetID) {
				t.Errorf(`the NO %d transaction input, assetID got=%#v; want=%#v`, j, gotTxs[i].Inputs[j].AssetID, input.AssetID)
			}
		}

		for k, output := range want.Outputs {
			if !testutil.DeepEqual(gotTxs[i].Outputs[k].AccountID, output.AccountID) {
				t.Errorf(`the NO %d transaction input, accountID got=%#v; want=%#v`, k, gotTxs[i].Inputs[k].AccountID, output.AccountID)
			}

			if !testutil.DeepEqual(gotTxs[i].Outputs[k].AssetID, output.AssetID) {
				t.Errorf(`the NO %d transaction input, assetID got=%#v; want=%#v`, k, gotTxs[i].Inputs[k].AssetID, output.AssetID)
			}
		}
	}
}

func AnnotatedTxs(txs []*types.Tx, w *Wallet) []*query.AnnotatedTx {
	// annotate account and asset
	annotatedTxs := []*query.AnnotatedTx{}
	for _, tx := range txs {
		annotatedTx := &query.AnnotatedTx{
			ID:      tx.ID,
			Inputs:  make([]*query.AnnotatedInput, 0, len(tx.Inputs)),
			Outputs: make([]*query.AnnotatedOutput, 0, len(tx.Outputs)),
			Size:    tx.SerializedSize,
		}

		for i := range tx.Inputs {
			annotatedTx.Inputs = append(annotatedTx.Inputs, w.BuildAnnotatedInput(tx, uint32(i)))
		}
		for i := range tx.Outputs {
			annotatedTx.Outputs = append(annotatedTx.Outputs, w.BuildAnnotatedOutput(tx, i))
		}
		annotatedTxs = append(annotatedTxs, annotatedTx)
	}

	annotateTxsAccount(annotatedTxs, w.store)
	annotateTxsAsset(w, annotatedTxs)

	return annotatedTxs
}
