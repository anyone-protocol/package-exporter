package fetcher

import (
	"context"

	"github.com/ATOR-Development/downloads-exporter/internal/counter"
)

type Fetcher interface {
	// FetchCount fetches and returns downloads count
	FetchCount(ctx context.Context) ([]*counter.Result, error)

	// Name returns fetcher name
	Name() string
}
