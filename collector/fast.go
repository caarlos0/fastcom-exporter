package collector

import (
	"sync"
	"time"

	"github.com/caarlos0/fastcom-exporter/fast"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type fastCollector struct {
	mutex sync.Mutex
	cache *cache.Cache

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
	downloadBytes  *prometheus.Desc
}

// NewFastCollector returns a fast.com collector
func NewFastCollector(cache *cache.Cache) prometheus.Collector {
	const namespace = "fastcom"
	return &fastCollector{
		cache: cache,
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Exporter is up",
			nil,
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
			"Returns how long the probe took to complete in seconds",
			nil,
			nil,
		),
		downloadBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "download", "bytes_second"),
			"Download speed in B/s",
			nil,
			nil,
		),
	}
}

// Describe all metrics
func (c *fastCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.downloadBytes
}

// Collect all metrics
func (c *fastCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	success := 1
	defer func() {
		ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, float64(success))
	}()

	result, err := c.cachedOrCollect()
	if err != nil {
		success = 0
		log.Error().Err(err).Msg("fast.com collector failed")
	}

	ch <- prometheus.MustNewConstMetric(c.downloadBytes, prometheus.GaugeValue, result)
}

func (c *fastCollector) cachedOrCollect() (float64, error) {
	cold, ok := c.cache.Get("result")
	if ok {
		log.Debug().Msg("returning results from cache")
		return cold.(float64), nil
	}

	hot, err := c.collect()
	if err != nil {
		return hot, err
	}
	log.Debug().Msg("returning results from api")
	c.cache.Set("result", hot, cache.DefaultExpiration)
	return hot, nil
}

func (c *fastCollector) collect() (float64, error) {
	log.Debug().Msg("collecting fast.com metrics")
	return fast.Measure()
}
