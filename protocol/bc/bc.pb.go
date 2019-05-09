// Code generated by protoc-gen-go. DO NOT EDIT.
// source: bc.proto

/*
Package bc is a generated protocol buffer package.

It is generated from these files:
	bc.proto

It has these top-level messages:
	Hash
	Program
	AssetID
	AssetAmount
	AssetDefinition
	ValueSource
	ValueDestination
	BlockHeader
	TxHeader
	TxVerifyResult
	TransactionStatus
	Mux
	Coinbase
	Output
	Retirement
	Issuance
	Spend
*/
package bc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Hash struct {
	V0 uint64 `protobuf:"fixed64,1,opt,name=v0" json:"v0,omitempty"`
	V1 uint64 `protobuf:"fixed64,2,opt,name=v1" json:"v1,omitempty"`
	V2 uint64 `protobuf:"fixed64,3,opt,name=v2" json:"v2,omitempty"`
	V3 uint64 `protobuf:"fixed64,4,opt,name=v3" json:"v3,omitempty"`
}

func (m *Hash) Reset()                    { *m = Hash{} }
func (m *Hash) String() string            { return proto.CompactTextString(m) }
func (*Hash) ProtoMessage()               {}
func (*Hash) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Hash) GetV0() uint64 {
	if m != nil {
		return m.V0
	}
	return 0
}

func (m *Hash) GetV1() uint64 {
	if m != nil {
		return m.V1
	}
	return 0
}

func (m *Hash) GetV2() uint64 {
	if m != nil {
		return m.V2
	}
	return 0
}

func (m *Hash) GetV3() uint64 {
	if m != nil {
		return m.V3
	}
	return 0
}

type Program struct {
	VmVersion uint64 `protobuf:"varint,1,opt,name=vm_version,json=vmVersion" json:"vm_version,omitempty"`
	Code      []byte `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
}

func (m *Program) Reset()                    { *m = Program{} }
func (m *Program) String() string            { return proto.CompactTextString(m) }
func (*Program) ProtoMessage()               {}
func (*Program) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Program) GetVmVersion() uint64 {
	if m != nil {
		return m.VmVersion
	}
	return 0
}

func (m *Program) GetCode() []byte {
	if m != nil {
		return m.Code
	}
	return nil
}

// This message type duplicates Hash, above. One alternative is to
// embed a Hash inside an AssetID. But it's useful for AssetID to be
// plain old data (without pointers). Another alternative is use Hash
// in any protobuf types where an AssetID is called for, but it's
// preferable to have type safety.
type AssetID struct {
	V0 uint64 `protobuf:"fixed64,1,opt,name=v0" json:"v0,omitempty"`
	V1 uint64 `protobuf:"fixed64,2,opt,name=v1" json:"v1,omitempty"`
	V2 uint64 `protobuf:"fixed64,3,opt,name=v2" json:"v2,omitempty"`
	V3 uint64 `protobuf:"fixed64,4,opt,name=v3" json:"v3,omitempty"`
}

func (m *AssetID) Reset()                    { *m = AssetID{} }
func (m *AssetID) String() string            { return proto.CompactTextString(m) }
func (*AssetID) ProtoMessage()               {}
func (*AssetID) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *AssetID) GetV0() uint64 {
	if m != nil {
		return m.V0
	}
	return 0
}

func (m *AssetID) GetV1() uint64 {
	if m != nil {
		return m.V1
	}
	return 0
}

func (m *AssetID) GetV2() uint64 {
	if m != nil {
		return m.V2
	}
	return 0
}

func (m *AssetID) GetV3() uint64 {
	if m != nil {
		return m.V3
	}
	return 0
}

type AssetAmount struct {
	AssetId *AssetID `protobuf:"bytes,1,opt,name=asset_id,json=assetId" json:"asset_id,omitempty"`
	Amount  uint64   `protobuf:"varint,2,opt,name=amount" json:"amount,omitempty"`
}

func (m *AssetAmount) Reset()                    { *m = AssetAmount{} }
func (m *AssetAmount) String() string            { return proto.CompactTextString(m) }
func (*AssetAmount) ProtoMessage()               {}
func (*AssetAmount) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *AssetAmount) GetAssetId() *AssetID {
	if m != nil {
		return m.AssetId
	}
	return nil
}

func (m *AssetAmount) GetAmount() uint64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

type AssetDefinition struct {
	IssuanceProgram *Program `protobuf:"bytes,1,opt,name=issuance_program,json=issuanceProgram" json:"issuance_program,omitempty"`
	Data            *Hash    `protobuf:"bytes,2,opt,name=data" json:"data,omitempty"`
}

func (m *AssetDefinition) Reset()                    { *m = AssetDefinition{} }
func (m *AssetDefinition) String() string            { return proto.CompactTextString(m) }
func (*AssetDefinition) ProtoMessage()               {}
func (*AssetDefinition) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *AssetDefinition) GetIssuanceProgram() *Program {
	if m != nil {
		return m.IssuanceProgram
	}
	return nil
}

func (m *AssetDefinition) GetData() *Hash {
	if m != nil {
		return m.Data
	}
	return nil
}

type ValueSource struct {
	Ref      *Hash        `protobuf:"bytes,1,opt,name=ref" json:"ref,omitempty"`
	Value    *AssetAmount `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	Position uint64       `protobuf:"varint,3,opt,name=position" json:"position,omitempty"`
}

