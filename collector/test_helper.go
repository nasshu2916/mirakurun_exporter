package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// テスト用のヘルパー関数
type metricInfo struct {
	Type   prometheus.ValueType
	Value  float64
	Labels map[string]string
}

func getMetricInfo(metric prometheus.Metric) metricInfo {
	dto := &dto.Metric{}
	err := metric.Write(dto)
	if err != nil {
		panic(err)
	}

	var value float64
	var metricType prometheus.ValueType

	switch {
	case dto.Gauge != nil:
		value = dto.Gauge.GetValue()
		metricType = prometheus.GaugeValue
	case dto.Counter != nil:
		value = dto.Counter.GetValue()
		metricType = prometheus.CounterValue
	default:
		value = 0
		metricType = prometheus.UntypedValue
	}

	labels := make(map[string]string)
	for _, label := range dto.GetLabel() {
		labels[label.GetName()] = label.GetValue()
	}

	return metricInfo{
		Type:   metricType,
		Value:  value,
		Labels: labels,
	}
}
