// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.12
// source: todo_message.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Task_TaskState int32

const (
	Task_TASK_STATE_UNKNOWN Task_TaskState = 0
	Task_UNRESOLVED         Task_TaskState = 1
	Task_COMPLETED          Task_TaskState = 2
)

// Enum value maps for Task_TaskState.
var (
	Task_TaskState_name = map[int32]string{
		0: "TASK_STATE_UNKNOWN",
		1: "UNRESOLVED",
		2: "COMPLETED",
	}
	Task_TaskState_value = map[string]int32{
		"TASK_STATE_UNKNOWN": 0,
		"UNRESOLVED":         1,
		"COMPLETED":          2,
	}
)

func (x Task_TaskState) Enum() *Task_TaskState {
	p := new(Task_TaskState)
	*p = x
	return p
}

func (x Task_TaskState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Task_TaskState) Descriptor() protoreflect.EnumDescriptor {
	return file_todo_message_proto_enumTypes[0].Descriptor()
}

func (Task_TaskState) Type() protoreflect.EnumType {
	return &file_todo_message_proto_enumTypes[0]
}

func (x Task_TaskState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Task_TaskState.Descriptor instead.
func (Task_TaskState) EnumDescriptor() ([]byte, []int) {
	return file_todo_message_proto_rawDescGZIP(), []int{0, 0}
}

type Task struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Title         string                 `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Description   string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	State         Task_TaskState         `protobuf:"varint,4,opt,name=state,proto3,enum=proto.Task_TaskState" json:"state,omitempty"`
	Created       int64                  `protobuf:"varint,5,opt,name=created,proto3" json:"created,omitempty"`
	Modified      int64                  `protobuf:"varint,6,opt,name=modified,proto3" json:"modified,omitempty"`
	Parent        string                 `protobuf:"bytes,7,opt,name=parent,proto3" json:"parent,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Task) Reset() {
	*x = Task{}
	mi := &file_todo_message_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Task) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Task) ProtoMessage() {}

func (x *Task) ProtoReflect() protoreflect.Message {
	mi := &file_todo_message_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Task.ProtoReflect.Descriptor instead.
func (*Task) Descriptor() ([]byte, []int) {
	return file_todo_message_proto_rawDescGZIP(), []int{0}
}

func (x *Task) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Task) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Task) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Task) GetState() Task_TaskState {
	if x != nil {
		return x.State
	}
	return Task_TASK_STATE_UNKNOWN
}

func (x *Task) GetCreated() int64 {
	if x != nil {
		return x.Created
	}
	return 0
}

func (x *Task) GetModified() int64 {
	if x != nil {
		return x.Modified
	}
	return 0
}

func (x *Task) GetParent() string {
	if x != nil {
		return x.Parent
	}
	return ""
}

type ScheduledTask struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Title         string                 `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Description   string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Expression    string                 `protobuf:"bytes,4,opt,name=expression,proto3" json:"expression,omitempty"`
	Parent        string                 `protobuf:"bytes,5,opt,name=parent,proto3" json:"parent,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ScheduledTask) Reset() {
	*x = ScheduledTask{}
	mi := &file_todo_message_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ScheduledTask) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScheduledTask) ProtoMessage() {}

func (x *ScheduledTask) ProtoReflect() protoreflect.Message {
	mi := &file_todo_message_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScheduledTask.ProtoReflect.Descriptor instead.
func (*ScheduledTask) Descriptor() ([]byte, []int) {
	return file_todo_message_proto_rawDescGZIP(), []int{1}
}

func (x *ScheduledTask) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ScheduledTask) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *ScheduledTask) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *ScheduledTask) GetExpression() string {
	if x != nil {
		return x.Expression
	}
	return ""
}

func (x *ScheduledTask) GetParent() string {
	if x != nil {
		return x.Parent
	}
	return ""
}

var File_todo_message_proto protoreflect.FileDescriptor

const file_todo_message_proto_rawDesc = "" +
	"\n" +
	"\x12todo_message.proto\x12\x05proto\"\x8d\x02\n" +
	"\x04Task\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x14\n" +
	"\x05title\x18\x02 \x01(\tR\x05title\x12 \n" +
	"\vdescription\x18\x03 \x01(\tR\vdescription\x12+\n" +
	"\x05state\x18\x04 \x01(\x0e2\x15.proto.Task.TaskStateR\x05state\x12\x18\n" +
	"\acreated\x18\x05 \x01(\x03R\acreated\x12\x1a\n" +
	"\bmodified\x18\x06 \x01(\x03R\bmodified\x12\x16\n" +
	"\x06parent\x18\a \x01(\tR\x06parent\"B\n" +
	"\tTaskState\x12\x16\n" +
	"\x12TASK_STATE_UNKNOWN\x10\x00\x12\x0e\n" +
	"\n" +
	"UNRESOLVED\x10\x01\x12\r\n" +
	"\tCOMPLETED\x10\x02\"\x8f\x01\n" +
	"\rScheduledTask\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x14\n" +
	"\x05title\x18\x02 \x01(\tR\x05title\x12 \n" +
	"\vdescription\x18\x03 \x01(\tR\vdescription\x12\x1e\n" +
	"\n" +
	"expression\x18\x04 \x01(\tR\n" +
	"expression\x12\x16\n" +
	"\x06parent\x18\x05 \x01(\tR\x06parentB%Z#github.com/clintjedwards/todo/protob\x06proto3"

var (
	file_todo_message_proto_rawDescOnce sync.Once
	file_todo_message_proto_rawDescData []byte
)

func file_todo_message_proto_rawDescGZIP() []byte {
	file_todo_message_proto_rawDescOnce.Do(func() {
		file_todo_message_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_todo_message_proto_rawDesc), len(file_todo_message_proto_rawDesc)))
	})
	return file_todo_message_proto_rawDescData
}

var file_todo_message_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_todo_message_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_todo_message_proto_goTypes = []any{
	(Task_TaskState)(0),   // 0: proto.Task.TaskState
	(*Task)(nil),          // 1: proto.Task
	(*ScheduledTask)(nil), // 2: proto.ScheduledTask
}
var file_todo_message_proto_depIdxs = []int32{
	0, // 0: proto.Task.state:type_name -> proto.Task.TaskState
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_todo_message_proto_init() }
func file_todo_message_proto_init() {
	if File_todo_message_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_todo_message_proto_rawDesc), len(file_todo_message_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_todo_message_proto_goTypes,
		DependencyIndexes: file_todo_message_proto_depIdxs,
		EnumInfos:         file_todo_message_proto_enumTypes,
		MessageInfos:      file_todo_message_proto_msgTypes,
	}.Build()
	File_todo_message_proto = out.File
	file_todo_message_proto_goTypes = nil
	file_todo_message_proto_depIdxs = nil
}
