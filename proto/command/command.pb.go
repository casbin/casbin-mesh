// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.4
// source: command.proto

package command

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type EnforceRequest_Level int32

const (
	EnforceRequest_QUERY_REQUEST_LEVEL_NONE   EnforceRequest_Level = 0
	EnforceRequest_QUERY_REQUEST_LEVEL_WEAK   EnforceRequest_Level = 1
	EnforceRequest_QUERY_REQUEST_LEVEL_STRONG EnforceRequest_Level = 2
)

// Enum value maps for EnforceRequest_Level.
var (
	EnforceRequest_Level_name = map[int32]string{
		0: "QUERY_REQUEST_LEVEL_NONE",
		1: "QUERY_REQUEST_LEVEL_WEAK",
		2: "QUERY_REQUEST_LEVEL_STRONG",
	}
	EnforceRequest_Level_value = map[string]int32{
		"QUERY_REQUEST_LEVEL_NONE":   0,
		"QUERY_REQUEST_LEVEL_WEAK":   1,
		"QUERY_REQUEST_LEVEL_STRONG": 2,
	}
)

func (x EnforceRequest_Level) Enum() *EnforceRequest_Level {
	p := new(EnforceRequest_Level)
	*p = x
	return p
}

func (x EnforceRequest_Level) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EnforceRequest_Level) Descriptor() protoreflect.EnumDescriptor {
	return file_command_proto_enumTypes[0].Descriptor()
}

func (EnforceRequest_Level) Type() protoreflect.EnumType {
	return &file_command_proto_enumTypes[0]
}

func (x EnforceRequest_Level) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EnforceRequest_Level.Descriptor instead.
func (EnforceRequest_Level) EnumDescriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{4, 0}
}

type Command_Type int32

const (
	Command_COMMAND_TYPE_ENFORCE_REQUEST Command_Type = 0
	Command_COMMAND_TYPE_MODIFY_REQUEST  Command_Type = 1
	Command_COMMAND_TYPE_METADATA_SET    Command_Type = 2
	Command_COMMAND_TYPE_METADATA_DELETE Command_Type = 3
	Command_COMMAND_TYPE_NOOP            Command_Type = 4
)

// Enum value maps for Command_Type.
var (
	Command_Type_name = map[int32]string{
		0: "COMMAND_TYPE_ENFORCE_REQUEST",
		1: "COMMAND_TYPE_MODIFY_REQUEST",
		2: "COMMAND_TYPE_METADATA_SET",
		3: "COMMAND_TYPE_METADATA_DELETE",
		4: "COMMAND_TYPE_NOOP",
	}
	Command_Type_value = map[string]int32{
		"COMMAND_TYPE_ENFORCE_REQUEST": 0,
		"COMMAND_TYPE_MODIFY_REQUEST":  1,
		"COMMAND_TYPE_METADATA_SET":    2,
		"COMMAND_TYPE_METADATA_DELETE": 3,
		"COMMAND_TYPE_NOOP":            4,
	}
)

func (x Command_Type) Enum() *Command_Type {
	p := new(Command_Type)
	*p = x
	return p
}

func (x Command_Type) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Command_Type) Descriptor() protoreflect.EnumDescriptor {
	return file_command_proto_enumTypes[1].Descriptor()
}

func (Command_Type) Type() protoreflect.EnumType {
	return &file_command_proto_enumTypes[1]
}

func (x Command_Type) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Command_Type.Descriptor instead.
func (Command_Type) EnumDescriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{6, 0}
}

type StringArray struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	S []string `protobuf:"bytes,1,rep,name=s,proto3" json:"s,omitempty"`
}

func (x *StringArray) Reset() {
	*x = StringArray{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringArray) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringArray) ProtoMessage() {}

func (x *StringArray) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringArray.ProtoReflect.Descriptor instead.
func (*StringArray) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{0}
}

func (x *StringArray) GetS() []string {
	if x != nil {
		return x.S
	}
	return nil
}

