package database

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"math"

	"github.com/vapor/application/mov/common"
	dbm "github.com/vapor/database/leveldb"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

const (
	order byte = iota
	tradePair
	matchStatus

	tradePairsNum = 1024
	ordersNum     = 10240
	assetIDLen    = 32
	rateByteLen   = 8
)

var (
	movStore         = []byte("MOV:")
	ordersPrefix     = append(movStore, order)
	tradePairsPrefix = append(movStore, tradePair)
	bestMatchStore   = append(movStore, matchStatus)
)

func calcOrderKey(fromAssetID, toAssetID *bc.AssetID, utxoHash *bc.Hash, rate float64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, math.Float64bits(rate))
	key := append(ordersPrefix, fromAssetID.Bytes()...)
	key = append(key, toAssetID.Bytes()...)
	key = append(key, buf...)
	return append(key, utxoHash.Bytes()...)
}

func calcTradePairKey(fromAssetID, toAssetID *bc.AssetID) []byte {
	key := append(tradePairsPrefix, fromAssetID.Bytes()...)
	return append(key, toAssetID.Bytes()...)
}

func calcUTXOHash(order *common.Order) bc.Hash {
	prog := &bc.Program{VmVersion: 1, Code: order.Utxo.ControlProgram}
	src := &bc.ValueSource{
		Ref:      order.Utxo.SourceID,
		Value:    &bc.AssetAmount{AssetId: order.FromAssetID, Amount: order.Utxo.Amount},
		Position: order.Utxo.SourcePos,
	}
	o := bc.NewIntraChainOutput(src, prog, 0)
	return bc.EntryID(o)
}

func getAssetIDFromKey(key []byte, preFix []byte, posIndex int) *bc.AssetID {
	b := [32]byte{}
	pos := len(preFix) + assetIDLen*posIndex
	copy(b[:], key[pos:pos+assetIDLen])
	assetID := bc.NewAssetID(b)
	return &assetID
}

func getRateFromKey(key []byte, preFix []byte) float64 {
	ratePos := len(preFix) + assetIDLen*2
	return math.Float64frombits(binary.BigEndian.Uint64(key[ratePos : ratePos+rateByteLen]))
}

type tradePairData struct {
	Count int
}

type MovStore struct {
	db dbm.DB
}

func NewMovStore(db dbm.DB, height uint64, hash *bc.Hash) (*MovStore, error) {
	value := db.Get(bestMatchStore)
	if value == nil {
		state := &common.MovDatabaseState{Height: height, Hash: hash}
		value, err := json.Marshal(state)
		if err != nil {
			return nil, err
		}

		db.Set(bestMatchStore, value)
	}

	return &MovStore{db: db}, nil
}

