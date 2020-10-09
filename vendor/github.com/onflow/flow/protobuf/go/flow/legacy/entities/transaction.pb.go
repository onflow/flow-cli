// Code generated by protoc-gen-go. DO NOT EDIT.
// source: flow/legacy/entities/transaction.proto

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

type TransactionStatus int32

const (
	TransactionStatus_UNKNOWN   TransactionStatus = 0
	TransactionStatus_PENDING   TransactionStatus = 1
	TransactionStatus_FINALIZED TransactionStatus = 2
	TransactionStatus_EXECUTED  TransactionStatus = 3
	TransactionStatus_SEALED    TransactionStatus = 4
	TransactionStatus_EXPIRED   TransactionStatus = 5
)

var TransactionStatus_name = map[int32]string{
	0: "UNKNOWN",
	1: "PENDING",
	2: "FINALIZED",
	3: "EXECUTED",
	4: "SEALED",
	5: "EXPIRED",
}

var TransactionStatus_value = map[string]int32{
	"UNKNOWN":   0,
	"PENDING":   1,
	"FINALIZED": 2,
	"EXECUTED":  3,
	"SEALED":    4,
	"EXPIRED":   5,
}

func (x TransactionStatus) String() string {
	return proto.EnumName(TransactionStatus_name, int32(x))
}

func (TransactionStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_1525bb76cf88cd72, []int{0}
}

type Transaction struct {
	Script               []byte                   `protobuf:"bytes,1,opt,name=script,proto3" json:"script,omitempty"`
	Arguments            [][]byte                 `protobuf:"bytes,2,rep,name=arguments,proto3" json:"arguments,omitempty"`
	ReferenceBlockId     []byte                   `protobuf:"bytes,3,opt,name=reference_block_id,json=referenceBlockId,proto3" json:"reference_block_id,omitempty"`
	GasLimit             uint64                   `protobuf:"varint,4,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
	ProposalKey          *Transaction_ProposalKey `protobuf:"bytes,5,opt,name=proposal_key,json=proposalKey,proto3" json:"proposal_key,omitempty"`
	Payer                []byte                   `protobuf:"bytes,6,opt,name=payer,proto3" json:"payer,omitempty"`
	Authorizers          [][]byte                 `protobuf:"bytes,7,rep,name=authorizers,proto3" json:"authorizers,omitempty"`
	PayloadSignatures    []*Transaction_Signature `protobuf:"bytes,8,rep,name=payload_signatures,json=payloadSignatures,proto3" json:"payload_signatures,omitempty"`
	EnvelopeSignatures   []*Transaction_Signature `protobuf:"bytes,9,rep,name=envelope_signatures,json=envelopeSignatures,proto3" json:"envelope_signatures,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_1525bb76cf88cd72, []int{0}
}

