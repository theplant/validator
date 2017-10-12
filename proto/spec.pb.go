// Code generated by protoc-gen-go.
// source: spec.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	spec.proto

It has these top-level messages:
	BadRequest
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type BadRequest struct {
	// Describes all violations in a client request.
	FieldViolations []*BadRequest_FieldViolation `protobuf:"bytes,1,rep,name=field_violations,json=fieldViolations" json:"field_violations,omitempty"`
}

func (m *BadRequest) Reset()                    { *m = BadRequest{} }
func (m *BadRequest) String() string            { return proto1.CompactTextString(m) }
func (*BadRequest) ProtoMessage()               {}
func (*BadRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BadRequest) GetFieldViolations() []*BadRequest_FieldViolation {
	if m != nil {
		return m.FieldViolations
	}
	return nil
}

// A message type used to describe a single bad request field.
type BadRequest_FieldViolation struct {
	// A path leading to a field in the request body. The value will be a
	// sequence of dot-separated identifiers that identify a protocol buffer
	// field. E.g., "field_violations.field" would identify this field.
	Field string `protobuf:"bytes,1,opt,name=field" json:"field,omitempty"`
	Code  string `protobuf:"bytes,2,opt,name=code" json:"code,omitempty"`
	// A description of why the request element is bad.
	Message string `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
}

func (m *BadRequest_FieldViolation) Reset()                    { *m = BadRequest_FieldViolation{} }
func (m *BadRequest_FieldViolation) String() string            { return proto1.CompactTextString(m) }
func (*BadRequest_FieldViolation) ProtoMessage()               {}
func (*BadRequest_FieldViolation) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

func (m *BadRequest_FieldViolation) GetField() string {
	if m != nil {
		return m.Field
	}
	return ""
}

func (m *BadRequest_FieldViolation) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *BadRequest_FieldViolation) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto1.RegisterType((*BadRequest)(nil), "proto.BadRequest")
	proto1.RegisterType((*BadRequest_FieldViolation)(nil), "proto.BadRequest.FieldViolation")
}

func init() { proto1.RegisterFile("spec.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 157 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0x2e, 0x48, 0x4d,
	0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0xeb, 0x19, 0xb9, 0xb8, 0x9c,
	0x12, 0x53, 0x82, 0x52, 0x0b, 0x4b, 0x53, 0x8b, 0x4b, 0x84, 0xbc, 0xb9, 0x04, 0xd2, 0x32, 0x53,
	0x73, 0x52, 0xe2, 0xcb, 0x32, 0xf3, 0x73, 0x12, 0x4b, 0x32, 0xf3, 0xf3, 0x8a, 0x25, 0x18, 0x15,
	0x98, 0x35, 0xb8, 0x8d, 0x14, 0x20, 0xfa, 0xf4, 0x10, 0x8a, 0xf5, 0xdc, 0x40, 0x2a, 0xc3, 0x60,
	0x0a, 0x83, 0xf8, 0xd3, 0x50, 0xf8, 0xc5, 0x52, 0x21, 0x5c, 0x7c, 0xa8, 0x4a, 0x84, 0x44, 0xb8,
	0x58, 0xc1, 0x8a, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0x20, 0x1c, 0x21, 0x21, 0x2e, 0x96,
	0xe4, 0xfc, 0x94, 0x54, 0x09, 0x26, 0xb0, 0x20, 0x98, 0x2d, 0x24, 0xc1, 0xc5, 0x9e, 0x9b, 0x5a,
	0x5c, 0x9c, 0x98, 0x9e, 0x2a, 0xc1, 0x0c, 0x16, 0x86, 0x71, 0x93, 0xd8, 0xc0, 0xee, 0x30, 0x06,
	0x04, 0x00, 0x00, 0xff, 0xff, 0x69, 0x52, 0x43, 0xb3, 0xcd, 0x00, 0x00, 0x00,
}
