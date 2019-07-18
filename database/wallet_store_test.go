package database

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"

	acc "github.com/vapor/account"
	"github.com/vapor/asset"
	"github.com/vapor/blockchain/pseudohsm"
	"github.com/vapor/blockchain/query"
	"github.com/vapor/crypto/ed25519/chainkd"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/testutil"
	"github.com/vapor/wallet"
)

func TestAccountIndexKey(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestAccount")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("TestAccountIndex1", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	xpub2, _, err := hsm.XCreate("TestAccountIndex2", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	xpubs1 := []chainkd.XPub{xpub1.XPub, xpub2.XPub}
	xpubs2 := []chainkd.XPub{xpub2.XPub, xpub1.XPub}
	if !reflect.DeepEqual(accountIndexKey(xpubs1), accountIndexKey(xpubs2)) {
		t.Fatal("accountIndexKey test err")
	}

	if reflect.DeepEqual(xpubs1, xpubs2) {
		t.Fatal("accountIndexKey test err")
	}
}

func TestDeleteContractUTXO(t *testing.T) {
	cases := []struct {
		utxos      []*acc.UTXO
		deleteUTXO *acc.UTXO
		want       []*acc.UTXO
	}{
		{
			utxos:      []*acc.UTXO{},
			deleteUTXO: &acc.UTXO{},
			want:       []*acc.UTXO{},
		},
		{
			utxos: []*acc.UTXO{
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x3e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
			deleteUTXO: &acc.UTXO{
				OutputID: bc.NewHash([32]byte{0x3e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
			},
			want: []*acc.UTXO{},
		},
		{
			utxos: []*acc.UTXO{
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x3e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x2e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x3f, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x5e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x6e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x7e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
			deleteUTXO: &acc.UTXO{},
			want: []*acc.UTXO{
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x3e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x2e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x3f, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x5e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x6e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x7e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
		},
		{
			utxos: []*acc.UTXO{},
			deleteUTXO: &acc.UTXO{
				OutputID: bc.NewHash([32]byte{0x3e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
			},
			want: []*acc.UTXO{},
		},
		{
			utxos: []*acc.UTXO{
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x0e, 0x04, 0x50, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x00, 0x01, 0x02, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x01, 0x01, 0x02, 0x39, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
			deleteUTXO: &acc.UTXO{
				OutputID: bc.NewHash([32]byte{0x01, 0x01, 0x02, 0x39, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
			},
			want: []*acc.UTXO{
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x0e, 0x04, 0x50, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x00, 0x01, 0x02, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
		},
	}

	for _, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		accountStore := NewAccountStore(testDB)
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		// store mock utxos
		for _, utxo := range c.utxos {
			if err := ws.SetContractUTXO(utxo.OutputID, utxo); err != nil {
				t.Fatal(err)
			}
		}

		// delete utxo
		ws.DeleteContractUTXO(c.deleteUTXO.OutputID)
		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		// get utxo by outputID
		for _, utxo := range c.want {
			if _, err := accountStore.GetUTXO(utxo.OutputID); err != nil {
				t.Fatal(err)
			}
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}

func TestDeleteRecoveryStatus(t *testing.T) {
	cases := []struct {
		state *wallet.RecoveryState
	}{
		{
			state: &wallet.RecoveryState{},
		},
		{
			state: &wallet.RecoveryState{
				XPubs: []chainkd.XPub{
					chainkd.XPub([64]byte{0x00, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
		},
		{
			state: &wallet.RecoveryState{
				XPubs: []chainkd.XPub{
					chainkd.XPub([64]byte{0x00, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x01, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x02, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x03, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x04, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x05, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x06, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x07, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x08, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x09, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x0a, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					chainkd.XPub([64]byte{0x0b, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				},
			},
		},
	}

	for i, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		if err := ws.SetRecoveryStatus(c.state); err != nil {
			t.Fatal(err)
		}

		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotState, err := ws.GetRecoveryStatus()
		if err != nil {
			t.Fatal(err)
		}

		if !testutil.DeepEqual(gotState, c.state) {
			t.Errorf("case %v: got recovery status, got: %v, want: %v.", i, gotState, c.state)
		}

		ws = walletStore.InitBatch()
		ws.DeleteRecoveryStatus()
		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotState, err = ws.GetRecoveryStatus()
		if err == nil {
			t.Errorf("case %v: got state should fail.", i)
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}

func TestGetWalletInfo(t *testing.T) {
	cases := []struct {
		state *wallet.StatusInfo
	}{
		{
			state: &wallet.StatusInfo{},
		},
		{
			state: &wallet.StatusInfo{
				Version:    uint(0),
				WorkHeight: uint64(9),
				WorkHash:   bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				BestHash:   bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				BestHeight: uint64(0),
			},
		},
		{
			state: &wallet.StatusInfo{
				Version:    uint(9),
				WorkHeight: uint64(9999999),
				WorkHash:   bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				BestHash:   bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				BestHeight: uint64(100000),
			},
		},
	}

	for i, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		if err := ws.SetWalletInfo(c.state); err != nil {
			t.Fatal(err)
		}

		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotState, err := ws.GetWalletInfo()
		if err != nil {
			t.Fatal(err)
		}

		if !testutil.DeepEqual(gotState, c.state) {
			t.Errorf("case %v: got wallet info, got: %v, want: %v.", i, gotState, c.state)
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}

func TestDeleteUnconfirmedTransaction(t *testing.T) {
	cases := []struct {
		tx *query.AnnotatedTx
	}{
		{
			tx: &query.AnnotatedTx{},
		},
		{
			tx: &query.AnnotatedTx{
				ID:          bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				Timestamp:   uint64(100),
				BlockID:     bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				BlockHeight: uint64(100),
				Position:    uint32(100),
			},
		},
	}

	for i, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		if err := ws.SetUnconfirmedTransaction(c.tx.ID.String(), c.tx); err != nil {
			t.Fatal(err)
		}

		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotTx, err := ws.GetUnconfirmedTransaction(c.tx.ID.String())
		if err != nil {
			t.Fatal(err)
		}

		if !testutil.DeepEqual(gotTx, c.tx) {
			t.Errorf("case %v: got unconfirmed transaction, got: %v, want: %v.", i, gotTx, c.tx)
		}

		ws = walletStore.InitBatch()
		ws.DeleteUnconfirmedTransaction(c.tx.ID.String())
		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotTx, err = ws.GetUnconfirmedTransaction(c.tx.ID.String())
		if err == nil {
			t.Errorf("case %v: got unconfirmed transaction should fail.", i)
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}

func TestListUnconfirmedTransactions(t *testing.T) {
	cases := []struct {
		txs []*query.AnnotatedTx
	}{
		{
			txs: []*query.AnnotatedTx{},
		},
		{
			txs: []*query.AnnotatedTx{
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
			},
		},
		{
			txs: []*query.AnnotatedTx{
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x02, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x02, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x03, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x03, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x04, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x04, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x05, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x05, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x06, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x06, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x07, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x07, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
				&query.AnnotatedTx{
					ID:          bc.NewHash([32]byte{0x08, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					Timestamp:   uint64(100),
					BlockID:     bc.NewHash([32]byte{0x08, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
					BlockHeight: uint64(100),
					Position:    uint32(100),
				},
			},
		},
	}

	for i, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		for _, tx := range c.txs {
			if err := ws.SetUnconfirmedTransaction(tx.ID.String(), tx); err != nil {
				t.Fatal(err)
			}
		}

		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotTxs, err := ws.ListUnconfirmedTransactions()
		if err != nil {
			t.Fatal(err)
		}

		sort.Sort(SortUnconfirmedTxs(gotTxs))
		sort.Sort(SortUnconfirmedTxs(c.txs))

		if !testutil.DeepEqual(gotTxs, c.txs) {
			t.Errorf("case %v: list unconfirmed transactions, got: %v, want: %v.", i, gotTxs, c.txs)
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}

type SortUnconfirmedTxs []*query.AnnotatedTx

func (s SortUnconfirmedTxs) Len() int { return len(s) }
func (s SortUnconfirmedTxs) Less(i, j int) bool {
	return bytes.Compare(s[i].ID.Bytes(), s[j].ID.Bytes()) < 0
}
func (s SortUnconfirmedTxs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func TestGetGlobalTransactionIndex(t *testing.T) {
	cases := []struct {
		globalTxID string
		blockHash  bc.Hash
		position   uint64
	}{
		{
			globalTxID: "",
			blockHash:  bc.NewHash([32]byte{}),
			position:   uint64(0),
		},
		{
			globalTxID: "8838474e18e945e2f648745a362ac6988bad172b0eef5c593038a2fc8b46261e",
			blockHash:  bc.NewHash([32]byte{0x00, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
			position:   uint64(0),
		},
		{
			globalTxID: "d8de517917a591daa447d6be28ffb2fac866703e4feb65e86221be9a22d3033a",
			blockHash:  bc.NewHash([32]byte{0x00, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
			position:   uint64(99),
		},
		{
			globalTxID: "f4c64e9f72623c26b06e17f909aec7c03c927fcf681781489257097e537b8d5a",
			blockHash:  bc.NewHash([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
			position:   uint64(99),
		},
	}

	for i, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		ws.SetGlobalTransactionIndex(c.globalTxID, &c.blockHash, c.position)

		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotIndex := ws.GetGlobalTransactionIndex(c.globalTxID)
		wantIndex := CalcGlobalTxIndex(&c.blockHash, c.position)

		if !testutil.DeepEqual(gotIndex, wantIndex) {
			t.Errorf("case %v: got global transaction index, got: %v, want: %v.", i, gotIndex, wantIndex)
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}

func TestGetAsset(t *testing.T) {
	cases := []struct {
		asset *asset.Asset
	}{
		{
			asset: &asset.Asset{
				AssetID:       bc.NewAssetID([32]byte{}),
				DefinitionMap: map[string]interface{}{},
			},
		},
		{
			asset: &asset.Asset{
				AssetID:       bc.NewAssetID([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				DefinitionMap: map[string]interface{}{},
			},
		},
		{
			asset: &asset.Asset{
				AssetID: bc.NewAssetID([32]byte{}),
				DefinitionMap: map[string]interface{}{
					"Name": "assetname",
				},
			},
		},
		{
			asset: &asset.Asset{
				AssetID: bc.NewAssetID([32]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				DefinitionMap: map[string]interface{}{
					"Name": "assetname",
				},
			},
		},
		{
			asset: &asset.Asset{
				AssetID: bc.NewAssetID([32]byte{0x02, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
				DefinitionMap: map[string]interface{}{
					"Name":   "assetname",
					"Amount": "1000000",
				},
			},
		},
	}

	for i, c := range cases {
		testDB := dbm.NewDB("testdb", "leveldb", "temp")
		walletStore := NewWalletStore(testDB)
		ws := walletStore.InitBatch()
		definitionByte, err := json.Marshal(c.asset.DefinitionMap)
		if err != nil {
			t.Fatal(err)
		}

		ws.SetAssetDefinition(&c.asset.AssetID, definitionByte)
		if err := ws.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		gotAsset, err := ws.GetAsset(&c.asset.AssetID)
		if err != nil {
			t.Fatal(err)
		}

		if !testutil.DeepEqual(gotAsset.DefinitionMap, c.asset.DefinitionMap) {
			t.Errorf("case %v: got asset, got: %v, want: %v.", i, gotAsset.DefinitionMap, c.asset.DefinitionMap)
		}

		testDB.Close()
		os.RemoveAll("temp")
	}
}
