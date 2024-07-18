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
	"github.com/ATOR-Development/downloads-exporter/internal/counter"
	"github.com/ATOR-Development/downloads-exporter/internal/fetcher"
)

const (
	namespace    = "package_downloads"
	fetchTimeout = time.Second * 3
)

type Exporter struct {
	labels   []string
	fetchers []fetcher.Fetcher
	logger   log.Logger

	upDesc                    *prometheus.Desc
	countDesc                 *prometheus.Desc
	scrapeDurationSecondsDesc *prometheus.Desc
}

func FromConfig(cfg *config.Config, logger log.Logger) (*Exporter, error) {
	var fetchers []fetcher.Fetcher
	for _, c := range cfg.Fetchers.DockerHub {
		if len(c.Name) == 0 {
			return nil, errors.New("exporter: docker hub name cannot be empty")
		}

		if len(c.Owner) == 0 {
			return nil, errors.New("exporter: docker hub owner cannot be empty")
		}

		if len(c.Repo) == 0 {
			return nil, errors.New("exporter: docker hub repo cannot be empty")
		}

		fetchers = append(fetchers, fetcher.NewDockerhubPullsFetcher(
			c.Name,
			c.Owner,
			c.Repo,
		))
	}

	for _, c := range cfg.Fetchers.GithubReleases {
		if len(c.Name) == 0 {
			return nil, errors.New("exporter: github releases name cannot be empty")
		}

		if len(c.Owner) == 0 {
			return nil, errors.New("exporter: github releases owner cannot be empty")
		}

		if len(c.Repo) == 0 {
			return nil, errors.New("exporter: github releases repo cannot be empty")
		}

		var err error
		var assetsRegexp *regexp.Regexp
		if len(c.AssetsRegexp) > 0 {
			assetsRegexp, err = regexp.Compile(c.AssetsRegexp)
			if err != nil {
				return nil, err
			}
		}

		labels := make(map[string]*regexp.Regexp)
		for labelName, labelRegexStr := range c.Labels {
			labelRegex, err := regexp.Compile(labelRegexStr)
			if err != nil {
				return nil, err
			}

			labels[labelName] = labelRegex
		}

		fetchers = append(fetchers, fetcher.NewGithubReleasesFetcher(
			c.Name,
			c.Owner,
			c.Repo,
			assetsRegexp,
			labels,
			counter.New(cfg.Labels),
		))
	}

	for _, c := range cfg.Fetchers.NginxAccessLogs {
		if len(c.Name) == 0 {
			return nil, errors.New("exporter: nginx access log name cannot be empty")
		}

		if len(c.AccessLogPath) == 0 {
			return nil, errors.New("exporter: nginx access log path cannot be empty")
		}

		if len(c.AccessLogRegexp) == 0 {
			return nil, errors.New("exporter: nginx access log regexp cannot be empty")
		}

		accessLogRegexp, err := regexp.Compile(c.AccessLogRegexp)
		if err != nil {
			return nil, err
		}

		labels := make(map[string]*regexp.Regexp)
		for labelName, labelRegexStr := range c.Labels {
			labelRegex, err := regexp.Compile(labelRegexStr)
			if err != nil {
				return nil, err
			}

			labels[labelName] = labelRegex
		}

		fetchers = append(fetchers, fetcher.NewNginxAccessLogFetcher(
			c.Name,
			c.AccessLogPath,
			accessLogRegexp,
			labels,
			counter.New(cfg.Labels),
		))
	}

	if len(fetchers) == 0 {
		return nil, errors.New("exporter: no configs specified")
	}

	return &Exporter{
		labels:   cfg.Labels,
		fetchers: fetchers,
		logger:   logger,

		upDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Was the last downloads count fetch successful.",
			[]string{"name"}, nil,
		),
		countDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "count"),
			"How many times item was downloaded.",
			append([]string{"name"}, cfg.Labels...), nil,
		),
		scrapeDurationSecondsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
			"How many seconds it took to fetch downloads count.",
			[]string{"name"}, nil,
		),
	}, nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last downloads count fetch successful.",
		[]string{"name"}, nil,
	)
	ch <- prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "count"),
		"How many times item was downloaded.",
		append([]string{"name"}, e.labels...), nil,
	)
	ch <- prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
		"How many seconds it took to fetch downloads count.",
		[]string{"name"}, nil,
	)
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

			results, err := f.FetchCount(ctx)
			if err != nil {
				ch <- prometheus.MustNewConstMetric(
					e.upDesc, prometheus.GaugeValue, 0, f.Name(),
				)
				level.Error(e.logger).Log("msg", "failed to fetch metric", "name", f.Name(), "err", err.Error())
				return
			}

			scrapeDuration := time.Since(scrapeStart)
			scrapeDurationSeconds := float64(scrapeDuration.Milliseconds()) / 1000

			ch <- prometheus.MustNewConstMetric(e.upDesc, prometheus.GaugeValue, 1, f.Name())
			ch <- prometheus.MustNewConstMetric(e.scrapeDurationSecondsDesc, prometheus.GaugeValue, scrapeDurationSeconds, f.Name())

			for _, result := range results {
				var labelValues []string = []string{f.Name()}
				for _, labelName := range e.labels {
					if value, ok := result.Labels[labelName]; ok {
						labelValues = append(labelValues, value)
					} else {
						labelValues = append(labelValues, "unknown")
					}
				}

				level.Info(e.logger).Log("msg", "fetched new metric", "name", f.Name(), "key", result.Key, "count", result.Count, "scrape_duration", scrapeDuration)
				ch <- prometheus.MustNewConstMetric(e.countDesc, prometheus.CounterValue, float64(result.Count), labelValues...)
			}
		}(f)
	}

	wg.Wait()
}
