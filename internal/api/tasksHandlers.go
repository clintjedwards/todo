package api

import (
	"context"
	"errors"
	"time"

	"github.com/clintjedwards/todo/internal/models"
	"github.com/clintjedwards/todo/internal/storage"
	proto "github.com/clintjedwards/todo/proto"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (api *API) GetTask(ctx context.Context, request *proto.GetTaskRequest) (*proto.GetTaskResponse, error) {
	if request.Id == "" {
		return nil, status.Error(codes.FailedPrecondition, "id required")
	}

	task, err := api.db.GetTask(api.db, request.Id)
	if err != nil {
		if errors.Is(err, storage.ErrEntityNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "task not found")
		}
		log.Error().Err(err).Msg("could not get task")
		return nil, status.Errorf(codes.FailedPrecondition, "failed to retrieve task %s from database", request.Id)
	}

	return &proto.GetTaskResponse{Task: task.ToProto()}, nil
}

func (api *API) ListTasks(ctx context.Context, request *proto.ListTasksRequest) (*proto.ListTasksResponse, error) {
	tasks, err := api.db.ListTasks(api.db, int(request.Offset), int(request.Limit), request.ExcludeCompleted)
	if err != nil {
		log.Error().Err(err).Msg("could not get tasks")
		return &proto.ListTasksResponse{}, status.Error(codes.Internal, "failed to retrieve tasks from database")
	}

	protoTasks := []*proto.Task{}
	for _, task := range tasks {
		protoTasks = append(protoTasks, task.ToProto())
	}

	return &proto.ListTasksResponse{
		Tasks: protoTasks,
	}, nil
}

func (api *API) CreateTask(ctx context.Context, request *proto.CreateTaskRequest) (*proto.CreateTaskResponse, error) {
	if request.Title == "" {
		return nil, status.Error(codes.FailedPrecondition, "title required")
	}

	newTask := models.NewTask(request.Title, request.Description, request.Parent)

	err := api.db.InsertTask(api.db, taskModelToStorage(newTask))
	if err != nil {
		if errors.Is(err, storage.ErrEntityExists) {
			return &proto.CreateTaskResponse{},
				status.Error(codes.AlreadyExists, "task already exists")
		}

		return &proto.CreateTaskResponse{},
			status.Error(codes.Internal, "could not insert task")
	}

	return &proto.CreateTaskResponse{Id: newTask.ID}, nil
}

func (api *API) UpdateTask(ctx context.Context, request *proto.UpdateTaskRequest) (*proto.UpdateTaskResponse, error) {
	if request.Id == "" {
		return &proto.UpdateTaskResponse{}, status.Error(codes.FailedPrecondition, "id required")
	}

	err := api.db.UpdateTask(api.db, request.Id, storage.UpdatableTaskFields{
		Title:       &request.Title,
		Description: &request.Description,
		Modified:    ptr(time.Now().UnixMilli()),
		Parent:      &request.Parent,
	})
	if err != nil {
		if errors.Is(err, storage.ErrEntityNotFound) {
			return &proto.UpdateTaskResponse{}, status.Error(codes.FailedPrecondition, "could not find task")
		}
		return &proto.UpdateTaskResponse{}, err
	}

	log.Info().Interface("task", request.Id).Msg("updated task")
	return &proto.UpdateTaskResponse{}, nil
}

func (api *API) DeleteTask(ctx context.Context, request *proto.DeleteTaskRequest) (*proto.DeleteTaskResponse, error) {
	if request.Id == "" {
		return nil, status.Error(codes.FailedPrecondition, "id required")
	}

	// If you delete a parent task we also need to delete all the children tasks.
	deletedTasks, err := api.DeleteTaskTree(request.Id)
	if err != nil {
		if errors.Is(err, storage.ErrEntityNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "could not find task")
		}
		return nil, err
	}

	return &proto.DeleteTaskResponse{
		Ids: deletedTasks,
	}, nil
}
