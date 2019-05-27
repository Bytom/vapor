// Code generated by protoc-gen-go. DO NOT EDIT.
// source: storage.proto

/*
Package storage is a generated protocol buffer package.

It is generated from these files:
	storage.proto

It has these top-level messages:
	UtxoEntry
*/
package storage

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

type UtxoEntry struct {
	IsCoinBase     bool   `protobuf:"varint,1,opt,name=isCoinBase" json:"isCoinBase,omitempty"`
	BlockHeight    uint64 `protobuf:"varint,2,opt,name=blockHeight" json:"blockHeight,omitempty"`
	Spent          bool   `protobuf:"varint,3,opt,name=spent" json:"spent,omitempty"`
	MainchainOutID string `protobuf:"bytes,4,opt,name=mainchainOutID" json:"mainchainOutID,omitempty"`
}

func (m *UtxoEntry) Reset()                    { *m = UtxoEntry{} }
func (m *UtxoEntry) String() string            { return proto.CompactTextString(m) }
func (*UtxoEntry) ProtoMessage()               {}
func (*UtxoEntry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *UtxoEntry) GetIsCoinBase() bool {
	if m != nil {
		return m.IsCoinBase
	}
	return false
}

func (m *UtxoEntry) GetBlockHeight() uint64 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

func (m *UtxoEntry) GetSpent() bool {
	if m != nil {
		return m.Spent
	}
	return false
}

func (m *UtxoEntry) GetMainchainOutID() string {
	if m != nil {
		return m.MainchainOutID
	}
	return ""
}

func init() {
	proto.RegisterType((*UtxoEntry)(nil), "chain.core.txdb.internal.storage.UtxoEntry")
}

func init() { proto.RegisterFile("storage.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 180 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x2e, 0xc9, 0x2f,
	0x4a, 0x4c, 0x4f, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x52, 0x48, 0xce, 0x48, 0xcc, 0xcc,
	0xd3, 0x4b, 0xce, 0x2f, 0x4a, 0xd5, 0x2b, 0xa9, 0x48, 0x49, 0xd2, 0xcb, 0xcc, 0x2b, 0x49, 0x2d,
	0xca, 0x4b, 0xcc, 0xd1, 0x83, 0xaa, 0x53, 0xea, 0x66, 0xe4, 0xe2, 0x0c, 0x2d, 0xa9, 0xc8, 0x77,
	0xcd, 0x2b, 0x29, 0xaa, 0x14, 0x92, 0xe3, 0xe2, 0xca, 0x2c, 0x76, 0xce, 0xcf, 0xcc, 0x73, 0x4a,
	0x2c, 0x4e, 0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x08, 0x42, 0x12, 0x11, 0x52, 0xe0, 0xe2, 0x4e,
	0xca, 0xc9, 0x4f, 0xce, 0xf6, 0x48, 0xcd, 0x4c, 0xcf, 0x28, 0x91, 0x60, 0x52, 0x60, 0xd4, 0x60,
	0x09, 0x42, 0x16, 0x12, 0x12, 0xe1, 0x62, 0x2d, 0x2e, 0x48, 0xcd, 0x2b, 0x91, 0x60, 0x06, 0x6b,
	0x86, 0x70, 0x84, 0xd4, 0xb8, 0xf8, 0x72, 0x13, 0x33, 0xf3, 0xc0, 0xae, 0xf1, 0x2f, 0x2d, 0xf1,
	0x74, 0x91, 0x60, 0x51, 0x60, 0xd4, 0xe0, 0x0c, 0x42, 0x13, 0x75, 0xe2, 0x8c, 0x62, 0x87, 0x3a,
	0x2c, 0x89, 0x0d, 0xec, 0x03, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x68, 0xac, 0x4c, 0xf7,
	0xd2, 0x00, 0x00, 0x00,
}
