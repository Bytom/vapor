package wallet

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/vapor/account"
	"github.com/vapor/asset"
	"github.com/vapor/blockchain/query"
	"github.com/vapor/common"
	"github.com/vapor/consensus"
	"github.com/vapor/consensus/segwit"
	"github.com/vapor/crypto/sha3pool"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

// annotateTxs adds asset data to transactions
func annotateTxsAsset(w *Wallet, txs []*query.AnnotatedTx) {
	for i, tx := range txs {
		for j, input := range tx.Inputs {
			alias, definition := w.getAliasDefinition(input.AssetID)
			txs[i].Inputs[j].AssetAlias, txs[i].Inputs[j].AssetDefinition = alias, &definition
		}
		for k, output := range tx.Outputs {
			alias, definition := w.getAliasDefinition(output.AssetID)
			txs[i].Outputs[k].AssetAlias, txs[i].Outputs[k].AssetDefinition = alias, &definition
		}
	}
}

func (w *Wallet) getExternalDefinition(assetID *bc.AssetID) json.RawMessage {
	externalAsset, err := w.store.GetAssetDefinition(assetID)
	if err != nil {
		log.WithFields(log.Fields{"module": logModule, "err": err}).Warning("fail on get asset definition.")
	}
	if externalAsset == nil {
		return nil
	}

	if err := w.AssetReg.SaveAsset(externalAsset, *externalAsset.Alias); err != nil {
		log.WithFields(log.Fields{"module": logModule, "err": err, "assetAlias": *externalAsset.Alias}).Warning("fail on save external asset to internal asset DB")
	}
	return json.RawMessage(externalAsset.RawDefinitionByte)
}

func (w *Wallet) getAliasDefinition(assetID bc.AssetID) (string, json.RawMessage) {
	//btm
	if assetID.String() == consensus.BTMAssetID.String() {
		alias := consensus.BTMAlias
		definition := []byte(asset.DefaultNativeAsset.RawDefinitionByte)

		return alias, definition
	}

	//local asset and saved external asset
	if localAsset, err := w.AssetReg.FindByID(nil, &assetID); err == nil {
		alias := *localAsset.Alias
		definition := []byte(localAsset.RawDefinitionByte)
		return alias, definition
	}

	//external asset
	if definition := w.getExternalDefinition(&assetID); definition != nil {
		return assetID.String(), definition
	}

	return "", nil
}

// annotateTxs adds account data to transactions
func annotateTxsAccount(txs []*query.AnnotatedTx, store WalletStorer) {
	for i, tx := range txs {
		for j, input := range tx.Inputs {
			//issue asset tx input SpentOutputID is nil
			if input.SpentOutputID == nil {
				continue
			}
			localAccount, err := getAccountFromACP(input.ControlProgram, store)
			if localAccount == nil || err != nil {
				continue
			}
			txs[i].Inputs[j].AccountAlias = localAccount.Alias
			txs[i].Inputs[j].AccountID = localAccount.ID
		}
		for j, output := range tx.Outputs {
			localAccount, err := getAccountFromACP(output.ControlProgram, store)
			if localAccount == nil || err != nil {
				continue
			}
			txs[i].Outputs[j].AccountAlias = localAccount.Alias
			txs[i].Outputs[j].AccountID = localAccount.ID
		}
	}
}

func getAccountFromACP(program []byte, store WalletStorer) (*account.Account, error) {
	var hash common.Hash

	sha3pool.Sum256(hash[:], program)
	accountCP, err := store.GetControlProgram(hash)
	if err != nil {
		return nil, err
	}
	if accountCP == nil {
		return nil, fmt.Errorf("failed get account control program:%x ", hash)
	}

	account, err := store.GetAccountByAccountID(accountCP.AccountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("failed get account:%s ", accountCP.AccountID)
	}
	return account, nil
}

var emptyJSONObject = json.RawMessage(`{}`)

func (w *Wallet) buildAnnotatedTransaction(orig *types.Tx, b *types.Block, statusFail bool, indexInBlock int) *query.AnnotatedTx {
	tx := &query.AnnotatedTx{
		ID:                     orig.ID,
		Timestamp:              b.Timestamp,
		BlockID:                b.Hash(),
		BlockHeight:            b.Height,
		Position:               uint32(indexInBlock),
		BlockTransactionsCount: uint32(len(b.Transactions)),
		Inputs:                 make([]*query.AnnotatedInput, 0, len(orig.Inputs)),
		Outputs:                make([]*query.AnnotatedOutput, 0, len(orig.Outputs)),
		StatusFail:             statusFail,
		Size:                   orig.SerializedSize,
	}
	for i := range orig.Inputs {
		tx.Inputs = append(tx.Inputs, w.BuildAnnotatedInput(orig, uint32(i)))
	}
	for i := range orig.Outputs {
		tx.Outputs = append(tx.Outputs, w.BuildAnnotatedOutput(orig, i))
	}
	return tx
}

