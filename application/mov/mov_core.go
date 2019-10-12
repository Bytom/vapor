package mov

import (
	"encoding/hex"
	"fmt"
	"github.com/vapor/protocol"

	"github.com/vapor/application/mov/common"
	"github.com/vapor/application/mov/database"
	"github.com/vapor/application/mov/match"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/protocol/vm"
)

type MovCore struct {
	movStore *database.MovStore
	chain    *protocol.Chain
}

// ChainStatus return the current block hegiht and block hash in dex core
func (m *MovCore) ChainStatus() (uint64, *bc.Hash, error) {
	state, err := m.movStore.GetMovDatabaseState()
	if err != nil {
		return 0, nil, err
	}

	return state.Height, state.Hash, nil
}

func (m *MovCore) ValidateBlock(block *types.Block) error {
	if err := m.ValidateTxs(block); err != nil {
		return err
	}
	return nil
}

// ValidateTxs validate the matched transactions is generated according to the matching rule.
func (m *MovCore) ValidateTxs(block *types.Block) error {
	deltaOrderMap, err := m.getDeltaOrders(block)
	if err != nil {
		return err
	}

	if err := m.validateMatchedTxs(block.Transactions, deltaOrderMap); err != nil {
		return err
	}
	return nil
}

func (m *MovCore) validateMatchedTxs(txs []*types.Tx, deltaOrderMap map[string]*database.DeltaOrders) error {
	var packagedMatchedTxs []string
	for _, tx := range txs {
		if isMatchedTx(tx) {
			packagedMatchedTxs = append(packagedMatchedTxs, tx.ID.String())
		}
	}

	var realMatchedTxs []string
	tradePairIterator := database.NewTradePairIterator(m.movStore)
	for tradePairIterator.HasNext() {
		matchedTxs, err := match.GenerateMatchedTxs(match.NewOrderTable(m.movStore, tradePairIterator.Next(), deltaOrderMap))
		if err != nil {
			return err
		}

		for _, matchedTx := range matchedTxs {
			realMatchedTxs = append(realMatchedTxs, matchedTx.ID.String())
		}
	}

	for i := 0; i < len(packagedMatchedTxs); i++ {
		if i >= len(realMatchedTxs) || packagedMatchedTxs[i] != realMatchedTxs[i] {
			return errors.New("fail to validate match transaction")
		}
	}
	return nil
}

func (m *MovCore) getDeltaOrders(block *types.Block) (map[string]*database.DeltaOrders, error) {
	_, bestBlockHash, err := m.ChainStatus()
	if err != nil {
		return nil, err
	}

	beginDetach, err := m.chain.GetHeaderByHash(bestBlockHash)
	if err != nil {
		return nil, err
	}

	beginAttach, err := m.chain.GetHeaderByHash(&block.PreviousBlockHash)
	if err != nil {
		return nil, err
	}

	attachBlockHeaders, detachBlockHeaders, err := m.chain.CalcReorganizeChain(beginAttach, beginDetach)
	if err != nil {
		return nil, err
	}

	return m.getDeltaOrdersHelper(attachBlockHeaders, detachBlockHeaders)
}