func (m *ValueSource) Reset()                    { *m = ValueSource{} }
func (m *ValueSource) String() string            { return proto.CompactTextString(m) }
func (*ValueSource) ProtoMessage()               {}
func (*ValueSource) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ValueSource) GetRef() *Hash {
	if m != nil {
		return m.Ref
	}
	return nil
}

func (m *ValueSource) GetValue() *AssetAmount {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *ValueSource) GetPosition() uint64 {
	if m != nil {
		return m.Position
	}
	return 0
}

type ValueDestination struct {
	Ref      *Hash        `protobuf:"bytes,1,opt,name=ref" json:"ref,omitempty"`
	Value    *AssetAmount `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	Position uint64       `protobuf:"varint,3,opt,name=position" json:"position,omitempty"`
}

func (m *ValueDestination) Reset()                    { *m = ValueDestination{} }
func (m *ValueDestination) String() string            { return proto.CompactTextString(m) }
func (*ValueDestination) ProtoMessage()               {}
func (*ValueDestination) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *ValueDestination) GetRef() *Hash {
	if m != nil {
		return m.Ref
	}
	return nil
}

func (m *ValueDestination) GetValue() *AssetAmount {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *ValueDestination) GetPosition() uint64 {
	if m != nil {
		return m.Position
	}
	return 0
}

type BlockHeader struct {
	Version               uint64             `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	Height                uint64             `protobuf:"varint,2,opt,name=height" json:"height,omitempty"`
	PreviousBlockId       *Hash              `protobuf:"bytes,3,opt,name=previous_block_id,json=previousBlockId" json:"previous_block_id,omitempty"`
	Timestamp             uint64             `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
	TransactionsRoot      *Hash              `protobuf:"bytes,5,opt,name=transactions_root,json=transactionsRoot" json:"transactions_root,omitempty"`
	TransactionStatusHash *Hash              `protobuf:"bytes,6,opt,name=transaction_status_hash,json=transactionStatusHash" json:"transaction_status_hash,omitempty"`
	Nonce                 uint64             `protobuf:"varint,7,opt,name=nonce" json:"nonce,omitempty"`
	Bits                  uint64             `protobuf:"varint,8,opt,name=bits" json:"bits,omitempty"`
	TransactionStatus     *TransactionStatus `protobuf:"bytes,9,opt,name=transaction_status,json=transactionStatus" json:"transaction_status,omitempty"`
	WitnessArguments      [][]byte           `protobuf:"bytes,10,rep,name=witness_arguments,json=witnessArguments,proto3" json:"witness_arguments,omitempty"`
}

func (m *BlockHeader) Reset()                    { *m = BlockHeader{} }
func (m *BlockHeader) String() string            { return proto.CompactTextString(m) }
func (*BlockHeader) ProtoMessage()               {}
func (*BlockHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *BlockHeader) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *BlockHeader) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *BlockHeader) GetPreviousBlockId() *Hash {
	if m != nil {
		return m.PreviousBlockId
	}
	return nil
}

func (m *BlockHeader) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *BlockHeader) GetTransactionsRoot() *Hash {
	if m != nil {
		return m.TransactionsRoot
	}
	return nil
}

func (m *BlockHeader) GetTransactionStatusHash() *Hash {
	if m != nil {
		return m.TransactionStatusHash
	}
	return nil
}

func (m *BlockHeader) GetNonce() uint64 {
	if m != nil {
		return m.Nonce
	}
	return 0
}

func (m *BlockHeader) GetBits() uint64 {
	if m != nil {
		return m.Bits
	}
	return 0
}

func (m *BlockHeader) GetTransactionStatus() *TransactionStatus {
	if m != nil {
		return m.TransactionStatus
	}
	return nil
}

func (m *BlockHeader) GetWitnessArguments() [][]byte {
	if m != nil {
		return m.WitnessArguments
	}
	return nil
}

type TxHeader struct {
	Version        uint64  `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	SerializedSize uint64  `protobuf:"varint,2,opt,name=serialized_size,json=serializedSize" json:"serialized_size,omitempty"`
	TimeRange      uint64  `protobuf:"varint,3,opt,name=time_range,json=timeRange" json:"time_range,omitempty"`
	ResultIds      []*Hash `protobuf:"bytes,4,rep,name=result_ids,json=resultIds" json:"result_ids,omitempty"`
}

