package task

import "asum/pkg/models"

type QueryOptions struct {
	WithUsers bool
	Status    *models.TaskStatus
	OrderBy   string
	Limit     int
	Offset    int
}

func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		WithUsers: false,
		OrderBy:   "created_at DESC",
		Limit:     20,
		Offset:    0,
	}
}
