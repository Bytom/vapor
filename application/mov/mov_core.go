package mov

import (
	"github.com/bytom/vapor/application/mov/common"
	"github.com/bytom/vapor/application/mov/contract"
	"github.com/bytom/vapor/application/mov/database"
	"github.com/bytom/vapor/application/mov/match"
	"github.com/bytom/vapor/consensus/segwit"
	dbm "github.com/bytom/vapor/database/leveldb"
	"github.com/bytom/vapor/errors"
	"github.com/bytom/vapor/protocol/bc"
	"github.com/bytom/vapor/protocol/bc/types"
)

const maxFeeRate = 0.05

var (
	errInvalidTradePairs             = errors.New("The trade pairs in the tx input is invalid")
	errStatusFailMustFalse           = errors.New("status fail of transaction does not allow to be true")
	errInputProgramMustP2WMCScript   = errors.New("input program of trade tx must p2wmc script")
	errExistCancelOrderInMatchedTx   = errors.New("can't exist cancel order in the matched transaction")
	errExistTradeInCancelOrderTx     = errors.New("can't exist trade in the cancel order transaction")
	errAmountOfFeeGreaterThanMaximum = errors.New("amount of fee greater than max fee amount")
	errAssetIDMustUniqueInMatchedTx  = errors.New("asset id must unique in matched transaction")
	errRatioOfTradeLessThanZero      = errors.New("ratio arguments must greater than zero")
	errSpendOutputIDIsIncorrect      = errors.New("spend output id of matched tx is not equals to actual matched tx")
	errRequestAmountMath             = errors.New("request amount of order less than one or big than max of int64")
	errNotMatchedOrder               = errors.New("order in matched tx is not matched")
)

// MovCore represent the core logic of the match module, which include generate match transactions before packing the block,
// verify the match transaction in block is correct, and update the order table according to the transaction.
type MovCore struct {
	movStore         database.MovStore
	startBlockHeight uint64
}

// NewMovCore return a instance of MovCore by path of mov db
func NewMovCore(dbBackend, dbDir string, startBlockHeight uint64) *MovCore {
	movDB := dbm.NewDB("mov", dbBackend, dbDir)
	return &MovCore{movStore: database.NewLevelDBMovStore(movDB), startBlockHeight: startBlockHeight}
}

// ApplyBlock parse pending order and cancel from the the transactions of block
// and add pending order to the dex db, remove cancel order from dex db.
func (m *MovCore) ApplyBlock(block *types.Block) error {
	if block.Height < m.startBlockHeight {
		return nil
	}

	if block.Height == m.startBlockHeight {
		blockHash := block.Hash()
		if err := m.movStore.InitDBState(block.Height, &blockHash); err != nil {
			return err
		}

		return nil
	}

	if err := m.validateMatchedTxSequence(block.Transactions); err != nil {
		return err
	}

	addOrders, deleteOrders, err := applyTransactions(block.Transactions)
	if err != nil {
		return err
	}

	return m.movStore.ProcessOrders(addOrders, deleteOrders, &block.BlockHeader)
}

// BeforeProposalBlock return all transactions than can be matched, and the number of transactions cannot exceed the given capacity.
func (m *MovCore) BeforeProposalBlock(txs []*types.Tx, nodeProgram []byte, blockHeight uint64, gasLeft int64, isTimeout func() bool) ([]*types.Tx, error) {
	if blockHeight <= m.startBlockHeight {
		return nil, nil
	}

	orderBook, err := buildOrderBook(m.movStore, txs)
	if err != nil {
		return nil, err
	}

	matchEngine := match.NewEngine(orderBook, maxFeeRate, nodeProgram)
	tradePairMap := make(map[string]bool)
	tradePairIterator := database.NewTradePairIterator(m.movStore)

	var packagedTxs []*types.Tx
	for gasLeft > 0 && !isTimeout() && tradePairIterator.HasNext() {
		tradePair := tradePairIterator.Next()
		if tradePairMap[tradePair.Key()] {
			continue
		}
		tradePairMap[tradePair.Key()] = true
		tradePairMap[tradePair.Reverse().Key()] = true

		for gasLeft > 0 && !isTimeout() && matchEngine.HasMatchedTx(tradePair, tradePair.Reverse()) {
			matchedTx, err := matchEngine.NextMatchedTx(tradePair, tradePair.Reverse())
			if err != nil {
				return nil, err
			}

			gasUsed := calcMatchedTxGasUsed(matchedTx)
			if gasLeft-gasUsed >= 0 {
				packagedTxs = append(packagedTxs, matchedTx)
			}
			gasLeft -= gasUsed
		}
	}
	return packagedTxs, nil
}

// ChainStatus return the current block height and block hash in dex core
func (m *MovCore) ChainStatus() (uint64, *bc.Hash, error) {
	state, err := m.movStore.GetMovDatabaseState()
	if err != nil {
		return 0, nil, err
	}

	return state.Height, state.Hash, nil
}

