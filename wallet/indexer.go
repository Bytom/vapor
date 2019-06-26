package wallet

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/account"
	"github.com/vapor/blockchain/query"
	"github.com/vapor/consensus"
	"github.com/vapor/crypto/sha3pool"
	chainjson "github.com/vapor/encoding/json"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

func parseGlobalTxIdx(globalTxIdx []byte) (*bc.Hash, uint64) {
	var hashBytes [32]byte
	copy(hashBytes[:], globalTxIdx[:32])
	hash := bc.NewHash(hashBytes)
	position := binary.BigEndian.Uint64(globalTxIdx[32:])
	return &hash, position
}

// saveExternalAssetDefinition save external and local assets definition,
// when query ,query local first and if have no then query external
// details see getAliasDefinition
func saveExternalAssetDefinition(b *types.Block, store WalletStorer) {
	store.InitBatch()
	defer store.CommitBatch()

	for _, tx := range b.Transactions {
		for _, orig := range tx.Inputs {
			if cci, ok := orig.TypedInput.(*types.CrossChainInput); ok {
				assetID := cci.AssetId
				if assetExist := store.GetAssetDefinition(assetID); assetExist == nil {
					store.SetAssetDefinition(assetID, cci.AssetDefinition)
				}
			}
		}
	}
}

// Summary is the struct of transaction's input and output summary
type Summary struct {
	Type         string             `json:"type"`
	AssetID      bc.AssetID         `json:"asset_id,omitempty"`
	AssetAlias   string             `json:"asset_alias,omitempty"`
	Amount       uint64             `json:"amount,omitempty"`
	AccountID    string             `json:"account_id,omitempty"`
	AccountAlias string             `json:"account_alias,omitempty"`
	Arbitrary    chainjson.HexBytes `json:"arbitrary,omitempty"`
}

// TxSummary is the struct of transaction summary
type TxSummary struct {
	ID        bc.Hash   `json:"tx_id"`
	Timestamp uint64    `json:"block_time"`
	Inputs    []Summary `json:"inputs"`
	Outputs   []Summary `json:"outputs"`
}

// indexTransactions saves all annotated transactions to the database.
func (w *Wallet) indexTransactions(b *types.Block, txStatus *bc.TransactionStatus, annotatedTxs []*query.AnnotatedTx) error {
	for _, tx := range annotatedTxs {
		rawTx, err := json.Marshal(tx)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("inserting annotated_txs to db")
			return err
		}

		w.store.SetTransaction(b.Height, tx.Position, tx.ID.String(), rawTx)
		w.store.DeleteUnconfirmedTransaction(tx.ID.String())
	}

	if !w.TxIndexFlag {
		return nil
	}

	for position, globalTx := range b.Transactions {
		blockHash := b.BlockHeader.Hash()
		w.store.SetGlobalTransactionIndex(globalTx.ID.String(), &blockHash, uint64(position))
	}

	return nil
}

// filterAccountTxs related and build the fully annotated transactions.
func (w *Wallet) filterAccountTxs(b *types.Block, txStatus *bc.TransactionStatus) []*query.AnnotatedTx {
	annotatedTxs := make([]*query.AnnotatedTx, 0, len(b.Transactions))

transactionLoop:
	for pos, tx := range b.Transactions {
		statusFail, _ := txStatus.GetStatus(pos)
		for _, v := range tx.Outputs {
			var hash [32]byte
			sha3pool.Sum256(hash[:], v.ControlProgram())

			if bytes := w.store.GetRawProgram(hash); bytes != nil {
				annotatedTxs = append(annotatedTxs, w.buildAnnotatedTransaction(tx, b, statusFail, pos))
				continue transactionLoop
			}
		}

		for _, v := range tx.Inputs {
			outid, err := v.SpentOutputID()
			if err != nil {
				continue
			}
			if bytes := w.store.GetStandardUTXO(outid); bytes != nil {
				annotatedTxs = append(annotatedTxs, w.buildAnnotatedTransaction(tx, b, statusFail, pos))
				continue transactionLoop
			}
		}
	}

	return annotatedTxs
}

// GetTransactionByTxID get transaction by txID
func (w *Wallet) GetTransactionByTxID(txID string) (*query.AnnotatedTx, error) {
	if annotatedTx, err := w.getAccountTxByTxID(txID); err == nil {
		return annotatedTx, nil
	} else if !w.TxIndexFlag {
		return nil, err
	}

	return w.getGlobalTxByTxID(txID)
}

