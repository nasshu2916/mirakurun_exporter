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

type mockTunersGetter struct {
	tuners *mirakurun.TunersResponse
	err    error
}

func (m *mockTunersGetter) GetTuners(ctx context.Context, logger *slog.Logger) (*mirakurun.TunersResponse, error) {
	return m.tuners, m.err
}

func TestTunerCollector_Collect(t *testing.T) {
	tests := []struct {
		name    string
		tuners  *mirakurun.TunersResponse
		wantErr bool
		checks  func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			tuners: &mirakurun.TunersResponse{
				{
					Index:       1,
					Name:        "Tuner1",
					Types:       []string{"GR", "BS"},
					IsAvailable: true,
					IsRemote:    false,
					IsFree:      true,
					IsUsing:     false,
					IsFault:     false,
					Users: []mirakurun.TunerUser{
						{
							ID:    "user1",
							Agent: "Chinachu",
							StreamInfo: map[int]mirakurun.TunerStreamInfo{
								0: {
									Packet: 1000,
									Drop:   10,
								},
								16: {
									Packet: 2000,
									Drop:   20,
								},
							},
						},
					},
				},
				{
					Index:       2,
					Name:        "Tuner2",
					Types:       []string{"GR"},
					IsAvailable: true,
					IsRemote:    true,
					IsFree:      false,
					IsUsing:     true,
					IsFault:     false,
					Users: []mirakurun.TunerUser{
						{
							ID:    "user2",
							Agent: "EPGStation",
							StreamInfo: map[int]mirakurun.TunerStreamInfo{
								0: {
									Packet: 2000,
									Drop:   20,
								},
								16: {
									Packet: 3000,
									Drop:   30,
								},
							},
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, metrics []prometheus.Metric) {
				// メトリクスの数を確認
				assert.Equal(t, 18, len(metrics))

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

				// チューナーデバイスメトリクスの検証
				deviceMetrics := metricMap["mirakurun_tuners_device"]
				require.Len(t, deviceMetrics, 2)

				// 各チューナーのメトリクスを検証
				deviceValues := make(map[string]map[string]string)
				for _, m := range deviceMetrics {
					deviceValues[m.Labels["index"]] = m.Labels
				}

				// チューナー1の検証
				tuner1 := deviceValues["1"]
				assert.Equal(t, "Tuner1", tuner1["name"])
				assert.Equal(t, "GR,BS", tuner1["type"])

				// チューナー2の検証
				tuner2 := deviceValues["2"]
				assert.Equal(t, "Tuner2", tuner2["name"])
				assert.Equal(t, "GR", tuner2["type"])

				// 利用可能チューナーメトリクスの検証
				availableMetrics := metricMap["mirakurun_tuners_available_tuner"]
				require.Len(t, availableMetrics, 2)
				availableValues := make(map[string]float64)
				for _, m := range availableMetrics {
					availableValues[m.Labels["index"]] = m.Value
				}
				assert.Equal(t, 1.0, availableValues["1"])
				assert.Equal(t, 1.0, availableValues["2"])

				// リモートチューナーメトリクスの検証
				remoteMetrics := metricMap["mirakurun_tuners_remote_tuner"]
				require.Len(t, remoteMetrics, 2)
				remoteValues := make(map[string]float64)
				for _, m := range remoteMetrics {
					remoteValues[m.Labels["index"]] = m.Value
				}
				assert.Equal(t, 0.0, remoteValues["1"])
				assert.Equal(t, 1.0, remoteValues["2"])

				// 空きチューナーメトリクスの検証
				freeMetrics := metricMap["mirakurun_tuners_free_tuner"]
				require.Len(t, freeMetrics, 2)
				freeValues := make(map[string]float64)
				for _, m := range freeMetrics {
					freeValues[m.Labels["index"]] = m.Value
				}
				assert.Equal(t, 1.0, freeValues["1"])
				assert.Equal(t, 0.0, freeValues["2"])

				// 使用中チューナーメトリクスの検証
				usingMetrics := metricMap["mirakurun_tuners_using_tuner"]
				require.Len(t, usingMetrics, 2)
				usingValues := make(map[string]float64)
				for _, m := range usingMetrics {
					usingValues[m.Labels["index"]] = m.Value
				}
				assert.Equal(t, 0.0, usingValues["1"])
				assert.Equal(t, 1.0, usingValues["2"])

				// 故障チューナーメトリクスの検証
				faultMetrics := metricMap["mirakurun_tuners_fault_tuner"]
				require.Len(t, faultMetrics, 2)
				faultValues := make(map[string]float64)
				for _, m := range faultMetrics {
					faultValues[m.Labels["index"]] = m.Value
				}
				assert.Equal(t, 0.0, faultValues["1"])
				assert.Equal(t, 0.0, faultValues["2"])

				// ユーザーメトリクスの検証
				userMetrics := metricMap["mirakurun_tuners_users"]
				require.Len(t, userMetrics, 2)
				userValues := make(map[string]map[string]string)
				for _, m := range userMetrics {
					userValues[m.Labels["index"]] = m.Labels
				}
				assert.Equal(t, "user1", userValues["1"]["user_id"])
				assert.Equal(t, "Chinachu", userValues["1"]["agent"])
				assert.Equal(t, "user2", userValues["2"]["user_id"])
				assert.Equal(t, "EPGStation", userValues["2"]["agent"])

				// ストリームパケットメトリクスの検証
				packetMetrics := metricMap["mirakurun_tuners_stream_packets"]
				require.Len(t, packetMetrics, 2)
				packetValues := make(map[string]float64)
				for _, m := range packetMetrics {
					packetValues[m.Labels["user_id"]] = m.Value
				}
				assert.Equal(t, 3000.0, packetValues["user1"])
				assert.Equal(t, 5000.0, packetValues["user2"])

				// ストリームドロップメトリクスの検証
				dropMetrics := metricMap["mirakurun_tuners_stream_drops"]
				require.Len(t, dropMetrics, 2)
				dropValues := make(map[string]float64)
				for _, m := range dropMetrics {
					dropValues[m.Labels["user_id"]] = m.Value
				}
				assert.Equal(t, 30.0, dropValues["user1"])
				assert.Equal(t, 50.0, dropValues["user2"])
			},
		},
		{
			name:    "エラー系",
			tuners:  nil,
			wantErr: true,
			checks:  func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockTunersGetter{
				tuners: tt.tuners,
				err:    nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newTunerCollector(context.Background(), nil, slog.Default())
			collector.(*tunerCollector).tunersGetter = mock

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

func TestTunerCollector_Describe(t *testing.T) {
	collector := newTunerCollector(context.Background(), nil, slog.Default())
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
	expectedDescs := 9
	assert.Equal(t, expectedDescs, len(descs))

	expectedDescsMap := map[string]string{
		"mirakurun_tuners_device":          "Tuner device information",
		"mirakurun_tuners_available_tuner": "Available tuner device",
		"mirakurun_tuners_remote_tuner":    "Remote tuner device",
		"mirakurun_tuners_free_tuner":      "Tuner device is free",
		"mirakurun_tuners_using_tuner":     "Tuner device is using",
		"mirakurun_tuners_fault_tuner":     "Tuner device is fault",
		"mirakurun_tuners_users":           "User using tuner device",
		"mirakurun_tuners_stream_packets":  "Stream packets by user",
		"mirakurun_tuners_stream_drops":    "Stream drops packets by user",
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
