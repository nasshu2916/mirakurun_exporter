package collector

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type jobsCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	jobCount        *prometheus.Desc
	retryJobCount   *prometheus.Desc
	abortJobCount   *prometheus.Desc
	skippedJobCount *prometheus.Desc
	failedJobCount  *prometheus.Desc
	durationAvg     *prometheus.Desc
}

func init() {
	registerCollector("jobs", defaultEnabled, newJobsCollector)
}

func newJobsCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "jobs"

	return &jobsCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		jobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "count"),
			"Count of jobs",
			[]string{"status"},
			nil,
		),
		retryJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "retry_count"),
			"Count of retried jobs",
			nil,
			nil,
		),
		abortJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "abort_count"),
			"Count of aborted jobs",
			nil,
			nil,
		),
		skippedJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "skipped_count"),
			"Count of skipped jobs",
			nil,
			nil,
		),
		failedJobCount: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "failed_count"),
			"Count of failed jobs",
			nil,
			nil,
		),
		durationAvg: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "duration_avg"),
			"Average duration of jobs",
			nil,
			nil,
		),
	}
}

func (c *jobsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jobCount
	ch <- c.retryJobCount
	ch <- c.abortJobCount
	ch <- c.skippedJobCount
	ch <- c.failedJobCount
	ch <- c.durationAvg
}

func (c *jobsCollector) Collect(ch chan<- prometheus.Metric) error {
	jobs, err := c.client.GetJobs(c.ctx, c.logger)
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
			c.jobCount,
			prometheus.GaugeValue,
			float64(count),
			status,
		)
	}
	ch <- prometheus.MustNewConstMetric(
		c.retryJobCount,
		prometheus.GaugeValue,
		float64(retryCount),
	)
	ch <- prometheus.MustNewConstMetric(
		c.abortJobCount,
		prometheus.GaugeValue,
		float64(abortCount),
	)
	ch <- prometheus.MustNewConstMetric(
		c.skippedJobCount,
		prometheus.GaugeValue,
		float64(skippedCount),
	)
	ch <- prometheus.MustNewConstMetric(
		c.failedJobCount,
		prometheus.GaugeValue,
		float64(failedCount),
	)

	var duration float64

	if finishedCount > 0 {
		duration = float64(durationSum) / float64(finishedCount)
	} else {
		duration = 0
	}
	ch <- prometheus.MustNewConstMetric(
		c.durationAvg,
		prometheus.GaugeValue,
		duration,
	)

	return nil
}