// DetachBlock parse pending order and cancel from the the transactions of block
// and add cancel order to the dex db, remove pending order from dex db.
func (m *MovCore) DetachBlock(block *types.Block) error {
	if block.Height <= m.startBlockHeight {
		return nil
	}

	deleteOrders, addOrders, err := applyTransactions(block.Transactions)
	if err != nil {
		return err
	}

	return m.movStore.ProcessOrders(addOrders, deleteOrders, &block.BlockHeader)
}

// IsDust block the transaction that are not generated by the match engine
func (m *MovCore) IsDust(tx *types.Tx) bool {
	for _, input := range tx.Inputs {
		if segwit.IsP2WMCScript(input.ControlProgram()) && !contract.IsCancelClauseSelector(input) {
			return true
		}
	}
	return false
}

// Name return the name of current module
func (m *MovCore) Name() string {
	return "MOV"
}

// StartHeight return the start block height of current module
func (m *MovCore) StartHeight() uint64 {
	return m.startBlockHeight
}

// ValidateBlock no need to verify the block header, because the first module has been verified.
// just need to verify the transactions in the block.
func (m *MovCore) ValidateBlock(block *types.Block, verifyResults []*bc.TxVerifyResult) error {
	return m.ValidateTxs(block.Transactions, verifyResults)
}

// ValidateTxs validate the trade transaction.
func (m *MovCore) ValidateTxs(txs []*types.Tx, verifyResults []*bc.TxVerifyResult) error {
	for i, tx := range txs {
		if err := m.ValidateTx(tx, verifyResults[i]); err != nil {
			return err
		}
	}
	return nil
}

// ValidateTxs validate one transaction.
func (m *MovCore) ValidateTx(tx *types.Tx, verifyResult *bc.TxVerifyResult) error {
	if common.IsMatchedTx(tx) {
		if err := validateMatchedTx(tx, verifyResult); err != nil {
			return err
		}
	}

	if common.IsCancelOrderTx(tx) {
		if err := validateCancelOrderTx(tx, verifyResult); err != nil {
			return err
		}
	}

	for _, output := range tx.Outputs {
		if !segwit.IsP2WMCScript(output.ControlProgram()) {
			continue
		}
		if verifyResult.StatusFail {
			return errStatusFailMustFalse
		}

		if err := validateMagneticContractArgs(output.AssetAmount(), output.ControlProgram()); err != nil {
			return err
		}
	}
	return nil
}

func validateCancelOrderTx(tx *types.Tx, verifyResult *bc.TxVerifyResult) error {
	if verifyResult.StatusFail {
		return errStatusFailMustFalse
	}

	for _, input := range tx.Inputs {
		if !segwit.IsP2WMCScript(input.ControlProgram()) {
			return errInputProgramMustP2WMCScript
		}

		if contract.IsTradeClauseSelector(input) {
			return errExistTradeInCancelOrderTx
		}
	}
	return nil
}

func validateMagneticContractArgs(fromAssetAmount bc.AssetAmount, program []byte) error {
	contractArgs, err := segwit.DecodeP2WMCProgram(program)
	if err != nil {
		return err
	}

	if *fromAssetAmount.AssetId == contractArgs.RequestedAsset {
		return errInvalidTradePairs
	}

	if contractArgs.RatioNumerator <= 0 || contractArgs.RatioDenominator <= 0 {
		return errRatioOfTradeLessThanZero
	}

	if match.CalcRequestAmount(fromAssetAmount.Amount, contractArgs) < 1 {
		return errRequestAmountMath
	}
	return nil
}

func validateMatchedTx(tx *types.Tx, verifyResult *bc.TxVerifyResult) error {
	if verifyResult.StatusFail {
		return errStatusFailMustFalse
	}

	fromAssetIDMap := make(map[string]bool)
	toAssetIDMap := make(map[string]bool)
	for i, input := range tx.Inputs {
		if !segwit.IsP2WMCScript(input.ControlProgram()) {
			return errInputProgramMustP2WMCScript
		}

		if contract.IsCancelClauseSelector(input) {
			return errExistCancelOrderInMatchedTx
		}

		order, err := common.NewOrderFromInput(tx, i)
		if err != nil {
			return err
		}

		fromAssetIDMap[order.FromAssetID.String()] = true
		toAssetIDMap[order.ToAssetID.String()] = true
	}

	if len(fromAssetIDMap) != len(tx.Inputs) || len(toAssetIDMap) != len(tx.Inputs) {
		return errAssetIDMustUniqueInMatchedTx
	}

	return validateMatchedTxFeeAmount(tx)
}

func validateMatchedTxFeeAmount(tx *types.Tx) error {
	txFee, err := match.CalcMatchedTxFee(&tx.TxData, maxFeeRate)
	if err != nil {
		return err
	}

	for _, amount := range txFee {
		if amount.FeeAmount > amount.MaxFeeAmount {
			return errAmountOfFeeGreaterThanMaximum
		}
	}
	return nil
}