func (w *Wallet) getAccountTxByTxID(txID string) (*query.AnnotatedTx, error) {
	annotatedTx := &query.AnnotatedTx{}
	txInfo, err := w.store.GetTransaction(txID)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(txInfo, annotatedTx); err != nil {
		return nil, err
	}

	annotateTxsAsset(w, []*query.AnnotatedTx{annotatedTx})
	return annotatedTx, nil
}

func (w *Wallet) getGlobalTxByTxID(txID string) (*query.AnnotatedTx, error) {
	globalTxIdx := w.store.GetGlobalTransaction(txID)
	if globalTxIdx == nil {
		return nil, fmt.Errorf("No transaction(tx_id=%s) ", txID)
	}

	blockHash, pos := parseGlobalTxIdx(globalTxIdx)
	block, err := w.chain.GetBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	txStatus, err := w.chain.GetTransactionStatus(blockHash)
	if err != nil {
		return nil, err
	}

	statusFail, err := txStatus.GetStatus(int(pos))
	if err != nil {
		return nil, err
	}

	tx := block.Transactions[int(pos)]
	return w.buildAnnotatedTransaction(tx, block, statusFail, int(pos)), nil
}

// GetTransactionsSummary get transactions summary
func (w *Wallet) GetTransactionsSummary(transactions []*query.AnnotatedTx) []TxSummary {
	Txs := []TxSummary{}

	for _, annotatedTx := range transactions {
		tmpTxSummary := TxSummary{
			Inputs:    make([]Summary, len(annotatedTx.Inputs)),
			Outputs:   make([]Summary, len(annotatedTx.Outputs)),
			ID:        annotatedTx.ID,
			Timestamp: annotatedTx.Timestamp,
		}

		for i, input := range annotatedTx.Inputs {
			tmpTxSummary.Inputs[i].Type = input.Type
			tmpTxSummary.Inputs[i].AccountID = input.AccountID
			tmpTxSummary.Inputs[i].AccountAlias = input.AccountAlias
			tmpTxSummary.Inputs[i].AssetID = input.AssetID
			tmpTxSummary.Inputs[i].AssetAlias = input.AssetAlias
			tmpTxSummary.Inputs[i].Amount = input.Amount
			tmpTxSummary.Inputs[i].Arbitrary = input.Arbitrary
		}
		for j, output := range annotatedTx.Outputs {
			tmpTxSummary.Outputs[j].Type = output.Type
			tmpTxSummary.Outputs[j].AccountID = output.AccountID
			tmpTxSummary.Outputs[j].AccountAlias = output.AccountAlias
			tmpTxSummary.Outputs[j].AssetID = output.AssetID
			tmpTxSummary.Outputs[j].AssetAlias = output.AssetAlias
			tmpTxSummary.Outputs[j].Amount = output.Amount
		}

		Txs = append(Txs, tmpTxSummary)
	}

	return Txs
}

func findTransactionsByAccount(annotatedTx *query.AnnotatedTx, accountID string) bool {
	for _, input := range annotatedTx.Inputs {
		if input.AccountID == accountID {
			return true
		}
	}

	for _, output := range annotatedTx.Outputs {
		if output.AccountID == accountID {
			return true
		}
	}

	return false
}

// GetTransactions get all walletDB transactions, and filter transactions by accountID optional
func (w *Wallet) GetTransactions(accountID string) ([]*query.AnnotatedTx, error) {
	annotatedTxs := []*query.AnnotatedTx{}
	annotatedTxs, err := w.store.GetTransactions()
	if err != nil {
		return nil, err
	}

	newAnnotatedTxs := []*query.AnnotatedTx{}
	for _, annotatedTx := range annotatedTxs {
		if accountID == "" || findTransactionsByAccount(annotatedTx, accountID) {
			annotateTxsAsset(w, []*query.AnnotatedTx{annotatedTx})
			newAnnotatedTxs = append([]*query.AnnotatedTx{annotatedTx}, newAnnotatedTxs...)
		}
	}

	return newAnnotatedTxs, nil
}

// GetAccountBalances return all account balances
func (w *Wallet) GetAccountBalances(accountID string, id string) ([]AccountBalance, error) {
	return w.indexBalances(w.GetAccountUtxos(accountID, "", false, false, false))
}