func (m *MovCore) getDeltaOrdersHelper(attachBlockHeaders, detachBlockHeaders []*types.BlockHeader) (map[string]*database.DeltaOrders, error) {
	var deleteOrders, addOrders []*common.Order
	for _, detachBlockHeader := range detachBlockHeaders {
		blockHash := detachBlockHeader.Hash()
		block, err := m.chain.GetBlockByHash(&blockHash)
		if err != nil {
			return nil, err
		}

		subDeleteOrders, subAddOrders, err := applyTransactions(block.Transactions)
		if err != nil {
			return nil, err
		}

		addOrders = append(addOrders, subAddOrders...)
		deleteOrders = append(deleteOrders, subDeleteOrders...)
	}

	for _, attachBlockHeader := range attachBlockHeaders {
		blockHash := attachBlockHeader.Hash()
		block, err := m.chain.GetBlockByHash(&blockHash)
		if err != nil {
			return nil, err
		}

		subAddOrders, subDeleteOrders, err := applyTransactions(block.Transactions)
		if err != nil {
			return nil, err
		}

		addOrders = append(addOrders, subAddOrders...)
		deleteOrders = append(deleteOrders, subDeleteOrders...)
	}

	deltaOrderMap := make(map[string]*database.DeltaOrders)

	for _, addOrder := range addOrders {
		key := fmt.Sprintf("%s:%s", addOrder.FromAssetID, addOrder.ToAssetID)
		if _, ok := deltaOrderMap[key]; !ok {
			deltaOrderMap[key] = &database.DeltaOrders{}
		}
		deltaOrderMap[key].AddOrders = append(deltaOrderMap[key].AddOrders, addOrder)
	}

	for _, deleteOrder := range deleteOrders {
		key := fmt.Sprintf("%s:%s", deleteOrder.FromAssetID, deleteOrder.ToAssetID)
		if _, ok := deltaOrderMap[key]; !ok {
			deltaOrderMap[key] = &database.DeltaOrders{}
		}
		deltaOrderMap[key].DeleteOrders = append(deltaOrderMap[key].DeleteOrders, deleteOrder)
	}
	return deltaOrderMap, nil
}

// ApplyBlock parse pending order and cancel from the the transactions of block
// and add pending order to the dex db, remove cancel order from dex db.
func (m *MovCore) ApplyBlock(block *types.Block) error {
	addOrders, deleteOrders, err := applyTransactions(block.Transactions)
	if err != nil {
		return err
	}

	return m.movStore.ProcessOrders(addOrders, deleteOrders, &block.BlockHeader)
}

// DetachBlock parse pending order and cancel from the the transactions of block
// and add cancel order to the dex db, remove pending order from dex db.
func (m *MovCore) DetachBlock(block *types.Block) error {
	deleteOrders, addOrders, err := applyTransactions(block.Transactions)
	if err != nil {
		return err
	}

	blockHeader, err := m.chain.GetHeaderByHash(&block.PreviousBlockHash)
	if err != nil {
		return err
	}

	return m.movStore.ProcessOrders(addOrders, deleteOrders, blockHeader)
}

// BeforeProposalBlock get all pending orders from the dex db, parse pending orders and cancel orders from transactions
// Then merge the two, use match engine to generate matched transactions, finally return them.
func (m *MovCore) BeforeProposalBlock(txs []*types.Tx, numOfPackage int) ([]*types.Tx, error) {
	var packagedTxs []*types.Tx
	tradePairIterator := database.NewTradePairIterator(m.movStore)

	for tradePairIterator.HasNext() {
		orderTable := match.NewOrderTable(m.movStore, tradePairIterator.Next(), nil)
		matchedTxs, err := match.GenerateMatchedTxs(orderTable)
		if err != nil {
			return nil, err
		}

		num := len(matchedTxs)
		if len(packagedTxs)+len(matchedTxs) > numOfPackage {
			num = numOfPackage - len(packagedTxs)
		}
		for i := 0; i < num; i++ {
			packagedTxs = append(packagedTxs, matchedTxs[i])
		}
	}
	return packagedTxs, nil
}

func (m *MovCore) IsDust(tx *types.Tx) bool {
	return false
}

func applyTransactions(txs []*types.Tx) ([]*common.Order, []*common.Order, error) {
	var addOrders []*common.Order
	var deleteOrders []*common.Order
	var matchedTxs []*types.Tx
	for _, tx := range txs {
		subAddOrders, err := getPendingOrderIfPresent(tx)
		if err != nil {
			return nil, nil, err
		}

		addOrders = append(addOrders, subAddOrders...)

		subDeleteOrders, err := getCancelOrderIfPresent(tx)
		if err != nil {
			return nil, nil, err
		}

		deleteOrders = append(deleteOrders, subDeleteOrders...)

		if isMatchedTx(tx) {
			matchedTxs = append(matchedTxs, tx)
		}
	}
	subAddOrders, subDeleteOrders, err := applyMatchedTxs(matchedTxs)
	if err != nil {
		return nil, nil, nil
	}

	addOrders = append(addOrders, subAddOrders...)
	deleteOrders = append(deleteOrders, subDeleteOrders...)
	return addOrders, deleteOrders, nil
}

