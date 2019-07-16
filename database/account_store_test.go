package database

import (
	"os"
	"testing"

	"github.com/vapor/common"
	"github.com/vapor/testutil"

	"github.com/vapor/blockchain/signers"
	"github.com/vapor/crypto/ed25519/chainkd"

	acc "github.com/vapor/account"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/protocol/bc"
)

func TestDeleteAccount(t *testing.T) {
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer func() {
		testDB.Close()
		os.RemoveAll("temp")
	}()

	cases := []struct {
		accounts      []*acc.Account
		deleteAccount *acc.Account
		want          []*acc.Account
	}{
		{
			accounts:      []*acc.Account{},
			deleteAccount: &acc.Account{},
			want:          []*acc.Account{},
		},
		{
			accounts: []*acc.Account{},
			deleteAccount: &acc.Account{
				ID:    "id-1",
				Alias: "alias-1",
			},
			want: []*acc.Account{},
		},
		{
			accounts: []*acc.Account{
				&acc.Account{
					ID:    "id-1",
					Alias: "alias-1",
				},
				&acc.Account{
					ID:    "id-2",
					Alias: "alias-2",
				},
			},
			deleteAccount: &acc.Account{},
			want: []*acc.Account{
				&acc.Account{
					ID:    "id-1",
					Alias: "alias-1",
				},
				&acc.Account{
					ID:    "id-2",
					Alias: "alias-2",
				},
			},
		},
		{
			accounts: []*acc.Account{
				&acc.Account{
					ID:    "id-1",
					Alias: "alias-1",
				},
				&acc.Account{
					ID:    "id-2",
					Alias: "alias-2",
				},
			},
			deleteAccount: &acc.Account{
				ID:    "id-3",
				Alias: "alias-3",
			},
			want: []*acc.Account{
				&acc.Account{
					ID:    "id-1",
					Alias: "alias-1",
				},
				&acc.Account{
					ID:    "id-2",
					Alias: "alias-2",
				},
			},
		},
		{
			accounts: []*acc.Account{
				&acc.Account{
					ID:    "id-1",
					Alias: "alias-1",
				},
				&acc.Account{
					ID:    "id-2",
					Alias: "alias-2",
				},
			},
			deleteAccount: &acc.Account{
				ID:    "id-1",
				Alias: "alias-1",
			},
			want: []*acc.Account{
				&acc.Account{
					ID:    "id-2",
					Alias: "alias-2",
				},
			},
		},
	}

	accountStore := NewAccountStore(testDB)
	for i, c := range cases {
		as := accountStore.InitBatch()
		// store mock accounts
		for _, a := range c.accounts {
			if err := as.SetAccount(a); err != nil {
				t.Fatal(err)
			}
		}

		// delete account
		if err := as.DeleteAccount(c.deleteAccount); err != nil {
			t.Fatal(err)
		}

		if err := as.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		// get account by deleteAccount.ID, it should print ErrFindAccount
		if _, err := as.GetAccountByID(c.deleteAccount.ID); err != acc.ErrFindAccount {
			t.Fatal(err)
		}

		for _, a := range c.want {
			if _, err := as.GetAccountByID(a.ID); err != nil {
				t.Errorf("case %v: cann't find account, err: %v", i, err)
			}

			if _, err := as.GetAccountByAlias(a.Alias); err != nil {
				t.Errorf("case %v: cann't find account, err: %v", i, err)
			}
		}
	}
}

func TestDeleteStandardUTXO(t *testing.T) {
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer func() {
		testDB.Close()
		os.RemoveAll("temp")
	}()

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
			},
			deleteUTXO: &acc.UTXO{},
			want: []*acc.UTXO{
				&acc.UTXO{
					OutputID: bc.NewHash([32]byte{0x3e, 0x94, 0x5d, 0x35, 0x70, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c}),
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

	accountStore := NewAccountStore(testDB)
	for _, c := range cases {
		as := accountStore.InitBatch()
		// store mock utxos
		for _, utxo := range c.utxos {
			if err := as.SetStandardUTXO(utxo.OutputID, utxo); err != nil {
				t.Fatal(err)
			}
		}

		// delete utxo
		as.DeleteStandardUTXO(c.deleteUTXO.OutputID)
		if err := as.CommitBatch(); err != nil {
			t.Fatal(err)
		}

		// get utxo by outputID
		for _, utxo := range c.want {
			if _, err := as.GetUTXO(utxo.OutputID); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestGetAccountIndex(t *testing.T) {
	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer func() {
		testDB.Close()
		os.RemoveAll("temp")
	}()

	cases := []struct {
		account      *acc.Account
		currentIndex uint64
		want         uint64
	}{
		{
			account: &acc.Account{
				Signer: &signers.Signer{
					XPubs: []chainkd.XPub{
						[64]byte{0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
						[64]byte{0x09, 0x09, 0x09, 0x01, 0x01, 0x00, 0x04, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
					},
					KeyIndex: uint64(0),
				},
			},
			currentIndex: uint64(0),
			want:         uint64(0),
		},
		{
			account: &acc.Account{
				Signer: &signers.Signer{
					XPubs: []chainkd.XPub{
						[64]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
						[64]byte{0x00, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
					},
					KeyIndex: uint64(1),
				},
			},
			currentIndex: uint64(1),
			want:         uint64(1),
		},
		{
			account: &acc.Account{
				Signer: &signers.Signer{
					XPubs: []chainkd.XPub{
						[64]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
						[64]byte{0x00, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
					},
					KeyIndex: uint64(9),
				},
			},
			currentIndex: uint64(1),
			want:         uint64(9),
		},
		{
			account: &acc.Account{
				Signer: &signers.Signer{
					XPubs: []chainkd.XPub{
						[64]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
						[64]byte{0x00, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
					},
					KeyIndex: uint64(10),
				},
			},
			currentIndex: uint64(88),
			want:         uint64(88),
		},
		{
			account: &acc.Account{
				Signer: &signers.Signer{
					XPubs:    []chainkd.XPub{},
					KeyIndex: uint64(0),
				},
			},
			currentIndex: uint64(0),
			want:         uint64(0),
		},
		{
			account: &acc.Account{
				Signer: &signers.Signer{
					XPubs: []chainkd.XPub{
						[64]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c, 0x01, 0x01, 0x51, 0x31, 0x71, 0x30, 0xd4, 0x3b, 0x3d, 0xe3, 0xdd, 0x80, 0x67, 0x29, 0x9a, 0x5e, 0x09, 0xf9, 0xfb, 0x2b, 0xad, 0x5f, 0x92, 0xc8, 0x69, 0xd1, 0x42, 0x39, 0x74, 0x9a, 0xd1, 0x1c},
					},
					KeyIndex: uint64(1),
				},
			},
			currentIndex: uint64(77),
			want:         uint64(77),
		},
	}

	accountStore := NewAccountStore(testDB)
	for i, c := range cases {
		as := accountStore.InitBatch()
		v := as.(*AccountStore)
		v.db.Set(accountIndexKey(c.account.XPubs), common.Unit64ToBytes(c.currentIndex))
		as.SetAccountIndex(c.account)
		if err := as.CommitBatch(); err != nil {
			t.Fatal(err)
		}
		gotIndex := as.GetAccountIndex(c.account.XPubs)
		if !testutil.DeepEqual(gotIndex, c.want) {
			t.Errorf("case %v: got incorrect account index, got: %v, want: %v.", i, gotIndex, c.want)
		}
	}
}
