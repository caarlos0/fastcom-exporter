package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/fastcom-exporter/collector"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// nolint: gochecknoglobals
var (
	bind     = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9877").String()
	debug    = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	interval = kingpin.Flag("refresh.interval", "time between refreshes with fast.com").Default("15m").Duration()
	version  = "master"
)

func main() {
	kingpin.Version("fastcom-exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := log.NewLogfmtLogger(os.Stderr)
	if *debug {
		logger = level.NewFilter(logger, level.AllowDebug())
	} else {
		logger = level.NewFilter(logger, level.AllowInfo())
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	level.Info(logger).Log("msg", "starting fastcom-exporter", "version", version)

	prometheus.MustRegister(collector.NewFastCollector(logger, cache.New(*interval, *interval)))
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w, `
			<html>
			<head><title>Fast.com Exporter</title></head>
			<body>
				<h1>Fast.com Exporter</h1>
				<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>
			`,
		)
	})
	level.Info(logger).Log("msg", "listening on "+*bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		level.Error(logger).Log("msg", "error listening", "addr", *bind, "err", err)
	}
}
