package mov

import (
	"github.com/vapor/application/mov/common"
	"github.com/vapor/application/mov/database"
	"github.com/vapor/application/mov/match"
	"github.com/vapor/application/mov/util"
	"github.com/vapor/consensus/segwit"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

type MovCore struct {
	movStore database.MovStore
}

// ChainStatus return the current block height and block hash in dex core
func (m *MovCore) ChainStatus() (uint64, *bc.Hash, error) {
	state, err := m.movStore.GetMovDatabaseState()
	if err != nil {
		return 0, nil, err
	}

	return state.Height, state.Hash, nil
}

func (m *MovCore) ValidateBlock(block *types.Block) error {
	if err := m.ValidateTxs(block.Transactions); err != nil {
		return err
	}
	return nil
}

// ValidateTxs validate the matched transactions is generated according to the matching rule.
func (m *MovCore) ValidateTxs(txs []*types.Tx) error {
	return nil
}

// ApplyBlock parse pending order and cancel from the the transactions of block
// and add pending order to the dex db, remove cancel order from dex db.
func (m *MovCore) ApplyBlock(block *types.Block) error {
	if err := m.validateMatchedTxs(block.Transactions); err != nil {
		return err
	}

	addOrders, deleteOrders, err := applyTransactions(block.Transactions)
	if err != nil {
		return err
	}

	return m.movStore.ProcessOrders(addOrders, deleteOrders, &block.BlockHeader)
}

func (m *MovCore) validateMatchedTxs(txs []*types.Tx) error {
	matchEngine := match.NewEngine(m.movStore)
	for _, matchedTx := range txs {
		if !isMatchedTx(matchedTx) {
			continue
		}

		tradePair := getTradePairFromMatchedTx(matchedTx)
		realMatchedTx, err := matchEngine.NextMatchedTx(tradePair, tradePair.Reverse())
		if err != nil {
			return err
		}

		if matchedTx.ID != realMatchedTx.ID {
			return errors.New("fail to validate match transaction")
		}
	}
	return nil
}

func getTradePairFromMatchedTx(tx *types.Tx) *common.TradePair {
	fromAssetID := tx.Inputs[0].AssetID()
	toAssetID := tx.Inputs[1].AssetID()
	return &common.TradePair{FromAssetID: &fromAssetID, ToAssetID: &toAssetID}
}

// DetachBlock parse pending order and cancel from the the transactions of block
// and add cancel order to the dex db, remove pending order from dex db.
func (m *MovCore) DetachBlock(block *types.Block) error {
	deleteOrders, addOrders, err := applyTransactions(block.Transactions)
	if err != nil {
		return err
	}

	return m.movStore.ProcessOrders(addOrders, deleteOrders, &block.BlockHeader)
}

// BeforeProposalBlock return all transactions than can be matched, and the number of transactions cannot exceed the given capacity.
func (m *MovCore) BeforeProposalBlock(capacity int) ([]*types.Tx, error) {
	matchEngine := match.NewEngine(m.movStore)
	tradePairMap := make(map[string]bool)
	tradePairIterator := database.NewTradePairIterator(m.movStore)
	remainder := capacity

	var packagedTxs []*types.Tx
	for remainder > 0 && tradePairIterator.HasNext() {
		tradePair := tradePairIterator.Next()
		if tradePairMap[tradePair.Key()] {
			continue
		}
		tradePairMap[tradePair.Key()] = true
		tradePairMap[tradePair.Reverse().Key()] = true

		for {
			matchedTx, err := matchEngine.NextMatchedTx(tradePair, tradePair.Reverse())
			if err != nil {
				return nil, err
			}

			if matchedTx == nil {
				break
			}
			packagedTxs = append(packagedTxs, matchedTx)
			remainder--
		}
	}
	return packagedTxs, nil
}

func (m *MovCore) IsDust(tx *types.Tx) bool {
	return false
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

func isMatchedTx(tx *types.Tx) bool {
	p2wmCount := 0
	for _, input := range tx.Inputs {
		if input.InputType() == types.SpendInputType && util.IsTradeClauseSelector(input) && segwit.IsP2WMCScript(input.ControlProgram()) {
			p2wmCount++
		}
	}
	return p2wmCount >= 2
}
