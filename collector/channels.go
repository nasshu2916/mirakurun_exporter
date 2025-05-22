package collector

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type ChannelsCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	channel *prometheus.Desc
}

func init() {
	registerCollector("channel", defaultEnabled, newChannelsCollector)
}

func newChannelsCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "channel"

	return &ChannelsCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		channel: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "mirakurun_channels"),
			"Channel information",
			[]string{"name", "type", "channel"},
			nil,
		),
	}
}

func (c *ChannelsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.channel
}

func (c *ChannelsCollector) Collect(ch chan<- prometheus.Metric) error {
	channels, err := c.client.GetChannels(c.ctx, c.logger)
	if err != nil {
		return err
	}

	for _, channel := range *channels {
		ch <- prometheus.MustNewConstMetric(
			c.channel,
			prometheus.GaugeValue,
			1,
			channel.Name,
			channel.Type,
			channel.Channel,
		)
	}
	return nil
}
