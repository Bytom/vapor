package netsync

import (
	"bytes"

	log "github.com/sirupsen/logrus"
	"github.com/tendermint/go-wire"

	"github.com/vapor/errors"
	msgs "github.com/vapor/netsync/messages"
	"github.com/vapor/p2p"
	"github.com/vapor/p2p/connection"
)

var (
	errProtocolHandshakeTimeout = errors.New("Protocol handshake timeout")
	errStatusRequest            = errors.New("Status request error")
)

//ProtocolReactor handles new coming protocol message.
type ProtocolReactor struct {
	p2p.BaseReactor

	cm *ChainManager
}

// NewProtocolReactor returns the reactor of whole blockchain.
func NewProtocolReactor(cm *ChainManager) *ProtocolReactor {
	pr := &ProtocolReactor{
		cm: cm,
	}
	pr.BaseReactor = *p2p.NewBaseReactor("ProtocolReactor", pr)
	return pr
}

// GetChannels implements Reactor
func (pr *ProtocolReactor) GetChannels() []*connection.ChannelDescriptor {
	return []*connection.ChannelDescriptor{
		{
			ID:                msgs.BlockchainChannel,
			Priority:          5,
			SendQueueCapacity: 100,
		},
	}
}

// OnStart implements BaseService
func (pr *ProtocolReactor) OnStart() error {
	pr.BaseReactor.OnStart()
	return nil
}

// OnStop implements BaseService
func (pr *ProtocolReactor) OnStop() {
	pr.BaseReactor.OnStop()
}

// AddPeer implements Reactor by sending our state to peer.
func (pr *ProtocolReactor) AddPeer(peer *p2p.Peer) error {
	pr.cm.AddPeer(peer)
	if err := pr.cm.SendStatus(peer); err != nil {
		return err
	}
	pr.cm.syncTransactions(peer.Key)
	return nil
}

// RemovePeer implements Reactor by removing peer from the pool.
func (pr *ProtocolReactor) RemovePeer(peer *p2p.Peer, reason interface{}) {
	pr.cm.RemovePeer(peer.Key)
}

//decodeMessage decode msg
func decodeMessage(bz []byte) (msgType byte, msg msgs.BlockchainMessage, err error) {
	msgType = bz[0]
	n := int(0)
	r := bytes.NewReader(bz)
	msg = wire.ReadBinary(struct{ msgs.BlockchainMessage }{}, r, msgs.MaxBlockchainResponseSize, &n, &err).(struct{ msgs.BlockchainMessage }).BlockchainMessage
	if err != nil && n != len(bz) {
		err = errors.New("DecodeMessage() had bytes left over")
	}
	return
}

// Receive implements Reactor by handling 4 types of messages (look below).
func (pr *ProtocolReactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	msgType, msg, err := decodeMessage(msgBytes)
	if err != nil {
		log.WithFields(log.Fields{"module": logModule, "err": err}).Error("fail on reactor decoding message")
		return
	}

	pr.cm.processMsg(src, msgType, msg)
}
