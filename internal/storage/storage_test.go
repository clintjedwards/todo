package storage

import (
	"os"
	"testing"
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
	_, err := New(path, 200)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
}