type BytesArray struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	B [][]byte `protobuf:"bytes,1,rep,name=b,proto3" json:"b,omitempty"`
}

func (x *BytesArray) Reset() {
	*x = BytesArray{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesArray) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesArray) ProtoMessage() {}

func (x *BytesArray) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesArray.ProtoReflect.Descriptor instead.
func (*BytesArray) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{1}
}

func (x *BytesArray) GetB() [][]byte {
	if x != nil {
		return x.B
	}
	return nil
}

type Parameter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Value:
	//	*Parameter_I
	//	*Parameter_D
	//	*Parameter_B
	//	*Parameter_Y
	//	*Parameter_S
	//	*Parameter_A
	Value isParameter_Value `protobuf_oneof:"value"`
}

func (x *Parameter) Reset() {
	*x = Parameter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Parameter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Parameter) ProtoMessage() {}

func (x *Parameter) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Parameter.ProtoReflect.Descriptor instead.
func (*Parameter) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{2}
}

func (m *Parameter) GetValue() isParameter_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (x *Parameter) GetI() int64 {
	if x, ok := x.GetValue().(*Parameter_I); ok {
		return x.I
	}
	return 0
}

func (x *Parameter) GetD() float64 {
	if x, ok := x.GetValue().(*Parameter_D); ok {
		return x.D
	}
	return 0
}

func (x *Parameter) GetB() bool {
	if x, ok := x.GetValue().(*Parameter_B); ok {
		return x.B
	}
	return false
}

func (x *Parameter) GetY() []byte {
	if x, ok := x.GetValue().(*Parameter_Y); ok {
		return x.Y
	}
	return nil
}

func (x *Parameter) GetS() string {
	if x, ok := x.GetValue().(*Parameter_S); ok {
		return x.S
	}
	return ""
}

func (x *Parameter) GetA() *BytesArray {
	if x, ok := x.GetValue().(*Parameter_A); ok {
		return x.A
	}
	return nil
}

type isParameter_Value interface {
	isParameter_Value()
}

type Parameter_I struct {
	I int64 `protobuf:"zigzag64,1,opt,name=i,proto3,oneof"`
}

type Parameter_D struct {
	D float64 `protobuf:"fixed64,2,opt,name=d,proto3,oneof"`
}

type Parameter_B struct {
	B bool `protobuf:"varint,3,opt,name=b,proto3,oneof"`
}

type Parameter_Y struct {
	Y []byte `protobuf:"bytes,4,opt,name=y,proto3,oneof"`
}

type Parameter_S struct {
	S string `protobuf:"bytes,5,opt,name=s,proto3,oneof"`
}

type Parameter_A struct {
	A *BytesArray `protobuf:"bytes,7,opt,name=a,proto3,oneof"`
}

func (*Parameter_I) isParameter_Value() {}

func (*Parameter_D) isParameter_Value() {}

func (*Parameter_B) isParameter_Value() {}

func (*Parameter_Y) isParameter_Value() {}

func (*Parameter_S) isParameter_Value() {}

func (*Parameter_A) isParameter_Value() {}

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ns         string       `protobuf:"bytes,1,opt,name=ns,proto3" json:"ns,omitempty"`
	Parameters []*Parameter `protobuf:"bytes,2,rep,name=parameters,proto3" json:"parameters,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{3}
}

func (x *Request) GetNs() string {
	if x != nil {
		return x.Ns
	}
	return ""
}

func (x *Request) GetParameters() []*Parameter {
	if x != nil {
		return x.Parameters
	}
	return nil
}

type EnforceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request   *Request             `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
	Timings   bool                 `protobuf:"varint,2,opt,name=timings,proto3" json:"timings,omitempty"`
	Level     EnforceRequest_Level `protobuf:"varint,3,opt,name=level,proto3,enum=command.EnforceRequest_Level" json:"level,omitempty"`
	Freshness int64                `protobuf:"varint,4,opt,name=freshness,proto3" json:"freshness,omitempty"`
}

