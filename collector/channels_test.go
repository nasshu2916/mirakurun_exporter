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

type mockChannelsGetter struct {
	channels *mirakurun.ChannelsResponse
	err      error
}

func (m *mockChannelsGetter) GetChannels(ctx context.Context, logger *slog.Logger) (*mirakurun.ChannelsResponse, error) {
	return m.channels, m.err
}

func TestChannelsCollector_Collect(t *testing.T) {
	tests := []struct {
		name     string
		channels *mirakurun.ChannelsResponse
		wantErr  bool
		checks   func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			channels: &mirakurun.ChannelsResponse{
				{
					Name:    "NHK総合",
					Type:    "GR",
					Channel: "1",
				},
				{
					Name:    "NHK Eテレ",
					Type:    "GR",
					Channel: "2",
				},
			},
			wantErr: false,
			checks: func(t *testing.T, metrics []prometheus.Metric) {
				// メトリクスの数を確認
				assert.Equal(t, 2, len(metrics))

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

					if fqName == "mirakurun_channel_channel" {
						metricMap[fqName] = append(metricMap[fqName], info)
					}
				}

				// チャンネルメトリクスの検証
				channelMetrics := metricMap["mirakurun_channel_channel"]
				require.Len(t, channelMetrics, 2)

				// 各チャンネルのメトリクスを検証
				channelValues := make(map[string]map[string]string)
				for _, m := range channelMetrics {
					channelValues[m.Labels["name"]] = m.Labels
				}

				// NHK総合の検証
				nhkG := channelValues["NHK総合"]
				assert.Equal(t, "GR", nhkG["type"])
				assert.Equal(t, "1", nhkG["channel"])

				// NHK Eテレの検証
				nhkE := channelValues["NHK Eテレ"]
				assert.Equal(t, "GR", nhkE["type"])
				assert.Equal(t, "2", nhkE["channel"])
			},
		},
		{
			name:     "エラー系",
			channels: nil,
			wantErr:  true,
			checks:   func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockChannelsGetter{
				channels: tt.channels,
				err:      nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newChannelsCollector(context.Background(), nil, slog.Default())
			collector.(*channelsCollector).channelsGetter = mock

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

func TestChannelsCollector_Describe(t *testing.T) {
	collector := newChannelsCollector(context.Background(), nil, slog.Default())
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
	expectedDescs := 1
	assert.Equal(t, expectedDescs, len(descs))

	expectedDescsMap := map[string]string{
		"mirakurun_channel_channel": "Channel information",
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