func (d *MovStore) ListOrders(orderAfter *common.Order) ([]*common.Order, error) {
	if orderAfter.FromAssetID == nil || orderAfter.ToAssetID == nil {
		return nil, errors.New("assetID is nil")
	}

	orderPreFix := append(ordersPrefix, orderAfter.FromAssetID.Bytes()...)
	orderPreFix = append(orderPreFix, orderAfter.ToAssetID.Bytes()...)

	var startKey []byte
	if orderAfter.Rate > 0 {
		h := calcUTXOHash(orderAfter)
		startKey = calcOrderKey(orderAfter.FromAssetID, orderAfter.ToAssetID, &h, orderAfter.Rate)
	}

	itr := d.db.IteratorPrefixWithStart(orderPreFix, startKey, false)
	defer itr.Release()

	var orders []*common.Order
	for txNum := ordersNum; itr.Next() && txNum > 0; txNum-- {
		rate := getRateFromKey(itr.Key(), ordersPrefix)

		movUtxo := &common.MovUtxo{}
		if err := json.Unmarshal(itr.Value(), movUtxo); err != nil {
			return nil, err
		}

		order := &common.Order{
			FromAssetID: orderAfter.FromAssetID,
			ToAssetID:   orderAfter.ToAssetID,
			Rate:        rate,
			Utxo:        movUtxo,
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (d *MovStore) ProcessOrders(addOrders []*common.Order, delOreders []*common.Order, blockHeader *types.BlockHeader) error {
	if err := d.checkMovDatabaseState(blockHeader); err != nil {
		return err
	}

	batch := d.db.NewBatch()

	if err := d.addOrders(batch, addOrders); err != nil {
		return err
	}

	if err := d.deleteOrders(batch, delOreders); err != nil {
		return err
	}

	hash := blockHeader.Hash()
	if err := d.saveMovDatabaseState(batch, &common.MovDatabaseState{Height: blockHeader.Height, Hash: &hash}); err != nil {
		return err
	}

	batch.Write()
	return nil
}

func (d *MovStore) addOrders(batch dbm.Batch, orders []*common.Order) error {
	tradePairsCnt := make(map[common.TradePair]int)
	for _, order := range orders {
		utxoHash := calcUTXOHash(order)
		key := calcOrderKey(order.FromAssetID, order.ToAssetID, &utxoHash, order.Rate)

		data, err := json.Marshal(order.Utxo)
		if err != nil {
			return err
		}

		batch.Set(key, data)

		tradePair := common.TradePair{
			FromAssetID: order.FromAssetID,
			ToAssetID:   order.ToAssetID,
		}
		tradePairsCnt[tradePair] += 1
	}

	return d.updateTradePairs(batch, tradePairsCnt)
}

func (d *MovStore) deleteOrders(batch dbm.Batch, orders []*common.Order) error {
	tradePairs := make(map[common.TradePair]int)
	for _, order := range orders {
		utxoHash := calcUTXOHash(order)
		key := calcOrderKey(order.FromAssetID, order.ToAssetID, &utxoHash, order.Rate)
		batch.Delete(key)

		tradePair := common.TradePair{
			FromAssetID: order.FromAssetID,
			ToAssetID:   order.ToAssetID,
		}
		tradePairs[tradePair] -= 1
	}

	return d.updateTradePairs(batch, tradePairs)
}

func (d *MovStore) GetMovDatabaseState() (*common.MovDatabaseState, error) {
	if value := d.db.Get(bestMatchStore); value != nil {
		state := &common.MovDatabaseState{}
		return state, json.Unmarshal(value, state)
	}

	return nil, errors.New("don't find state of mov-database")
}

func (d *MovStore) ListTradePairsWithStart(fromAssetIDAfter, toAssetIDAfter *bc.AssetID) ([]*common.TradePair, error) {
	var startKey []byte
	if fromAssetIDAfter != nil && toAssetIDAfter != nil {
		startKey = calcTradePairKey(fromAssetIDAfter, toAssetIDAfter)
	}

	itr := d.db.IteratorPrefixWithStart(tradePairsPrefix, startKey, false)
	defer itr.Release()

	var tradePairs []*common.TradePair
	for txNum := tradePairsNum; itr.Next() && txNum > 0; txNum-- {
		key := itr.Key()
		fromAssetID := getAssetIDFromKey(key, tradePairsPrefix, 0)
		toAssetID := getAssetIDFromKey(key, tradePairsPrefix, 1)

		tradePairData := &tradePairData{}
		if err := json.Unmarshal(itr.Value(), tradePairData); err != nil {
			return nil, err
		}

		tradePairs = append(tradePairs, &common.TradePair{FromAssetID: fromAssetID, ToAssetID: toAssetID, Count: tradePairData.Count})
	}

	return tradePairs, nil
}

func (d *MovStore) updateTradePairs(batch dbm.Batch, tradePairs map[common.TradePair]int) error {
	for k, v := range tradePairs {
		key := calcTradePairKey(k.FromAssetID, k.ToAssetID)
		tradePairData := &tradePairData{}
		if value := d.db.Get(key); value != nil {
			if err := json.Unmarshal(value, tradePairData); err != nil {
				return err
			}
		} else if v < 0 {
			return errors.New("don't find trade pair")
		}

		tradePairData.Count += v
		if tradePairData.Count > 0 {
			value, err := json.Marshal(tradePairData)
			if err != nil {
				return err
			}

			batch.Set(key, value)
		} else {
			batch.Delete(key)
		}
	}
	return nil
}

func (d *MovStore) checkMovDatabaseState(header *types.BlockHeader) error {
	state, err := d.GetMovDatabaseState()
	if err != nil {
		return err
	}

	if (state.Hash.String() == header.PreviousBlockHash.String() && (state.Height+1) == header.Height) || state.Height == (header.Height+1) {
		return nil
	}

	return errors.New("the status of the block is inconsistent with that of mov-database")
}

func (d *MovStore) saveMovDatabaseState(batch dbm.Batch, state *common.MovDatabaseState) error {
	value, err := json.Marshal(state)
	if err != nil {
		return err
	}

	batch.Set(bestMatchStore, value)
	return nil
}