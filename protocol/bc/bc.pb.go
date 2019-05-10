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
	TransactionStatus     *TransactionStatus `protobuf:"bytes,7,opt,name=transaction_status,json=transactionStatus" json:"transaction_status,omitempty"`
	WitnessArguments      [][]byte           `protobuf:"bytes,8,rep,name=witness_arguments,json=witnessArguments,proto3" json:"witness_arguments,omitempty"`
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
	// 903 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x56, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0x86, 0x28, 0x5a, 0xa2, 0x47, 0x89, 0x65, 0xad, 0x93, 0x94, 0x08, 0x52, 0xd4, 0x20, 0x90,
	0x3a, 0x45, 0x01, 0xc3, 0x96, 0xd3, 0xf6, 0xd2, 0x43, 0xdd, 0xba, 0x69, 0x74, 0x30, 0x52, 0xac,
	0x0d, 0x5f, 0x89, 0x15, 0xb9, 0x92, 0x16, 0xa5, 0xb8, 0xea, 0xee, 0x52, 0x75, 0x7c, 0xeb, 0x43,
	0xf4, 0x59, 0xfa, 0x08, 0x3d, 0xf5, 0x99, 0x5a, 0xec, 0x70, 0x29, 0x51, 0x3f, 0xce, 0x0f, 0x8a,
	0xde, 0x34, 0x3f, 0xfb, 0xcd, 0xcc, 0x37, 0x3f, 0x14, 0x04, 0xc3, 0xe4, 0x78, 0xa6, 0xa4, 0x91,
	0xc4, 0x1b, 0x26, 0xd1, 0x2b, 0xf0, 0x5f, 0x33, 0x3d, 0x21, 0x7b, 0xe0, 0xcd, 0x4f, 0xc2, 0xc6,
	0x61, 0xe3, 0x45, 0x8b, 0x7a, 0xf3, 0x13, 0x94, 0x4f, 0x43, 0xcf, 0xc9, 0xa7, 0x28, 0xf7, 0xc3,
	0xa6, 0x93, 0xfb, 0x28, 0x9f, 0x85, 0xbe, 0x93, 0xcf, 0xa2, 0x6f, 0xa1, 0xfd, 0xb3, 0x92, 0x63,
	0xc5, 0xa6, 0xe4, 0x53, 0x80, 0xf9, 0x34, 0x9e, 0x73, 0xa5, 0x85, 0xcc, 0x11, 0xd2, 0xa7, 0xbb,
	0xf3, 0xe9, 0x4d, 0xa9, 0x20, 0x04, 0xfc, 0x44, 0xa6, 0x1c, 0xb1, 0x1f, 0x50, 0xfc, 0x1d, 0x0d,
	0xa0, 0x7d, 0xae, 0x35, 0x37, 0x83, 0x8b, 0xff, 0x9c, 0xc8, 0x25, 0x74, 0x10, 0xea, 0x7c, 0x2a,
	0x8b, 0xdc, 0x90, 0xcf, 0x21, 0x60, 0x56, 0x8c, 0x45, 0x8a, 0xa0, 0x9d, 0x7e, 0xe7, 0x78, 0x98,
	0x1c, 0xbb, 0x68, 0xb4, 0x8d, 0xc6, 0x41, 0x4a, 0x9e, 0x40, 0x8b, 0xe1, 0x0b, 0x0c, 0xe5, 0x53,
	0x27, 0x45, 0x63, 0xe8, 0xa2, 0xef, 0x05, 0x1f, 0x89, 0x5c, 0x18, 0x5b, 0xc0, 0xd7, 0xb0, 0x2f,
	0xb4, 0x2e, 0x58, 0x9e, 0xf0, 0x78, 0x56, 0xd6, 0x5c, 0x87, 0x76, 0x34, 0xd0, 0x6e, 0xe5, 0x54,
	0xf1, 0xf2, 0x0c, 0xfc, 0x94, 0x19, 0x86, 0x01, 0x3a, 0xfd, 0xc0, 0xfa, 0x5a, 0xea, 0x29, 0x6a,
	0xa3, 0x0c, 0x3a, 0x37, 0x2c, 0x2b, 0xf8, 0x95, 0x2c, 0x54, 0xc2, 0xc9, 0x53, 0x68, 0x2a, 0x3e,
	0x72, 0xb8, 0x4b, 0x5f, 0xab, 0x24, 0xcf, 0x61, 0x67, 0x6e, 0x5d, 0x1d, 0x52, 0x77, 0x51, 0x50,
	0x59, 0x33, 0x2d, 0xad, 0xe4, 0x29, 0x04, 0x33, 0xa9, 0x31, 0x67, 0xe4, 0xcb, 0xa7, 0x0b, 0x39,
	0xfa, 0x15, 0xf6, 0x31, 0xda, 0x05, 0xd7, 0x46, 0xe4, 0x0c, 0xeb, 0xfa, 0x9f, 0x43, 0xfe, 0xe3,
	0x41, 0xe7, 0xfb, 0x4c, 0x26, 0xbf, 0xbc, 0xe6, 0x2c, 0xe5, 0x8a, 0x84, 0xd0, 0x5e, 0x9d, 0x91,
	0x4a, 0xb4, 0xbd, 0x98, 0x70, 0x31, 0x9e, 0x2c, 0x7a, 0x51, 0x4a, 0xe4, 0x25, 0xf4, 0x66, 0x8a,
	0xcf, 0x85, 0x2c, 0x74, 0x3c, 0xb4, 0x48, 0xb6, 0xa9, 0xcd, 0xb5, 0x74, 0xbb, 0x95, 0x0b, 0xc6,
	0x1a, 0xa4, 0xe4, 0x19, 0xec, 0x1a, 0x31, 0xe5, 0xda, 0xb0, 0xe9, 0x0c, 0xe7, 0xc4, 0xa7, 0x4b,
	0x05, 0xf9, 0x0a, 0x7a, 0x46, 0xb1, 0x5c, 0xb3, 0xc4, 0x26, 0xa9, 0x63, 0x25, 0xa5, 0x09, 0x77,
	0xd6, 0x30, 0xf7, 0xeb, 0x2e, 0x54, 0x4a, 0x43, 0xbe, 0x83, 0x4f, 0x6a, 0xba, 0x58, 0x1b, 0x66,
	0x0a, 0x1d, 0x4f, 0x98, 0x9e, 0x84, 0xad, 0xb5, 0xc7, 0x8f, 0x6b, 0x8e, 0x57, 0xe8, 0x87, 0x0b,
	0x77, 0x01, 0x64, 0x13, 0x21, 0x6c, 0xe3, 0xe3, 0xc7, 0xf6, 0xf1, 0xf5, 0xfa, 0x33, 0xda, 0xdb,
	0x40, 0x22, 0x5f, 0x42, 0xef, 0x37, 0x61, 0x72, 0xae, 0x75, 0xcc, 0xd4, 0xb8, 0x98, 0xf2, 0xdc,
	0xe8, 0x30, 0x38, 0x6c, 0xbe, 0x78, 0x40, 0xf7, 0x9d, 0xe1, 0xbc, 0xd2, 0x47, 0x7f, 0x34, 0x20,
	0xb8, 0xbe, 0x7d, 0x2f, 0xfd, 0x47, 0xd0, 0xd5, 0x5c, 0x09, 0x96, 0x89, 0x3b, 0x9e, 0xc6, 0x5a,
	0xdc, 0x71, 0xd7, 0x87, 0xbd, 0xa5, 0xfa, 0x4a, 0xdc, 0x71, 0xbb, 0xe8, 0x96, 0xc8, 0x58, 0xb1,
	0x7c, 0xcc, 0x5d, 0xbf, 0x91, 0x5a, 0x6a, 0x15, 0xe4, 0x08, 0x40, 0x71, 0x5d, 0x64, 0x76, 0xf7,
	0x74, 0xe8, 0x1f, 0x36, 0x57, 0x68, 0xd9, 0x2d, 0x6d, 0x83, 0x54, 0x47, 0xa7, 0xb0, 0x77, 0x7d,
	0x7b, 0xc3, 0x95, 0x18, 0xbd, 0xa5, 0xa8, 0x24, 0x9f, 0x41, 0xc7, 0x51, 0x3a, 0x62, 0x22, 0xc3,
	0x04, 0x03, 0x0a, 0xa5, 0xea, 0x15, 0x13, 0x59, 0x34, 0x82, 0xde, 0x06, 0x3f, 0xef, 0x28, 0xe9,
	0x1b, 0x78, 0x38, 0x47, 0xfc, 0x8a, 0x67, 0x0f, 0xb3, 0x21, 0xc8, 0xf3, 0x4a, 0x68, 0xfa, 0xa0,
	0x74, 0x2c, 0x21, 0xa3, 0xbf, 0x1b, 0xd0, 0xbc, 0x2c, 0x6e, 0xc9, 0x17, 0xd0, 0xd6, 0xb8, 0x98,
	0x3a, 0x6c, 0xe0, 0x53, 0xdc, 0x80, 0xda, 0xc2, 0xd2, 0xca, 0x4e, 0x9e, 0x43, 0xbb, 0xba, 0x0a,
	0xde, 0xe6, 0x55, 0xa8, 0x6c, 0xe4, 0x27, 0x78, 0x54, 0x75, 0x2e, 0x5d, 0x2e, 0xa1, 0x0e, 0x9b,
	0x08, 0xff, 0x68, 0x01, 0x5f, 0xdb, 0x50, 0x7a, 0xe0, 0x5e, 0xd4, 0x74, 0xf7, 0x8c, 0x80, 0x7f,
	0xcf, 0x08, 0x48, 0x08, 0x7e, 0x90, 0x22, 0x1f, 0x32, 0xcd, 0xc9, 0x8f, 0x70, 0xb0, 0x25, 0x03,
	0xb7, 0xff, 0xdb, 0x13, 0x20, 0x9b, 0x09, 0xd8, 0xfd, 0x62, 0x6a, 0x28, 0x8c, 0x62, 0xea, 0xad,
	0x3b, 0xea, 0x4b, 0x45, 0xf4, 0x7b, 0x03, 0x5a, 0x6f, 0x0a, 0x33, 0x2b, 0x0c, 0x39, 0x82, 0x56,
	0xc9, 0x91, 0x0b, 0xb1, 0x41, 0xa1, 0x33, 0x93, 0x97, 0xd0, 0x4d, 0x64, 0x6e, 0x94, 0xcc, 0xe2,
	0x77, 0x30, 0xb9, 0xe7, 0x7c, 0xaa, 0xf3, 0x1a, 0x42, 0x5b, 0xaa, 0x54, 0xe4, 0x2c, 0x73, 0xa3,
	0x58, 0x89, 0xd1, 0x1b, 0x00, 0xca, 0x8d, 0x50, 0xdc, 0x72, 0xf0, 0xe1, 0x69, 0xd4, 0x00, 0xbd,
	0x55, 0xc0, 0x3f, 0x3d, 0x08, 0x06, 0xee, 0xba, 0xdb, 0x31, 0xcf, 0xa5, 0xfd, 0x16, 0xe0, 0xf6,
	0xaf, 0x5f, 0xcf, 0x5d, 0xb4, 0xe1, 0xc6, 0x7f, 0xe0, 0x0d, 0xbd, 0xa7, 0x2d, 0xcd, 0x8f, 0x6c,
	0xcb, 0x25, 0x84, 0x8b, 0xb1, 0xc0, 0x0f, 0x60, 0xba, 0xf8, 0x82, 0xe1, 0x15, 0xec, 0xf4, 0x0f,
	0x16, 0x09, 0x2c, 0x3f, 0x6e, 0xf4, 0x49, 0x35, 0x32, 0x6b, 0x1f, 0xbd, 0xad, 0x53, 0xb6, 0xb3,
	0x7d, 0xca, 0xea, 0xcc, 0xb5, 0x56, 0x99, 0xfb, 0xab, 0x01, 0x3b, 0x57, 0x33, 0x9e, 0xa7, 0xe4,
	0x04, 0xba, 0x7a, 0xc6, 0x73, 0x13, 0x4b, 0x9c, 0x8e, 0xe5, 0xf7, 0x79, 0xc9, 0xdd, 0x43, 0x74,
	0x28, 0xa7, 0x67, 0x90, 0xde, 0x47, 0x8c, 0xf7, 0x91, 0xc4, 0x6c, 0xad, 0xa4, 0xf9, 0xfe, 0x4a,
	0xfc, 0x95, 0x4a, 0x86, 0x2d, 0xfc, 0x0f, 0x75, 0xf6, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x78,
	0xac, 0x41, 0xa8, 0x4f, 0x09, 0x00, 0x00,
}