func (m *MovCore) validateMatchedTxSequence(txs []*types.Tx) error {
	orderBook, err := buildOrderBook(m.movStore, txs)
	if err != nil {
		return err
	}

	for _, matchedTx := range txs {
		if !common.IsMatchedTx(matchedTx) {
			continue
		}

		tradePairs, err := getTradePairsFromMatchedTx(matchedTx)
		if err != nil {
			return err
		}

		orders := orderBook.PeekOrders(tradePairs)
		if !match.IsMatched(orders) {
			return errNotMatchedOrder
		}

		if err := validateSpendOrders(matchedTx, orders); err != nil {
			return err
		}

		orderBook.PopOrders(tradePairs)

		for i, output := range matchedTx.Outputs {
			if !segwit.IsP2WMCScript(output.ControlProgram()) {
				continue
			}

			order, err := common.NewOrderFromOutput(matchedTx, i)
			if err != nil {
				return err
			}

			if err := orderBook.AddOrder(order); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSpendOrders(matchedTx *types.Tx, orders []*common.Order) error {
	spendOutputIDs := make(map[string]bool)
	for _, input := range matchedTx.Inputs {
		spendOutputID, err := input.SpentOutputID()
		if err != nil {
			return err
		}

		spendOutputIDs[spendOutputID.String()] = true
	}

	for _, order := range orders {
		outputID := order.UTXOHash().String()
		if _, ok := spendOutputIDs[outputID]; !ok {
			return errSpendOutputIDIsIncorrect
		}
	}
	return nil
}

func applyTransactions(txs []*types.Tx) ([]*common.Order, []*common.Order, error) {
	deleteOrderMap := make(map[string]*common.Order)
	addOrderMap := make(map[string]*common.Order)
	for _, tx := range txs {
		addOrders, err := getAddOrdersFromTx(tx)
		if err != nil {
			return nil, nil, err
		}

		for _, order := range addOrders {
			addOrderMap[order.Key()] = order
		}

		deleteOrders, err := getDeleteOrdersFromTx(tx)
		if err != nil {
			return nil, nil, err
		}

		for _, order := range deleteOrders {
			deleteOrderMap[order.Key()] = order
		}
	}

	addOrders, deleteOrders := mergeOrders(addOrderMap, deleteOrderMap)
	return addOrders, deleteOrders, nil
}

func buildOrderBook(store database.MovStore, txs []*types.Tx) (*match.OrderBook, error) {
	var nonMatchedTxs []*types.Tx
	for _, tx := range txs {
		if !common.IsMatchedTx(tx) {
			nonMatchedTxs = append(nonMatchedTxs, tx)
		}
	}

	var arrivalAddOrders, arrivalDelOrders []*common.Order
	for _, tx := range nonMatchedTxs {
		addOrders, err := getAddOrdersFromTx(tx)
		if err != nil {
			return nil, err
		}

		delOrders, err := getDeleteOrdersFromTx(tx)
		if err != nil {
			return nil, err
		}

		arrivalAddOrders = append(arrivalAddOrders, addOrders...)
		arrivalDelOrders = append(arrivalDelOrders, delOrders...)
	}

	return match.NewOrderBook(store, arrivalAddOrders, arrivalDelOrders), nil
}

func calcMatchedTxGasUsed(tx *types.Tx) int64 {
	return int64(len(tx.Inputs))*150 + int64(tx.SerializedSize)
}

func getAddOrdersFromTx(tx *types.Tx) ([]*common.Order, error) {
	var orders []*common.Order
	for i, output := range tx.Outputs {
		if output.OutputType() != types.IntraChainOutputType || !segwit.IsP2WMCScript(output.ControlProgram()) {
			continue
		}

		order, err := common.NewOrderFromOutput(tx, i)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}
	return orders, nil
}

func getDeleteOrdersFromTx(tx *types.Tx) ([]*common.Order, error) {
	var orders []*common.Order
	for i, input := range tx.Inputs {
		if input.InputType() != types.SpendInputType || !segwit.IsP2WMCScript(input.ControlProgram()) {
			continue
		}

		order, err := common.NewOrderFromInput(tx, i)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}
	return orders, nil
}

func getTradePairsFromMatchedTx(tx *types.Tx) ([]*common.TradePair, error) {
	var tradePairs []*common.TradePair
	for _, tx := range tx.Inputs {
		contractArgs, err := segwit.DecodeP2WMCProgram(tx.ControlProgram())
		if err != nil {
			return nil, err
		}

		tradePairs = append(tradePairs, &common.TradePair{FromAssetID: tx.AssetAmount().AssetId, ToAssetID: &contractArgs.RequestedAsset})
	}
	return tradePairs, nil
}

func mergeOrders(addOrderMap, deleteOrderMap map[string]*common.Order) ([]*common.Order, []*common.Order) {
	var deleteOrders, addOrders []*common.Order
	for orderID, order := range addOrderMap {
		if _, ok := deleteOrderMap[orderID]; ok {
			delete(deleteOrderMap, orderID)
			continue
		}
		addOrders = append(addOrders, order)
	}

	for _, order := range deleteOrderMap {
		deleteOrders = append(deleteOrders, order)
	}
	return addOrders, deleteOrders
}
