package collector

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type programsGetter interface {
	GetPrograms(ctx context.Context, logger *slog.Logger) (*mirakurun.ProgramsResponse, error)
}

type programsCollector struct {
	ctx    context.Context
	logger *slog.Logger

	programsGetter programsGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("programs", defaultEnabled, newProgramsCollector)
}

func newProgramsCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "programs"

	metricDefs := map[string]metricDefinition{
		"count": {
			name:       "count",
			help:       "Count of programs by service",
			labelNames: []string{"service_id"},
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

	return &programsCollector{
		ctx:            ctx,
		programsGetter: client,
		logger:         logger,
		metrics:        metrics,
		metricTypes:    metricTypes,
	}
}

func (c *programsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *programsCollector) Collect(ch chan<- prometheus.Metric) error {
	programs, err := c.programsGetter.GetPrograms(c.ctx, c.logger)
	if err != nil {
		return err
	}

	programCount := make(map[int]int)
	for _, program := range *programs {
		programCount[program.ServiceID]++
	}

	for serviceID, count := range programCount {
		ch <- prometheus.MustNewConstMetric(
			c.metrics["count"],
			c.metricTypes["count"],
			float64(count),
			strconv.Itoa(serviceID),
		)
	}

	return nil
}
