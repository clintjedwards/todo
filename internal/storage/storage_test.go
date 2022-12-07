package storage

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func tempFile() string {
	f, err := os.CreateTemp("", "todo-test-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

func TestCRUDTasks(t *testing.T) {
	path := tempFile()
	db, err := New(path, 200)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	task1 := Task{
		ID:          "test_task_1",
		Title:       "Test Task 1",
		Description: "This is a test task.",
		State:       "UNRESOLVED",
		Created:     0,
		Modified:    0,
	}

	task2 := Task{
		ID:          "test_task_2",
		Title:       "Test Task 2",
		Description: "This is a test task.",
		State:       "UNRESOLVED",
		Created:     0,
		Modified:    0,
	}

	task3 := Task{
		ID:          "test_task_3",
		Title:       "Test Task 3",
		Description: "A child of task 1",
		State:       "UNRESOLVED",
		Created:     0,
		Modified:    0,
		Parent:      "test_task_1",
	}

	err = db.InsertTask(db, &task1)
	if err != nil {
		t.Fatal(err)
	}

	err = db.InsertTask(db, &task2)
	if err != nil {
		t.Fatal(err)
	}

	err = db.InsertTask(db, &task3)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := db.ListTasks(db, 0, 0, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 3 {
		t.Fatalf("incorrect number of tasks retrieved from ListTasks")
	}

	tasks, err = db.GetTaskChildren(db, "test_task_1")
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fatalf("incorrect number of tasks retrieved from GetTaskChildren; got %d; want %d", len(tasks), 1)
	}

	err = db.UpdateTask(db, "test_task_2", UpdatableTaskFields{
		Modified: ptr(int64(100)),
		Parent:   ptr("test_task_1"),
	})
	if err != nil {
		t.Fatal(err)
	}

	task2.Modified = 100
	task2.Parent = "test_task_1"

	retrievedTask2, err := db.GetTask(db, "test_task_2")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(task2, retrievedTask2); diff != "" {
		t.Errorf("unexpected map values (-want +got):\n%s", diff)
	}

	err = db.DeleteTask(db, "test_task_2")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.GetTask(db, "test_task_2")
	if !errors.Is(err, ErrEntityNotFound) {
		t.Fatalf("expected error Not Found; found alternate error: %v", err)
	}
}
