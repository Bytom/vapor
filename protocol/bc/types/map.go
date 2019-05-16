package types

import (
	log "github.com/sirupsen/logrus"

	"github.com/vapor/consensus"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/vm"
	"github.com/vapor/protocol/vm/vmutil"
)

// MapTx converts a types TxData object into its entries-based
// representation.
func MapTx(oldTx *TxData) *bc.Tx {
	txID, txHeader, entries := mapTx(oldTx)
	tx := &bc.Tx{
		TxHeader: txHeader,
		ID:       txID,
		Entries:  entries,
		InputIDs: make([]bc.Hash, len(oldTx.Inputs)),
	}

	spentOutputIDs := make(map[bc.Hash]bool)
	for id, e := range entries {
		var ord uint64
		switch e := e.(type) {
		case *bc.Issuance:
			ord = e.Ordinal

		case *bc.CrossChainInput:
			ord = e.Ordinal
			if *e.WitnessDestination.Value.AssetId == *consensus.BTMAssetID {
				tx.GasInputIDs = append(tx.GasInputIDs, id)
			}

		case *bc.Spend:
			ord = e.Ordinal
			spentOutputIDs[*e.SpentOutputId] = true
			if *e.WitnessDestination.Value.AssetId == *consensus.BTMAssetID {
				tx.GasInputIDs = append(tx.GasInputIDs, id)
			}

		case *bc.Coinbase:
			ord = 0
			tx.GasInputIDs = append(tx.GasInputIDs, id)

		default:
			continue
		}

		if ord >= uint64(len(oldTx.Inputs)) {
			continue
		}
		tx.InputIDs[ord] = id
	}

	for id := range spentOutputIDs {
		tx.SpentOutputIDs = append(tx.SpentOutputIDs, id)
	}
	return tx
}

