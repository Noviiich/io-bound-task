package models

import (
	"time"

	resp "github.com/Noviiich/io-bound-task/internal/lib/api/response"
)

type Task struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      resp.TaskStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Duration    time.Duration   `json:"duration,omitempty"`
}

type TaskResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Status      resp.TaskStatus `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Duration    string          `json:"duration,omitempty"`
}

func (t *Task) IsActive() bool {
	return t.Status == resp.StatusPending || t.Status == resp.StatusRunning
}

func (t *Task) IsCompleted() bool {
	return t.Status == resp.StatusCompleted || t.Status == resp.StatusFailed || t.Status == resp.StatusCancelled
}

func (t *Task) ToResponse() *TaskResponse {
	taskResp := &TaskResponse{
		ID:          t.ID,
		Name:        t.Name,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		StartedAt:   t.StartedAt,
		CompletedAt: t.CompletedAt,
	}
	if t.Duration != 0 {
		taskResp.Duration = t.Duration.String()
	}

	return taskResp
}
