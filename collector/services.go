package collector

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type servicesGetter interface {
	GetServices(ctx context.Context, logger *slog.Logger) (*mirakurun.ServicesResponse, error)
}

type servicesCollector struct {
	ctx    context.Context
	logger *slog.Logger

	servicesGetter servicesGetter

	metrics     map[string]*prometheus.Desc
	metricTypes map[string]prometheus.ValueType
}

func init() {
	registerCollector("service", defaultEnabled, newServicesCollector)
}

func newServicesCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "service"

	metricDefs := map[string]metricDefinition{
		"service": {
			name:       "service",
			help:       "Service information",
			labelNames: []string{"id", "service_id", "service_name", "service_type", "channel_type", "channel_id"},
			metricType: prometheus.GaugeValue,
		},
		"epg_updated_at": {
			name:       "epg_updated_at",
			help:       "Service EPG updated at",
			labelNames: []string{"id"},
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

	return &servicesCollector{
		ctx:            ctx,
		servicesGetter: client,
		logger:         logger,
		metrics:        metrics,
		metricTypes:    metricTypes,
	}
}

func (c *servicesCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metrics {
		ch <- desc
	}
}

func (c *servicesCollector) Collect(ch chan<- prometheus.Metric) error {
	services, err := c.servicesGetter.GetServices(c.ctx, c.logger)
	if err != nil {
		return err
	}

	for _, service := range *services {
		ID := strconv.Itoa(int(service.ID))
		ch <- prometheus.MustNewConstMetric(
			c.metrics["service"],
			c.metricTypes["service"],
			1,
			ID,
			strconv.Itoa(service.ServiceID),
			service.Name,
			strconv.Itoa(service.Type),
			service.Channel.Type,
			service.Channel.Channel,
		)

		ch <- prometheus.MustNewConstMetric(
			c.metrics["epg_updated_at"],
			c.metricTypes["epg_updated_at"],
			float64(service.EpgUpdatedAt)/1000,
			ID,
		)
	}

	return nil
}
