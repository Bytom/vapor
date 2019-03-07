package types

import (
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/vapor/encoding/blockchain"
	"github.com/vapor/encoding/bufpool"
	"github.com/vapor/errors"
	"github.com/vapor/protocol/bc"
)

type Proof struct {
	Sign           []byte
	ControlProgram []byte
	Address        []byte
}

func (p *Proof) readFrom(r *blockchain.Reader) (err error) {
	if p.Sign, err = blockchain.ReadVarstr31(r); err != nil {
		return err
	}
	if p.ControlProgram, err = blockchain.ReadVarstr31(r); err != nil {
		return err
	}
	if p.Address, err = blockchain.ReadVarstr31(r); err != nil {
		return err
	}
	return nil
}

func (p *Proof) writeTo(w io.Writer) error {
	if _, err := blockchain.WriteVarstr31(w, p.Sign); err != nil {
		return err
	}

	if _, err := blockchain.WriteVarstr31(w, p.ControlProgram); err != nil {
		return err
	}
	if _, err := blockchain.WriteVarstr31(w, p.Address); err != nil {
		return err
	}
	return nil
}

// BlockHeader defines information about a block and is used in the Bytom
type BlockHeader struct {
	Version           uint64  // The version of the block.
	Height            uint64  // The height of the block.
	PreviousBlockHash bc.Hash // The hash of the previous block.
	Timestamp         uint64  // The time of the block in seconds.
	Coinbase          []byte
	Proof             Proof
	Extra             []byte
	CmtMsges          []*CommitMsg
	BlockCommitment
}

// Time returns the time represented by the Timestamp in block header.
func (bh *BlockHeader) Time() time.Time {
	return time.Unix(int64(bh.Timestamp), 0).UTC()
}

// Hash returns complete hash of the block header.
func (bh *BlockHeader) Hash() bc.Hash {
	h, _ := mapBlockHeader(bh)
	return h
}

// MarshalText fulfills the json.Marshaler interface. This guarantees that
// block headers will get deserialized correctly when being parsed from HTTP
// requests.
func (bh *BlockHeader) MarshalText() ([]byte, error) {
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	if _, err := bh.WriteTo(buf); err != nil {
		return nil, err
	}

	enc := make([]byte, hex.EncodedLen(buf.Len()))
	hex.Encode(enc, buf.Bytes())
	return enc, nil
}

// UnmarshalText fulfills the encoding.TextUnmarshaler interface.
func (bh *BlockHeader) UnmarshalText(text []byte) error {
	decoded := make([]byte, hex.DecodedLen(len(text)))
	if _, err := hex.Decode(decoded, text); err != nil {
		return err
	}

	_, err := bh.readFrom(blockchain.NewReader(decoded))
	return err
}

func (bh *BlockHeader) readFrom(r *blockchain.Reader) (serflag uint8, err error) {
	var serflags [1]byte
	io.ReadFull(r, serflags[:])
	serflag = serflags[0]
	switch serflag {
	case SerBlockHeader, SerBlockFull:
	default:
		return 0, fmt.Errorf("unsupported serialization flags 0x%x", serflags)
	}

	if bh.Version, err = blockchain.ReadVarint63(r); err != nil {
		return 0, err
	}
	if bh.Height, err = blockchain.ReadVarint63(r); err != nil {
		return 0, err
	}
	if _, err = bh.PreviousBlockHash.ReadFrom(r); err != nil {
		return 0, err
	}
	if bh.Timestamp, err = blockchain.ReadVarint63(r); err != nil {
		return 0, err
	}
	if bh.Coinbase, err = blockchain.ReadVarstr31(r); err != nil {
		return 0, err
	}
	if _, err = blockchain.ReadExtensibleString(r, bh.BlockCommitment.readFrom); err != nil {
		return 0, err
	}
	if _, err = blockchain.ReadExtensibleString(r, bh.Proof.readFrom); err != nil {
		return 0, err
	}
	if bh.Extra, err = blockchain.ReadVarstr31(r); err != nil {
		return 0, err
	}
	nelts := uint32(0)
	if nelts, err = blockchain.ReadVarint31(r); err != nil {
		return 0, err
	}
	if nelts == 0 {
		return 0, nil
	}

	cmtMsges := []*CommitMsg{}
	for ; nelts > 0 && err == nil; nelts-- {
		s := &CommitMsg{}
		_, err = blockchain.ReadExtensibleString(r, s.readFrom)
		cmtMsges = append(cmtMsges, s)
	}
	if len(cmtMsges) < int(nelts) {
		err = io.ErrUnexpectedEOF
	}
	if err != nil {
		return 0, err
	}
	bh.CmtMsges = cmtMsges

	return
}

// WriteTo writes the block header to the input io.Writer
func (bh *BlockHeader) WriteTo(w io.Writer) (int64, error) {
	ew := errors.NewWriter(w)
	if err := bh.writeTo(ew, SerBlockHeader); err != nil {
		return 0, err
	}
	return ew.Written(), ew.Err()
}

func (bh *BlockHeader) writeTo(w io.Writer, serflags uint8) (err error) {
	w.Write([]byte{serflags})
	if _, err = blockchain.WriteVarint63(w, bh.Version); err != nil {
		return err
	}
	if _, err = blockchain.WriteVarint63(w, bh.Height); err != nil {
		return err
	}
	if _, err = bh.PreviousBlockHash.WriteTo(w); err != nil {
		return err
	}
	if _, err = blockchain.WriteVarint63(w, bh.Timestamp); err != nil {
		return err
	}
	if _, err := blockchain.WriteVarstr31(w, bh.Coinbase); err != nil {
		return err
	}
	if _, err = blockchain.WriteExtensibleString(w, nil, bh.BlockCommitment.writeTo); err != nil {
		return err
	}
	if _, err = blockchain.WriteExtensibleString(w, nil, bh.Proof.writeTo); err != nil {
		return err
	}
	if _, err = blockchain.WriteVarstr31(w, bh.Extra); err != nil {
		return err
	}

	if _, err = blockchain.WriteVarint31(w, uint64(len(bh.CmtMsges))); err != nil {
		return err
	}
	for _, s := range bh.CmtMsges {
		if _, err = blockchain.WriteExtensibleString(w, nil, s.writeTo); err != nil {
			return err
		}
	}
	return nil
}
