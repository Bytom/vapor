package proposal

import (
	"sort"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/account"
	"github.com/vapor/blockchain/txbuilder"
	"github.com/vapor/consensus"
	"github.com/vapor/errors"
	"github.com/vapor/protocol"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/protocol/state"
	"github.com/vapor/protocol/validation"
	"github.com/vapor/protocol/vm/vmutil"
)

const (
	logModule = "mining"
)

// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.
func createCoinbaseTx(chain *protocol.Chain, accountManager *account.Manager, preBlockHeader *types.BlockHeader) (tx *types.Tx, err error) {
	preBlockHash := preBlockHeader.Hash()
	consensusResult, err := chain.GetConsensusResultByHash(&preBlockHash)
	if err != nil {
		return nil, err
	}

	rewards, err := consensusResult.GetCoinbaseRewards(preBlockHeader.Height)
	if err != nil {
		return nil, err
	}

	arbitrary := append([]byte{0x00}, []byte(strconv.FormatUint(preBlockHeader.Height + 1, 10))...)
	var script []byte
	if accountManager == nil {
		script, err = vmutil.DefaultCoinbaseProgram()
	} else {
		script, err = accountManager.GetCoinbaseControlProgram()
		arbitrary = append(arbitrary, accountManager.GetCoinbaseArbitrary()...)
	}
	if err != nil {
		return nil, err
	}

	if len(arbitrary) > consensus.ActiveNetParams.CoinbaseArbitrarySizeLimit {
		return nil, validation.ErrCoinbaseArbitraryOversize
	}

	builder := txbuilder.NewBuilder(time.Now())
	if err = builder.AddInput(types.NewCoinbaseInput(arbitrary), &txbuilder.SigningInstruction{}); err != nil {
		return nil, err
	}
	if err = builder.AddOutput(types.NewIntraChainOutput(*consensus.BTMAssetID, 0, script)); err != nil {
		return nil, err
	}

	for _, r := range rewards {
		if err = builder.AddOutput(types.NewIntraChainOutput(*consensus.BTMAssetID, r.Amount, r.ControlProgram)); err != nil {
			return nil, err
		}
	}

	_, txData, err := builder.Build()
	if err != nil {
		return nil, err
	}

	byteData, err := txData.MarshalText()
	if err != nil {
		return nil, err
	}

	txData.SerializedSize = uint64(len(byteData))
	tx = &types.Tx{
		TxData: *txData,
		Tx:     types.MapTx(txData),
	}
	return tx, nil
}

// NewBlockTemplate returns a new block template that is ready to be solved
func NewBlockTemplate(chain *protocol.Chain, txPool *protocol.TxPool, accountManager *account.Manager, timestamp uint64) (*types.Block, error) {
	block, err := createBasicBlock(chain, accountManager, timestamp)
	if err != nil {
		return nil, err
	}

	view := state.NewUtxoViewpoint()
	txStatus := bc.NewTransactionStatus()
	if err := txStatus.SetStatus(0, false); err != nil {
		return nil, err
	}

	gasLeft, err := applyTransactionFromPool(chain, view, block, txStatus, int64(consensus.ActiveNetParams.MaxBlockGas))
	if err != nil {
		return nil, err
	}
	
	if err := applyTransactionFromSubProtocol(chain, view, block, txStatus, accountManager, gasLeft); err != nil {
		return nil, err
	}

	var txEntries []*bc.Tx
	for _, tx := range block.Transactions {
		txEntries = append(txEntries, tx.Tx)
	}

	block.BlockHeader.BlockCommitment.TransactionsMerkleRoot, err = types.TxMerkleRoot(txEntries)
	if err != nil {
		return nil, err
	}

	block.BlockHeader.BlockCommitment.TransactionStatusHash, err = types.TxStatusMerkleRoot(txStatus.VerifyStatus)

	_, err = chain.SignBlock(block)
	return block, err
}

func createBasicBlock(chain *protocol.Chain, accountManager *account.Manager, timestamp uint64) (*types.Block, error) {
	preBlockHeader := chain.BestBlockHeader()
	coinbaseTx, err := createCoinbaseTx(chain, accountManager, preBlockHeader)
	if err != nil {
		return nil, errors.Wrap(err, "fail on create coinbase tx")
	}

	return &types.Block{
		BlockHeader: types.BlockHeader{
			Version:           1,
			Height:            preBlockHeader.Height + 1,
			PreviousBlockHash: preBlockHeader.Hash(),
			Timestamp:         timestamp,
			BlockCommitment:   types.BlockCommitment{},
			BlockWitness:      types.BlockWitness{Witness: make([][]byte, consensus.ActiveNetParams.NumOfConsensusNode)},
		},
		Transactions: []*types.Tx{coinbaseTx},
	}, nil
}

