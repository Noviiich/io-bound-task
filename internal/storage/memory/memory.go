package memory

import (
	"fmt"
	"sync"

	"github.com/Noviiich/io-bound-task/internal/domain/models"
	"github.com/Noviiich/io-bound-task/internal/storage"
)

type Storage struct {
	mu    sync.RWMutex
	tasks map[string]*models.Task
}

func New() *Storage {
	return &Storage{
		tasks: make(map[string]*models.Task),
	}
}

func (r *Storage) CreateTask(task *models.Task) error {
	const op = "storage.memory.CreateTask"

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; exists {
		return fmt.Errorf("%s: %w", op, storage.ErrTaskAlreadyExists)
	}

	taskCopy := *task
	r.tasks[task.ID] = &taskCopy
	return nil
}

func (r *Storage) GetByID(id string) (*models.Task, error) {
	const op = "storage.memory.GetByID"
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
	}

	taskCopy := *task
	return &taskCopy, nil
}

func (r *Storage) UpdateTask(task *models.Task) error {
	const op = "storage.memory.UpdateTask"

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[task.ID]; !exists {
		return fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
	}

	taskCopy := *task
	r.tasks[task.ID] = &taskCopy
	return nil
}

func (r *Storage) DeleteTask(id string) error {
	const op = "storage.memory.DeleteTask"

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tasks[id]; !exists {
		return fmt.Errorf("%s: %w", op, storage.ErrTaskNotFound)
	}

	delete(r.tasks, id)
	return nil
}