func (m *TxHeader) Reset()                    { *m = TxHeader{} }
func (m *TxHeader) String() string            { return proto.CompactTextString(m) }
func (*TxHeader) ProtoMessage()               {}
func (*TxHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *TxHeader) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *TxHeader) GetSerializedSize() uint64 {
	if m != nil {
		return m.SerializedSize
	}
	return 0
}

func (m *TxHeader) GetTimeRange() uint64 {
	if m != nil {
		return m.TimeRange
	}
	return 0
}

func (m *TxHeader) GetResultIds() []*Hash {
	if m != nil {
		return m.ResultIds
	}
	return nil
}

type TxVerifyResult struct {
	StatusFail bool `protobuf:"varint,1,opt,name=status_fail,json=statusFail" json:"status_fail,omitempty"`
}

func (m *TxVerifyResult) Reset()                    { *m = TxVerifyResult{} }
func (m *TxVerifyResult) String() string            { return proto.CompactTextString(m) }
func (*TxVerifyResult) ProtoMessage()               {}
func (*TxVerifyResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *TxVerifyResult) GetStatusFail() bool {
	if m != nil {
		return m.StatusFail
	}
	return false
}

type TransactionStatus struct {
	Version      uint64            `protobuf:"varint,1,opt,name=version" json:"version,omitempty"`
	VerifyStatus []*TxVerifyResult `protobuf:"bytes,2,rep,name=verify_status,json=verifyStatus" json:"verify_status,omitempty"`
}

func (m *TransactionStatus) Reset()                    { *m = TransactionStatus{} }
func (m *TransactionStatus) String() string            { return proto.CompactTextString(m) }
func (*TransactionStatus) ProtoMessage()               {}
func (*TransactionStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *TransactionStatus) GetVersion() uint64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *TransactionStatus) GetVerifyStatus() []*TxVerifyResult {
	if m != nil {
		return m.VerifyStatus
	}
	return nil
}

type Mux struct {
	Sources             []*ValueSource      `protobuf:"bytes,1,rep,name=sources" json:"sources,omitempty"`
	Program             *Program            `protobuf:"bytes,2,opt,name=program" json:"program,omitempty"`
	WitnessDestinations []*ValueDestination `protobuf:"bytes,3,rep,name=witness_destinations,json=witnessDestinations" json:"witness_destinations,omitempty"`
	WitnessArguments    [][]byte            `protobuf:"bytes,4,rep,name=witness_arguments,json=witnessArguments,proto3" json:"witness_arguments,omitempty"`
}

