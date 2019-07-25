package settlementvotereward

import (
	"bytes"
	"math/big"

	"github.com/jinzhu/gorm"

	"github.com/vapor/consensus"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc/types"
	"github.com/vapor/toolbar/apinode"
	"github.com/vapor/toolbar/common"
	"github.com/vapor/toolbar/vote_reward/config"
)

var (
	errFoundReward   = errors.New("No reward found")
	errNoStandbyNode = errors.New("No Standby Node")
	errNoRewardTx    = errors.New("No reward transaction")
)

const standbyNodesRewardForConsensusCycle = 7610350076 // 400000000000000 / (365 * 24 * 60 / (500 * 1200 / 1000 / 60))

type voteResult struct {
	VoteAddress string
	VoteNum     uint64
}

type SettlementReward struct {
	rewardCfg   *config.RewardConfig
	node        *apinode.Node
	db          *gorm.DB
	rewards     map[string]uint64
	startHeight uint64
	endHeight   uint64
}

func NewSettlementReward(db *gorm.DB, cfg *config.Config, startHeight, endHeight uint64) *SettlementReward {
	return &SettlementReward{
		db:          db,
		rewardCfg:   cfg.RewardConf,
		node:        apinode.NewNode(cfg.NodeIP),
		rewards:     make(map[string]uint64),
		startHeight: startHeight,
		endHeight:   endHeight,
	}
}

func (s *SettlementReward) getVoteResultFromDB(height uint64) (voteResults []*voteResult, err error) {
	query := s.db.Table("utxos").Select("vote_address, sum(vote_num) as vote_num")
	query = query.Where("(veto_height >= ? or veto_height = 0) and vote_height <= ? and xpub = ?", height-consensus.ActiveNetParams.RoundVoteBlockNums+1, height-consensus.ActiveNetParams.RoundVoteBlockNums, s.rewardCfg.XPub)
	query = query.Group("vote_address")
	if err := query.Scan(&voteResults).Error; err != nil {
		return nil, err
	}

	return voteResults, nil
}

func (s *SettlementReward) Settlement() error {
	for height := s.startHeight + consensus.ActiveNetParams.RoundVoteBlockNums; height <= s.endHeight; height += consensus.ActiveNetParams.RoundVoteBlockNums {
		totalReward, err := s.getCoinbaseReward(height + 1)
		if err == errFoundReward {
			totalReward, err = s.getStandbyNodeReward(height - consensus.ActiveNetParams.RoundVoteBlockNums)
		}

		if err == errNoStandbyNode {
			continue
		}

		if err != nil {
			return errors.Wrapf(err, "get total reward at height: %d", height)
		}

		voteResults, err := s.getVoteResultFromDB(height)
		if err != nil {
			return err
		}

		s.calcVoterRewards(voteResults, totalReward)
	}

	if len(s.rewards) == 0 {
		return errNoRewardTx
	}

	// send transactions
	return s.node.BatchSendBTM(s.rewardCfg.AccountID, s.rewardCfg.Password, s.rewards)
}

func (s *SettlementReward) getStandbyNodeReward(height uint64) (uint64, error) {
	voteInfos, err := s.node.GetVoteByHeight(height)
	if err != nil {
		return 0, errors.Wrapf(err, "get alternative node reward")
	}

	voteInfos = common.CalcStandByNodes(voteInfos)

	err = errNoStandbyNode
	totalVoteNum := uint64(0)
	xpubVoteNum := uint64(0)
	for _, voteInfo := range voteInfos {
		totalVoteNum += voteInfo.VoteNum
		if s.rewardCfg.XPub == voteInfo.Vote {
			xpubVoteNum = voteInfo.VoteNum
			err = nil
		}
	}

	if err != nil {
		return 0, err
	}

	amount := big.NewInt(0).SetUint64(standbyNodesRewardForConsensusCycle)
	rewardRatio := big.NewInt(0).SetUint64(s.rewardCfg.RewardRatio)
	amount.Mul(amount, rewardRatio).Div(amount, big.NewInt(100))

	total := big.NewInt(0).SetUint64(totalVoteNum)
	voteNum := big.NewInt(0).SetUint64(xpubVoteNum)

	return amount.Mul(amount, voteNum).Div(amount, total).Uint64(), nil
}

func (s *SettlementReward) getCoinbaseReward(height uint64) (uint64, error) {
	block, err := s.node.GetBlockByHeight(height)
	if err != nil {
		return 0, err
	}

	miningControl, err := common.GetControlProgramFromAddress(s.rewardCfg.MiningAddress)
	if err != nil {
		return 0, err
	}

	for _, output := range block.Transactions[0].Outputs {
		output, ok := output.TypedOutput.(*types.IntraChainOutput)
		if !ok {
			return 0, errors.New("Output type error")
		}

		if output.Amount == 0 {
			continue
		}

		if bytes.Equal(miningControl, output.ControlProgram) {
			amount := big.NewInt(0).SetUint64(output.Amount)
			rewardRatio := big.NewInt(0).SetUint64(s.rewardCfg.RewardRatio)
			amount.Mul(amount, rewardRatio).Div(amount, big.NewInt(100))

			return amount.Uint64(), nil
		}
	}
	return 0, errFoundReward
}

func (s *SettlementReward) calcVoterRewards(voteResults []*voteResult, totalReward uint64) {
	totalVoteNum := uint64(0)
	for _, voteResult := range voteResults {
		totalVoteNum += voteResult.VoteNum
	}

	for _, voteResult := range voteResults {
		// voteNum / totalVoteNum  * totalReward
		voteNum := big.NewInt(0).SetUint64(voteResult.VoteNum)
		total := big.NewInt(0).SetUint64(totalVoteNum)
		reward := big.NewInt(0).SetUint64(totalReward)

		amount := voteNum.Mul(voteNum, reward).Div(voteNum, total).Uint64()

		if amount != 0 {
			s.rewards[voteResult.VoteAddress] += amount
		}
	}
}
