package consensusmgr

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/tendermint/go-wire"

	"github.com/vapor/netsync/peers"
	"github.com/vapor/protocol/bc"
	"github.com/vapor/protocol/bc/types"
)

const (
	blockSignatureByte = byte(0x10)
	blockProposeByte   = byte(0x11)
)

//ConsensusMessage is a generic message for consensus reactor.
type ConsensusMessage interface {
	String() string
	BroadcastMarkSendRecord(ps *peers.PeerSet, peers []string)
	BroadcastFilterTargetPeers(ps *peers.PeerSet) []string
}

var _ = wire.RegisterInterface(
	struct{ ConsensusMessage }{},
	wire.ConcreteType{O: &BlockSignatureMsg{}, Byte: blockSignatureByte},
	wire.ConcreteType{O: &BlockProposeMsg{}, Byte: blockProposeByte},
)

//decodeMessage decode msg
func decodeMessage(bz []byte) (msgType byte, msg ConsensusMessage, err error) {
	msgType = bz[0]
	n := int(0)
	r := bytes.NewReader(bz)
	msg = wire.ReadBinary(struct{ ConsensusMessage }{}, r, maxBlockchainResponseSize, &n, &err).(struct{ ConsensusMessage }).ConsensusMessage
	if err != nil && n != len(bz) {
		err = errors.New("DecodeMessage() had bytes left over")
	}
	return
}

// BlockSignatureMsg block signature message transferred between nodes.
type BlockSignatureMsg struct {
	BlockHash [32]byte
	Height    uint64
	Signature []byte
	PubKey    [64]byte
}

//NewBlockSignatureMsg create new block signature msg.
func NewBlockSignatureMsg(blockHash bc.Hash, height uint64, signature []byte, pubKey [64]byte) ConsensusMessage {
	hash := blockHash.Byte32()
	return &BlockSignatureMsg{BlockHash: hash, Height: height, Signature: signature, PubKey: pubKey}
}

func (bs *BlockSignatureMsg) String() string {
	return fmt.Sprintf("{block_hash: %s,block_height:%d,signature:%s,pubkey:%s}", hex.EncodeToString(bs.BlockHash[:]), bs.Height, hex.EncodeToString(bs.Signature), hex.EncodeToString(bs.PubKey[:]))
}

// BroadcastMarkSendRecord mark send message record to prevent messages from being sent repeatedly.
func (bs *BlockSignatureMsg) BroadcastMarkSendRecord(ps *peers.PeerSet, peers []string) {
	for _, peer := range peers {
		ps.MarkBlockSignature(peer, bs.Signature)
	}
}

// BroadcastFilterTargetPeers filter target peers to filter the nodes that need to send messages.
func (bs *BlockSignatureMsg) BroadcastFilterTargetPeers(ps *peers.PeerSet) []string {
	return ps.PeersWithoutSign(bs.Signature)
}

// BlockProposeMsg block propose message transferred between nodes.
type BlockProposeMsg struct {
	RawBlock []byte
}

//NewBlockProposeMsg create new block propose msg.
func NewBlockProposeMsg(block *types.Block) (ConsensusMessage, error) {
	rawBlock, err := block.MarshalText()
	if err != nil {
		return nil, err
	}
	return &BlockProposeMsg{RawBlock: rawBlock}, nil
}

//GetProposeBlock get propose block from msg.
func (bp *BlockProposeMsg) GetProposeBlock() (*types.Block, error) {
	block := &types.Block{}
	if err := block.UnmarshalText(bp.RawBlock); err != nil {
		return nil, err
	}
	return block, nil
}

func (bp *BlockProposeMsg) String() string {
	block, err := bp.GetProposeBlock()
	if err != nil {
		return "{err: wrong message}"
	}
	blockHash := block.Hash()
	return fmt.Sprintf("{block_height: %d, block_hash: %s}", block.Height, blockHash.String())
}

// BroadcastMarkSendRecord mark send message record to prevent messages from being sent repeatedly.
func (bp *BlockProposeMsg) BroadcastMarkSendRecord(ps *peers.PeerSet, peers []string) {
	block, err := bp.GetProposeBlock()
	if err != nil {
		return
	}

	hash := block.Hash()
	height := block.Height
	for _, peer := range peers {
		ps.MarkBlock(peer, &hash)
		ps.MarkStatus(peer, height)
	}
}

// BroadcastFilterTargetPeers filter target peers to filter the nodes that need to send messages.
func (bp *BlockProposeMsg) BroadcastFilterTargetPeers(ps *peers.PeerSet) []string {
	block, err := bp.GetProposeBlock()
	if err != nil {
		return nil
	}

	return ps.PeersWithoutBlock(block.Hash())
}