func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (m *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(m, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetScript() []byte {
	if m != nil {
		return m.Script
	}
	return nil
}

func (m *Transaction) GetArguments() [][]byte {
	if m != nil {
		return m.Arguments
	}
	return nil
}

func (m *Transaction) GetReferenceBlockId() []byte {
	if m != nil {
		return m.ReferenceBlockId
	}
	return nil
}

func (m *Transaction) GetGasLimit() uint64 {
	if m != nil {
		return m.GasLimit
	}
	return 0
}

func (m *Transaction) GetProposalKey() *Transaction_ProposalKey {
	if m != nil {
		return m.ProposalKey
	}
	return nil
}

func (m *Transaction) GetPayer() []byte {
	if m != nil {
		return m.Payer
	}
	return nil
}

func (m *Transaction) GetAuthorizers() [][]byte {
	if m != nil {
		return m.Authorizers
	}
	return nil
}

func (m *Transaction) GetPayloadSignatures() []*Transaction_Signature {
	if m != nil {
		return m.PayloadSignatures
	}
	return nil
}

func (m *Transaction) GetEnvelopeSignatures() []*Transaction_Signature {
	if m != nil {
		return m.EnvelopeSignatures
	}
	return nil
}

type Transaction_ProposalKey struct {
	Address              []byte   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	KeyId                uint32   `protobuf:"varint,2,opt,name=key_id,json=keyId,proto3" json:"key_id,omitempty"`
	SequenceNumber       uint64   `protobuf:"varint,3,opt,name=sequence_number,json=sequenceNumber,proto3" json:"sequence_number,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Transaction_ProposalKey) Reset()         { *m = Transaction_ProposalKey{} }
func (m *Transaction_ProposalKey) String() string { return proto.CompactTextString(m) }
func (*Transaction_ProposalKey) ProtoMessage()    {}
func (*Transaction_ProposalKey) Descriptor() ([]byte, []int) {
	return fileDescriptor_1525bb76cf88cd72, []int{0, 0}
}

func (m *Transaction_ProposalKey) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction_ProposalKey.Unmarshal(m, b)
}
func (m *Transaction_ProposalKey) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction_ProposalKey.Marshal(b, m, deterministic)
}
func (m *Transaction_ProposalKey) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction_ProposalKey.Merge(m, src)
}
func (m *Transaction_ProposalKey) XXX_Size() int {
	return xxx_messageInfo_Transaction_ProposalKey.Size(m)
}
func (m *Transaction_ProposalKey) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction_ProposalKey.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction_ProposalKey proto.InternalMessageInfo

func (m *Transaction_ProposalKey) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Transaction_ProposalKey) GetKeyId() uint32 {
	if m != nil {
		return m.KeyId
	}
	return 0
}

func (m *Transaction_ProposalKey) GetSequenceNumber() uint64 {
	if m != nil {
		return m.SequenceNumber
	}
	return 0
}

type Transaction_Signature struct {
	Address              []byte   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	KeyId                uint32   `protobuf:"varint,2,opt,name=key_id,json=keyId,proto3" json:"key_id,omitempty"`
	Signature            []byte   `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Transaction_Signature) Reset()         { *m = Transaction_Signature{} }
func (m *Transaction_Signature) String() string { return proto.CompactTextString(m) }
func (*Transaction_Signature) ProtoMessage()    {}
func (*Transaction_Signature) Descriptor() ([]byte, []int) {
	return fileDescriptor_1525bb76cf88cd72, []int{0, 1}
}

func (m *Transaction_Signature) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction_Signature.Unmarshal(m, b)
}
func (m *Transaction_Signature) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction_Signature.Marshal(b, m, deterministic)
}
func (m *Transaction_Signature) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction_Signature.Merge(m, src)
}
func (m *Transaction_Signature) XXX_Size() int {
	return xxx_messageInfo_Transaction_Signature.Size(m)
}
func (m *Transaction_Signature) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction_Signature.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction_Signature proto.InternalMessageInfo

func (m *Transaction_Signature) GetAddress() []byte {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *Transaction_Signature) GetKeyId() uint32 {
	if m != nil {
		return m.KeyId
	}
	return 0
}

func (m *Transaction_Signature) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterEnum("entities.TransactionStatus", TransactionStatus_name, TransactionStatus_value)
	proto.RegisterType((*Transaction)(nil), "entities.Transaction")
	proto.RegisterType((*Transaction_ProposalKey)(nil), "entities.Transaction.ProposalKey")
	proto.RegisterType((*Transaction_Signature)(nil), "entities.Transaction.Signature")
}

func init() {
	proto.RegisterFile("flow/legacy/entities/transaction.proto", fileDescriptor_1525bb76cf88cd72)
}

