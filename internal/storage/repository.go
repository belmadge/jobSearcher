package storage

import (
	"context"
	"time"

	"jobsearcher/internal/domain"
)

type SaveResult struct { New bool; Updated bool; Duplicate bool }
type Stats struct {
	Total, New, Seen, Updated, Expired, Rejected, Recommended, Possible int
	Sources map[string]int
}

type JobRepository interface {
	SaveJob(context.Context, *domain.Job) (SaveResult, error)
	GetJob(context.Context, string) (domain.Job, error)
	GetJobs(context.Context) ([]domain.Job, error)
	MarkSeen(context.Context, string, string) error
	FindByCanonicalURL(context.Context, string) (domain.Job, error)
	FindBySourceID(context.Context, string, string) (domain.Job, error)
	MarkUnseenExpired(context.Context, string, time.Time) (int64, error)
	GetStats(context.Context) (Stats, error)
	Close() error
}
