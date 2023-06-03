package models

import (
	"math/rand"
	"time"

	"github.com/clintjedwards/todo/internal/storage"
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

// Returns a storage layer model from a domain-layer model.
func (t *Task) ToStorage() *storage.Task {
	return &storage.Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		State:       string(t.State),
		Created:     t.Created,
		Modified:    t.Modified,
		Parent:      t.Parent,
	}
}

func NewTask(title, description, parent string) *Task {
	return &Task{
		ID:          string(generateRandString(3)),
		Title:       title,
		Description: description,
		State:       TaskStateUnresolved,
		Created:     time.Now().UnixMilli(),
		Modified:    0,
		Parent:      parent,
	}
}

type ScheduledTask struct {
	ID          string
	Title       string
	Description string
	Parent      string
	Expression  string
}

func (t *ScheduledTask) ToProto() *proto.ScheduledTask {
	return &proto.ScheduledTask{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Parent:      t.Parent,
		Expression:  t.Expression,
	}
}

// Returns a storage layer model from a domain-layer model.
func (t *ScheduledTask) ToStorage() *storage.ScheduledTask {
	return &storage.ScheduledTask{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Parent:      t.Parent,
		Expression:  t.Expression,
	}
}

func NewScheduledTask(title, description, parent, expression string) *ScheduledTask {
	return &ScheduledTask{
		ID:          string(generateRandString(3)),
		Title:       title,
		Description: description,
		Parent:      parent,
		Expression:  expression,
	}
}

// generateRandString generates a variable length string; can be used for ids
func generateRandString(length int) []byte {
	const charset = "abcdefghijklmnopqrstuvwxyz" + "0123456789"

	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}

	return b
}
