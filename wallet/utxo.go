package wallet

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/account"
	"github.com/vapor/consensus"
	"github.com/vapor/consensus/segwit"
	"github.com/vapor/crypto/sha3pool"
	"github.com/vapor/database"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

// GetAccountUtxos return all account unspent outputs
func (w *Wallet) GetAccountUtxos(accountID string, id string, unconfirmed, isSmartContract bool, vote bool) []*account.UTXO {
	prefix := database.UTXOPrefix
	if isSmartContract {
		prefix = database.SUTXOPrefix
	}

	accountUtxos := []*account.UTXO{}
	if unconfirmed {
		accountUtxos = w.AccountMgr.ListUnconfirmedUtxo(accountID, isSmartContract)
	}

	rawConfirmedUTXOs := w.store.GetAccountUTXOs(string(prefix) + id)
	confirmedUTXOs := []*account.UTXO{}
	for _, rawConfirmedUTXO := range rawConfirmedUTXOs {
		confirmedUTXO := new(account.UTXO)
		if err := json.Unmarshal(rawConfirmedUTXO, confirmedUTXO); err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Warn("GetAccountUTXOs fail on unmarshal utxo")
			continue
		}
		confirmedUTXOs = append(confirmedUTXOs, confirmedUTXO)
	}
	accountUtxos = append(accountUtxos, confirmedUTXOs...)

	newAccountUtxos := []*account.UTXO{}
	for _, accountUtxo := range accountUtxos {
		if vote && accountUtxo.Vote == nil {
			continue
		}

		if accountID == accountUtxo.AccountID || accountID == "" {
			newAccountUtxos = append(newAccountUtxos, accountUtxo)
		}
	}
	return newAccountUtxos
}

func (w *Wallet) attachUtxos(b *types.Block, txStatus *bc.TransactionStatus) {
	for txIndex, tx := range b.Transactions {
		statusFail, err := txStatus.GetStatus(txIndex)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("attachUtxos fail on get tx status")
			continue
		}

		//hand update the transaction input utxos
		inputUtxos := txInToUtxos(tx, statusFail)
		for _, inputUtxo := range inputUtxos {
			if segwit.IsP2WScript(inputUtxo.ControlProgram) {
				w.store.DeleteStardardUTXO(inputUtxo.OutputID)
			} else {
				w.store.DeleteContractUTXO(inputUtxo.OutputID)
			}
		}

		//hand update the transaction output utxos
		outputUtxos := txOutToUtxos(tx, statusFail, b.Height)
		utxos := w.filterAccountUtxo(outputUtxos)
		if err := w.saveUtxos(utxos); err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("attachUtxos fail on saveUtxos")
		}
	}
}

func (w *Wallet) detachUtxos(b *types.Block, txStatus *bc.TransactionStatus) {
	for txIndex := len(b.Transactions) - 1; txIndex >= 0; txIndex-- {
		tx := b.Transactions[txIndex]
		for j := range tx.Outputs {
			code := []byte{}
			switch resOut := tx.Entries[*tx.ResultIds[j]].(type) {
			case *bc.IntraChainOutput:
				code = resOut.ControlProgram.Code
			case *bc.VoteOutput:
				code = resOut.ControlProgram.Code
			default:
				continue
			}

			if segwit.IsP2WScript(code) {
				w.store.DeleteStardardUTXO(*tx.ResultIds[j])
			} else {
				w.store.DeleteContractUTXO(*tx.ResultIds[j])
			}
		}

		statusFail, err := txStatus.GetStatus(txIndex)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("detachUtxos fail on get tx status")
			continue
		}

		inputUtxos := txInToUtxos(tx, statusFail)
		utxos := w.filterAccountUtxo(inputUtxos)
		if err := w.saveUtxos(utxos); err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("detachUtxos fail on batchSaveUtxos")
			return
		}
	}
}

func (w *Wallet) filterAccountUtxo(utxos []*account.UTXO) []*account.UTXO {
	outsByScript := make(map[string][]*account.UTXO, len(utxos))
	for _, utxo := range utxos {
		scriptStr := string(utxo.ControlProgram)
		outsByScript[scriptStr] = append(outsByScript[scriptStr], utxo)
	}

	result := make([]*account.UTXO, 0, len(utxos))
	for s := range outsByScript {
		if !segwit.IsP2WScript([]byte(s)) {
			for _, utxo := range outsByScript[s] {
				result = append(result, utxo)
			}
			continue
		}

		var hash [32]byte
		sha3pool.Sum256(hash[:], []byte(s))
		data := w.store.GetRawProgram(hash)
		if data == nil {
			continue
		}

		cp := &account.CtrlProgram{}
		if err := json.Unmarshal(data, cp); err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("filterAccountUtxo fail on unmarshal control program")
			continue
		}

		for _, utxo := range outsByScript[s] {
			utxo.AccountID = cp.AccountID
			utxo.Address = cp.Address
			utxo.ControlProgramIndex = cp.KeyIndex
			utxo.Change = cp.Change
			result = append(result, utxo)
		}
	}
	return result
}