// AccountBalance account balance
type AccountBalance struct {
	AccountID       string                 `json:"account_id"`
	Alias           string                 `json:"account_alias"`
	AssetAlias      string                 `json:"asset_alias"`
	AssetID         string                 `json:"asset_id"`
	Amount          uint64                 `json:"amount"`
	AssetDefinition map[string]interface{} `json:"asset_definition"`
}

func (w *Wallet) indexBalances(accountUTXOs []*account.UTXO) ([]AccountBalance, error) {
	accBalance := make(map[string]map[string]uint64)
	balances := []AccountBalance{}

	for _, accountUTXO := range accountUTXOs {
		assetID := accountUTXO.AssetID.String()
		if _, ok := accBalance[accountUTXO.AccountID]; ok {
			if _, ok := accBalance[accountUTXO.AccountID][assetID]; ok {
				accBalance[accountUTXO.AccountID][assetID] += accountUTXO.Amount
			} else {
				accBalance[accountUTXO.AccountID][assetID] = accountUTXO.Amount
			}
		} else {
			accBalance[accountUTXO.AccountID] = map[string]uint64{assetID: accountUTXO.Amount}
		}
	}

	var sortedAccount []string
	for k := range accBalance {
		sortedAccount = append(sortedAccount, k)
	}
	sort.Strings(sortedAccount)

	for _, id := range sortedAccount {
		var sortedAsset []string
		for k := range accBalance[id] {
			sortedAsset = append(sortedAsset, k)
		}
		sort.Strings(sortedAsset)

		for _, assetID := range sortedAsset {
			alias := w.AccountMgr.GetAliasByID(id)
			targetAsset, err := w.AssetReg.GetAsset(assetID)
			if err != nil {
				return nil, err
			}

			assetAlias := *targetAsset.Alias
			balances = append(balances, AccountBalance{
				Alias:           alias,
				AccountID:       id,
				AssetID:         assetID,
				AssetAlias:      assetAlias,
				Amount:          accBalance[id][assetID],
				AssetDefinition: targetAsset.DefinitionMap,
			})
		}
	}

	return balances, nil
}

// GetAccountVotes return all account votes
func (w *Wallet) GetAccountVotes(accountID string, id string) ([]AccountVotes, error) {
	return w.indexVotes(w.GetAccountUtxos(accountID, "", false, false, true))
}

type voteDetail struct {
	Vote       string `json:"vote"`
	VoteNumber uint64 `json:"vote_number"`
}

// AccountVotes account vote
type AccountVotes struct {
	AccountID       string       `json:"account_id"`
	Alias           string       `json:"account_alias"`
	TotalVoteNumber uint64       `json:"total_vote_number"`
	VoteDetails     []voteDetail `json:"vote_details"`
}

func (w *Wallet) indexVotes(accountUTXOs []*account.UTXO) ([]AccountVotes, error) {
	accVote := make(map[string]map[string]uint64)
	votes := []AccountVotes{}

	for _, accountUTXO := range accountUTXOs {
		if accountUTXO.AssetID != *consensus.BTMAssetID || accountUTXO.Vote == nil {
			continue
		}
		xpub := hex.EncodeToString(accountUTXO.Vote)
		if _, ok := accVote[accountUTXO.AccountID]; ok {
			accVote[accountUTXO.AccountID][xpub] += accountUTXO.Amount
		} else {
			accVote[accountUTXO.AccountID] = map[string]uint64{xpub: accountUTXO.Amount}

		}
	}

	var sortedAccount []string
	for k := range accVote {
		sortedAccount = append(sortedAccount, k)
	}
	sort.Strings(sortedAccount)

	for _, id := range sortedAccount {
		var sortedXpub []string
		for k := range accVote[id] {
			sortedXpub = append(sortedXpub, k)
		}
		sort.Strings(sortedXpub)

		voteDetails := []voteDetail{}
		voteTotal := uint64(0)
		for _, xpub := range sortedXpub {
			voteDetails = append(voteDetails, voteDetail{
				Vote:       xpub,
				VoteNumber: accVote[id][xpub],
			})
			voteTotal += accVote[id][xpub]
		}
		alias := w.AccountMgr.GetAliasByID(id)
		votes = append(votes, AccountVotes{
			Alias:           alias,
			AccountID:       id,
			VoteDetails:     voteDetails,
			TotalVoteNumber: voteTotal,
		})
	}

	return votes, nil
}
