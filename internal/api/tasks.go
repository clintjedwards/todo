package api

import (
	"time"

	"github.com/clintjedwards/avail/v2"
	"github.com/clintjedwards/todo/internal/models"
	"github.com/clintjedwards/todo/internal/storage"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

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

// Complete a parent task and all it's children recursively.
func (api *API) CompleteTaskTree(id string) ([]string, error) {
	completedTasks := []string{}

	err := storage.InsideTx(api.db.DB, func(tx *sqlx.Tx) error {
		err := api.recursivelyCompleteTasks(tx, id, &completedTasks)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return completedTasks, nil
}

// Completes a parent task and all it's children recursively.
func (api *API) recursivelyCompleteTasks(tx *sqlx.Tx, id string, completedTasks *[]string) error {
	err := api.db.UpdateTask(tx, id, storage.UpdatableTaskFields{
		State: ptr(string(models.TaskStateCompleted)),
	})
	if err != nil {
		return err
	}

	*completedTasks = append(*completedTasks, id)

	children, err := api.db.GetTaskChildren(tx, id)
	if err != nil {
		return err
	}

	for _, task := range children {
		err := api.recursivelyCompleteTasks(tx, task.ID, completedTasks)
		if err != nil {
			return err
		}
	}

	return nil
}

func (api *API) restoreReoccurringTasks() error {
	scheduledTasks, err := api.db.ListScheduledTasks(api.db, 0, 0)
	if err != nil {
		return err
	}

	for _, task := range scheduledTasks {
		task := task
		go func(scheduledTask storage.ScheduledTask) {
			avail, err := avail.New(task.Expression)
			if err != nil {
				log.Error().Err(err).Msg("could not start monitoring for scheduled task")
				return
			}

			for {
				if avail.Able(time.Now()) {
					newTask := models.NewTask(scheduledTask.Title, scheduledTask.Description, scheduledTask.Parent)

					err := api.db.InsertTask(api.db, newTask.ToStorage())
					if err != nil {
						log.Error().Err(err).Msg("could not create task")
					}

					log.Debug().Str("id", newTask.ID).Str("title", scheduledTask.Title).Msg("scheduled a new task")
				}

				time.Sleep(time.Minute * 1)
			}
		}(task)
	}

	return nil
}