var fileDescriptor_1525bb76cf88cd72 = []byte{
	// 459 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x5f, 0x8f, 0x93, 0x4e,
	0x14, 0xfd, 0xd1, 0xff, 0x5c, 0xba, 0x3f, 0xd9, 0xeb, 0x9f, 0x90, 0x75, 0x13, 0xd1, 0x07, 0x25,
	0xc6, 0xb4, 0xc9, 0xfa, 0x09, 0x56, 0x41, 0x43, 0xb6, 0xc1, 0x86, 0xee, 0xc6, 0xcd, 0xbe, 0x34,
	0x53, 0xb8, 0x8b, 0xa4, 0x94, 0xc1, 0x99, 0x41, 0x83, 0x5f, 0xdb, 0x2f, 0x60, 0x8a, 0xa5, 0xf0,
	0xe0, 0x83, 0x3e, 0x9e, 0x73, 0xe7, 0x9e, 0x39, 0x27, 0xe7, 0xc2, 0xcb, 0xfb, 0x8c, 0x7f, 0x9f,
	0x67, 0x94, 0xb0, 0xa8, 0x9a, 0x53, 0xae, 0x52, 0x95, 0x92, 0x9c, 0x2b, 0xc1, 0x72, 0xc9, 0x22,
	0x95, 0xf2, 0x7c, 0x56, 0x08, 0xae, 0x38, 0x4e, 0x9a, 0xd9, 0x8b, 0x9f, 0x03, 0x30, 0xae, 0xdb,
	0x39, 0x3e, 0x81, 0x91, 0x8c, 0x44, 0x5a, 0x28, 0x4b, 0xb3, 0x35, 0x67, 0x1a, 0x1e, 0x10, 0x9e,
	0x83, 0xce, 0x44, 0x52, 0xee, 0x28, 0x57, 0xd2, 0xea, 0xd9, 0x7d, 0x67, 0x1a, 0xb6, 0x04, 0xbe,
	0x01, 0x14, 0x74, 0x4f, 0x82, 0xf2, 0x88, 0xd6, 0x9b, 0x8c, 0x47, 0xdb, 0x75, 0x1a, 0x5b, 0xfd,
	0x5a, 0xc1, 0x3c, 0x4e, 0xde, 0xed, 0x07, 0x7e, 0x8c, 0x4f, 0x41, 0x4f, 0x98, 0x5c, 0x67, 0xe9,
	0x2e, 0x55, 0xd6, 0xc0, 0xd6, 0x9c, 0x41, 0x38, 0x49, 0x98, 0x5c, 0xec, 0x31, 0xba, 0x30, 0x2d,
	0x04, 0x2f, 0xb8, 0x64, 0xd9, 0x7a, 0x4b, 0x95, 0x35, 0xb4, 0x35, 0xc7, 0xb8, 0x78, 0x3e, 0x6b,
	0x1c, 0xcf, 0x3a, 0x6e, 0x67, 0xcb, 0xc3, 0xcb, 0x2b, 0xaa, 0x42, 0xa3, 0x68, 0x01, 0x3e, 0x82,
	0x61, 0xc1, 0x2a, 0x12, 0xd6, 0xa8, 0xf6, 0xf0, 0x1b, 0xa0, 0x0d, 0x06, 0x2b, 0xd5, 0x17, 0x2e,
	0xd2, 0x1f, 0x24, 0xa4, 0x35, 0xae, 0x63, 0x74, 0x29, 0x0c, 0x00, 0x0b, 0x56, 0x65, 0x9c, 0xc5,
	0x6b, 0x99, 0x26, 0x39, 0x53, 0xa5, 0x20, 0x69, 0x4d, 0xec, 0xbe, 0x63, 0x5c, 0x3c, 0xfb, 0xb3,
	0x87, 0x55, 0xf3, 0x2e, 0x3c, 0x3d, 0xac, 0x1e, 0x19, 0x89, 0x4b, 0x78, 0x48, 0xf9, 0x37, 0xca,
	0x78, 0x41, 0x5d, 0x41, 0xfd, 0xef, 0x04, 0xb1, 0xd9, 0x6d, 0x15, 0xcf, 0x12, 0x30, 0x3a, 0xa9,
	0xd1, 0x82, 0x31, 0x8b, 0x63, 0x41, 0x52, 0x1e, 0x0a, 0x6b, 0x20, 0x3e, 0x86, 0xd1, 0x96, 0xaa,
	0x7d, 0x0f, 0x3d, 0x5b, 0x73, 0x4e, 0xc2, 0xe1, 0x96, 0x2a, 0x3f, 0xc6, 0x57, 0xf0, 0x40, 0xd2,
	0xd7, 0xb2, 0x6e, 0x2a, 0x2f, 0x77, 0x1b, 0x12, 0x75, 0x4f, 0x83, 0xf0, 0xff, 0x86, 0x0e, 0x6a,
	0xf6, 0xec, 0x0e, 0xf4, 0xe3, 0xb7, 0xff, 0xfe, 0xcd, 0x39, 0xe8, 0xc7, 0xbc, 0x87, 0x43, 0x68,
	0x89, 0xd7, 0x11, 0x9c, 0x76, 0x12, 0xaf, 0x14, 0x53, 0xa5, 0x44, 0x03, 0xc6, 0x37, 0xc1, 0x55,
	0xf0, 0xe9, 0x73, 0x60, 0xfe, 0xb7, 0x07, 0x4b, 0x2f, 0x70, 0xfd, 0xe0, 0xa3, 0xa9, 0xe1, 0x09,
	0xe8, 0x1f, 0xfc, 0xe0, 0x72, 0xe1, 0xdf, 0x79, 0xae, 0xd9, 0xc3, 0x29, 0x4c, 0xbc, 0x5b, 0xef,
	0xfd, 0xcd, 0xb5, 0xe7, 0x9a, 0x7d, 0x04, 0x18, 0xad, 0xbc, 0xcb, 0x85, 0xe7, 0x9a, 0x83, 0xfd,
	0x96, 0x77, 0xbb, 0xf4, 0x43, 0xcf, 0x35, 0x87, 0x9b, 0x51, 0x7d, 0xeb, 0x6f, 0x7f, 0x05, 0x00,
	0x00, 0xff, 0xff, 0xe6, 0x7c, 0x07, 0xac, 0x15, 0x03, 0x00, 0x00,
}