func (m *Mux) Reset()                    { *m = Mux{} }
func (m *Mux) String() string            { return proto.CompactTextString(m) }
func (*Mux) ProtoMessage()               {}
func (*Mux) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *Mux) GetSources() []*ValueSource {
	if m != nil {
		return m.Sources
	}
	return nil
}

func (m *Mux) GetProgram() *Program {
	if m != nil {
		return m.Program
	}
	return nil
}

func (m *Mux) GetWitnessDestinations() []*ValueDestination {
	if m != nil {
		return m.WitnessDestinations
	}
	return nil
}

func (m *Mux) GetWitnessArguments() [][]byte {
	if m != nil {
		return m.WitnessArguments
	}
	return nil
}

type Coinbase struct {
	WitnessDestination *ValueDestination `protobuf:"bytes,1,opt,name=witness_destination,json=witnessDestination" json:"witness_destination,omitempty"`
	Arbitrary          []byte            `protobuf:"bytes,2,opt,name=arbitrary,proto3" json:"arbitrary,omitempty"`
}

func (m *Coinbase) Reset()                    { *m = Coinbase{} }
func (m *Coinbase) String() string            { return proto.CompactTextString(m) }
func (*Coinbase) ProtoMessage()               {}
func (*Coinbase) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *Coinbase) GetWitnessDestination() *ValueDestination {
	if m != nil {
		return m.WitnessDestination
	}
	return nil
}

func (m *Coinbase) GetArbitrary() []byte {
	if m != nil {
		return m.Arbitrary
	}
	return nil
}

type Output struct {
	Source         *ValueSource `protobuf:"bytes,1,opt,name=source" json:"source,omitempty"`
	ControlProgram *Program     `protobuf:"bytes,2,opt,name=control_program,json=controlProgram" json:"control_program,omitempty"`
	Ordinal        uint64       `protobuf:"varint,3,opt,name=ordinal" json:"ordinal,omitempty"`
}

func (m *Output) Reset()                    { *m = Output{} }
func (m *Output) String() string            { return proto.CompactTextString(m) }
func (*Output) ProtoMessage()               {}
func (*Output) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *Output) GetSource() *ValueSource {
	if m != nil {
		return m.Source
	}
	return nil
}

func (m *Output) GetControlProgram() *Program {
	if m != nil {
		return m.ControlProgram
	}
	return nil
}

func (m *Output) GetOrdinal() uint64 {
	if m != nil {
		return m.Ordinal
	}
	return 0
}

type Retirement struct {
	Source  *ValueSource `protobuf:"bytes,1,opt,name=source" json:"source,omitempty"`
	Ordinal uint64       `protobuf:"varint,2,opt,name=ordinal" json:"ordinal,omitempty"`
}

func (m *Retirement) Reset()                    { *m = Retirement{} }
func (m *Retirement) String() string            { return proto.CompactTextString(m) }
func (*Retirement) ProtoMessage()               {}
func (*Retirement) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *Retirement) GetSource() *ValueSource {
	if m != nil {
		return m.Source
	}
	return nil
}

func (m *Retirement) GetOrdinal() uint64 {
	if m != nil {
		return m.Ordinal
	}
	return 0
}

