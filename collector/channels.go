package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type channelsGetter interface {
	GetChannels(ctx context.Context, logger *slog.Logger) (*mirakurun.ChannelsResponse, error)
}

type channelsCollector struct {
	ctx    context.Context
	logger *slog.Logger

	channelsGetter channelsGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("channel", defaultEnabled, newChannelsCollector)
}

func newChannelsCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "channel"

	metricDefs := map[string]metricDefinition{
		"channel": {
			name:       "channel",
			help:       "Channel information",
			labelNames: []string{"name", "type", "channel"},
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

	return &channelsCollector{
		ctx:            ctx,
		channelsGetter: client,
		logger:         logger,
		metrics:        metrics,
		metricTypes:    metricTypes,
	}
}

func (c *channelsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *channelsCollector) Collect(ch chan<- prometheus.Metric) error {
	channels, err := c.channelsGetter.GetChannels(c.ctx, c.logger)
	if err != nil {
		return err
	}

	for _, channel := range *channels {
		ch <- prometheus.MustNewConstMetric(
			c.metrics["channel"],
			c.metricTypes["channel"],
			1,
			channel.Name,
			channel.Type,
			channel.Channel,
		)
	}
	return nil
}
