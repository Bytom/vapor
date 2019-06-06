package synchron

import (
	// "database/sql"
	// "encoding/hex"
	// "encoding/json"
	// "fmt"
	// "math/big"
	// "sort"

	// "github.com/bytom/consensus"
	// "github.com/bytom/errors"
	// TODO:
	btmBc "github.com/bytom/protocol/bc"
	btmTypes "github.com/bytom/protocol/bc/types"
	"github.com/jinzhu/gorm"
	// TODO:
	// "github.com/blockcenter/coin/btm"
	// "github.com/blockcenter/database/orm"
)

type attachBlockProcessor struct {
	db *gorm.DB
	// coin     *orm.Coin
	block    *btmTypes.Block
	txStatus *btmBc.TransactionStatus
}

func (p *attachBlockProcessor) processIssuing(db *gorm.DB, txs []*btmTypes.Tx) error {
	return addIssueAssets(db, txs)
}

/*
func (p *attachBlockProcessor) getBlock() *btmTypes.Block {
	return p.block
}

func (p *attachBlockProcessor) getTxStatus() *bc.TransactionStatus {
	return p.txStatus
}

func (p *attachBlockProcessor) getCoin() *orm.Coin {
	return p.coin
}

func (p *attachBlockProcessor) processCoinInfo() error {
	blockHash := p.block.Hash()
	p.coin.BlockHeight = p.block.Height
	p.coin.BlockHash = blockHash.String()
	db := p.db.Model(p.coin).Where("block_hash = ?", p.block.PreviousBlockHash.String()).Updates(p.coin)
	if err := db.Error; err != nil {
		return err
	}

	if db.RowsAffected != 1 {
		return ErrInconsistentDB
	}
	return nil
}

type addressTxSorter []*orm.AddressTransaction

func (a addressTxSorter) Len() int      { return len(a) }
func (a addressTxSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a addressTxSorter) Less(i, j int) bool {
	return a[i].TransactionID < a[j].TransactionID ||
		(a[i].TransactionID == a[j].TransactionID && a[i].AddressID < a[j].AddressID) ||
		(a[i].TransactionID == a[j].TransactionID && a[i].AddressID == a[j].AddressID && a[i].AssetID < a[j].AssetID)
}

func (p *attachBlockProcessor) processAddressTransaction(mappings []*addressTxMapping) error {
	txMap := make(map[string]int64)
	addressTxMap := make(map[string]*orm.AddressTransaction)

	for _, m := range mappings {
		txHash := m.transaction.ID.String()
		if _, ok := txMap[txHash]; !ok {
			txID, err := p.upsertTransaction(m)
			if err != nil {
				return err
			}

			txMap[txHash] = txID
		}

		// is smart contract
		if m.address == nil {
			continue
		}

		var amount int64
		switch m.source.(type) {
		case *btmTypes.TxInput:
			amount -= int64(m.amount)

		case *btmTypes.TxOutput:
			amount = int64(m.amount)
		}

		addressTxKey := fmt.Sprintf("%d:%d:%d", m.address.ID, txMap[txHash], m.asset.ID)
		addressTx, ok := addressTxMap[addressTxKey]
		if !ok {
			addressTx = &orm.AddressTransaction{
				AddressID:     m.address.ID,
				TransactionID: uint64(txMap[txHash]),
				AssetID:       m.asset.ID,
			}
			addressTxMap[addressTxKey] = addressTx
		}

		addressTx.Amount += amount
	}

	var mergedAddrTxs []*orm.AddressTransaction
	for _, addressTx := range addressTxMap {
		mergedAddrTxs = append(mergedAddrTxs, addressTx)
	}
	sort.Sort(addressTxSorter(mergedAddrTxs))

	for _, addressTx := range mergedAddrTxs {
		if err := p.db.Where(addressTx).FirstOrCreate(addressTx).Error; err != nil {
			return err
		}
	}
	return nil
}

func (p *attachBlockProcessor) upsertTransaction(mapping *addressTxMapping) (int64, error) {
	rawTx, err := mapping.transaction.MarshalText()
	if err != nil {
		return 0, err
	}

	tx := &orm.Transaction{Hash: mapping.transaction.ID.String()}
	p.db.Unscoped().Where(tx).First(tx)
	// collided confirmed tx hash
	if tx.BlockHeight > 0 {
		return int64(tx.ID), nil
	}

	tx.CoinID = p.coin.ID
	tx.TxIndex = mapping.txIndex
	tx.RawData = string(rawTx)
	tx.BlockHeight = p.block.Height
	tx.BlockTimestamp = p.block.Timestamp
	tx.StatusFail = mapping.statusFail
	return int64(tx.ID), p.db.Unscoped().Save(tx).Error
}

func (p *attachBlockProcessor) processSpendBalance(input *btmTypes.TxInput, deltaBalance *deltaBalance) {
	amount := big.NewInt(0)
	amount.SetUint64(input.Amount())
	deltaBalance.Balance.Sub(deltaBalance.Balance, amount)
	deltaBalance.TotalSent.Add(deltaBalance.TotalSent, amount)
}

func (p *attachBlockProcessor) processReceiveBalance(output *btmTypes.TxOutput, deltaBalance *deltaBalance) {
	amount := big.NewInt(0)
	amount.SetUint64(output.Amount)
	deltaBalance.Balance.Add(deltaBalance.Balance, amount)
	deltaBalance.TotalReceived.Add(deltaBalance.TotalReceived, amount)
}

func (p *attachBlockProcessor) processSpendUTXO(utxoIDList []string) error {
	return p.db.Model(&orm.Utxo{}).Where("hash in (?)", utxoIDList).Update("is_spend", true).Error
}

func (p *attachBlockProcessor) processReceiveUTXO(m *addressTxMapping) error {
	outputID := m.transaction.OutputID(m.sourceIndex)
	output, err := m.transaction.Output(*outputID)
	if err != nil {
		return err
	}

	rawUtxo := &btm.UTXO{
		SourceID:  output.Source.Ref,
		SourcePos: uint64(m.sourceIndex),
	}
	rawData, err := json.Marshal(rawUtxo)
	if err != nil {
		return err
	}

	validHeight := p.block.Height
	if m.txIndex == 0 && p.block.Height != 0 {
		validHeight += consensus.CoinbasePendingBlockNumber
	}

	var cp []byte
	switch source := m.source.(type) {
	case *btmTypes.TxOutput:
		cp = source.ControlProgram
	default:
		return errors.New("wrong source type for processReceiveUTXO")
	}

	utxo := &orm.Utxo{Hash: outputID.String()}
	err = p.db.Where(&orm.Utxo{Hash: outputID.String()}).First(utxo).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	utxo.BlockHeight = p.block.Height
	utxo.ValidHeight = validHeight
	utxo.IsSpend = false
	utxo.AssetID = m.asset.ID
	utxo.Amount = output.Source.Value.Amount
	utxo.RawData = string(rawData)
	utxo.ControlProgram = hex.EncodeToString(cp)

	if m.address != nil {
		utxo.AddressID = sql.NullInt64{Int64: int64(m.address.ID), Valid: true}
	}

	if err == gorm.ErrRecordNotFound {
		return p.db.Create(utxo).Error
	}
	return p.db.Model(&orm.Utxo{}).Update(utxo).Error
}
*/
