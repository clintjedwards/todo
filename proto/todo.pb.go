// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.21.12
// source: todo.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_todo_proto protoreflect.FileDescriptor

const file_todo_proto_rawDesc = "" +
	"\n" +
	"\n" +
	"todo.proto\x12\x05proto\x1a\x14todo_transport.proto2\xdf\x06\n" +
	"\x04Todo\x12J\n" +
	"\rGetSystemInfo\x12\x1b.proto.GetSystemInfoRequest\x1a\x1c.proto.GetSystemInfoResponse\x12>\n" +
	"\tListTasks\x12\x17.proto.ListTasksRequest\x1a\x18.proto.ListTasksResponse\x12A\n" +
	"\n" +
	"CreateTask\x12\x18.proto.CreateTaskRequest\x1a\x19.proto.CreateTaskResponse\x128\n" +
	"\aGetTask\x12\x15.proto.GetTaskRequest\x1a\x16.proto.GetTaskResponse\x12A\n" +
	"\n" +
	"UpdateTask\x12\x18.proto.UpdateTaskRequest\x1a\x19.proto.UpdateTaskResponse\x12A\n" +
	"\n" +
	"DeleteTask\x12\x18.proto.DeleteTaskRequest\x1a\x19.proto.DeleteTaskResponse\x12Y\n" +
	"\x12ListScheduledTasks\x12 .proto.ListScheduledTasksRequest\x1a!.proto.ListScheduledTasksResponse\x12\\\n" +
	"\x13CreateScheduledTask\x12!.proto.CreateScheduledTaskRequest\x1a\".proto.CreateScheduledTaskResponse\x12S\n" +
	"\x10GetScheduledTask\x12\x1e.proto.GetScheduledTaskRequest\x1a\x1f.proto.GetScheduledTaskResponse\x12\\\n" +
	"\x13UpdateScheduledTask\x12!.proto.UpdateScheduledTaskRequest\x1a\".proto.UpdateScheduledTaskResponse\x12\\\n" +
	"\x13DeleteScheduledTask\x12!.proto.DeleteScheduledTaskRequest\x1a\".proto.DeleteScheduledTaskResponseB%Z#github.com/clintjedwards/todo/protob\x06proto3"

var file_todo_proto_goTypes = []any{
	(*GetSystemInfoRequest)(nil),        // 0: proto.GetSystemInfoRequest
	(*ListTasksRequest)(nil),            // 1: proto.ListTasksRequest
	(*CreateTaskRequest)(nil),           // 2: proto.CreateTaskRequest
	(*GetTaskRequest)(nil),              // 3: proto.GetTaskRequest
	(*UpdateTaskRequest)(nil),           // 4: proto.UpdateTaskRequest
	(*DeleteTaskRequest)(nil),           // 5: proto.DeleteTaskRequest
	(*ListScheduledTasksRequest)(nil),   // 6: proto.ListScheduledTasksRequest
	(*CreateScheduledTaskRequest)(nil),  // 7: proto.CreateScheduledTaskRequest
	(*GetScheduledTaskRequest)(nil),     // 8: proto.GetScheduledTaskRequest
	(*UpdateScheduledTaskRequest)(nil),  // 9: proto.UpdateScheduledTaskRequest
	(*DeleteScheduledTaskRequest)(nil),  // 10: proto.DeleteScheduledTaskRequest
	(*GetSystemInfoResponse)(nil),       // 11: proto.GetSystemInfoResponse
	(*ListTasksResponse)(nil),           // 12: proto.ListTasksResponse
	(*CreateTaskResponse)(nil),          // 13: proto.CreateTaskResponse
	(*GetTaskResponse)(nil),             // 14: proto.GetTaskResponse
	(*UpdateTaskResponse)(nil),          // 15: proto.UpdateTaskResponse
	(*DeleteTaskResponse)(nil),          // 16: proto.DeleteTaskResponse
	(*ListScheduledTasksResponse)(nil),  // 17: proto.ListScheduledTasksResponse
	(*CreateScheduledTaskResponse)(nil), // 18: proto.CreateScheduledTaskResponse
	(*GetScheduledTaskResponse)(nil),    // 19: proto.GetScheduledTaskResponse
	(*UpdateScheduledTaskResponse)(nil), // 20: proto.UpdateScheduledTaskResponse
	(*DeleteScheduledTaskResponse)(nil), // 21: proto.DeleteScheduledTaskResponse
}
var file_todo_proto_depIdxs = []int32{
	0,  // 0: proto.Todo.GetSystemInfo:input_type -> proto.GetSystemInfoRequest
	1,  // 1: proto.Todo.ListTasks:input_type -> proto.ListTasksRequest
	2,  // 2: proto.Todo.CreateTask:input_type -> proto.CreateTaskRequest
	3,  // 3: proto.Todo.GetTask:input_type -> proto.GetTaskRequest
	4,  // 4: proto.Todo.UpdateTask:input_type -> proto.UpdateTaskRequest
	5,  // 5: proto.Todo.DeleteTask:input_type -> proto.DeleteTaskRequest
	6,  // 6: proto.Todo.ListScheduledTasks:input_type -> proto.ListScheduledTasksRequest
	7,  // 7: proto.Todo.CreateScheduledTask:input_type -> proto.CreateScheduledTaskRequest
	8,  // 8: proto.Todo.GetScheduledTask:input_type -> proto.GetScheduledTaskRequest
	9,  // 9: proto.Todo.UpdateScheduledTask:input_type -> proto.UpdateScheduledTaskRequest
	10, // 10: proto.Todo.DeleteScheduledTask:input_type -> proto.DeleteScheduledTaskRequest
	11, // 11: proto.Todo.GetSystemInfo:output_type -> proto.GetSystemInfoResponse
	12, // 12: proto.Todo.ListTasks:output_type -> proto.ListTasksResponse
	13, // 13: proto.Todo.CreateTask:output_type -> proto.CreateTaskResponse
	14, // 14: proto.Todo.GetTask:output_type -> proto.GetTaskResponse
	15, // 15: proto.Todo.UpdateTask:output_type -> proto.UpdateTaskResponse
	16, // 16: proto.Todo.DeleteTask:output_type -> proto.DeleteTaskResponse
	17, // 17: proto.Todo.ListScheduledTasks:output_type -> proto.ListScheduledTasksResponse
	18, // 18: proto.Todo.CreateScheduledTask:output_type -> proto.CreateScheduledTaskResponse
	19, // 19: proto.Todo.GetScheduledTask:output_type -> proto.GetScheduledTaskResponse
	20, // 20: proto.Todo.UpdateScheduledTask:output_type -> proto.UpdateScheduledTaskResponse
	21, // 21: proto.Todo.DeleteScheduledTask:output_type -> proto.DeleteScheduledTaskResponse
	11, // [11:22] is the sub-list for method output_type
	0,  // [0:11] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_todo_proto_init() }
func file_todo_proto_init() {
	if File_todo_proto != nil {
		return
	}
	file_todo_transport_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_todo_proto_rawDesc), len(file_todo_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_todo_proto_goTypes,
		DependencyIndexes: file_todo_proto_depIdxs,
	}.Build()
	File_todo_proto = out.File
	file_todo_proto_goTypes = nil
	file_todo_proto_depIdxs = nil
}
