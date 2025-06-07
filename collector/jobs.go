package collector

import (
	"context"
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type jobsGetter interface {
	GetJobs(ctx context.Context, logger *slog.Logger) (*mirakurun.JobsResponse, error)
}

type jobsCollector struct {
	ctx    context.Context
	logger *slog.Logger

	jobsGetter jobsGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("jobs", defaultEnabled, newJobsCollector)
}

func newJobsCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "jobs"

	metricDefs := map[string]metricDefinition{
		"count": {
			name:       "count",
			help:       "Count of jobs",
			labelNames: []string{"status"},
			metricType: prometheus.GaugeValue,
		},
		"retry_count": {
			name:       "retry_count",
			help:       "Count of retried jobs",
			metricType: prometheus.GaugeValue,
		},
		"abort_count": {
			name:       "abort_count",
			help:       "Count of aborted jobs",
			metricType: prometheus.GaugeValue,
		},
		"skipped_count": {
			name:       "skipped_count",
			help:       "Count of skipped jobs",
			metricType: prometheus.GaugeValue,
		},
		"failed_count": {
			name:       "failed_count",
			help:       "Count of failed jobs",
			metricType: prometheus.GaugeValue,
		},
		"duration_avg": {
			name:       "duration_avg",
			help:       "Average duration of jobs",
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

	return &jobsCollector{
		ctx:         ctx,
		jobsGetter:  client,
		logger:      logger,
		metrics:     metrics,
		metricTypes: metricTypes,
	}
}

func (c *jobsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *jobsCollector) Collect(ch chan<- prometheus.Metric) error {
	jobs, err := c.jobsGetter.GetJobs(c.ctx, c.logger)
	if err != nil {
		return err
	}

	jobCount := make(map[string]int)
	var retryCount, abortCount, skippedCount, failedCount int
	var durationSum, finishedCount int
	for _, job := range *jobs {
		jobCount[job.Status]++
		retryCount += job.RetryCount
		if job.IsAborting {
			abortCount++
		}
		if job.HasSkipped {
			skippedCount++
		}
		if job.HasFailed {
			failedCount++
		}
		if job.Status == "finished" && !job.HasSkipped {
			durationSum += job.Duration
			finishedCount++
		}
	}

	for status, count := range jobCount {
		ch <- prometheus.MustNewConstMetric(
			c.metrics["count"],
			c.metricTypes["count"],
			float64(count),
			status,
		)
	}

	ch <- prometheus.MustNewConstMetric(
		c.metrics["retry_count"],
		c.metricTypes["retry_count"],
		float64(retryCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.metrics["abort_count"],
		c.metricTypes["abort_count"],
		float64(abortCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.metrics["skipped_count"],
		c.metricTypes["skipped_count"],
		float64(skippedCount),
	)

	ch <- prometheus.MustNewConstMetric(
		c.metrics["failed_count"],
		c.metricTypes["failed_count"],
		float64(failedCount),
	)

	var duration float64
	if finishedCount > 0 {
		duration = float64(durationSum) / float64(finishedCount)
	}

	ch <- prometheus.MustNewConstMetric(
		c.metrics["duration_avg"],
		c.metricTypes["duration_avg"],
		duration,
	)

	return nil
}
