package collector

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"strconv"
)

type servicesCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	service             *prometheus.Desc
	serviceEpgUpdatedAt *prometheus.Desc
}

func init() {
	registerCollector("service", defaultEnabled, newServicesCollector)
}

func newServicesCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "service"

	return &servicesCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		service: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "service"),
			"Service information",
			[]string{"id", "service_id", "service_name", "service_type", "channel_type", "channel_id"},
			nil,
		),
		serviceEpgUpdatedAt: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "epg_updated_at"),
			"Service EPG updated at",
			[]string{"id"},
			nil,
		),
	}
}

func (c *servicesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.service
	ch <- c.serviceEpgUpdatedAt
}

func (c *servicesCollector) Collect(ch chan<- prometheus.Metric) error {
	services, err := c.client.GetServices(c.ctx, c.logger)
	if err != nil {
		return err
	}

	for _, service := range *services {
		ID := strconv.Itoa(int(service.ID))
		ch <- prometheus.MustNewConstMetric(
			c.service,
			prometheus.GaugeValue,
			1,
			ID,
			strconv.Itoa(service.ServiceID),
			service.Name,
			strconv.Itoa(service.Type),
			service.Channel.Type,
			service.Channel.Channel,
		)

		ch <- prometheus.MustNewConstMetric(
			c.serviceEpgUpdatedAt,
			prometheus.GaugeValue,
			float64(service.EpgUpdatedAt)/1000,
			ID,
		)
	}

	return nil
}