func (w *Wallet) saveUtxos(utxos []*account.UTXO) error {
	for _, utxo := range utxos {
		data, err := json.Marshal(utxo)
		if err != nil {
			return errors.Wrap(err, "failed marshal accountutxo")
		}

		if segwit.IsP2WScript(utxo.ControlProgram) {
			w.store.SetStandardUTXO(utxo.OutputID, data)
		} else {
			w.store.SetContractUTXO(utxo.OutputID, data)
		}
	}
	return nil
}

func txInToUtxos(tx *types.Tx, statusFail bool) []*account.UTXO {
	utxos := []*account.UTXO{}
	for _, inpID := range tx.Tx.InputIDs {

		e, err := tx.Entry(inpID)
		if err != nil {
			continue
		}
		utxo := &account.UTXO{}
		switch inp := e.(type) {
		case *bc.Spend:
			resOut, err := tx.IntraChainOutput(*inp.SpentOutputId)
			if err != nil {
				log.WithFields(log.Fields{"module": logModule, "err": err}).Error("txInToUtxos fail on get resOut for spedn")
				continue
			}
			if statusFail && *resOut.Source.Value.AssetId != *consensus.BTMAssetID {
				continue
			}
			utxo = &account.UTXO{
				OutputID:       *inp.SpentOutputId,
				AssetID:        *resOut.Source.Value.AssetId,
				Amount:         resOut.Source.Value.Amount,
				ControlProgram: resOut.ControlProgram.Code,
				SourceID:       *resOut.Source.Ref,
				SourcePos:      resOut.Source.Position,
			}
		case *bc.VetoInput:
			resOut, err := tx.VoteOutput(*inp.SpentOutputId)
			if err != nil {
				log.WithFields(log.Fields{"module": logModule, "err": err}).Error("txInToUtxos fail on get resOut for vetoInput")
				continue
			}
			if statusFail && *resOut.Source.Value.AssetId != *consensus.BTMAssetID {
				continue
			}
			utxo = &account.UTXO{
				OutputID:       *inp.SpentOutputId,
				AssetID:        *resOut.Source.Value.AssetId,
				Amount:         resOut.Source.Value.Amount,
				ControlProgram: resOut.ControlProgram.Code,
				SourceID:       *resOut.Source.Ref,
				SourcePos:      resOut.Source.Position,
				Vote:           resOut.Vote,
			}
		default:
			continue
		}
		utxos = append(utxos, utxo)
	}
	return utxos
}

func txOutToUtxos(tx *types.Tx, statusFail bool, blockHeight uint64) []*account.UTXO {
	validHeight := uint64(0)
	if tx.Inputs[0].InputType() == types.CoinbaseInputType {
		validHeight = blockHeight + consensus.CoinbasePendingBlockNumber
	}

	utxos := []*account.UTXO{}
	for i, out := range tx.Outputs {
		entryOutput, err := tx.Entry(*tx.ResultIds[i])
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("txOutToUtxos fail on get entryOutput")
			continue
		}

		utxo := &account.UTXO{}
		switch bcOut := entryOutput.(type) {
		case *bc.IntraChainOutput:
			if statusFail && *out.AssetAmount().AssetId != *consensus.BTMAssetID {
				continue
			}
			utxo = &account.UTXO{
				OutputID:       *tx.OutputID(i),
				AssetID:        *out.AssetAmount().AssetId,
				Amount:         out.AssetAmount().Amount,
				ControlProgram: out.ControlProgram(),
				SourceID:       *bcOut.Source.Ref,
				SourcePos:      bcOut.Source.Position,
				ValidHeight:    validHeight,
			}

		case *bc.VoteOutput:
			if statusFail && *out.AssetAmount().AssetId != *consensus.BTMAssetID {
				continue
			}

			voteValidHeight := blockHeight + consensus.VotePendingBlockNumber
			if validHeight < voteValidHeight {
				validHeight = voteValidHeight
			}

			utxo = &account.UTXO{
				OutputID:       *tx.OutputID(i),
				AssetID:        *out.AssetAmount().AssetId,
				Amount:         out.AssetAmount().Amount,
				ControlProgram: out.ControlProgram(),
				SourceID:       *bcOut.Source.Ref,
				SourcePos:      bcOut.Source.Position,
				ValidHeight:    validHeight,
				Vote:           bcOut.Vote,
			}

		default:
			log.WithFields(log.Fields{"module": logModule}).Warn("txOutToUtxos fail on get bcOut")
			continue
		}

		utxos = append(utxos, utxo)
	}
	return utxos
}
