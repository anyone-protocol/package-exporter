package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/anyone-protocol/package-exporter/internal/config"
	"github.com/anyone-protocol/package-exporter/internal/exporter"
)

var (
	configFile    = flag.String("config", "config.yml", "Config file.")
	listenAddress = flag.String("listen-address", ":8080", "Exporter HTTP listen address.")
	metricsPath   = flag.String("metrics-path", "/metrics", "URL path for metrics endpoint.")
)

func main() {
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.WithPrefix(logger, "ts", log.TimestampFormat(time.Now, time.Stamp))

	cfg, err := config.FromFile(*configFile)
	if err != nil {
		level.Error(logger).Log("msg", "cannot read config", "err", err.Error())
		os.Exit(1)
	}

	exp, err := exporter.FromConfig(cfg, logger)
	if err != nil {
		level.Error(logger).Log("msg", "cannot create exporter", "err", err.Error())
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(exp)

	html := strings.Join([]string{
		`<html>`,
		`  <head>`,
		`    <title>Downloads Exporter</title>`,
		`  </head>`,
		`  <body>`,
		`    <h1>Downloads Exporter</h1>`,
		`    <p>`,
		`      <a href="` + *metricsPath + `">` + *metricsPath + `</a>`,
		`    </p>`,
		`  </body>`,
		`</html>`,
	}, "\n")

	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	})

	http.ListenAndServe(*listenAddress, nil)
}
