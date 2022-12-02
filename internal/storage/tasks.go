package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	qb "github.com/Masterminds/squirrel"
	"github.com/clintjedwards/todo/internal/models"
	"github.com/clintjedwards/todo/proto"
)

type Task struct {
	ID          string
	Title       string
	Description string
	State       string
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

type UpdatableTaskFields struct {
	Title       *string
	Description *string
	State       *string
	Modified    *int64
	Parent      *string
}

func (db *DB) ListTasks(conn Queryable, offset, limit int, excludeCompleted bool) ([]Task, error) {
	if limit == 0 || limit > db.maxResultsLimit {
		limit = db.maxResultsLimit
	}

	statement := qb.Select("id", "title", "description", "state", "created", "modified", "parent").
		From("tasks").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	if excludeCompleted {
		statement = statement.Where(qb.NotEq{"state": models.TaskStateCompleted})
	}

	query, args := statement.MustSql()

	tasks := []Task{}
	err := conn.Select(&tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return tasks, nil
}

func (db *DB) GetTask(conn Queryable, id string) (Task, error) {
	query, args := qb.Select("id", "title", "description", "state", "created", "modified", "parent").
		From("tasks").
		Where(qb.Eq{"id": id}).MustSql()

	task := Task{}
	err := conn.Get(&task, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrEntityNotFound
		}

		return Task{}, fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return task, nil
}

func (db *DB) InsertTask(conn Queryable, task *Task) error {
	_, err := conn.NamedExec(`INSERT INTO tasks (id, title, description, state, created, modified, parent) VALUES
	(:id, :title, :description, :state, :created, :modified, :parent)`, task)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrEntityExists
		}

		return fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return nil
}

func (db *DB) UpdateTask(conn Queryable, id string, fields UpdatableTaskFields) error {
	statement := qb.Update("tasks")

	if fields.Title != nil {
		statement = statement.Set("title", fields.Title)
	}

	if fields.Description != nil {
		statement = statement.Set("description", fields.Description)
	}

	if fields.State != nil {
		statement = statement.Set("state", fields.State)
	}

	if fields.Modified != nil {
		statement = statement.Set("modified", fields.Modified)
	}

	if fields.Parent != nil {
		statement = statement.Set("parent", fields.Parent)
	}

	query, args := statement.Where(qb.Eq{"id": id}).MustSql()

	_, err := conn.Exec(query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrEntityNotFound
		}

		return fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return nil
}

func (db *DB) DeleteTask(conn Queryable, id string) error {
	query, args := qb.Delete("tasks").Where(qb.Eq{"id": id}).MustSql()
	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return nil
}
