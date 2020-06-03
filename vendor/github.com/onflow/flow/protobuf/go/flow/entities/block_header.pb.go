// Code generated by protoc-gen-go. DO NOT EDIT.
// source: flow/entities/block_header.proto

package entities

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type BlockHeader struct {
	Id                   []byte               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ParentId             []byte               `protobuf:"bytes,2,opt,name=parent_id,json=parentId,proto3" json:"parent_id,omitempty"`
	Height               uint64               `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Timestamp            *timestamp.Timestamp `protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *BlockHeader) Reset()         { *m = BlockHeader{} }
func (m *BlockHeader) String() string { return proto.CompactTextString(m) }
func (*BlockHeader) ProtoMessage()    {}
func (*BlockHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_b9d8363da0276a74, []int{0}
}

func (m *BlockHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockHeader.Unmarshal(m, b)
}
func (m *BlockHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockHeader.Marshal(b, m, deterministic)
}
func (m *BlockHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockHeader.Merge(m, src)
}
func (m *BlockHeader) XXX_Size() int {
	return xxx_messageInfo_BlockHeader.Size(m)
}
func (m *BlockHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockHeader.DiscardUnknown(m)
}

var xxx_messageInfo_BlockHeader proto.InternalMessageInfo

func (m *BlockHeader) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *BlockHeader) GetParentId() []byte {
	if m != nil {
		return m.ParentId
	}
	return nil
}

func (m *BlockHeader) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *BlockHeader) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockHeader)(nil), "entities.BlockHeader")
}

func init() { proto.RegisterFile("flow/entities/block_header.proto", fileDescriptor_b9d8363da0276a74) }

var fileDescriptor_b9d8363da0276a74 = []byte{
	// 188 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x48, 0xcb, 0xc9, 0x2f,
	0xd7, 0x4f, 0xcd, 0x2b, 0xc9, 0x2c, 0xc9, 0x4c, 0x2d, 0xd6, 0x4f, 0xca, 0xc9, 0x4f, 0xce, 0x8e,
	0xcf, 0x48, 0x4d, 0x4c, 0x49, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0x49,
	0x4a, 0xc9, 0xa7, 0xe7, 0xe7, 0xa7, 0xe7, 0xa4, 0xea, 0x83, 0xc5, 0x93, 0x4a, 0xd3, 0xf4, 0x4b,
	0x32, 0x73, 0x53, 0x8b, 0x4b, 0x12, 0x73, 0x0b, 0x20, 0x4a, 0x95, 0x7a, 0x18, 0xb9, 0xb8, 0x9d,
	0x40, 0x26, 0x78, 0x80, 0x0d, 0x10, 0xe2, 0xe3, 0x62, 0xca, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4,
	0xe0, 0x09, 0x62, 0xca, 0x4c, 0x11, 0x92, 0xe6, 0xe2, 0x2c, 0x48, 0x2c, 0x4a, 0xcd, 0x2b, 0x89,
	0xcf, 0x4c, 0x91, 0x60, 0x02, 0x0b, 0x73, 0x40, 0x04, 0x3c, 0x53, 0x84, 0xc4, 0xb8, 0xd8, 0x32,
	0x52, 0x33, 0xd3, 0x33, 0x4a, 0x24, 0x98, 0x15, 0x18, 0x35, 0x58, 0x82, 0xa0, 0x3c, 0x21, 0x0b,
	0x2e, 0x4e, 0xb8, 0x3d, 0x12, 0x2c, 0x0a, 0x8c, 0x1a, 0xdc, 0x46, 0x52, 0x7a, 0x10, 0x97, 0xe8,
	0xc1, 0x5c, 0xa2, 0x17, 0x02, 0x53, 0x11, 0x84, 0x50, 0x9c, 0xc4, 0x06, 0x96, 0x36, 0x06, 0x04,
	0x00, 0x00, 0xff, 0xff, 0x93, 0xa0, 0x02, 0xb3, 0xe4, 0x00, 0x00, 0x00,
}
