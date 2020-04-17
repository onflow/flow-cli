// Code generated by protoc-gen-go. DO NOT EDIT.
// source: flow/entities/collection.proto

package entities

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type Collection struct {
	Id                   []byte   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	TransactionIds       [][]byte `protobuf:"bytes,2,rep,name=transaction_ids,json=transactionIds,proto3" json:"transaction_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Collection) Reset()         { *m = Collection{} }
func (m *Collection) String() string { return proto.CompactTextString(m) }
func (*Collection) ProtoMessage()    {}
func (*Collection) Descriptor() ([]byte, []int) {
	return fileDescriptor_b302551359ed99bf, []int{0}
}

func (m *Collection) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Collection.Unmarshal(m, b)
}
func (m *Collection) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Collection.Marshal(b, m, deterministic)
}
func (m *Collection) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Collection.Merge(m, src)
}
func (m *Collection) XXX_Size() int {
	return xxx_messageInfo_Collection.Size(m)
}
func (m *Collection) XXX_DiscardUnknown() {
	xxx_messageInfo_Collection.DiscardUnknown(m)
}

var xxx_messageInfo_Collection proto.InternalMessageInfo

func (m *Collection) GetId() []byte {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *Collection) GetTransactionIds() [][]byte {
	if m != nil {
		return m.TransactionIds
	}
	return nil
}

type CollectionGuarantee struct {
	CollectionId         []byte   `protobuf:"bytes,1,opt,name=collection_id,json=collectionId,proto3" json:"collection_id,omitempty"`
	Signatures           [][]byte `protobuf:"bytes,2,rep,name=signatures,proto3" json:"signatures,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CollectionGuarantee) Reset()         { *m = CollectionGuarantee{} }
func (m *CollectionGuarantee) String() string { return proto.CompactTextString(m) }
func (*CollectionGuarantee) ProtoMessage()    {}
func (*CollectionGuarantee) Descriptor() ([]byte, []int) {
	return fileDescriptor_b302551359ed99bf, []int{1}
}

func (m *CollectionGuarantee) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CollectionGuarantee.Unmarshal(m, b)
}
func (m *CollectionGuarantee) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CollectionGuarantee.Marshal(b, m, deterministic)
}
func (m *CollectionGuarantee) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CollectionGuarantee.Merge(m, src)
}
func (m *CollectionGuarantee) XXX_Size() int {
	return xxx_messageInfo_CollectionGuarantee.Size(m)
}
func (m *CollectionGuarantee) XXX_DiscardUnknown() {
	xxx_messageInfo_CollectionGuarantee.DiscardUnknown(m)
}

var xxx_messageInfo_CollectionGuarantee proto.InternalMessageInfo

func (m *CollectionGuarantee) GetCollectionId() []byte {
	if m != nil {
		return m.CollectionId
	}
	return nil
}

func (m *CollectionGuarantee) GetSignatures() [][]byte {
	if m != nil {
		return m.Signatures
	}
	return nil
}

func init() {
	proto.RegisterType((*Collection)(nil), "entities.Collection")
	proto.RegisterType((*CollectionGuarantee)(nil), "entities.CollectionGuarantee")
}

func init() { proto.RegisterFile("flow/entities/collection.proto", fileDescriptor_b302551359ed99bf) }

var fileDescriptor_b302551359ed99bf = []byte{
	// 163 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4b, 0xcb, 0xc9, 0x2f,
	0xd7, 0x4f, 0xcd, 0x2b, 0xc9, 0x2c, 0xc9, 0x4c, 0x2d, 0xd6, 0x4f, 0xce, 0xcf, 0xc9, 0x49, 0x4d,
	0x2e, 0xc9, 0xcc, 0xcf, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0x49, 0x29, 0xb9,
	0x72, 0x71, 0x39, 0xc3, 0x65, 0x85, 0xf8, 0xb8, 0x98, 0x32, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35,
	0x78, 0x82, 0x98, 0x32, 0x53, 0x84, 0xd4, 0xb9, 0xf8, 0x4b, 0x8a, 0x12, 0xf3, 0x8a, 0x13, 0xc1,
	0xd2, 0xf1, 0x99, 0x29, 0xc5, 0x12, 0x4c, 0x0a, 0xcc, 0x1a, 0x3c, 0x41, 0x7c, 0x48, 0xc2, 0x9e,
	0x29, 0xc5, 0x4a, 0x51, 0x5c, 0xc2, 0x08, 0x63, 0xdc, 0x4b, 0x13, 0x8b, 0x12, 0xf3, 0x4a, 0x52,
	0x53, 0x85, 0x94, 0xb9, 0x78, 0x11, 0x76, 0xc7, 0xc3, 0x8d, 0xe6, 0x41, 0x08, 0x7a, 0xa6, 0x08,
	0xc9, 0x71, 0x71, 0x15, 0x67, 0xa6, 0xe7, 0x25, 0x96, 0x94, 0x16, 0xa5, 0xc2, 0xcc, 0x47, 0x12,
	0x49, 0x62, 0x03, 0xbb, 0xd9, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xdc, 0x28, 0xf2, 0x98, 0xd5,
	0x00, 0x00, 0x00,
}