func (x *EnforceRequest) Reset() {
	*x = EnforceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EnforceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnforceRequest) ProtoMessage() {}

func (x *EnforceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnforceRequest.ProtoReflect.Descriptor instead.
func (*EnforceRequest) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{4}
}

func (x *EnforceRequest) GetRequest() *Request {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *EnforceRequest) GetTimings() bool {
	if x != nil {
		return x.Timings
	}
	return false
}

func (x *EnforceRequest) GetLevel() EnforceRequest_Level {
	if x != nil {
		return x.Level
	}
	return EnforceRequest_QUERY_REQUEST_LEVEL_NONE
}

func (x *EnforceRequest) GetFreshness() int64 {
	if x != nil {
		return x.Freshness
	}
	return 0
}

type ModifyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Request *Request `protobuf:"bytes,1,opt,name=request,proto3" json:"request,omitempty"`
}

func (x *ModifyRequest) Reset() {
	*x = ModifyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ModifyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModifyRequest) ProtoMessage() {}

func (x *ModifyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModifyRequest.ProtoReflect.Descriptor instead.
func (*ModifyRequest) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{5}
}

func (x *ModifyRequest) GetRequest() *Request {
	if x != nil {
		return x.Request
	}
	return nil
}

type Command struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type       Command_Type `protobuf:"varint,1,opt,name=type,proto3,enum=command.Command_Type" json:"type,omitempty"`
	SubCommand []byte       `protobuf:"bytes,2,opt,name=sub_command,json=subCommand,proto3" json:"sub_command,omitempty"`
	Compressed bool         `protobuf:"varint,3,opt,name=compressed,proto3" json:"compressed,omitempty"`
}

func (x *Command) Reset() {
	*x = Command{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Command) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Command) ProtoMessage() {}

func (x *Command) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Command.ProtoReflect.Descriptor instead.
func (*Command) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{6}
}

func (x *Command) GetType() Command_Type {
	if x != nil {
		return x.Type
	}
	return Command_COMMAND_TYPE_ENFORCE_REQUEST
}

func (x *Command) GetSubCommand() []byte {
	if x != nil {
		return x.SubCommand
	}
	return nil
}

func (x *Command) GetCompressed() bool {
	if x != nil {
		return x.Compressed
	}
	return false
}

type MetadataSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RaftId string            `protobuf:"bytes,1,opt,name=raft_id,json=raftId,proto3" json:"raft_id,omitempty"`
	Data   map[string]string `protobuf:"bytes,2,rep,name=data,proto3" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *MetadataSet) Reset() {
	*x = MetadataSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetadataSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetadataSet) ProtoMessage() {}

func (x *MetadataSet) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetadataSet.ProtoReflect.Descriptor instead.
func (*MetadataSet) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{7}
}

func (x *MetadataSet) GetRaftId() string {
	if x != nil {
		return x.RaftId
	}
	return ""
}

func (x *MetadataSet) GetData() map[string]string {
	if x != nil {
		return x.Data
	}
	return nil
}

type MetadataDelete struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RaftId string `protobuf:"bytes,1,opt,name=raft_id,json=raftId,proto3" json:"raft_id,omitempty"`
}

func (x *MetadataDelete) Reset() {
	*x = MetadataDelete{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetadataDelete) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetadataDelete) ProtoMessage() {}

func (x *MetadataDelete) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetadataDelete.ProtoReflect.Descriptor instead.
func (*MetadataDelete) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{8}
}

func (x *MetadataDelete) GetRaftId() string {
	if x != nil {
		return x.RaftId
	}
	return ""
}

type Noop struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Noop) Reset() {
	*x = Noop{}
	if protoimpl.UnsafeEnabled {
		mi := &file_command_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Noop) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Noop) ProtoMessage() {}

func (x *Noop) ProtoReflect() protoreflect.Message {
	mi := &file_command_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Noop.ProtoReflect.Descriptor instead.
func (*Noop) Descriptor() ([]byte, []int) {
	return file_command_proto_rawDescGZIP(), []int{9}
}

func (x *Noop) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_command_proto protoreflect.FileDescriptor

var file_command_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x22, 0x1b, 0x0a, 0x0b, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x41, 0x72, 0x72, 0x61, 0x79, 0x12, 0x0c, 0x0a, 0x01, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x01, 0x73, 0x22, 0x1a, 0x0a, 0x0a, 0x42, 0x79, 0x74, 0x65, 0x73, 0x41, 0x72,
	0x72, 0x61, 0x79, 0x12, 0x0c, 0x0a, 0x01, 0x62, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x01,
	0x62, 0x22, 0x89, 0x01, 0x0a, 0x09, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x12,
	0x0e, 0x0a, 0x01, 0x69, 0x18, 0x01, 0x20, 0x01, 0x28, 0x12, 0x48, 0x00, 0x52, 0x01, 0x69, 0x12,
	0x0e, 0x0a, 0x01, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x48, 0x00, 0x52, 0x01, 0x64, 0x12,
	0x0e, 0x0a, 0x01, 0x62, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x01, 0x62, 0x12,
	0x0e, 0x0a, 0x01, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x01, 0x79, 0x12,
	0x0e, 0x0a, 0x01, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x01, 0x73, 0x12,
	0x23, 0x0a, 0x01, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x63, 0x6f, 0x6d,
	0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x41, 0x72, 0x72, 0x61, 0x79, 0x48,
	0x00, 0x52, 0x01, 0x61, 0x42, 0x07, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x4d, 0x0a,
	0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x6e, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x6e, 0x73, 0x12, 0x32, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61,
	0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72,
	0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x22, 0x8e, 0x02, 0x0a,
	0x0e, 0x45, 0x6e, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x2a, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x10, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x74,
	0x69, 0x6d, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x74, 0x69,
	0x6d, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x33, 0x0a, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x1d, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x45,
	0x6e, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4c, 0x65,
	0x76, 0x65, 0x6c, 0x52, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x1c, 0x0a, 0x09, 0x66, 0x72,
	0x65, 0x73, 0x68, 0x6e, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x66,
	0x72, 0x65, 0x73, 0x68, 0x6e, 0x65, 0x73, 0x73, 0x22, 0x63, 0x0a, 0x05, 0x4c, 0x65, 0x76, 0x65,
	0x6c, 0x12, 0x1c, 0x0a, 0x18, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45,
	0x53, 0x54, 0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x4e, 0x4f, 0x4e, 0x45, 0x10, 0x00, 0x12,
	0x1c, 0x0a, 0x18, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54,
	0x5f, 0x4c, 0x45, 0x56, 0x45, 0x4c, 0x5f, 0x57, 0x45, 0x41, 0x4b, 0x10, 0x01, 0x12, 0x1e, 0x0a,
	0x1a, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x4c,
	0x45, 0x56, 0x45, 0x4c, 0x5f, 0x53, 0x54, 0x52, 0x4f, 0x4e, 0x47, 0x10, 0x02, 0x22, 0x3b, 0x0a,
	0x0d, 0x4d, 0x6f, 0x64, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2a,
	0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x99, 0x02, 0x0a, 0x07, 0x43,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x29, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x43,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x75, 0x62, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x73, 0x75, 0x62, 0x43, 0x6f, 0x6d, 0x6d, 0x61,
	0x6e, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73, 0x65, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x70, 0x72, 0x65, 0x73, 0x73,
	0x65, 0x64, 0x22, 0xa1, 0x01, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x1c, 0x43,
	0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x45, 0x4e, 0x46, 0x4f,
	0x52, 0x43, 0x45, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x00, 0x12, 0x1f, 0x0a,
	0x1b, 0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4d, 0x4f,
	0x44, 0x49, 0x46, 0x59, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x01, 0x12, 0x1d,
	0x0a, 0x19, 0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4d,
	0x45, 0x54, 0x41, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x53, 0x45, 0x54, 0x10, 0x02, 0x12, 0x20, 0x0a,
	0x1c, 0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4d, 0x45,
	0x54, 0x41, 0x44, 0x41, 0x54, 0x41, 0x5f, 0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x03, 0x12,
	0x15, 0x0a, 0x11, 0x43, 0x4f, 0x4d, 0x4d, 0x41, 0x4e, 0x44, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x4e, 0x4f, 0x4f, 0x50, 0x10, 0x04, 0x22, 0x93, 0x01, 0x0a, 0x0b, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x53, 0x65, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x72, 0x61, 0x66, 0x74, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x61, 0x66, 0x74, 0x49, 0x64, 0x12,
	0x32, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e,
	0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x53, 0x65, 0x74, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x1a, 0x37, 0x0a, 0x09, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x29, 0x0a, 0x0e,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x17,
	0x0a, 0x07, 0x72, 0x61, 0x66, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x72, 0x61, 0x66, 0x74, 0x49, 0x64, 0x22, 0x16, 0x0a, 0x04, 0x4e, 0x6f, 0x6f, 0x70, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x42,
	0x0b, 0x5a, 0x09, 0x2f, 0x3b, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_command_proto_rawDescOnce sync.Once
	file_command_proto_rawDescData = file_command_proto_rawDesc
)

func file_command_proto_rawDescGZIP() []byte {
	file_command_proto_rawDescOnce.Do(func() {
		file_command_proto_rawDescData = protoimpl.X.CompressGZIP(file_command_proto_rawDescData)
	})
	return file_command_proto_rawDescData
}

var file_command_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_command_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_command_proto_goTypes = []interface{}{
	(EnforceRequest_Level)(0), // 0: command.EnforceRequest.Level
	(Command_Type)(0),         // 1: command.Command.Type
	(*StringArray)(nil),       // 2: command.StringArray
	(*BytesArray)(nil),        // 3: command.BytesArray
	(*Parameter)(nil),         // 4: command.Parameter
	(*Request)(nil),           // 5: command.Request
	(*EnforceRequest)(nil),    // 6: command.EnforceRequest
	(*ModifyRequest)(nil),     // 7: command.ModifyRequest
	(*Command)(nil),           // 8: command.Command
	(*MetadataSet)(nil),       // 9: command.MetadataSet
	(*MetadataDelete)(nil),    // 10: command.MetadataDelete
	(*Noop)(nil),              // 11: command.Noop
	nil,                       // 12: command.MetadataSet.DataEntry
}
var file_command_proto_depIdxs = []int32{
	3,  // 0: command.Parameter.a:type_name -> command.BytesArray
	4,  // 1: command.Request.parameters:type_name -> command.Parameter
	5,  // 2: command.EnforceRequest.request:type_name -> command.Request
	0,  // 3: command.EnforceRequest.level:type_name -> command.EnforceRequest.Level
	5,  // 4: command.ModifyRequest.request:type_name -> command.Request
	1,  // 5: command.Command.type:type_name -> command.Command.Type
	12, // 6: command.MetadataSet.data:type_name -> command.MetadataSet.DataEntry
	7,  // [7:7] is the sub-list for method output_type
	7,  // [7:7] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_command_proto_init() }
func file_command_proto_init() {
	if File_command_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_command_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringArray); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesArray); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Parameter); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EnforceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ModifyRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Command); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetadataSet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetadataDelete); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_command_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Noop); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_command_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Parameter_I)(nil),
		(*Parameter_D)(nil),
		(*Parameter_B)(nil),
		(*Parameter_Y)(nil),
		(*Parameter_S)(nil),
		(*Parameter_A)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_command_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_command_proto_goTypes,
		DependencyIndexes: file_command_proto_depIdxs,
		EnumInfos:         file_command_proto_enumTypes,
		MessageInfos:      file_command_proto_msgTypes,
	}.Build()
	File_command_proto = out.File
	file_command_proto_rawDesc = nil
	file_command_proto_goTypes = nil
	file_command_proto_depIdxs = nil
}
