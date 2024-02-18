package fetcher

import (
	"context"
	"regexp"
)

// NewNginxAccessLogFetcher creates new Nginx access log downloads fetcher from config
func NewNginxAccessLogFetcher(name, accessLogPath string, accessLogRegexp *regexp.Regexp) Fetcher {
	return &nginxAccessLogFetcher{
		name:            name,
		accessLogPath:   accessLogPath,
		accessLogRegexp: accessLogRegexp,
	}
}

type nginxAccessLogFetcher struct {
	name            string
	accessLogPath   string
	accessLogRegexp *regexp.Regexp
}

// FetchCount fetches download count from nginx access logs and returns it
func (f *nginxAccessLogFetcher) FetchCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (f *nginxAccessLogFetcher) Name() string {
	return f.name
}
