package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Noviiich/io-bound-task/internal/domain/models"
	resp "github.com/Noviiich/io-bound-task/internal/lib/api/response"
	"github.com/Noviiich/io-bound-task/internal/processors"
	"github.com/google/uuid"
)

type TaskStorage interface {
	CreateTask(task *models.Task) error
	GetByID(id string) (*models.Task, error)
	UpdateTask(task *models.Task) error
	DeleteTask(id string) error
}

type TaskProcessor interface {
	Process(ctx context.Context, task *models.Task) error
}

type TaskService struct {
	taskStorage   TaskStorage
	taskProcessor TaskProcessor
	mu            sync.RWMutex
	cancelMap     map[string]context.CancelFunc
}

func New(taskStorage TaskStorage) *TaskService {
	processor := processors.New()

	service := &TaskService{
		taskStorage:   taskStorage,
		taskProcessor: processor,
		cancelMap:     make(map[string]context.CancelFunc),
	}

	return service
}

func (s *TaskService) RegisterTask(name string) (string, error) {
	const op = "services.RegisterTask"

	task := &models.Task{
		ID:        uuid.New().String(),
		Name:      name,
		Status:    resp.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.taskStorage.CreateTask(task); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	go s.processTask(task)

	return task.ID, nil
}

func (s *TaskService) GetTask(id string) (*models.Task, error) {
	const op = "services.GetTask"

	task, err := s.taskStorage.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if task.Status == resp.StatusRunning && task.StartedAt != nil {
		task.Duration = time.Since(*task.StartedAt)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(id string) error {
	const op = "services.DeleteTask"

	task, err := s.taskStorage.GetByID(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if task.IsActive() {
		s.CancelTask(id)
		go func() {
			time.Sleep(10 * time.Second)
			s.taskStorage.DeleteTask(id)
		}()
		return nil
	}

	return s.taskStorage.DeleteTask(id)
}

func (s *TaskService) CancelTask(id string) error {
	s.mu.Lock()
	cancelFunc, exists := s.cancelMap[id]
	if exists {
		cancelFunc()
		delete(s.cancelMap, id)
	}
	s.mu.Unlock()

	if exists {
		task, err := s.taskStorage.GetByID(id)
		if err != nil {
			return err
		}

		if task.IsActive() {
			task.Status = resp.StatusCancelled
			now := time.Now()
			task.CompletedAt = &now
			if task.StartedAt != nil {
				task.Duration = now.Sub(*task.StartedAt)
			}
			return s.taskStorage.UpdateTask(task)
		}
	}

	return nil
}

func (s *TaskService) processTask(task *models.Task) {
	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.cancelMap[task.ID] = cancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.cancelMap, task.ID)
		s.mu.Unlock()
		cancel()
	}()

	time.Sleep(10 * time.Second)
	task.Status = resp.StatusRunning
	now := time.Now()
	task.StartedAt = &now
	if err := s.taskStorage.UpdateTask(task); err != nil {
		return
	}

	err := s.taskProcessor.Process(ctx, task)

	currentTask, getErr := s.taskStorage.GetByID(task.ID)
	if getErr != nil {
		return
	}

	completedAt := time.Now()
	currentTask.CompletedAt = &completedAt
	if currentTask.StartedAt != nil {
		currentTask.Duration = completedAt.Sub(*currentTask.StartedAt)
	}

	if err != nil {
		if ctx.Err() == context.Canceled {
			currentTask.Status = resp.StatusCancelled
		} else {
			currentTask.Status = resp.StatusFailed
		}
	} else {
		currentTask.Status = resp.StatusCompleted
	}

	s.taskStorage.UpdateTask(currentTask)
}
