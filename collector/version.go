package collector

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type versionCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	version *prometheus.Desc
}

func init() {
	registerCollector("version", defaultDisabled, newVersionCollector)
}

func newVersionCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "version"

	return &versionCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		version: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "mirakurun_version"),
			"Mirakurun version",
			[]string{"current", "latest"},
			nil,
		),
	}
}

func (c *versionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.version
}

func (c *versionCollector) Collect(ch chan<- prometheus.Metric) error {
	version, err := c.client.GetVersion(c.ctx, c.logger)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.version,
		prometheus.GaugeValue,
		1,
		[]string{version.Current, version.Latest}...,
	)
	return nil
}
