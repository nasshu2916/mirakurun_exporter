package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type versionGetter interface {
	GetVersion(ctx context.Context, logger *slog.Logger) (*mirakurun.VersionResponse, error)
}

type versionCollector struct {
	ctx    context.Context
	logger *slog.Logger

	versionGetter versionGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("version", defaultDisabled, newVersionCollector)
}

func newVersionCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "version"

	metricDefs := map[string]metricDefinition{
		"mirakurun_version": {
			name:       "mirakurun_version",
			help:       "Mirakurun version",
			labelNames: []string{"current", "latest"},
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

	return &versionCollector{
		ctx:           ctx,
		versionGetter: client,
		logger:        logger,
		metrics:       metrics,
		metricTypes:   metricTypes,
	}
}

func (c *versionCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *versionCollector) Collect(ch chan<- prometheus.Metric) error {
	version, err := c.versionGetter.GetVersion(c.ctx, c.logger)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		c.metrics["mirakurun_version"],
		c.metricTypes["mirakurun_version"],
		1,
		[]string{version.Current, version.Latest}...,
	)
	return nil
}
