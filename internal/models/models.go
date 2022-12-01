package models

type TaskState string

const (
	TaskStateUnknown    TaskState = "UNKNOWN"
	TaskStateUnresolved TaskState = "UNRESOLVED"
	TaskStateCompleted  TaskState = "COMPLETED"
)

type Task struct{}
