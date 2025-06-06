package collector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

const (
	namespace       = "mirakurun"
	defaultEnabled  = true
	defaultDisabled = false
)

type CollectorFactory func(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector

type metricDefinition struct {
	name       string
	help       string
	labelNames []string
	metricType prometheus.ValueType
}

var (
	factories        = make(map[string]CollectorFactory)
	collectorState   = make(map[string]*bool)
	forcedCollectors = make(map[string]bool)

	enableScrapeCollector = kingpin.Flag("collector.scrape", "Enable the scrape collector (default: true).").Default("true").Bool()
)

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"mirakurun_exporter: Duration of a collector scrape",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"mirakurun_exporter: Whether a collector succeeded",
		[]string{"collector"},
		nil,
	)
)

type Collector interface {
	Describe(ch chan<- *prometheus.Desc)
	Collect(ch chan<- prometheus.Metric) error
}

type MirakurunCollector struct {
	Collectors map[string]Collector
	logger     *slog.Logger
}

func registerCollector(collector string, isDefaultEnabled bool, factory CollectorFactory) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(collectorFlagAction(collector)).Bool()

	collectorState[collector] = flag
	factories[collector] = factory
}

func DisableDefaultCollectors() {
	for c := range collectorState {
		if _, ok := forcedCollectors[c]; !ok {
			*collectorState[c] = false
		}
	}
}

func collectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

func MetricsHandler(client *mirakurun.Client, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("metrics request", "url", r.URL.String())
		registry := prometheus.NewRegistry()
		mirakurunCollector, err := NewMirakurunCollector(r.Context(), client, logger)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create collector: %s", err), http.StatusInternalServerError)
			return
		}
		registry.MustRegister(mirakurunCollector)

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorLog:      slog.NewLogLogger(logger.Handler(), slog.LevelError),
			ErrorHandling: promhttp.ContinueOnError,
		})
		h.ServeHTTP(w, r)
	}
}

func NewMirakurunCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) (*MirakurunCollector, error) {
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if !*enabled {
			continue
		}
		collectors[key] = factories[key](ctx, client, logger)
	}
	return &MirakurunCollector{Collectors: collectors, logger: logger}, nil
}

func (mirakurunCollector *MirakurunCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, collector := range mirakurunCollector.Collectors {
		collector.Describe(ch)
	}
}

func (mirakurunCollector *MirakurunCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(mirakurunCollector.Collectors))
	for name, c := range mirakurunCollector.Collectors {
		go func(name string, c Collector) {
			executeCollect(name, c, ch, mirakurunCollector.logger)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func executeCollect(name string, c Collector, ch chan<- prometheus.Metric, logger *slog.Logger) {
	begin := time.Now()
	err := c.Collect(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		logger.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		success = 0
	} else {
		logger.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	if *enableScrapeCollector {
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
	}
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
