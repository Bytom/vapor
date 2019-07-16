package synchron

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/toolbar/common"
	"github.com/vapor/toolbar/common/service"
	"github.com/vapor/toolbar/reward/config"
	"github.com/vapor/toolbar/reward/database/orm"
)

type chainKeeper struct {
	cfg   *config.Chain
	db    *gorm.DB
	node  *service.Node
	XPubs map[string]bool
}

func NewChainKeeper(db *gorm.DB, cfg *config.Config) *chainKeeper {
	xpubs := map[string]bool{}
	for _, xpub := range cfg.XPubs {
		xpubs[xpub.String()] = true
	}

	return &chainKeeper{
		cfg:   &cfg.Chain,
		db:    db,
		node:  service.NewNode(cfg.Chain.Upstream),
		XPubs: xpubs,
	}
}

func (c *chainKeeper) Run() {
	ticker := time.NewTicker(time.Duration(c.cfg.SyncSeconds) * time.Second)
	for ; true; <-ticker.C {
		for {
			isUpdate, err := c.syncBlock()
			if err != nil {
				log.WithField("error", err).Errorln("blockKeeper fail on process block")
				break
			}

			if !isUpdate {
				break
			}
		}
	}
}

func (c *chainKeeper) syncBlock() (bool, error) {
	blockState := &orm.BlockState{}
	if err := c.db.First(blockState).Error; err != nil {
		return false, errors.Wrap(err, "query chain")
	}

	height, err := c.node.GetBlockCount()
	if err != nil {
		return false, err
	}

	if height == blockState.Height {
		return false, nil
	}

	nextBlockStr, txStatus, err := c.node.GetBlockByHeight(blockState.Height + 1)
	if err != nil {
		return false, err
	}

	nextBlock := &types.Block{}
	if err := nextBlock.UnmarshalText([]byte(nextBlockStr)); err != nil {
		return false, errors.New("Unmarshal nextBlock")
	}

	// Normal case, the previous hash of next block equals to the hash of current block,
	// just sync to database directly.
	if nextBlock.PreviousBlockHash.String() == blockState.BlockHash {
		return true, c.AttachBlock(nextBlock, txStatus)
	}

	log.WithField("block height", blockState.Height).Debug("the prev hash of remote is not equals the hash of current best block, must rollback")
	currentBlockStr, txStatus, err := c.node.GetBlockByHash(blockState.BlockHash)
	if err != nil {
		return false, err
	}

	currentBlock := &types.Block{}
	if err := nextBlock.UnmarshalText([]byte(currentBlockStr)); err != nil {
		return false, errors.New("Unmarshal currentBlock")
	}

	return true, c.DetachBlock(currentBlock, txStatus)
}

func (c *chainKeeper) AttachBlock(block *types.Block, txStatus *bc.TransactionStatus) error {
	ormDB := c.db.Begin()
	for pos, tx := range block.Transactions {
		statusFail, _ := txStatus.GetStatus(pos)
		if statusFail {
			log.WithFields(log.Fields{"block height": block.Height, "statusFail": statusFail}).Debug("AttachBlock")
			continue
		}

		for _, input := range tx.Inputs {
			vetoInput, ok := input.TypedInput.(*types.VetoInput)
			if !ok {
				continue
			}

			pubkey := hex.EncodeToString(vetoInput.Vote)
			if ok := c.XPubs[pubkey]; ok {
				prog := &bc.Program{VmVersion: vetoInput.VMVersion, Code: vetoInput.ControlProgram}
				src := &bc.ValueSource{
					Ref:      &vetoInput.SourceID,
					Value:    &vetoInput.AssetAmount,
					Position: vetoInput.SourcePosition,
				}
				prevout := bc.NewVoteOutput(src, prog, 0, vetoInput.Vote) // ordinal doesn't matter for prevouts, only for result outputs
				outputID := bc.EntryID(prevout)
				utxo := &orm.Utxo{
					VoterAddress: common.GetAddressFromControlProgram(vetoInput.ControlProgram),
					OutputID:     outputID.String(),
				}
				// update data
				if err := ormDB.Where(utxo).Update("veto_height", block.Height).Error; err != nil && err != gorm.ErrRecordNotFound {
					ormDB.Rollback()
					return err
				}
			}
		}

		for index, output := range tx.Outputs {
			voteOutput, ok := output.TypedOutput.(*types.VoteOutput)
			if !ok {
				continue
			}
			pubkey := hex.EncodeToString(voteOutput.Vote)
			fmt.Println(pubkey)
			if ok := c.XPubs[pubkey]; ok {
				outputID := tx.OutputID(index)
				utxo := &orm.Utxo{
					Xpub:         pubkey,
					VoterAddress: common.GetAddressFromControlProgram(voteOutput.ControlProgram),
					VoteHeight:   block.Height,
					VoteNum:      voteOutput.Amount,
					VetoHeight:   0,
					OutputID:     outputID.String(),
				}
				// insert data
				if err := ormDB.Save(utxo).Error; err != nil {
					return err
				}
			}
		}
	}

	return ormDB.Commit().Error
}

func (c *chainKeeper) DetachBlock(block *types.Block, txStatus *bc.TransactionStatus) error {
	ormDB := c.db.Begin()
	for txIndex := len(block.Transactions) - 1; txIndex >= 0; txIndex-- {
		statusFail, _ := txStatus.GetStatus(txIndex)
		if statusFail {
			log.WithFields(log.Fields{"block height": block.Height, "statusFail": statusFail}).Debug("DetachBlock")
			continue
		}
		tx := block.Transactions[txIndex]

		for _, input := range tx.Inputs {
			vetoInput, ok := input.TypedInput.(*types.VetoInput)
			if !ok {
				continue
			}

			pubkey := hex.EncodeToString(vetoInput.Vote)
			if ok := c.XPubs[pubkey]; ok {
				prog := &bc.Program{VmVersion: vetoInput.VMVersion, Code: vetoInput.ControlProgram}
				src := &bc.ValueSource{
					Ref:      &vetoInput.SourceID,
					Value:    &vetoInput.AssetAmount,
					Position: vetoInput.SourcePosition,
				}
				prevout := bc.NewVoteOutput(src, prog, 0, vetoInput.Vote) // ordinal doesn't matter for prevouts, only for result outputs
				outputID := bc.EntryID(prevout)
				utxo := &orm.Utxo{
					VoterAddress: common.GetAddressFromControlProgram(vetoInput.ControlProgram),
					OutputID:     outputID.String(),
				}
				// update data
				if err := ormDB.Where(utxo).Update("veto_height", 0).Error; err != nil && err != gorm.ErrRecordNotFound {
					ormDB.Rollback()
					return err
				}
			}
		}

		for index, output := range tx.Outputs {
			voteOutput, ok := output.TypedOutput.(*types.VoteOutput)
			if !ok {
				continue
			}
			pubkey := hex.EncodeToString(voteOutput.Vote)
			fmt.Println(pubkey)
			if ok := c.XPubs[pubkey]; ok {
				outputID := tx.OutputID(index)
				utxo := &orm.Utxo{
					Xpub:         pubkey,
					VoterAddress: common.GetAddressFromControlProgram(voteOutput.ControlProgram),
					VoteHeight:   block.Height,
					VoteNum:      voteOutput.Amount,
					OutputID:     outputID.String(),
				}
				// insert data
				if err := ormDB.Where(utxo).Delete(&orm.Utxo{}).Error; err != nil {
					return err
				}
			}
		}

	}

	return ormDB.Commit().Error
}
