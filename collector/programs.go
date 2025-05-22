package collector

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"strconv"
)

type programsCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	programCount *prometheus.Desc
}

func init() {
	registerCollector("programs", defaultEnabled, newProgramsCollector)
}

func newProgramsCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "programs"

	return &programsCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		programCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "count"),
			"Count of programs by service",
			[]string{"service_id"},
			nil,
		),
	}
}

func (c *programsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.programCount
}

func (c *programsCollector) Collect(ch chan<- prometheus.Metric) error {
	programs, err := c.client.GetPrograms(c.ctx, c.logger)
	if err != nil {
		return err
	}

	programCount := make(map[int]int)
	for _, program := range *programs {
		programCount[program.ServiceID]++
	}

	for serviceID, count := range programCount {
		ch <- prometheus.MustNewConstMetric(
			c.programCount,
			prometheus.GaugeValue,
			float64(count),
			strconv.Itoa(serviceID),
		)
	}

	return nil
}
