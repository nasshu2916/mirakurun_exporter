package collector

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasshu2916/mirakurun_exporter/mirakurun"
)

type mockServicesGetter struct {
	services *mirakurun.ServicesResponse
	err      error
}

func (m *mockServicesGetter) GetServices(ctx context.Context, logger *slog.Logger) (*mirakurun.ServicesResponse, error) {
	return m.services, m.err
}

func TestServicesCollector_Collect(t *testing.T) {
	tests := []struct {
		name     string
		services *mirakurun.ServicesResponse
		wantErr  bool
		checks   func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			services: &mirakurun.ServicesResponse{
				{
					ID:           1,
					ServiceID:    101,
					Name:         "NHK総合",
					Type:         1,
					EpgUpdatedAt: 1600000000000,
					Channel: mirakurun.ServiceChannel{
						Type:    "GR",
						Channel: "1",
					},
				},
				{
					ID:           2,
					ServiceID:    102,
					Name:         "NHK Eテレ",
					Type:         1,
					EpgUpdatedAt: 1600000000000,
					Channel: mirakurun.ServiceChannel{
						Type:    "GR",
						Channel: "2",
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, metrics []prometheus.Metric) {
				// メトリクスの数を確認
				assert.Equal(t, 4, len(metrics))

				// メトリクスを種類ごとに分類
				metricMap := make(map[string][]metricInfo)
				for _, metric := range metrics {
					info := getMetricInfo(metric)
					desc := metric.Desc().String()

					// fqNameを抽出
					fqName := ""
					parts := strings.SplitN(desc, "\"", 3)
					if len(parts) > 1 {
						fqName = parts[1]
					}

					metricMap[fqName] = append(metricMap[fqName], info)
				}

				// サービスメトリクスの検証
				serviceMetrics := metricMap["mirakurun_service_service"]
				require.Len(t, serviceMetrics, 2)

				// 各サービスのメトリクスを検証
				serviceValues := make(map[string]map[string]string)
				for _, m := range serviceMetrics {
					serviceValues[m.Labels["id"]] = m.Labels
				}

				// サービス1の検証
				service1 := serviceValues["1"]
				assert.Equal(t, "101", service1["service_id"])
				assert.Equal(t, "NHK総合", service1["service_name"])
				assert.Equal(t, "1", service1["service_type"])
				assert.Equal(t, "GR", service1["channel_type"])
				assert.Equal(t, "1", service1["channel_id"])

				// サービス2の検証
				service2 := serviceValues["2"]
				assert.Equal(t, "102", service2["service_id"])
				assert.Equal(t, "NHK Eテレ", service2["service_name"])
				assert.Equal(t, "1", service2["service_type"])
				assert.Equal(t, "GR", service2["channel_type"])
				assert.Equal(t, "2", service2["channel_id"])

				// EPG更新時刻メトリクスの検証
				epgMetrics := metricMap["mirakurun_service_epg_updated_at"]
				require.Len(t, epgMetrics, 2)

				// 各サービスのEPG更新時刻を検証
				epgValues := make(map[string]float64)
				for _, m := range epgMetrics {
					epgValues[m.Labels["id"]] = m.Value
				}
				assert.Equal(t, 1600000000.0, epgValues["1"])
				assert.Equal(t, 1600000000.0, epgValues["2"])
			},
		},
		{
			name:     "エラー系",
			services: nil,
			wantErr:  true,
			checks:   func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockServicesGetter{
				services: tt.services,
				err:      nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newServicesCollector(context.Background(), nil, slog.Default())
			collector.(*servicesCollector).servicesGetter = mock

			ch := make(chan prometheus.Metric, 100)
			err := collector.Collect(ch)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			metrics := make([]prometheus.Metric, 0)
			for {
				select {
				case metric := <-ch:
					metrics = append(metrics, metric)
				default:
					tt.checks(t, metrics)
					return
				}
			}
		})
	}
}

func TestServicesCollector_Describe(t *testing.T) {
	collector := newServicesCollector(context.Background(), nil, slog.Default())
	ch := make(chan *prometheus.Desc, 64)
	collector.Describe(ch)

	descs := make([]*prometheus.Desc, 0)
	for {
		select {
		case desc := <-ch:
			descs = append(descs, desc)
		default:
			goto done
		}
	}
done:
	expectedDescs := 2
	assert.Equal(t, expectedDescs, len(descs))

	expectedDescsMap := map[string]string{
		"mirakurun_service_service":        "Service information",
		"mirakurun_service_epg_updated_at": "Service EPG updated at",
	}

	found := map[string]bool{}
	for _, desc := range descs {
		descStr := desc.String()
		for fqName, help := range expectedDescsMap {
			if strings.Contains(descStr, fqName) && strings.Contains(descStr, help) {
				found[fqName] = true
			}
		}
	}
	for fqName := range expectedDescsMap {
		assert.True(t, found[fqName], fqName+" not found in described metrics")
	}
}
