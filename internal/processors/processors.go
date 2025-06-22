package processors

import (
	"context"
	"time"

	"github.com/Noviiich/io-bound-task/internal/domain/models"
)

// LongRunningProcessor реализует TaskProcessor для долгих задач
type LongRunningProcessor struct{}

func New() *LongRunningProcessor {
	return &LongRunningProcessor{}
}

// Process выполняет длительную I/O операцию
func (p *LongRunningProcessor) Process(ctx context.Context, task *models.Task) error {
	// Симуляция долгой I/O операции (10-12 секунд)
	// duration := time.Duration(3+rand.Intn(3)) * time.Minute
	duration := time.Duration(1) * time.Minute

	select {
	case <-ctx.Done():
		time.Sleep(10 * time.Second)
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
