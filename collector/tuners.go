package collector

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type tunersGetter interface {
	GetTuners(ctx context.Context, logger *slog.Logger) (*mirakurun.TunersResponse, error)
}

type tunerCollector struct {
	ctx    context.Context
	logger *slog.Logger

	tunersGetter tunersGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("tuners", defaultEnabled, newTunerCollector)
}

func newTunerCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "tuners"

	metricDefs := map[string]metricDefinition{
		"device": {
			name:       "device",
			help:       "Tuner device information",
			labelNames: []string{"index", "name", "type"},
			metricType: prometheus.GaugeValue,
		},
		"available_tuner": {
			name:       "available_tuner",
			help:       "Available tuner device",
			labelNames: []string{"index"},
			metricType: prometheus.GaugeValue,
		},
		"remote_tuner": {
			name:       "remote_tuner",
			help:       "Remote tuner device",
			labelNames: []string{"index"},
			metricType: prometheus.GaugeValue,
		},
		"free_tuner": {
			name:       "free_tuner",
			help:       "Tuner device is free",
			labelNames: []string{"index"},
			metricType: prometheus.GaugeValue,
		},
		"using_tuner": {
			name:       "using_tuner",
			help:       "Tuner device is using",
			labelNames: []string{"index"},
			metricType: prometheus.GaugeValue,
		},
		"fault_tuner": {
			name:       "fault_tuner",
			help:       "Tuner device is fault",
			labelNames: []string{"index"},
			metricType: prometheus.GaugeValue,
		},
		"users": {
			name:       "users",
			help:       "User using tuner device",
			labelNames: []string{"index", "user_id", "agent"},
			metricType: prometheus.GaugeValue,
		},
		"stream_packets": {
			name:       "stream_packets",
			help:       "Stream packets by user",
			labelNames: []string{"user_id"},
			metricType: prometheus.CounterValue,
		},
		"stream_drops": {
			name:       "stream_drops",
			help:       "Stream drops packets by user",
			labelNames: []string{"user_id"},
			metricType: prometheus.CounterValue,
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

	return &tunerCollector{
		ctx:          ctx,
		tunersGetter: client,
		logger:       logger,
		metrics:      metrics,
		metricTypes:  metricTypes,
	}
}

func (c *tunerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *tunerCollector) Collect(ch chan<- prometheus.Metric) error {
	tuners, err := c.tunersGetter.GetTuners(c.ctx, c.logger)
	if err != nil {
		return err
	}

	for _, tuner := range *tuners {
		index := strconv.Itoa(tuner.Index)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["device"],
			c.metricTypes["device"],
			1,
			index, tuner.Name, strings.Join(tuner.Types, ","),
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["available_tuner"],
			c.metricTypes["available_tuner"],
			boolToFloat64(tuner.IsAvailable),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["remote_tuner"],
			c.metricTypes["remote_tuner"],
			boolToFloat64(tuner.IsRemote),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["free_tuner"],
			c.metricTypes["free_tuner"],
			boolToFloat64(tuner.IsFree),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["using_tuner"],
			c.metricTypes["using_tuner"],
			boolToFloat64(tuner.IsUsing),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["fault_tuner"],
			c.metricTypes["fault_tuner"],
			boolToFloat64(tuner.IsFault),
			index,
		)
		for _, user := range tuner.Users {
			ch <- prometheus.MustNewConstMetric(
				c.metrics["users"],
				c.metricTypes["users"],
				1,
				index, user.ID, user.Agent,
			)

			var streamPackets, streamDrops int64
			if user.StreamInfo != nil {
				for _, stream := range user.StreamInfo {
					streamPackets += stream.Packet
					streamDrops += stream.Drop
				}
			} else {
				c.logger.Warn("StreamInfo is nil", "user_id", user.ID)
			}

			ch <- prometheus.MustNewConstMetric(
				c.metrics["stream_packets"],
				c.metricTypes["stream_packets"],
				float64(streamPackets),
				user.ID,
			)
			ch <- prometheus.MustNewConstMetric(
				c.metrics["stream_drops"],
				c.metricTypes["stream_drops"],
				float64(streamDrops),
				user.ID,
			)
		}
	}
	return nil
}
