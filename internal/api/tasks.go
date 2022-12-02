package api

import (
	"github.com/clintjedwards/todo/internal/models"
	"github.com/clintjedwards/todo/internal/storage"
)

// Returns a storage layer model from a domain-layer model.
func taskModelToStorage(task *models.Task) *storage.Task {
	return &storage.Task{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		State:       string(task.State),
		Created:     task.Created,
		Modified:    task.Modified,
		Parent:      task.Parent,
	}
}
