package tasks

import (
	"context"
	"time"

	"github.com/8run0/kllla2/backend/pkg/db/tasks"
)

var _ taskService = Service{}

type CreateTaskRequest struct {
	Task Task
}
type DeleteTaskRequest struct {
	ID uint64
}
type GetTaskRequest struct {
	ID uint64
}

type CompleteTaskRequest struct {
	ID uint64
}
type ListTasksRequest struct {
	Limit int64
}
type UpdateTaskRequest struct {
	ID   uint64
	Task Task
}
type Task struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CompletedOn time.Time `json:"completed_on"`
}

type taskService interface {
	Create(context.Context, CreateTaskRequest) (uint64, error)
	Update(context.Context, UpdateTaskRequest) error
	Delete(context.Context, DeleteTaskRequest) error
	List(context.Context, ListTasksRequest) ([]*Task, error)
	GetByID(context.Context, GetTaskRequest) (*Task, error)
	Complete(context.Context, CompleteTaskRequest) error
}

type Service struct {
	DB *tasks.DB
}

// Complete implements taskService
func (s Service) Complete(ctx context.Context, req CompleteTaskRequest) error {
	return s.DB.Complete(ctx, req.ID)
}

// Create implements taskService
func (s Service) Create(ctx context.Context, req CreateTaskRequest) (uint64, error) {
	return s.DB.Create(ctx, tasks.Task{
		Title:       req.Task.Title,
		Description: req.Task.Description,
	})
}

// Delete implements taskService
func (s Service) Delete(ctx context.Context, req DeleteTaskRequest) error {
	return s.DB.Delete(ctx, req.ID)
}

func dbTaskToTask(dbTask tasks.Task) *Task {
	task := &Task{
		ID:          dbTask.ID,
		Title:       dbTask.Title,
		Description: dbTask.Description,
		Completed:   dbTask.Completed,
		CompletedOn: dbTask.CompletedOn,
	}
	return task
}

// Get implements taskService
func (s Service) GetByID(ctx context.Context, req GetTaskRequest) (*Task, error) {
	dbTask, err := s.DB.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return dbTaskToTask(dbTask), nil
}

// List implements taskService
func (s Service) List(ctx context.Context, req ListTasksRequest) ([]*Task, error) {
	dbTasks, err := s.DB.List(ctx, req.Limit)
	if err != nil {
		return nil, err
	}
	tasks := make([]*Task, len(dbTasks))
	for pos, dbTask := range dbTasks {
		t := dbTaskToTask(dbTask)
		tasks[pos] = t
	}
	return tasks, nil
}

// Update implements taskService
func (s Service) Update(ctx context.Context, req UpdateTaskRequest) error {
	return s.DB.Update(ctx, tasks.Task{
		ID:          req.ID,
		Title:       req.Task.Title,
		Description: req.Task.Description,
	})
}
