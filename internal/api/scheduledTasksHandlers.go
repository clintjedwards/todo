package api

import (
	"context"
	"errors"
	"time"

	"github.com/clintjedwards/avail/v2"
	"github.com/clintjedwards/todo/internal/models"
	"github.com/clintjedwards/todo/internal/storage"
	proto "github.com/clintjedwards/todo/proto"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (api *API) GetScheduledTask(ctx context.Context, request *proto.GetScheduledTaskRequest) (*proto.GetScheduledTaskResponse, error) {
	if request.Id == "" {
		return nil, status.Error(codes.FailedPrecondition, "id required")
	}

	scheduledTask, err := api.db.GetScheduledTask(api.db, request.Id)
	if err != nil {
		if errors.Is(err, storage.ErrEntityNotFound) {
			return nil, status.Error(codes.FailedPrecondition, "scheduled task not found")
		}
		log.Error().Err(err).Msg("could not get scheduled task")
		return nil, status.Errorf(codes.FailedPrecondition, "failed to retrieve scheduled task %s from database", request.Id)
	}

	return &proto.GetScheduledTaskResponse{ScheduledTask: scheduledTask.ToProto()}, nil
}

func (api *API) ListScheduledTasks(ctx context.Context, request *proto.ListScheduledTasksRequest) (*proto.ListScheduledTasksResponse, error) {
	scheduledTask, err := api.db.ListScheduledTasks(api.db, int(request.Offset), int(request.Limit))
	if err != nil {
		log.Error().Err(err).Msg("could not get scheduledTask")
		return &proto.ListScheduledTasksResponse{}, status.Error(codes.Internal, "failed to retrieve scheduledTask from database")
	}

	protoScheduledTasks := []*proto.ScheduledTask{}
	for _, scheduledTask := range scheduledTask {
		protoScheduledTasks = append(protoScheduledTasks, scheduledTask.ToProto())
	}

	return &proto.ListScheduledTasksResponse{
		ScheduledTasks: protoScheduledTasks,
	}, nil
}

func (api *API) CreateScheduledTask(ctx context.Context, request *proto.CreateScheduledTaskRequest) (*proto.CreateScheduledTaskResponse, error) {
	if request.Title == "" {
		return nil, status.Error(codes.FailedPrecondition, "title required")
	}

	if request.Expression == "" {
		return nil, status.Error(codes.FailedPrecondition, "expression required")
	}

	newScheduledTask := models.NewScheduledTask(request.Title, request.Description, request.Parent, request.Expression)
	// Register it in our in-memory cache
	innerCtx, cancelFunc := context.WithCancel(context.Background())
	api.scheduledTasks[newScheduledTask.ID] = cancelFunc

	avail, err := avail.New(request.Expression)
	if err != nil {
		return &proto.CreateScheduledTaskResponse{}, status.Errorf(codes.FailedPrecondition, "incorrect expression used; %v", err)
	}

	go func(ctx context.Context, scheduledTask models.ScheduledTask) {
		// Sleep for a minute first in case the user's time-frame matches exactly the time he created the task.
		time.Sleep(time.Minute * 1)

		for {
			select {
			case <-ctx.Done():
				log.Debug().Str("id", scheduledTask.ID).Msg("scheduled task processing cancelled")
				return
			default:
				if avail.Able(time.Now()) {
					newTask := models.NewTask(scheduledTask.Title, scheduledTask.Description, scheduledTask.Parent)

					err := api.db.InsertTask(api.db, newTask.ToStorage())
					if err != nil {
						log.Error().Err(err).Msg("could not create task")
					}

					log.Debug().Str("id", newTask.ID).Str("title", scheduledTask.Title).Str("scheduled_task_id", scheduledTask.ID).Msg("scheduled a new task")
				}

				time.Sleep(time.Minute * 1)
			}
		}
	}(innerCtx, *newScheduledTask)

	err = api.db.InsertScheduledTask(api.db, newScheduledTask.ToStorage())
	if err != nil {
		if errors.Is(err, storage.ErrEntityExists) {
			return &proto.CreateScheduledTaskResponse{}, status.Error(codes.AlreadyExists, "scheduled task already exists")
		}

		return &proto.CreateScheduledTaskResponse{},
			status.Error(codes.Internal, "could not insert scheduled task")
	}

	return &proto.CreateScheduledTaskResponse{Id: newScheduledTask.ID}, nil
}

func (api *API) UpdateScheduledTask(ctx context.Context, request *proto.UpdateScheduledTaskRequest) (*proto.UpdateScheduledTaskResponse, error) {
	if request.Id == "" {
		return &proto.UpdateScheduledTaskResponse{}, status.Error(codes.FailedPrecondition, "id required")
	}

	err := api.db.UpdateScheduledTask(api.db, request.Id, storage.UpdatableScheduledTaskFields{
		Title:       &request.Title,
		Description: &request.Description,
		Parent:      &request.Parent,
		Expression:  &request.Expression,
	})
	if err != nil {
		if errors.Is(err, storage.ErrEntityNotFound) {
			return &proto.UpdateScheduledTaskResponse{}, status.Error(codes.FailedPrecondition, "could not find scheduled task")
		}
		return &proto.UpdateScheduledTaskResponse{}, err
	}

	log.Info().Str("id", request.Id).Msg("updated scheduled task")
	return &proto.UpdateScheduledTaskResponse{}, nil
}

func (api *API) DeleteScheduledTask(ctx context.Context, request *proto.DeleteScheduledTaskRequest) (*proto.DeleteScheduledTaskResponse, error) {
	if request.Id == "" {
		return nil, status.Error(codes.FailedPrecondition, "id required")
	}

	cancel, exists := api.scheduledTasks[request.Id]
	if !exists {
		return &proto.DeleteScheduledTaskResponse{Id: request.Id}, nil
	}

	cancel()

	err := api.db.DeleteScheduledTask(api.db, request.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not delete scheduled task; %v", err)
	}

	return &proto.DeleteScheduledTaskResponse{
		Id: request.Id,
	}, nil
}
