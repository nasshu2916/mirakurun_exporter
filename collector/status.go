package collector

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type statusCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	version             *prometheus.Desc
	process             *prometheus.Desc
	memoryUsage         *prometheus.Desc
	epgStoredEventCount *prometheus.Desc
	streamCount         *prometheus.Desc
	errorCount          *prometheus.Desc
	timerAccuracyM1     *prometheus.Desc
	timerAccuracyM5     *prometheus.Desc
	timerAccuracyM15    *prometheus.Desc
}

func init() {
	registerCollector("status", defaultEnabled, newStatusCollector)
}

func newStatusCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "status"

	return &statusCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "version"),
			"Version of Mirakurun",
			[]string{"mirakurun", "node"},
			nil,
		),
		process: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "process"),
			"Process information of Mirakurun",
			[]string{"arch", "platform"},
			nil,
		),
		memoryUsage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "memory_usage"),
			"Memory usage of Mirakurun",
			[]string{"type"},
			nil,
		),
		epgStoredEventCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "epg_stored_events"),
			"Count of stored EPG events",
			nil,
			nil,
		),
		streamCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "stream_count"),
			"Count of streams",
			[]string{"type"},
			nil,
		),
		errorCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "error_count"),
			"Count of errors",
			[]string{"type"},
			nil,
		),
		timerAccuracyM1: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "timer_accuracy_m1"),
			"Timer accuracy for 1 minute",
			[]string{"type"},
			nil,
		),
		timerAccuracyM5: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "timer_accuracy_m5"),
			"Timer accuracy for 5 minutes",
			[]string{"type"},
			nil,
		),
		timerAccuracyM15: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "timer_accuracy_m15"),
			"Timer accuracy for 15 minutes",
			[]string{"type"},
			nil,
		),
	}
}

func (c *statusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.version
	ch <- c.process
	ch <- c.memoryUsage
	ch <- c.epgStoredEventCount
	ch <- c.streamCount
	ch <- c.errorCount
	ch <- c.timerAccuracyM1
	ch <- c.timerAccuracyM5
	ch <- c.timerAccuracyM15
}

func (c *statusCollector) Collect(ch chan<- prometheus.Metric) error {
	status, err := c.client.GetStatus(c.ctx, c.logger)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.version,
		prometheus.GaugeValue,
		1,
		status.Version, status.Process.Versions["node"],
	)
	ch <- prometheus.MustNewConstMetric(
		c.process,
		prometheus.GaugeValue,
		1,
		status.Process.Arch, status.Process.Platform,
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsage,
		prometheus.GaugeValue,
		float64(status.Process.MemoryUsage.RSS),
		"RSS",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsage,
		prometheus.GaugeValue,
		float64(status.Process.MemoryUsage.HeapTotal),
		"HeapTotal",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsage,
		prometheus.GaugeValue,
		float64(status.Process.MemoryUsage.HeapUsed),
		"HeapUsed",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsage,
		prometheus.GaugeValue,
		float64(status.Process.MemoryUsage.External),
		"External",
	)
	ch <- prometheus.MustNewConstMetric(
		c.memoryUsage,
		prometheus.GaugeValue,
		float64(status.Process.MemoryUsage.ArrayBuffers),
		"ArrayBuffers",
	)
	ch <- prometheus.MustNewConstMetric(
		c.epgStoredEventCount,
		prometheus.GaugeValue,
		float64(status.EPG.StoredEvents),
	)
	ch <- prometheus.MustNewConstMetric(
		c.streamCount,
		prometheus.GaugeValue,
		float64(status.StreamCount.TunerDevice),
		"TunerDevice",
	)
	ch <- prometheus.MustNewConstMetric(
		c.streamCount,
		prometheus.GaugeValue,
		float64(status.StreamCount.TSFilter),
		"TSFilter",
	)
	ch <- prometheus.MustNewConstMetric(
		c.streamCount,
		prometheus.GaugeValue,
		float64(status.StreamCount.Decoder),
		"Decoder",
	)
	ch <- prometheus.MustNewConstMetric(
		c.errorCount,
		prometheus.CounterValue,
		float64(status.ErrorCount.UncaughtException),
		"UncaughtException",
	)
	ch <- prometheus.MustNewConstMetric(
		c.errorCount,
		prometheus.CounterValue,
		float64(status.ErrorCount.UnhandledRejection),
		"UnhandledRejection",
	)
	ch <- prometheus.MustNewConstMetric(
		c.errorCount,
		prometheus.CounterValue,
		float64(status.ErrorCount.BufferOverflow),
		"BufferOverflow",
	)
	ch <- prometheus.MustNewConstMetric(
		c.errorCount,
		prometheus.CounterValue,
		float64(status.ErrorCount.TunerDeviceRespawn),
		"TunerDeviceRespawn",
	)
	ch <- prometheus.MustNewConstMetric(
		c.errorCount,
		prometheus.CounterValue,
		float64(status.ErrorCount.DecoderRespawn),
		"DecoderRespawn",
	)

	for _, field := range []string{"avg", "min", "max"} {
		ch <- prometheus.MustNewConstMetric(
			c.timerAccuracyM1,
			prometheus.GaugeValue,
			status.TimerAccuracy.M1.GetValue(field),
			field,
		)
	}

	for _, field := range []string{"avg", "min", "max"} {
		ch <- prometheus.MustNewConstMetric(
			c.timerAccuracyM5,
			prometheus.GaugeValue,
			status.TimerAccuracy.M5.GetValue(field),
			field,
		)
	}

	for _, field := range []string{"avg", "min", "max"} {
		ch <- prometheus.MustNewConstMetric(
			c.timerAccuracyM15,
			prometheus.GaugeValue,
			status.TimerAccuracy.M15.GetValue(field),
			field,
		)
	}

	return nil
}