func applyTransactionFromPool(chain *protocol.Chain, view *state.UtxoViewpoint, block *types.Block, txStatus *bc.TransactionStatus, gasLeft int64) (int64, error) {
	poolTxs := getAllTxsFromPool(chain.GetTxPool())
	results, gasLeft := preValidateTxs(poolTxs, chain, view, gasLeft)
	for _, result := range results {
		if result.err != nil {
			blkGenSkipTxForErr(chain.GetTxPool(), &result.tx.ID, result.err)
			continue
		}

		if err := txStatus.SetStatus(len(block.Transactions), result.gasOnly); err != nil {
			return 0, err
		}

		block.Transactions = append(block.Transactions, result.tx)
	}
	return gasLeft, nil
}

func applyTransactionFromSubProtocol(chain *protocol.Chain, view *state.UtxoViewpoint, block *types.Block, txStatus *bc.TransactionStatus, accountManager *account.Manager, gasLeft int64) error {
	txs, err := getTxsFromSubProtocols(chain, accountManager, gasLeft)
	if err != nil {
		return err
	}

	results, gasLeft := preValidateTxs(txs, chain, view, gasLeft)
	for _, result := range results {
		if result.err != nil {
			continue
		}

		if err := txStatus.SetStatus(len(block.Transactions), result.gasOnly); err != nil {
			return err
		}

		block.Transactions = append(block.Transactions, result.tx)
	}
	return nil
}

type validateTxResult struct {
	tx      *types.Tx
	gasOnly bool
	err     error
}

func preValidateTxs(txs []*types.Tx, chain *protocol.Chain, view *state.UtxoViewpoint, gasLeft int64) ([]*validateTxResult, int64) {
	var results []*validateTxResult

	bcBlock := &bc.Block{BlockHeader: &bc.BlockHeader{Height: chain.BestBlockHeight() + 1}}
	bcTxs := make([]*bc.Tx, len(txs))
	for i, tx := range txs {
		bcTxs[i] = tx.Tx
	}

	validateResults := validation.ValidateTxs(bcTxs, bcBlock)
	for i := 0; i < len(validateResults) && gasLeft > 0; i++ {
		gasOnlyTx := false
		gasStatus := validateResults[i].GetGasState()
		if err := validateResults[i].GetError(); err != nil {
			if !gasStatus.GasValid {
				results = append(results, &validateTxResult{tx: txs[i], err: err})
				continue
			}
			gasOnlyTx = true
		}

		if err := chain.GetTransactionsUtxo(view, []*bc.Tx{bcTxs[i]}); err != nil {
			results = append(results, &validateTxResult{tx: txs[i], err: err})
			continue
		}

		if gasLeft-gasStatus.GasUsed < 0 {
			break
		}

		if err := view.ApplyTransaction(bcBlock, bcTxs[i], gasOnlyTx); err != nil {
			results = append(results, &validateTxResult{tx: txs[i], err: err})
			continue
		}

		results = append(results, &validateTxResult{tx: txs[i], gasOnly: gasOnlyTx})
		gasLeft -= gasStatus.GasUsed
	}
	return results, gasLeft
}

func getAllTxsFromPool(txPool *protocol.TxPool) []*types.Tx {
	txDescList := txPool.GetTransactions()
	sort.Sort(byTime(txDescList))

	poolTxs := make([]*types.Tx, len(txDescList))
	for i, txDesc := range txDescList {
		poolTxs[i] = txDesc.Tx
	}
	return poolTxs
}

func getTxsFromSubProtocols(chain *protocol.Chain, accountManager *account.Manager, gasLeft int64) ([]*types.Tx, error) {
	cp, err := accountManager.GetCoinbaseControlProgram()
	if err != nil {
		return nil, err
	}

	var result []*types.Tx
	var subTxs []*types.Tx
	for i, p := range chain.SubProtocols() {
		if gasLeft <= 0 {
			break
		}

		subTxs, gasLeft, err = p.BeforeProposalBlock(cp, gasLeft, calcGasUsed(&bc.Block{BlockHeader: &bc.BlockHeader{Height: chain.BestBlockHeight() + 1}}))
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "index": i, "error": err}).Error("failed on sub protocol txs package")
			continue
		}

		result = append(result, subTxs...)
	}
	return result, nil
}

func calcGasUsed(bcBlock *bc.Block) func (*types.Tx) (int64, error) {
	return func (tx *types.Tx) (int64, error) {
		gasState, err := validation.ValidateTx(tx.Tx, bcBlock)
		if err != nil {
			return 0, err
		}

		return gasState.GasUsed, nil
	}
}

func blkGenSkipTxForErr(txPool *protocol.TxPool, txHash *bc.Hash, err error) {
	log.WithFields(log.Fields{"module": logModule, "error": err}).Error("mining block generation: skip tx due to")
	txPool.RemoveTransaction(txHash)
}
