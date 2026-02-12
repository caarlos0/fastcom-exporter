package main

import (
	"fmt"
	"net/http"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/fastcom-exporter/collector"
	"github.com/charmbracelet/log"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// nolint: gochecknoglobals
var (
	bind     = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9877").String()
	debug    = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	format   = kingpin.Flag("logFormat", "log format to use").Default("console").Enum("json", "console")
	interval = kingpin.Flag("refresh.interval", "time between refreshes with fast.com").Default("30m").Duration()
	version  = "master"
)

func main() {
	kingpin.Version("fastcom-exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.SetLevel(log.InfoLevel)
	if *format == "json" {
		log.SetFormatter(log.JSONFormatter)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("enabled debug mode")
	}
	log.Infof("starting fastcom-exporter %s", version)

	prometheus.MustRegister(collector.NewFastCollector(cache.New(*interval, *interval)))
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
	log.Infof("listening on %s", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatal("error starting server", "err", err)
	}
}
