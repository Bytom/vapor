package protocol

import (
	log "github.com/sirupsen/logrus"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/protocol/state"
	"github.com/vapor/protocol/validation"
)

// ErrBadTx is returned for transactions failing validation
var ErrBadTx = errors.New("invalid transaction")

// GetTransactionStatus return the transaction status of give block
func (c *Chain) GetTransactionStatus(hash *bc.Hash) (*bc.TransactionStatus, error) {
	return c.store.GetTransactionStatus(hash)
}

// GetTransactionsUtxo return all the utxos that related to the txs' inputs
func (c *Chain) GetTransactionsUtxo(view *state.UtxoViewpoint, txs []*bc.Tx) error {
	return c.store.GetTransactionsUtxo(view, txs)
}

// ValidateTx validates the given transaction. A cache holds
// per-transaction validation results and is consulted before
// performing full validation.
func (c *Chain) ValidateTx(tx *types.Tx) (bool, error) {
	if ok := c.txPool.HaveTransaction(&tx.ID); ok {
		return false, c.txPool.GetErrCache(&tx.ID)
	}

	if ok := c.txPool.IsWithdrawSpent(tx); ok {
		log.WithFields(log.Fields{"module": "ValidateTx", "error": "Cliam transactions already exist in the trading pool"}).Error("chain error")
		return false, errors.New("Cliam transactions already exist in the trading pool")
	}

	bh := c.BestBlockHeader()
	block := types.MapBlock(&types.Block{BlockHeader: *bh})
	gasStatus, err := validation.ValidateTx(tx.Tx, block)
	if gasStatus.GasValid == false {
		c.txPool.AddErrCache(&tx.ID, err)
		return false, err
	}

	if err != nil {
		log.WithFields(log.Fields{"tx_id": tx.Tx.ID.String(), "error": err}).Info("transaction status fail")
	}

	return c.txPool.ProcessTransaction(tx, err != nil, block.BlockHeader.Height, gasStatus.BTMValue)
}
