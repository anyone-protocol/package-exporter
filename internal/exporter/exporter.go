package exporter

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ATOR-Development/downloads-exporter/internal/config"
	"github.com/ATOR-Development/downloads-exporter/internal/fetcher"
)

const (
	namespace    = "package_downloads"
	fetchTimeout = time.Second * 3
)

var (
	upDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last downloads count fetch successful.",
		[]string{"name"}, nil,
	)
	countDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "count"),
		"How many times item was downloaded.",
		[]string{"name"}, nil,
	)
	scrapeDurationSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
		"How many seconds it took to fetch downloads count.",
		[]string{"name"}, nil,
	)
)

type Exporter struct {
	fetchers []fetcher.Fetcher
	logger   log.Logger
}

func FromConfig(cfg *config.Config, logger log.Logger) (*Exporter, error) {
	var fetchers []fetcher.Fetcher
	for _, cfg := range cfg.Fetchers.DockerHub {
		if len(cfg.Name) == 0 {
			return nil, errors.New("exporter: docker hub name cannot be empty")
		}

		if len(cfg.Owner) == 0 {
			return nil, errors.New("exporter: docker hub owner cannot be empty")
		}

		if len(cfg.Repo) == 0 {
			return nil, errors.New("exporter: docker hub repo cannot be empty")
		}

		fetchers = append(fetchers, fetcher.NewDockerhubPullsFetcher(
			cfg.Name,
			cfg.Owner,
			cfg.Repo,
		))
	}

	for _, cfg := range cfg.Fetchers.GithubReleases {
		if len(cfg.Name) == 0 {
			return nil, errors.New("exporter: github releases name cannot be empty")
		}

		if len(cfg.Owner) == 0 {
			return nil, errors.New("exporter: github releases owner cannot be empty")
		}

		if len(cfg.Repo) == 0 {
			return nil, errors.New("exporter: github releases repo cannot be empty")
		}

		var err error
		var assetsRegexp *regexp.Regexp
		if len(cfg.AssetsRegexp) > 0 {
			assetsRegexp, err = regexp.Compile(cfg.AssetsRegexp)
			if err != nil {
				return nil, err
			}
		}

		fetchers = append(fetchers, fetcher.NewGithubReleasesFetcher(
			cfg.Name,
			cfg.Owner,
			cfg.Repo,
			assetsRegexp,
		))
	}

	for _, cfg := range cfg.Fetchers.NginxAccessLogs {
		if len(cfg.Name) == 0 {
			return nil, errors.New("exporter: nginx access log name cannot be empty")
		}

		if len(cfg.AccessLogPath) == 0 {
			return nil, errors.New("exporter: nginx access log path cannot be empty")
		}

		if len(cfg.AccessLogRegexp) == 0 {
			return nil, errors.New("exporter: nginx access log regexp cannot be empty")
		}

		accessLogRegexp, err := regexp.Compile(cfg.AccessLogRegexp)
		if err != nil {
			return nil, err
		}

		fetchers = append(fetchers, fetcher.NewNginxAccessLogFetcher(
			cfg.Name,
			cfg.AccessLogPath,
			accessLogRegexp,
		))
	}

	if len(fetchers) == 0 {
		return nil, errors.New("exporter: no configs specified")
	}

	return &Exporter{fetchers: fetchers, logger: logger}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- countDesc
	ch <- scrapeDurationSecondsDesc
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	scrapeStart := time.Now()

	var wg sync.WaitGroup
	for _, f := range e.fetchers {
		wg.Add(1)

		go func(f fetcher.Fetcher) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
			defer cancel()

			count, err := f.FetchCount(ctx)
			if err != nil {
				ch <- prometheus.MustNewConstMetric(
					upDesc, prometheus.GaugeValue, 0, f.Name(),
				)
				level.Error(e.logger).Log("msg", "failed to fetch metric", "name", f.Name(), "err", err.Error())
				return
			}

			scrapeDuration := time.Since(scrapeStart)
			scrapeDurationSeconds := float64(scrapeDuration.Milliseconds()) / 1000

			level.Info(e.logger).Log("msg", "fetched new metric", "name", f.Name(), "count", count, "scrape_duration", scrapeDuration)

			ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1, f.Name())
			ch <- prometheus.MustNewConstMetric(countDesc, prometheus.CounterValue, float64(count), f.Name())
			ch <- prometheus.MustNewConstMetric(scrapeDurationSecondsDesc, prometheus.GaugeValue, scrapeDurationSeconds, f.Name())
		}(f)
	}

	wg.Wait()
}
