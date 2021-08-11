package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/caarlos0/fastcom-exporter/collector"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// nolint: gochecknoglobals
var (
	bind     = kingpin.Flag("bind", "addr to bind the server").Short('b').Default(":9877").String()
	debug    = kingpin.Flag("debug", "show debug logs").Default("false").Bool()
	format   = kingpin.Flag("logFormat", "log format to use").Default("console").Enum("json", "console")
	interval = kingpin.Flag("refresh.interval", "time between refreshes with fast.com").Default("15m").Duration()
	version  = "master"
)

func main() {
	kingpin.Version("fastcom-exporter version " + version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("enabled debug mode")
	}
	log.Info().Msgf("starting fastcom-exporter %s", version)

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
	log.Info().Msgf("listening on %s", *bind)
	if err := http.ListenAndServe(*bind, nil); err != nil {
		log.Fatal().Err(err).Msg("error starting server")
	}
}
