package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type statusGetter interface {
	GetStatus(ctx context.Context, logger *slog.Logger) (*mirakurun.StatusResponse, error)
}

type statusCollector struct {
	ctx    context.Context
	logger *slog.Logger

	statusGetter statusGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("status", defaultEnabled, newStatusCollector)
}

func newStatusCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "status"

	metricDefs := map[string]metricDefinition{
		"version": {
			name:       "version",
			help:       "Version of Mirakurun",
			labelNames: []string{"mirakurun", "node"},
			metricType: prometheus.GaugeValue,
		},
		"process": {
			name:       "process",
			help:       "Process information of Mirakurun",
			labelNames: []string{"arch", "platform"},
			metricType: prometheus.GaugeValue,
		},
		"memory_usage": {
			name:       "memory_usage",
			help:       "Memory usage of Mirakurun",
			labelNames: []string{"type"},
			metricType: prometheus.GaugeValue,
		},
		"epg_stored_events": {
			name:       "epg_stored_events",
			help:       "Count of stored EPG events",
			metricType: prometheus.GaugeValue,
		},
		"stream_count": {
			name:       "stream_count",
			help:       "Count of streams",
			labelNames: []string{"type"},
			metricType: prometheus.GaugeValue,
		},
		"error_count": {
			name:       "error_count",
			help:       "Count of errors",
			labelNames: []string{"type"},
			metricType: prometheus.CounterValue,
		},
		"timer_accuracy_m1": {
			name:       "timer_accuracy_m1",
			help:       "Timer accuracy for 1 minute",
			labelNames: []string{"type"},
			metricType: prometheus.GaugeValue,
		},
		"timer_accuracy_m5": {
			name:       "timer_accuracy_m5",
			help:       "Timer accuracy for 5 minutes",
			labelNames: []string{"type"},
			metricType: prometheus.GaugeValue,
		},
		"timer_accuracy_m15": {
			name:       "timer_accuracy_m15",
			help:       "Timer accuracy for 15 minutes",
			labelNames: []string{"type"},
			metricType: prometheus.GaugeValue,
		},
	}

	metrics := make(map[string]*prometheus.Desc)
	metricTypes := make(map[string]prometheus.ValueType)
	for name, def := range metricDefs {
		metrics[name] = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, def.name),
			def.help,
			def.labelNames,
			nil,
		)
		metricTypes[name] = def.metricType
	}

	return &statusCollector{
		ctx:          ctx,
		statusGetter: client,
		logger:       logger,
		metrics:      metrics,
		metricTypes:  metricTypes,
	}
}

func (c *statusCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *statusCollector) Collect(ch chan<- prometheus.Metric) error {
	status, err := c.statusGetter.GetStatus(c.ctx, c.logger)
	if err != nil {
		return err
	}

	// Version metrics
	ch <- prometheus.MustNewConstMetric(
		c.metrics["version"],
		c.metricTypes["version"],
		1,
		status.Version, status.Process.Versions["node"],
	)

	// Process metrics
	ch <- prometheus.MustNewConstMetric(
		c.metrics["process"],
		c.metricTypes["process"],
		1,
		status.Process.Arch, status.Process.Platform,
	)

	// Memory usage metrics
	memoryTypes := map[string]float64{
		"RSS":          float64(status.Process.MemoryUsage.RSS),
		"HeapTotal":    float64(status.Process.MemoryUsage.HeapTotal),
		"HeapUsed":     float64(status.Process.MemoryUsage.HeapUsed),
		"External":     float64(status.Process.MemoryUsage.External),
		"ArrayBuffers": float64(status.Process.MemoryUsage.ArrayBuffers),
	}
	for memType, value := range memoryTypes {
		ch <- prometheus.MustNewConstMetric(
			c.metrics["memory_usage"],
			c.metricTypes["memory_usage"],
			value,
			memType,
		)
	}

	// EPG metrics
	ch <- prometheus.MustNewConstMetric(
		c.metrics["epg_stored_events"],
		c.metricTypes["epg_stored_events"],
		float64(status.EPG.StoredEvents),
	)

	// Stream count metrics
	streamTypes := map[string]float64{
		"TunerDevice": float64(status.StreamCount.TunerDevice),
		"TSFilter":    float64(status.StreamCount.TSFilter),
		"Decoder":     float64(status.StreamCount.Decoder),
	}
	for streamType, value := range streamTypes {
		ch <- prometheus.MustNewConstMetric(
			c.metrics["stream_count"],
			c.metricTypes["stream_count"],
			value,
			streamType,
		)
	}

	// Error count metrics
	errorTypes := map[string]float64{
		"UncaughtException":  float64(status.ErrorCount.UncaughtException),
		"UnhandledRejection": float64(status.ErrorCount.UnhandledRejection),
		"BufferOverflow":     float64(status.ErrorCount.BufferOverflow),
		"TunerDeviceRespawn": float64(status.ErrorCount.TunerDeviceRespawn),
		"DecoderRespawn":     float64(status.ErrorCount.DecoderRespawn),
	}
	for errorType, value := range errorTypes {
		ch <- prometheus.MustNewConstMetric(
			c.metrics["error_count"],
			c.metricTypes["error_count"],
			value,
			errorType,
		)
	}

	// Timer accuracy metrics
	timerFields := []string{"avg", "min", "max"}
	timerPeriods := map[string]struct {
		metric string
		value  func(string) float64
	}{
		"M1":  {metric: "timer_accuracy_m1", value: status.TimerAccuracy.M1.GetValue},
		"M5":  {metric: "timer_accuracy_m5", value: status.TimerAccuracy.M5.GetValue},
		"M15": {metric: "timer_accuracy_m15", value: status.TimerAccuracy.M15.GetValue},
	}

	for _, data := range timerPeriods {
		for _, field := range timerFields {
			ch <- prometheus.MustNewConstMetric(
				c.metrics[data.metric],
				c.metricTypes[data.metric],
				data.value(field),
				field,
			)
		}
	}

	return nil
}
