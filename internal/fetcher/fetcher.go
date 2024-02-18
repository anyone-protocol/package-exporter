package fetcher

import (
	"context"
)

type Fetcher interface {
	// FetchCount fetches and returns downloads count
	FetchCount(ctx context.Context) (int, error)

	// Name returns fetcher name
	Name() string
}
