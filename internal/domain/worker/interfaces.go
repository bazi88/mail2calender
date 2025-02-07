package worker

import (
	"context"
	"time"

	"mail2calendar/internal/domain/email_auth"
)

// EmailProcessor handles email processing tasks
type EmailProcessor interface {
	// ProcessEmails fetches and processes new emails
	ProcessEmails(ctx context.Context, token *email_auth.EmailToken) error

	// SyncCalendar synchronizes processed events with calendar
	SyncCalendar(ctx context.Context, token *email_auth.EmailToken) error
}

// WorkerConfig contains configuration for the worker
type WorkerConfig struct {
	Concurrency      int
	QueueSize        int
	RetryAttempts    int
	RetryDelay       time.Duration
	FetchInterval    time.Duration
	ProcessBatchSize int
}

// Worker represents the background worker service
type Worker interface {
	// Start starts the worker service
	Start(ctx context.Context) error

	// Stop gracefully stops the worker service
	Stop(ctx context.Context) error

	// AddTask adds a new task to the worker queue
	AddTask(ctx context.Context, task interface{}) error

	// GetStats returns current worker statistics
	GetStats(ctx context.Context) (WorkerStats, error)
}

// WorkerStats contains worker service statistics
type WorkerStats struct {
	ActiveWorkers   int
	QueuedTasks     int
	ProcessedTasks  int64
	FailedTasks     int64
	AverageLatency  time.Duration
	LastProcessedAt time.Time
}
