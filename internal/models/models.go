package models

import (
	"math/rand"
	"time"

	"github.com/clintjedwards/todo/proto"
)

type TaskState string

const (
	TaskStateUnknown    TaskState = "UNKNOWN"
	TaskStateUnresolved TaskState = "UNRESOLVED"
	TaskStateCompleted  TaskState = "COMPLETED"
)

type Task struct {
	ID          string
	Title       string
	Description string
	State       TaskState
	Created     int64
	Modified    int64
	Parent      string
}

func (t *Task) ToProto() *proto.Task {
	return &proto.Task{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		State:       proto.Task_TaskState(proto.Task_TaskState_value[string(t.State)]),
		Created:     t.Created,
		Modified:    t.Modified,
		Parent:      t.Parent,
	}
}

func NewTask(title, description, parent string) *Task {
	return &Task{
		ID:          string(generateRandString(5)),
		Title:       title,
		Description: description,
		State:       TaskStateUnresolved,
		Created:     time.Now().UnixMilli(),
		Modified:    0,
		Parent:      parent,
	}
}

// generateRandString generates a variable length string; can be used for ids
func generateRandString(length int) []byte {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return b
}