func applyMatchedTxs(txs []*types.Tx) ([]*common.Order, []*common.Order, error) {
	deleteOrderMap := make(map[string]*common.Order)
	addOrderMap := make(map[string]*common.Order)
	for _, tx := range txs {
		tradeOrders, err := getTradeOrderIfPresent(tx)
		if err != nil {
			return nil, nil, err
		}

		for _, order := range tradeOrders {
			orderID := fmt.Sprintf("%s:%d", order.Utxo.SourceID, order.Utxo.SourcePos)
			deleteOrderMap[orderID] = order
		}

		pendingOrders, err := getPendingOrderIfPresent(tx)
		if err != nil {
			return nil, nil, err
		}

		for _, order := range pendingOrders {
			orderID := fmt.Sprintf("%s:%d", order.Utxo.SourceID, order.Utxo.SourcePos)
			addOrderMap[orderID] = order
		}
	}

	var deleteOrders, addOrders []*common.Order
	for orderID, order := range addOrderMap {
		if deleteOrderMap[orderID] != nil {
			delete(deleteOrderMap, orderID)
			continue
		}
		addOrders = append(addOrders, order)
	}
	for _, order := range deleteOrderMap {
		deleteOrders = append(deleteOrders, order)
	}
	return addOrders, deleteOrders, nil
}

func getPendingOrderIfPresent(tx *types.Tx) ([]*common.Order, error) {
	var orders []*common.Order
	for i, output := range tx.Outputs {
		if output.OutputType() == types.IntraChainOutputType && IsP2WMCScript(output.ControlProgram()) {
			order, err := common.OutputToOrder(tx, i)
			if err != nil {
				return nil, err
			}

			orders = append(orders, order)
		}
	}
	return orders, nil
}

func getTradeOrderIfPresent(tx *types.Tx) ([]*common.Order, error) {
	return getInputOrderByClauseSelector(tx, isTradeClauseSelector)
}

func getCancelOrderIfPresent(tx *types.Tx) ([]*common.Order, error) {
	return getInputOrderByClauseSelector(tx, isCancelClauseSelector)
}

func getInputOrderByClauseSelector(tx *types.Tx, checkClauseSelector func(*types.TxInput) bool) ([]*common.Order, error) {
	var orders []*common.Order
	for _, input := range tx.Inputs {
		if input.InputType() != types.SpendInputType || IsP2WMCScript(input.ControlProgram()) {
			continue
		}

		if checkClauseSelector(input) {
			order, err := common.InputToOrder(input)
			if err != nil {
				return nil, err
			}

			orders = append(orders, order)
		}
	}
	return orders, nil
}

func isMatchedTx(tx *types.Tx) bool {
	if len(tx.Inputs) != 2 {
		return false
	}

	if !IsP2WMCScript(tx.Inputs[0].ControlProgram()) || !IsP2WMCScript(tx.Inputs[1].ControlProgram()) {
		return false
	}

	if tx.Inputs[0].InputType() != types.SpendInputType || tx.Inputs[1].InputType() != types.SpendInputType {
		return false
	}

	if !isTradeClauseSelector(tx.Inputs[0]) {
		return false
	}

	return isTradeClauseSelector(tx.Inputs[1])
}

// TODO
func isCancelClauseSelector(input *types.TxInput) bool {
	return len(input.Arguments()) >= 2 && hex.EncodeToString(input.Arguments()[1]) == hex.EncodeToString(vm.Int64Bytes(2))
}

// TODO
func isTradeClauseSelector(input *types.TxInput) bool {
	if len(input.Arguments()) < 2 {
		return false
	}
	clauseSelector := hex.EncodeToString(input.Arguments()[1])
	return clauseSelector == hex.EncodeToString(vm.Int64Bytes(0)) || clauseSelector == hex.EncodeToString(vm.Int64Bytes(1))
}

// -------------------- mock -------------------

func IsP2WMCScript(prog []byte) bool {
	return false
}