func mapTx(tx *TxData) (headerID bc.Hash, hdr *bc.TxHeader, entryMap map[bc.Hash]bc.Entry) {
	entryMap = make(map[bc.Hash]bc.Entry)
	addEntry := func(e bc.Entry) bc.Hash {
		id := bc.EntryID(e)
		entryMap[id] = e
		return id
	}

	var (
		spends    []*bc.Spend
		issuances []*bc.Issuance
		crossIns  []*bc.CrossChainInput
		coinbase  *bc.Coinbase
	)

	muxSources := make([]*bc.ValueSource, len(tx.Inputs))
	for i, input := range tx.Inputs {
		switch inp := input.TypedInput.(type) {
		case *IssuanceInput:
			nonceHash := inp.NonceHash()
			assetDefHash := inp.AssetDefinitionHash()
			value := input.AssetAmount()

			issuance := bc.NewIssuance(&nonceHash, &value, uint64(i))
			issuance.WitnessAssetDefinition = &bc.AssetDefinition{
				Data: &assetDefHash,
				IssuanceProgram: &bc.Program{
					VmVersion: inp.VMVersion,
					Code:      inp.IssuanceProgram,
				},
			}
			issuance.WitnessArguments = inp.Arguments
			issuanceID := addEntry(issuance)

			muxSources[i] = &bc.ValueSource{
				Ref:   &issuanceID,
				Value: &value,
			}
			issuances = append(issuances, issuance)

		case *SpendInput:
			// create entry for prevout
			prog := &bc.Program{VmVersion: inp.VMVersion, Code: inp.ControlProgram}
			src := &bc.ValueSource{
				Ref:      &inp.SourceID,
				Value:    &inp.AssetAmount,
				Position: inp.SourcePosition,
			}
			prevout := bc.NewIntraChainOutput(src, prog, 0) // ordinal doesn't matter for prevouts, only for result outputs
			prevoutID := addEntry(prevout)
			// create entry for spend
			spend := bc.NewSpend(&prevoutID, uint64(i))
			spend.WitnessArguments = inp.Arguments
			spendID := addEntry(spend)
			// setup mux
			muxSources[i] = &bc.ValueSource{
				Ref:   &spendID,
				Value: &inp.AssetAmount,
			}
			spends = append(spends, spend)

		case *CoinbaseInput:
			coinbase = bc.NewCoinbase(inp.Arbitrary)
			coinbaseID := addEntry(coinbase)

			out := tx.Outputs[0]
			value := out.AssetAmount()
			muxSources[i] = &bc.ValueSource{
				Ref:   &coinbaseID,
				Value: &value,
			}

		case *CrossChainInput:
			crossIn := bc.NewCrossChainInput(&inp.MainchainSourceID, &inp.AssetAmount, uint64(i))
			crossIn.WitnessArguments = inp.Arguments
			crossInID := addEntry(crossIn)
			muxSources[i] = &bc.ValueSource{
				Ref:   &crossInID,
				Value: &inp.AssetAmount,
			}
			crossIns = append(crossIns, crossIn)
		}
	}

	mux := bc.NewMux(muxSources, &bc.Program{VmVersion: 1, Code: []byte{byte(vm.OP_TRUE)}})
	muxID := addEntry(mux)

	// connect the inputs to the mux
	for _, spend := range spends {
		spentOutput := entryMap[*spend.SpentOutputId].(*bc.IntraChainOutput)
		spend.SetDestination(&muxID, spentOutput.Source.Value, spend.Ordinal)
	}
	for _, issuance := range issuances {
		issuance.SetDestination(&muxID, issuance.Value, issuance.Ordinal)
	}
	for _, crossIn := range crossIns {
		crossIn.SetDestination(&muxID, crossIn.Value, crossIn.Ordinal)
	}

	if coinbase != nil {
		coinbase.SetDestination(&muxID, mux.Sources[0].Value, 0)
	}

	// convert types.outputs to the bc.output
	var resultIDs []*bc.Hash
	for i, out := range tx.Outputs {
		value := out.AssetAmount()
		src := &bc.ValueSource{
			Ref:      &muxID,
			Value:    &value,
			Position: uint64(i),
		}
		var resultID bc.Hash
		switch {
		// must deal with retirement first due to cases' priorities in the switch statement
		case vmutil.IsUnspendable(out.ControlProgram()):
			// retirement
			r := bc.NewRetirement(src, uint64(i))
			resultID = addEntry(r)

		case out.OutputType() == IntraChainOutputType:
			// non-retirement intra-chain tx
			prog := &bc.Program{out.VMVersion(), out.ControlProgram()}
			o := bc.NewIntraChainOutput(src, prog, uint64(i))
			resultID = addEntry(o)

		case out.OutputType() == CrossChainOutputType:
			// non-retirement cross-chain tx
			prog := &bc.Program{out.VMVersion(), out.ControlProgram()}
			o := bc.NewCrossChainOutput(src, prog, uint64(i))
			resultID = addEntry(o)

		default:
			log.Warn("unknown outType")
		}

		dest := &bc.ValueDestination{
			Value:    src.Value,
			Ref:      &resultID,
			Position: 0,
		}
		resultIDs = append(resultIDs, &resultID)
		mux.WitnessDestinations = append(mux.WitnessDestinations, dest)
	}

	h := bc.NewTxHeader(tx.Version, tx.SerializedSize, tx.TimeRange, resultIDs)
	return addEntry(h), h, entryMap
}

func mapBlockHeader(old *BlockHeader) (bc.Hash, *bc.BlockHeader) {
	bh := bc.NewBlockHeader(old.Version, old.Height, &old.PreviousBlockHash, old.Timestamp, &old.TransactionsMerkleRoot, &old.TransactionStatusHash, old.Witness)
	return bc.EntryID(bh), bh
}

// MapBlock converts a types block to bc block
func MapBlock(old *Block) *bc.Block {
	if old == nil {
		return nil
	}

	b := new(bc.Block)
	b.ID, b.BlockHeader = mapBlockHeader(&old.BlockHeader)
	for _, oldTx := range old.Transactions {
		b.Transactions = append(b.Transactions, oldTx.Tx)
	}
	return b
}