// BuildAnnotatedInput build the annotated input.
func (w *Wallet) BuildAnnotatedInput(tx *types.Tx, i uint32) *query.AnnotatedInput {
	orig := tx.Inputs[i]
	in := &query.AnnotatedInput{
		AssetDefinition: &emptyJSONObject,
	}
	if orig.InputType() != types.CoinbaseInputType {
		in.AssetID = orig.AssetID()
		in.Amount = orig.Amount()
	}

	id := tx.Tx.InputIDs[i]
	in.InputID = id
	e := tx.Entries[id]
	switch e := e.(type) {
	case *bc.VetoInput:
		in.Type = "veto"
		in.ControlProgram = orig.ControlProgram()
		in.Address = w.getAddressFromControlProgram(in.ControlProgram, false)
		in.SpentOutputID = e.SpentOutputId
		arguments := orig.Arguments()
		for _, arg := range arguments {
			in.WitnessArguments = append(in.WitnessArguments, arg)
		}

	case *bc.CrossChainInput:
		in.Type = "cross_chain_in"
		in.ControlProgram = orig.ControlProgram()
		in.Address = w.getAddressFromControlProgram(in.ControlProgram, true)
		in.SpentOutputID = e.MainchainOutputId
		arguments := orig.Arguments()
		for _, arg := range arguments {
			in.WitnessArguments = append(in.WitnessArguments, arg)
		}

	case *bc.Spend:
		in.Type = "spend"
		in.ControlProgram = orig.ControlProgram()
		in.Address = w.getAddressFromControlProgram(in.ControlProgram, false)
		in.SpentOutputID = e.SpentOutputId
		arguments := orig.Arguments()
		for _, arg := range arguments {
			in.WitnessArguments = append(in.WitnessArguments, arg)
		}

	case *bc.Coinbase:
		in.Type = "coinbase"
		in.Arbitrary = e.Arbitrary
	}
	return in
}

func (w *Wallet) getAddressFromControlProgram(prog []byte, isMainchain bool) string {
	netParams := &consensus.ActiveNetParams
	if isMainchain {
		netParams = &consensus.MainNetParams
	}

	if segwit.IsP2WPKHScript(prog) {
		if pubHash, err := segwit.GetHashFromStandardProg(prog); err == nil {
			return BuildP2PKHAddress(pubHash, netParams)
		}
	} else if segwit.IsP2WSHScript(prog) {
		if scriptHash, err := segwit.GetHashFromStandardProg(prog); err == nil {
			return BuildP2SHAddress(scriptHash, netParams)
		}
	}

	return ""
}

func BuildP2PKHAddress(pubHash []byte, netParams *consensus.Params) string {
	address, err := common.NewAddressWitnessPubKeyHash(pubHash, netParams)
	if err != nil {
		return ""
	}

	return address.EncodeAddress()
}

func BuildP2SHAddress(scriptHash []byte, netParams *consensus.Params) string {
	address, err := common.NewAddressWitnessScriptHash(scriptHash, netParams)
	if err != nil {
		return ""
	}

	return address.EncodeAddress()
}

// BuildAnnotatedOutput build the annotated output.
func (w *Wallet) BuildAnnotatedOutput(tx *types.Tx, idx int) *query.AnnotatedOutput {
	orig := tx.Outputs[idx]
	outid := tx.OutputID(idx)
	out := &query.AnnotatedOutput{
		OutputID:        *outid,
		Position:        idx,
		AssetID:         *orig.AssetAmount().AssetId,
		AssetDefinition: &emptyJSONObject,
		Amount:          orig.AssetAmount().Amount,
		ControlProgram:  orig.ControlProgram(),
	}

	var isMainchainAddress bool
	switch e := tx.Entries[*outid].(type) {
	case *bc.IntraChainOutput:
		out.Type = "control"
		isMainchainAddress = false

	case *bc.CrossChainOutput:
		out.Type = "cross_chain_out"
		isMainchainAddress = true

	case *bc.VoteOutput:
		out.Type = "vote"
		out.Vote = e.Vote
		isMainchainAddress = false
	}

	out.Address = w.getAddressFromControlProgram(orig.ControlProgram(), isMainchainAddress)
	return out
}
