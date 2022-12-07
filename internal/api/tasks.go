package api

import (
	"github.com/clintjedwards/todo/internal/models"
	"github.com/clintjedwards/todo/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
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

// Deletes a parent task and all it's children recursively.
func (api *API) DeleteTaskTree(id string) ([]string, error) {
	deletedTasks := []string{}

	err := storage.InsideTx(api.db.DB, func(tx *sqlx.Tx) error {
		err := api.recursivelyDeleteTasks(tx, id, &deletedTasks)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return deletedTasks, nil
}

// Deletes a parent task and all it's children recursively.
func (api *API) recursivelyDeleteTasks(tx *sqlx.Tx, id string, deletedTasks *[]string) error {
	err := api.db.DeleteTask(tx, id)
	if err != nil {
		return err
	}

	log.Info().Str("id", id).Msg("deleted task")
	*deletedTasks = append(*deletedTasks, id)

	children, err := api.db.GetTaskChildren(tx, id)
	if err != nil {
		return err
	}

	for _, task := range children {
		err := api.recursivelyDeleteTasks(tx, task.ID, deletedTasks)
		if err != nil {
			return err
		}
	}

	return nil
}
