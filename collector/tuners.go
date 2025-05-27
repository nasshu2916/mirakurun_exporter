package collector

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
	"github.com/prometheus/client_golang/prometheus"
)

type tunerCollector struct {
	ctx    context.Context
	client *mirakurun.Client
	logger *slog.Logger

	tunerDevice     *prometheus.Desc
	availableTuners *prometheus.Desc
	remoteTuners    *prometheus.Desc
	freeTuners      *prometheus.Desc
	usingTuners     *prometheus.Desc
	faultTuners     *prometheus.Desc
	users           *prometheus.Desc
	streamPackets   *prometheus.Desc
	streamDrops     *prometheus.Desc
}

func init() {
	registerCollector("tuners", defaultEnabled, newTunerCollector)
}

func newTunerCollector(ctx context.Context, client *mirakurun.Client, logger *slog.Logger) Collector {
	const subsystem = "tuners"

	return &tunerCollector{
		ctx:    ctx,
		client: client,
		logger: logger,

		tunerDevice: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "device"),
			"",
			[]string{"index", "name", "type"},
			nil,
		),
		availableTuners: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "available_tuner"),
			"available tuner device",
			[]string{"index"},
			nil,
		),
		remoteTuners: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "remote_tuner"),
			"remote tuner device",
			[]string{"index"},
			nil,
		),
		freeTuners: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "free_tuner"),
			"tuner device is free",
			[]string{"index"},
			nil,
		),
		usingTuners: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "using_tuner"),
			"tuner device is using",
			[]string{"index"},
			nil,
		),
		faultTuners: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "fault_tuner"),
			"tuner device is fault",
			[]string{"index"},
			nil,
		),
		users: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "users"),
			"user using tuner device",
			[]string{"index", "user_id", "agent"},
			nil,
		),
		streamPackets: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "stream_packets"),
			"stream packets by user",
			[]string{"user_id"},
			nil,
		),
		streamDrops: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "stream_drops"),
			"stream drops packets by user",
			[]string{"user_id"},
			nil,
		),
	}
}

func (c *tunerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.tunerDevice
	ch <- c.availableTuners
	ch <- c.remoteTuners
	ch <- c.freeTuners
	ch <- c.usingTuners
	ch <- c.faultTuners
	ch <- c.users
	ch <- c.streamPackets
	ch <- c.streamDrops
}

func (c *tunerCollector) Collect(ch chan<- prometheus.Metric) error {
	tuners, err := c.client.GetTuners(c.ctx, c.logger)
	if err != nil {
		return err
	}

	for _, tuner := range *tuners {
		index := strconv.Itoa(tuner.Index)
		ch <- prometheus.MustNewConstMetric(
			c.tunerDevice,
			prometheus.GaugeValue,
			1,
			index, tuner.Name, strings.Join(tuner.Types, ","),
		)
		ch <- prometheus.MustNewConstMetric(
			c.availableTuners,
			prometheus.GaugeValue,
			boolToFloat64(tuner.IsAvailable),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.remoteTuners,
			prometheus.GaugeValue,
			boolToFloat64(tuner.IsRemote),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.freeTuners,
			prometheus.GaugeValue,
			boolToFloat64(tuner.IsFree),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.usingTuners,
			prometheus.GaugeValue,
			boolToFloat64(tuner.IsUsing),
			index,
		)
		ch <- prometheus.MustNewConstMetric(
			c.faultTuners,
			prometheus.GaugeValue,
			boolToFloat64(tuner.IsFault),
			index,
		)
		for _, user := range tuner.Users {
			ch <- prometheus.MustNewConstMetric(
				c.users,
				prometheus.GaugeValue,
				1,
				index, user.ID, user.Agent,
			)

			var streamPackets, streamDrops int64
			if user.StreamInfo != nil {
				for _, stream := range user.StreamInfo {
					streamPackets += stream.Packet
					streamDrops += stream.Drop
				}
			} else {
				c.logger.Warn("StreamInfo is nil", "user_id", user.ID)
			}

			ch <- prometheus.MustNewConstMetric(
				c.streamPackets,
				prometheus.CounterValue,
				float64(streamPackets),
				user.ID,
			)
			ch <- prometheus.MustNewConstMetric(
				c.streamDrops,
				prometheus.CounterValue,
				float64(streamDrops),
				user.ID,
			)
		}
	}
	return nil
}
