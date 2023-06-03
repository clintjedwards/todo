package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	qb "github.com/Masterminds/squirrel"
	"github.com/clintjedwards/todo/proto"
)

type ScheduledTask struct {
	ID          string `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Expression  string `db:"expression"`
	Parent      string `db:"parent"`
}

func (t *ScheduledTask) ToProto() *proto.ScheduledTask {
	return &proto.ScheduledTask{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Expression:  t.Expression,
		Parent:      t.Parent,
	}
}

type UpdatableScheduledTaskFields struct {
	Title       *string
	Description *string
	Expression  *string
	Parent      *string
}

func (db *DB) ListScheduledTasks(conn Queryable, offset, limit int) ([]ScheduledTask, error) {
	if limit == 0 || limit > db.maxResultsLimit {
		limit = db.maxResultsLimit
	}

	statement := qb.Select("id", "title", "description", "expression", "parent").
		From("scheduled_tasks").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	query, args := statement.MustSql()

	tasks := []ScheduledTask{}
	err := conn.Select(&tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return tasks, nil
}

func (db *DB) GetScheduledTask(conn Queryable, id string) (ScheduledTask, error) {
	query, args := qb.Select("id", "title", "description", "expression", "parent").
		From("scheduled_tasks").
		Where(qb.Eq{"id": id}).MustSql()

	task := ScheduledTask{}
	err := conn.Get(&task, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ScheduledTask{}, ErrEntityNotFound
		}

		return ScheduledTask{}, fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return task, nil
}

func (db *DB) InsertScheduledTask(conn Queryable, task *ScheduledTask) error {
	_, err := conn.NamedExec(`INSERT INTO scheduled_tasks (id, title, description, expression, parent) VALUES
	(:id, :title, :description, :expression, :parent)`, task)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrEntityExists
		}

		return fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return nil
}

func (db *DB) UpdateScheduledTask(conn Queryable, id string, fields UpdatableScheduledTaskFields) error {
	statement := qb.Update("scheduled_tasks")

	if fields.Title != nil {
		statement = statement.Set("title", fields.Title)
	}

	if fields.Description != nil {
		statement = statement.Set("description", fields.Description)
	}

	if fields.Parent != nil {
		statement = statement.Set("parent", fields.Parent)
	}

	if fields.Expression != nil {
		statement = statement.Set("expression", fields.Expression)
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

func (db *DB) DeleteScheduledTask(conn Queryable, id string) error {
	query, args := qb.Delete("scheduled_tasks").Where(qb.Eq{"id": id}).MustSql()
	_, err := conn.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("database error occurred: %v; %w", err, ErrInternal)
	}

	return nil
}