type Issuance struct {
	NonceHash              *Hash             `protobuf:"bytes,1,opt,name=nonce_hash,json=nonceHash" json:"nonce_hash,omitempty"`
	Value                  *AssetAmount      `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	WitnessDestination     *ValueDestination `protobuf:"bytes,3,opt,name=witness_destination,json=witnessDestination" json:"witness_destination,omitempty"`
	WitnessAssetDefinition *AssetDefinition  `protobuf:"bytes,4,opt,name=witness_asset_definition,json=witnessAssetDefinition" json:"witness_asset_definition,omitempty"`
	WitnessArguments       [][]byte          `protobuf:"bytes,5,rep,name=witness_arguments,json=witnessArguments,proto3" json:"witness_arguments,omitempty"`
	Ordinal                uint64            `protobuf:"varint,6,opt,name=ordinal" json:"ordinal,omitempty"`
}

func (m *Issuance) Reset()                    { *m = Issuance{} }
func (m *Issuance) String() string            { return proto.CompactTextString(m) }
func (*Issuance) ProtoMessage()               {}
func (*Issuance) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *Issuance) GetNonceHash() *Hash {
	if m != nil {
		return m.NonceHash
	}
	return nil
}

func (m *Issuance) GetValue() *AssetAmount {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Issuance) GetWitnessDestination() *ValueDestination {
	if m != nil {
		return m.WitnessDestination
	}
	return nil
}

func (m *Issuance) GetWitnessAssetDefinition() *AssetDefinition {
	if m != nil {
		return m.WitnessAssetDefinition
	}
	return nil
}

func (m *Issuance) GetWitnessArguments() [][]byte {
	if m != nil {
		return m.WitnessArguments
	}
	return nil
}

func (m *Issuance) GetOrdinal() uint64 {
	if m != nil {
		return m.Ordinal
	}
	return 0
}

type Spend struct {
	SpentOutputId      *Hash             `protobuf:"bytes,1,opt,name=spent_output_id,json=spentOutputId" json:"spent_output_id,omitempty"`
	WitnessDestination *ValueDestination `protobuf:"bytes,2,opt,name=witness_destination,json=witnessDestination" json:"witness_destination,omitempty"`
	WitnessArguments   [][]byte          `protobuf:"bytes,3,rep,name=witness_arguments,json=witnessArguments,proto3" json:"witness_arguments,omitempty"`
	Ordinal            uint64            `protobuf:"varint,4,opt,name=ordinal" json:"ordinal,omitempty"`
}

func (m *Spend) Reset()                    { *m = Spend{} }
func (m *Spend) String() string            { return proto.CompactTextString(m) }
func (*Spend) ProtoMessage()               {}
func (*Spend) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *Spend) GetSpentOutputId() *Hash {
	if m != nil {
		return m.SpentOutputId
	}
	return nil
}

func (m *Spend) GetWitnessDestination() *ValueDestination {
	if m != nil {
		return m.WitnessDestination
	}
	return nil
}

func (m *Spend) GetWitnessArguments() [][]byte {
	if m != nil {
		return m.WitnessArguments
	}
	return nil
}

func (m *Spend) GetOrdinal() uint64 {
	if m != nil {
		return m.Ordinal
	}
	return 0
}

func init() {
	proto.RegisterType((*Hash)(nil), "bc.Hash")
	proto.RegisterType((*Program)(nil), "bc.Program")
	proto.RegisterType((*AssetID)(nil), "bc.AssetID")
	proto.RegisterType((*AssetAmount)(nil), "bc.AssetAmount")
	proto.RegisterType((*AssetDefinition)(nil), "bc.AssetDefinition")
	proto.RegisterType((*ValueSource)(nil), "bc.ValueSource")
	proto.RegisterType((*ValueDestination)(nil), "bc.ValueDestination")
	proto.RegisterType((*BlockHeader)(nil), "bc.BlockHeader")
	proto.RegisterType((*TxHeader)(nil), "bc.TxHeader")
	proto.RegisterType((*TxVerifyResult)(nil), "bc.TxVerifyResult")
	proto.RegisterType((*TransactionStatus)(nil), "bc.TransactionStatus")
	proto.RegisterType((*Mux)(nil), "bc.Mux")
	proto.RegisterType((*Coinbase)(nil), "bc.Coinbase")
	proto.RegisterType((*Output)(nil), "bc.Output")
	proto.RegisterType((*Retirement)(nil), "bc.Retirement")
	proto.RegisterType((*Issuance)(nil), "bc.Issuance")
	proto.RegisterType((*Spend)(nil), "bc.Spend")
}

func init() { proto.RegisterFile("bc.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 924 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x56, 0xdd, 0x6e, 0xe3, 0x44,
	0x14, 0x56, 0x62, 0x37, 0x3f, 0x27, 0xdd, 0xa6, 0x99, 0x76, 0x17, 0x6b, 0xb5, 0x88, 0xca, 0xd2,
	0xd2, 0x45, 0x48, 0x55, 0x9b, 0x2e, 0x70, 0xc3, 0x05, 0x85, 0xb2, 0x6c, 0x2e, 0xaa, 0x45, 0xd3,
	0xaa, 0xb7, 0xd6, 0xc4, 0x9e, 0x24, 0x23, 0x1c, 0x4f, 0x98, 0x19, 0x87, 0x6e, 0xef, 0x78, 0x08,
	0x1e, 0x82, 0x27, 0xe0, 0x11, 0xb8, 0xe2, 0x9d, 0xd0, 0x1c, 0x8f, 0x13, 0xe7, 0xa7, 0xfb, 0x23,
	0xc4, 0x5d, 0xce, 0x8f, 0xbf, 0x73, 0xce, 0x77, 0x7e, 0x26, 0xd0, 0x1a, 0xc6, 0x27, 0x33, 0x25,
	0x8d, 0x24, 0xf5, 0x61, 0x1c, 0xbe, 0x02, 0xff, 0x35, 0xd3, 0x13, 0xb2, 0x07, 0xf5, 0xf9, 0x69,
	0x50, 0x3b, 0xaa, 0xbd, 0x68, 0xd0, 0xfa, 0xfc, 0x14, 0xe5, 0xb3, 0xa0, 0xee, 0xe4, 0x33, 0x94,
	0xfb, 0x81, 0xe7, 0xe4, 0x3e, 0xca, 0xe7, 0x81, 0xef, 0xe4, 0xf3, 0xf0, 0x5b, 0x68, 0xfe, 0xac,
	0xe4, 0x58, 0xb1, 0x29, 0xf9, 0x14, 0x60, 0x3e, 0x8d, 0xe6, 0x5c, 0x69, 0x21, 0x33, 0x84, 0xf4,
	0x69, 0x7b, 0x3e, 0xbd, 0x2d, 0x14, 0x84, 0x80, 0x1f, 0xcb, 0x84, 0x23, 0xf6, 0x2e, 0xc5, 0xdf,
	0xe1, 0x00, 0x9a, 0x17, 0x5a, 0x73, 0x33, 0xb8, 0xfc, 0xcf, 0x89, 0x5c, 0x41, 0x07, 0xa1, 0x2e,
	0xa6, 0x32, 0xcf, 0x0c, 0xf9, 0x1c, 0x5a, 0xcc, 0x8a, 0x91, 0x48, 0x10, 0xb4, 0xd3, 0xef, 0x9c,
	0x0c, 0xe3, 0x13, 0x17, 0x8d, 0x36, 0xd1, 0x38, 0x48, 0xc8, 0x13, 0x68, 0x30, 0xfc, 0x02, 0x43,
	0xf9, 0xd4, 0x49, 0xe1, 0x18, 0xba, 0xe8, 0x7b, 0xc9, 0x47, 0x22, 0x13, 0xc6, 0x16, 0xf0, 0x35,
	0xec, 0x0b, 0xad, 0x73, 0x96, 0xc5, 0x3c, 0x9a, 0x15, 0x35, 0x57, 0xa1, 0x1d, 0x0d, 0xb4, 0x5b,
	0x3a, 0x95, 0xbc, 0x3c, 0x03, 0x3f, 0x61, 0x86, 0x61, 0x80, 0x4e, 0xbf, 0x65, 0x7d, 0x2d, 0xf5,
	0x14, 0xb5, 0x61, 0x0a, 0x9d, 0x5b, 0x96, 0xe6, 0xfc, 0x5a, 0xe6, 0x2a, 0xe6, 0xe4, 0x29, 0x78,
	0x8a, 0x8f, 0x1c, 0xee, 0xd2, 0xd7, 0x2a, 0xc9, 0x73, 0xd8, 0x99, 0x5b, 0x57, 0x87, 0xd4, 0x5d,
	0x14, 0x54, 0xd4, 0x4c, 0x0b, 0x2b, 0x79, 0x0a, 0xad, 0x99, 0xd4, 0x98, 0x33, 0xf2, 0xe5, 0xd3,
	0x85, 0x1c, 0xfe, 0x0a, 0xfb, 0x18, 0xed, 0x92, 0x6b, 0x23, 0x32, 0x86, 0x75, 0xfd, 0xcf, 0x21,
	0xff, 0xf4, 0xa0, 0xf3, 0x7d, 0x2a, 0xe3, 0x5f, 0x5e, 0x73, 0x96, 0x70, 0x45, 0x02, 0x68, 0xae,
	0xce, 0x48, 0x29, 0xda, 0x5e, 0x4c, 0xb8, 0x18, 0x4f, 0x16, 0xbd, 0x28, 0x24, 0xf2, 0x12, 0x7a,
	0x33, 0xc5, 0xe7, 0x42, 0xe6, 0x3a, 0x1a, 0x5a, 0x24, 0xdb, 0x54, 0x6f, 0x2d, 0xdd, 0x6e, 0xe9,
	0x82, 0xb1, 0x06, 0x09, 0x79, 0x06, 0x6d, 0x23, 0xa6, 0x5c, 0x1b, 0x36, 0x9d, 0xe1, 0x9c, 0xf8,
	0x74, 0xa9, 0x20, 0x5f, 0x41, 0xcf, 0x28, 0x96, 0x69, 0x16, 0xdb, 0x24, 0x75, 0xa4, 0xa4, 0x34,
	0xc1, 0xce, 0x1a, 0xe6, 0x7e, 0xd5, 0x85, 0x4a, 0x69, 0xc8, 0x77, 0xf0, 0x49, 0x45, 0x17, 0x69,
	0xc3, 0x4c, 0xae, 0xa3, 0x09, 0xd3, 0x93, 0xa0, 0xb1, 0xf6, 0xf1, 0xe3, 0x8a, 0xe3, 0x35, 0xfa,
	0xe1, 0xc2, 0x1d, 0xc2, 0x4e, 0x26, 0xb3, 0x98, 0x07, 0x4d, 0x4c, 0xa9, 0x10, 0xec, 0x72, 0x0c,
	0x85, 0xd1, 0x41, 0x0b, 0x95, 0xf8, 0x9b, 0x5c, 0x02, 0xd9, 0x8c, 0x15, 0xb4, 0x31, 0xcc, 0x63,
	0x1b, 0xe6, 0x66, 0x3d, 0x00, 0xed, 0x6d, 0xc4, 0x24, 0x5f, 0x42, 0xef, 0x37, 0x61, 0x32, 0xae,
	0x75, 0xc4, 0xd4, 0x38, 0x9f, 0xf2, 0xcc, 0xe8, 0x00, 0x8e, 0xbc, 0x17, 0xbb, 0x74, 0xdf, 0x19,
	0x2e, 0x4a, 0x7d, 0xf8, 0x47, 0x0d, 0x5a, 0x37, 0x77, 0xef, 0x6d, 0xd4, 0x31, 0x74, 0x35, 0x57,
	0x82, 0xa5, 0xe2, 0x9e, 0x27, 0x91, 0x16, 0xf7, 0xdc, 0x75, 0x6c, 0x6f, 0xa9, 0xbe, 0x16, 0xf7,
	0xdc, 0x9e, 0x04, 0x4b, 0x79, 0xa4, 0x58, 0x36, 0xe6, 0x6e, 0x32, 0xb0, 0x09, 0xd4, 0x2a, 0xc8,
	0x31, 0x80, 0xe2, 0x3a, 0x4f, 0xed, 0x96, 0xea, 0xc0, 0x3f, 0xf2, 0x56, 0x08, 0x6c, 0x17, 0xb6,
	0x41, 0xa2, 0xc3, 0x33, 0xd8, 0xbb, 0xb9, 0xbb, 0xe5, 0x4a, 0x8c, 0xde, 0x52, 0x54, 0x92, 0xcf,
	0xa0, 0xe3, 0xc8, 0x1f, 0x31, 0x91, 0x62, 0x82, 0x2d, 0x0a, 0x85, 0xea, 0x15, 0x13, 0x69, 0x38,
	0x82, 0xde, 0x06, 0x3f, 0xef, 0x28, 0xe9, 0x1b, 0x78, 0x34, 0x47, 0xfc, 0x92, 0xe7, 0x3a, 0x66,
	0x43, 0x90, 0xe7, 0x95, 0xd0, 0x74, 0xb7, 0x70, 0x2c, 0x20, 0xc3, 0x7f, 0x6a, 0xe0, 0x5d, 0xe5,
	0x77, 0xe4, 0x0b, 0x68, 0x6a, 0x5c, 0x61, 0x1d, 0xd4, 0xf0, 0x53, 0xdc, 0x95, 0xca, 0x6a, 0xd3,
	0xd2, 0x4e, 0x9e, 0x43, 0xb3, 0xbc, 0x1f, 0xf5, 0xcd, 0xfb, 0x51, 0xda, 0xc8, 0x4f, 0x70, 0x58,
	0x76, 0x2e, 0x59, 0xae, 0xab, 0x0e, 0x3c, 0x84, 0x3f, 0x5c, 0xc0, 0x57, 0x76, 0x99, 0x1e, 0xb8,
	0x2f, 0x2a, 0xba, 0x07, 0x46, 0xc0, 0x7f, 0x60, 0x04, 0x24, 0xb4, 0x7e, 0x90, 0x22, 0x1b, 0x32,
	0xcd, 0xc9, 0x8f, 0x70, 0xb0, 0x25, 0x03, 0x77, 0x29, 0xb6, 0x27, 0x40, 0x36, 0x13, 0xb0, 0x9b,
	0xc8, 0xd4, 0x50, 0x18, 0xc5, 0xd4, 0x5b, 0x77, 0xfe, 0x97, 0x8a, 0xf0, 0xf7, 0x1a, 0x34, 0xde,
	0xe4, 0x66, 0x96, 0x1b, 0x72, 0x0c, 0x8d, 0x82, 0x23, 0x17, 0x62, 0x83, 0x42, 0x67, 0x26, 0x2f,
	0xa1, 0x1b, 0xcb, 0xcc, 0x28, 0x99, 0x46, 0xef, 0x60, 0x72, 0xcf, 0xf9, 0x94, 0x87, 0x38, 0x80,
	0xa6, 0x54, 0x89, 0xc8, 0x58, 0xea, 0x46, 0xb1, 0x14, 0xc3, 0x37, 0x00, 0x94, 0x1b, 0xa1, 0xb8,
	0xe5, 0xe0, 0xc3, 0xd3, 0xa8, 0x00, 0xd6, 0x57, 0x01, 0xff, 0xaa, 0x43, 0x6b, 0xe0, 0xde, 0x01,
	0x3b, 0xe6, 0xb8, 0xe5, 0xc5, 0x9d, 0x58, 0xbf, 0xb3, 0x6d, 0xb4, 0xe1, 0x6d, 0xf8, 0xc0, 0x6b,
	0xfb, 0x40, 0x5b, 0xbc, 0x8f, 0x6c, 0xcb, 0x15, 0x04, 0x8b, 0xb1, 0xc0, 0xa7, 0x32, 0x59, 0xbc,
	0x75, 0x78, 0x2f, 0x3b, 0xfd, 0x83, 0x45, 0x02, 0xcb, 0x67, 0x90, 0x3e, 0x29, 0x47, 0x66, 0xed,
	0x79, 0xdc, 0x3a, 0x65, 0x3b, 0xdb, 0xa7, 0xac, 0xca, 0x5c, 0x63, 0x95, 0xb9, 0xbf, 0x6b, 0xb0,
	0x73, 0x3d, 0xe3, 0x59, 0x42, 0x4e, 0xa1, 0xab, 0x67, 0x3c, 0x33, 0x91, 0xc4, 0xe9, 0x58, 0xbe,
	0xe4, 0x4b, 0xee, 0x1e, 0xa1, 0x43, 0x31, 0x3d, 0x83, 0xe4, 0x21, 0x62, 0xea, 0x1f, 0x49, 0xcc,
	0xd6, 0x4a, 0xbc, 0xf7, 0x57, 0xe2, 0xaf, 0x54, 0x32, 0x6c, 0xe0, 0xbf, 0xad, 0xf3, 0x7f, 0x03,
	0x00, 0x00, 0xff, 0xff, 0xed, 0x23, 0x9c, 0xc3, 0x79, 0x09, 0x00, 0x00,
}
