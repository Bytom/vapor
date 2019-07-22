package reward

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"

	"github.com/vapor/consensus"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/toolbar/common"
	"github.com/vapor/toolbar/common/service"
	"github.com/vapor/toolbar/reward/config"
	"github.com/vapor/toolbar/reward/database/orm"
)

type voterReward struct {
	rewards             map[string]*big.Int
	totalCoinbaseReward uint64
	height              uint64
}

type voteResult struct {
	votes     map[string]*big.Int
	voteTotal *big.Int
}

type coinBaseReward struct {
	totalReward     uint64
	voteTotalReward *big.Int
}

type Vote struct {
	nodeConfig         *config.VoteRewardConfig
	node               *service.Node
	db                 *gorm.DB
	reward             *voterReward
	coinBaseRewards    map[uint64]*coinBaseReward
	roundVoteBlockNums uint64
	rewardStartHeight  uint64
	rewardEndHeight    uint64
}

func NewVote(db *gorm.DB, nodeConfig *config.VoteRewardConfig, rewardStartHeight, rewardEndHeight uint64) *Vote {
	return &Vote{
		db:                 db,
		nodeConfig:         nodeConfig,
		node:               service.NewNode(nodeConfig.Upstream),
		reward:             &voterReward{rewards: make(map[string]*big.Int)},
		coinBaseRewards:    make(map[uint64]*coinBaseReward),
		roundVoteBlockNums: consensus.ActiveNetParams.DPOSConfig.RoundVoteBlockNums,
		rewardStartHeight:  rewardStartHeight,
		rewardEndHeight:    rewardEndHeight,
	}
}

func (v *Vote) getVoteByXpub(height uint64) ([]*orm.Utxo, error) {
	utxos := []*orm.Utxo{}
	if err := v.db.Where("(veto_height >= ? or veto_height = 0) and vote_height <= ? and xpub = ?", height-v.roundVoteBlockNums+1, height-v.roundVoteBlockNums, v.nodeConfig.XPub).Find(&utxos).Error; err != nil {
		return nil, err
	}
	return utxos, nil
}

func (v *Vote) Start() error {
	// get coinbase reward
	if err := v.calcCoinbaseReward(); err != nil {
		return err
	}
	for height := v.rewardStartHeight + v.roundVoteBlockNums; height <= v.rewardEndHeight; height += v.roundVoteBlockNums {
		voteInfo, err := v.getVoteByXpub(height)
		if err != nil {
			return errors.Wrapf(err, "get vote from db at coinbase_height: %d", height)
		}

		voteResults := v.countVotes(voteInfo, height)
		if err := v.countReward(voteResults, height+1); err != nil {
			return errors.Wrapf(err, "count reaward at coinbase_height: %d", height+1)
		}
	}

	// send transactions
	return v.sendRewardTransaction()
}

func (v *Vote) calcCoinbaseReward() error {
	for height := v.rewardStartHeight + v.roundVoteBlockNums; height <= v.rewardEndHeight; height += v.roundVoteBlockNums {
		coinbaseHeight := height + 1
		coinbaseTx, err := v.node.GetCoinbaseTx(coinbaseHeight)
		if err != nil {
			return errors.Wrapf(err, "get coinbase reward at coinbase_height: %d", coinbaseHeight)
		}

		for _, output := range coinbaseTx.Outputs {
			output, ok := output.TypedOutput.(*types.IntraChainOutput)
			if !ok {
				return errors.New("Output type error")
			}

			if output.Amount == 0 {
				continue
			}

			address := common.GetAddressFromControlProgram(output.ControlProgram)

			if address == v.nodeConfig.MiningAddress {
				reward := &coinBaseReward{
					totalReward: output.Amount,
				}
				ratioNumerator := big.NewInt(int64(v.nodeConfig.RewardRatio))
				ratioDenominator := big.NewInt(100)
				coinBaseReward := big.NewInt(0).SetUint64(output.Amount)
				reward.voteTotalReward = coinBaseReward.Mul(coinBaseReward, ratioNumerator).Div(coinBaseReward, ratioDenominator)
				v.coinBaseRewards[coinbaseHeight] = reward
			}
		}
	}

	return nil
}

func (v *Vote) countVotes(utxos []*orm.Utxo, coinBaseHeight uint64) (voteResults *voteResult) {
	voteResults = &voteResult{
		votes:     make(map[string]*big.Int),
		voteTotal: big.NewInt(0),
	}
	for _, utxo := range utxos {
		voteBlockNum := uint64(0)
		if utxo.VetoHeight == 0 || (utxo.VetoHeight > coinBaseHeight && utxo.VoteHeight < coinBaseHeight) {
			voteBlockNum = coinBaseHeight - utxo.VoteHeight
		} else if utxo.VetoHeight > (coinBaseHeight-v.roundVoteBlockNums+1) && utxo.VetoHeight <= coinBaseHeight {
			voteBlockNum = utxo.VetoHeight - utxo.VoteHeight
		} else {
			continue
		}

		bigBlockNum := big.NewInt(0).SetUint64(voteBlockNum)
		bigVoteNum := big.NewInt(0).SetUint64(utxo.VoteNum)
		bigVoteNum.Mul(bigVoteNum, bigBlockNum)

		if vote, ok := voteResults.votes[utxo.VoterAddress]; ok {
			vote.Add(vote, bigVoteNum)
		} else {
			voteResults.votes[utxo.VoterAddress] = bigVoteNum
		}

		voteTotal := voteResults.voteTotal
		voteTotal.Add(voteTotal, bigVoteNum)
		voteResults.voteTotal = voteTotal
	}
	return
}

func (v *Vote) countReward(votes *voteResult, height uint64) error {
	coinBaseReward, ok := v.coinBaseRewards[height]
	if !ok {
		return errors.New("Doesn't have a coinbase reward")
	}

	for address, vote := range votes.votes {
		if reward, ok := v.reward.rewards[address]; ok {
			mul := vote.Mul(vote, coinBaseReward.voteTotalReward)
			amount := big.NewInt(0)
			amount.Div(mul, votes.voteTotal)
			reward.Add(reward, amount)
		} else {
			mul := vote.Mul(vote, coinBaseReward.voteTotalReward)
			amount := big.NewInt(0)
			amount.Div(mul, votes.voteTotal)
			if amount.Uint64() > 0 {
				v.reward.rewards[address] = amount
				v.reward.totalCoinbaseReward = coinBaseReward.totalReward
				v.reward.height = height
			}
		}
	}
	return nil
}

func (v *Vote) sendRewardTransaction() error {
	var outputActions []string
	inputAction := fmt.Sprintf(service.InputActionFmt, v.reward.totalCoinbaseReward, v.nodeConfig.AccountID)
	for address, amount := range v.reward.rewards {
		outputActions = append(outputActions, fmt.Sprintf(service.OutputActionFmt, amount.Uint64(), address))
	}

	outputAction := strings.Join(outputActions, ",")
	txID, err := v.node.SendTransaction(inputAction, outputAction, v.nodeConfig.Passwd)
	if err != nil {
		return err
	}

	log.Info("tx_id: ", txID)
	return nil

}
