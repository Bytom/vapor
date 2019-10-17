package database

import (
	"testing"

	"github.com/vapor/application/mov/common"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/testutil"

)

var (
	asset1 = bc.NewAssetID([32]byte{1})
	asset2 = bc.NewAssetID([32]byte{2})
	asset3 = bc.NewAssetID([32]byte{3})
	asset4 = bc.NewAssetID([32]byte{4})

	order1 = &common.Order{FromAssetID: assetID1, ToAssetID: assetID2, Rate: 0.1}
	order2 = &common.Order{FromAssetID: assetID1, ToAssetID: assetID2, Rate: 0.2}
	order3 = &common.Order{FromAssetID: assetID1, ToAssetID: assetID2, Rate: 0.3}
	order4 = &common.Order{FromAssetID: assetID1, ToAssetID: assetID2, Rate: 0.4}
	order5 = &common.Order{FromAssetID: assetID1, ToAssetID: assetID2, Rate: 0.5}
)

func TestTradePairIterator(t *testing.T) {
	cases := []struct {
		desc            string
		storeTradePairs []*common.TradePair
		wantTradePairs  []*common.TradePair
	}{
		{
			desc: "normal case",
			storeTradePairs: []*common.TradePair{
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset2,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset3,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset4,
				},
			},
			wantTradePairs: []*common.TradePair{
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset2,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset3,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset4,
				},
			},
		},
		{
			desc: "num of trade pairs more than one return",
			storeTradePairs: []*common.TradePair{
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset2,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset3,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset4,
				},
				{
					FromAssetID: &asset2,
					ToAssetID:   &asset1,
				},
			},
			wantTradePairs: []*common.TradePair{
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset2,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset3,
				},
				{
					FromAssetID: &asset1,
					ToAssetID:   &asset4,
				},
				{
					FromAssetID: &asset2,
					ToAssetID:   &asset1,
				},
			},
		},
		{
			desc:            "store is empty",
			storeTradePairs: []*common.TradePair{},
			wantTradePairs:  []*common.TradePair{},
		},
	}

	for i, c := range cases {
		store := &MockMovStore{TradePairs: c.storeTradePairs}
		var gotTradePairs []*common.TradePair
		iterator := NewTradePairIterator(store)
		for tradePair, err := iterator.Next();;tradePair, err = iterator.Next() {
			if err != nil {
				t.Fatal(err)
			}
			if tradePair == nil {
				break
			}


			gotTradePairs = append(gotTradePairs, tradePair)
		}
		if !testutil.DeepEqual(c.wantTradePairs, gotTradePairs) {
			t.Errorf("#%d(%s):got trade pairs it not equals want trade pairs", i, c.desc)
		}
	}
}

func TestOrderIterator(t *testing.T) {
	cases := []struct {
		desc        string
		tradePair   *common.TradePair
		storeOrders []*common.Order
		wantOrders  []*common.Order
	}{
		{
			desc: "normal case",
			tradePair: &common.TradePair{FromAssetID: assetID1, ToAssetID: assetID2},
			storeOrders: []*common.Order{order1, order2, order3},
			wantOrders:  []*common.Order{order1, order2, order3},
		},
		{
			desc: "num of orders more than one return",
			tradePair: &common.TradePair{FromAssetID: assetID1, ToAssetID: assetID2},
			storeOrders: []*common.Order{order1, order2, order3, order4, order5},
			wantOrders:  []*common.Order{order1, order2, order3, order4, order5},
		},
		{
			desc: "only one order",
			tradePair: &common.TradePair{FromAssetID: assetID1, ToAssetID: assetID2},
			storeOrders: []*common.Order{order1},
			wantOrders:  []*common.Order{order1},
		},
		{
			desc: "store is empty",
			tradePair: &common.TradePair{FromAssetID: assetID1, ToAssetID: assetID2},
			storeOrders: []*common.Order{},
			wantOrders:  []*common.Order{},
		},
	}

	for i, c := range cases {
		orderMap := map[string][]*common.Order{c.tradePair.Key(): c.storeOrders}
		store := &MockMovStore{OrderMap: orderMap}

		var gotOrders []*common.Order
		iterator := NewOrderIterator(store, c.tradePair)
		for orders, err := iterator.NextBatch();;orders, err = iterator.NextBatch() {
			if err != nil {
				t.Fatal(err)
			}
			if orders == nil {
				break
			}
			gotOrders = append(gotOrders, orders...)
		}
		if !testutil.DeepEqual(c.wantOrders, gotOrders) {
			t.Errorf("#%d(%s):got orders it not equals want orders", i, c.desc)
		}
	}
}
