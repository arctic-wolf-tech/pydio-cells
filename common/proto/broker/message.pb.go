// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

/*
Package broker is a generated protocol buffer package.

It is generated from these files:
	message.proto

It has these top-level messages:
	Empty
	PublishRequest
	SubscribeRequest
	Message
*/
package broker

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

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type PublishRequest struct {
	Topic   string   `protobuf:"bytes,1,opt,name=topic" json:"topic,omitempty"`
	Message *Message `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *PublishRequest) Reset()                    { *m = PublishRequest{} }
func (m *PublishRequest) String() string            { return proto.CompactTextString(m) }
func (*PublishRequest) ProtoMessage()               {}
func (*PublishRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PublishRequest) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func (m *PublishRequest) GetMessage() *Message {
	if m != nil {
		return m.Message
	}
	return nil
}

type SubscribeRequest struct {
	Topic string `protobuf:"bytes,1,opt,name=topic" json:"topic,omitempty"`
	Queue string `protobuf:"bytes,2,opt,name=queue" json:"queue,omitempty"`
}

func (m *SubscribeRequest) Reset()                    { *m = SubscribeRequest{} }
func (m *SubscribeRequest) String() string            { return proto.CompactTextString(m) }
func (*SubscribeRequest) ProtoMessage()               {}
func (*SubscribeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SubscribeRequest) GetTopic() string {
	if m != nil {
		return m.Topic
	}
	return ""
}

func (m *SubscribeRequest) GetQueue() string {
	if m != nil {
		return m.Queue
	}
	return ""
}

type Message struct {
	Header map[string]string `protobuf:"bytes,1,rep,name=header" json:"header,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Body   []byte            `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Message) GetHeader() map[string]string {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Message) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto.RegisterType((*Empty)(nil), "broker.Empty")
	proto.RegisterType((*PublishRequest)(nil), "broker.PublishRequest")
	proto.RegisterType((*SubscribeRequest)(nil), "broker.SubscribeRequest")
	proto.RegisterType((*Message)(nil), "broker.Message")
}

func init() { proto.RegisterFile("message.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 277 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x91, 0xcd, 0x4a, 0xc3, 0x40,
	0x10, 0xc7, 0xbb, 0xad, 0x4d, 0xc8, 0xc4, 0x6a, 0x19, 0x8a, 0x84, 0x78, 0x09, 0x39, 0xc5, 0x4b,
	0x90, 0xf4, 0xa2, 0x3d, 0x78, 0x10, 0x0a, 0x5e, 0x04, 0x5d, 0x9f, 0x20, 0xdb, 0x0e, 0x36, 0xb4,
	0x35, 0xe9, 0x26, 0x2b, 0xe6, 0x21, 0x7c, 0x67, 0xe9, 0xee, 0xc6, 0x8f, 0x1c, 0xbc, 0xed, 0x7c,
	0xf0, 0x9b, 0x3f, 0xbf, 0x85, 0xc9, 0x9e, 0xea, 0x3a, 0x7f, 0xa5, 0xb4, 0x92, 0x65, 0x53, 0xa2,
	0x23, 0x64, 0xb9, 0x25, 0x19, 0xbb, 0x30, 0x5e, 0xee, 0xab, 0xa6, 0x8d, 0x9f, 0xe1, 0xec, 0x49,
	0x89, 0x5d, 0x51, 0x6f, 0x38, 0x1d, 0x14, 0xd5, 0x0d, 0xce, 0x60, 0xdc, 0x94, 0x55, 0xb1, 0x0a,
	0x58, 0xc4, 0x12, 0x8f, 0x9b, 0x02, 0xaf, 0xc0, 0xb5, 0xa4, 0x60, 0x18, 0xb1, 0xc4, 0xcf, 0xce,
	0x53, 0x83, 0x4a, 0x1f, 0x4d, 0x9b, 0x77, 0xf3, 0xf8, 0x0e, 0xa6, 0x2f, 0x4a, 0xd4, 0x2b, 0x59,
	0x08, 0xfa, 0x1f, 0x3a, 0x83, 0xf1, 0x41, 0x91, 0x32, 0x48, 0x8f, 0x9b, 0x22, 0xfe, 0x64, 0xe0,
	0x5a, 0x28, 0xce, 0xc1, 0xd9, 0x50, 0xbe, 0x26, 0x19, 0xb0, 0x68, 0x94, 0xf8, 0xd9, 0x65, 0xef,
	0x6a, 0xfa, 0xa0, 0xa7, 0xcb, 0xb7, 0x46, 0xb6, 0xdc, 0xae, 0x22, 0xc2, 0x89, 0x28, 0xd7, 0xad,
	0xa6, 0x9e, 0x72, 0xfd, 0x0e, 0x6f, 0xc1, 0xff, 0xb5, 0x8a, 0x53, 0x18, 0x6d, 0xa9, 0xb5, 0x69,
	0x8e, 0xcf, 0x63, 0x96, 0xf7, 0x7c, 0xf7, 0x93, 0x45, 0x17, 0x8b, 0xe1, 0x0d, 0xcb, 0x3e, 0xc0,
	0xb9, 0xd7, 0x47, 0x31, 0x03, 0xd7, 0xca, 0xc2, 0x8b, 0x2e, 0xc8, 0x5f, 0x7b, 0xe1, 0xa4, 0xeb,
	0x1b, 0xbd, 0x03, 0x5c, 0x80, 0xf7, 0x6d, 0x03, 0x83, 0x6e, 0xda, 0x17, 0x14, 0xf6, 0x75, 0xc6,
	0x83, 0x6b, 0x26, 0x1c, 0xfd, 0x69, 0xf3, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x18, 0xfc, 0xae,
	0x92, 0xc5, 0x01, 0x00, 0x00,
}