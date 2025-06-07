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

type mockStatusGetter struct {
	status *mirakurun.StatusResponse
	err    error
}

func (m *mockStatusGetter) GetStatus(ctx context.Context, logger *slog.Logger) (*mirakurun.StatusResponse, error) {
	return m.status, m.err
}

func TestStatusCollector_Collect(t *testing.T) {
	tests := []struct {
		name    string
		status  *mirakurun.StatusResponse
		wantErr bool
		checks  func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			status: &mirakurun.StatusResponse{
				Version: "1.0.0",
				Process: mirakurun.Process{
					Versions: map[string]string{
						"node": "v16.0.0",
					},
					Arch:     "x64",
					Platform: "linux",
					MemoryUsage: mirakurun.MemoryUsage{
						RSS:          1000,
						HeapTotal:    2000,
						HeapUsed:     1500,
						External:     500,
						ArrayBuffers: 300,
					},
				},
				EPG: mirakurun.EPG{
					StoredEvents: 100,
				},
				StreamCount: mirakurun.StreamCount{
					TunerDevice: 2,
					TSFilter:    3,
					Decoder:     1,
				},
				ErrorCount: mirakurun.ErrorCount{
					UncaughtException:  1,
					UnhandledRejection: 2,
					BufferOverflow:     3,
					TunerDeviceRespawn: 4,
					DecoderRespawn:     5,
				},
				TimerAccuracy: mirakurun.TimerAccuracy{
					M1: mirakurun.TimerAccuracyValue{
						Avg: 1.0,
						Min: 0.5,
						Max: 1.5,
					},
					M5: mirakurun.TimerAccuracyValue{
						Avg: 2.0,
						Min: 1.5,
						Max: 2.5,
					},
					M15: mirakurun.TimerAccuracyValue{
						Avg: 3.0,
						Min: 2.5,
						Max: 3.5,
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, metrics []prometheus.Metric) {
				// メトリクスの数を確認
				assert.Equal(t, 25, len(metrics))

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

					for _, prefix := range []string{
						"mirakurun_status_version",
						"mirakurun_status_process",
						"mirakurun_status_memory_usage",
						"mirakurun_status_epg_stored_events",
						"mirakurun_status_stream_count",
						"mirakurun_status_error_count",
						"mirakurun_status_timer_accuracy_m1",
						"mirakurun_status_timer_accuracy_m5",
						"mirakurun_status_timer_accuracy_m15",
					} {
						if fqName == prefix {
							metricMap[prefix] = append(metricMap[prefix], info)
							break
						}
					}
				}

				// バージョンメトリクスの検証
				versionMetrics := metricMap["mirakurun_status_version"]
				require.Len(t, versionMetrics, 1)
				assert.Equal(t, prometheus.GaugeValue, versionMetrics[0].Type)
				assert.Equal(t, 1.0, versionMetrics[0].Value)
				assert.Equal(t, "1.0.0", versionMetrics[0].Labels["mirakurun"])
				assert.Equal(t, "v16.0.0", versionMetrics[0].Labels["node"])

				// プロセスメトリクスの検証
				processMetrics := metricMap["mirakurun_status_process"]
				require.Len(t, processMetrics, 1)
				assert.Equal(t, prometheus.GaugeValue, processMetrics[0].Type)
				assert.Equal(t, 1.0, processMetrics[0].Value)
				assert.Equal(t, "x64", processMetrics[0].Labels["arch"])
				assert.Equal(t, "linux", processMetrics[0].Labels["platform"])

				// メモリ使用量メトリクスの検証
				memoryMetrics := metricMap["mirakurun_status_memory_usage"]
				require.Len(t, memoryMetrics, 5)
				memoryValues := make(map[string]float64)
				for _, m := range memoryMetrics {
					assert.Equal(t, prometheus.GaugeValue, m.Type)
					memoryValues[m.Labels["type"]] = m.Value
				}
				assert.Equal(t, 1000.0, memoryValues["RSS"])
				assert.Equal(t, 2000.0, memoryValues["HeapTotal"])
				assert.Equal(t, 1500.0, memoryValues["HeapUsed"])
				assert.Equal(t, 500.0, memoryValues["External"])
				assert.Equal(t, 300.0, memoryValues["ArrayBuffers"])

				// EPGメトリクスの検証
				epgMetrics := metricMap["mirakurun_status_epg_stored_events"]
				require.Len(t, epgMetrics, 1)
				assert.Equal(t, prometheus.GaugeValue, epgMetrics[0].Type)
				assert.Equal(t, 100.0, epgMetrics[0].Value)

				// ストリーム数メトリクスの検証
				streamMetrics := metricMap["mirakurun_status_stream_count"]
				require.Len(t, streamMetrics, 3)
				streamValues := make(map[string]float64)
				for _, m := range streamMetrics {
					assert.Equal(t, prometheus.GaugeValue, m.Type)
					streamValues[m.Labels["type"]] = m.Value
				}
				assert.Equal(t, 2.0, streamValues["TunerDevice"])
				assert.Equal(t, 3.0, streamValues["TSFilter"])
				assert.Equal(t, 1.0, streamValues["Decoder"])

				// エラー数メトリクスの検証
				errorMetrics := metricMap["mirakurun_status_error_count"]
				require.Len(t, errorMetrics, 5)
				errorValues := make(map[string]float64)
				for _, m := range errorMetrics {
					assert.Equal(t, prometheus.CounterValue, m.Type)
					errorValues[m.Labels["type"]] = m.Value
				}
				assert.Equal(t, 1.0, errorValues["UncaughtException"])
				assert.Equal(t, 2.0, errorValues["UnhandledRejection"])
				assert.Equal(t, 3.0, errorValues["BufferOverflow"])
				assert.Equal(t, 4.0, errorValues["TunerDeviceRespawn"])
				assert.Equal(t, 5.0, errorValues["DecoderRespawn"])

				// タイマー精度メトリクスの検証
				timerValues := make(map[string]map[string]float64)
				for period, prefix := range map[string]string{
					"M1":  "mirakurun_status_timer_accuracy_m1",
					"M5":  "mirakurun_status_timer_accuracy_m5",
					"M15": "mirakurun_status_timer_accuracy_m15",
				} {
					metrics := metricMap[prefix]
					require.Len(t, metrics, 3)
					timerValues[period] = make(map[string]float64)
					for _, m := range metrics {
						assert.Equal(t, prometheus.GaugeValue, m.Type)
						timerValues[period][m.Labels["type"]] = m.Value
					}
				}

				// M1の検証
				assert.Equal(t, 1.0, timerValues["M1"]["avg"])
				assert.Equal(t, 0.5, timerValues["M1"]["min"])
				assert.Equal(t, 1.5, timerValues["M1"]["max"])

				// M5の検証
				assert.Equal(t, 2.0, timerValues["M5"]["avg"])
				assert.Equal(t, 1.5, timerValues["M5"]["min"])
				assert.Equal(t, 2.5, timerValues["M5"]["max"])

				// M15の検証
				assert.Equal(t, 3.0, timerValues["M15"]["avg"])
				assert.Equal(t, 2.5, timerValues["M15"]["min"])
				assert.Equal(t, 3.5, timerValues["M15"]["max"])
			},
		},
		{
			name:    "エラー系",
			status:  nil,
			wantErr: true,
			checks:  func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockStatusGetter{
				status: tt.status,
				err:    nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newStatusCollector(context.Background(), nil, slog.Default())
			collector.(*statusCollector).statusGetter = mock

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

func TestStatusCollector_Describe(t *testing.T) {
	collector := newStatusCollector(context.Background(), nil, slog.Default())
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
		"mirakurun_status_version":            "Version of Mirakurun",
		"mirakurun_status_process":            "Process information of Mirakurun",
		"mirakurun_status_memory_usage":       "Memory usage of Mirakurun",
		"mirakurun_status_epg_stored_events":  "Count of stored EPG events",
		"mirakurun_status_stream_count":       "Count of streams",
		"mirakurun_status_error_count":        "Count of errors",
		"mirakurun_status_timer_accuracy_m1":  "Timer accuracy for 1 minute",
		"mirakurun_status_timer_accuracy_m5":  "Timer accuracy for 5 minutes",
		"mirakurun_status_timer_accuracy_m15": "Timer accuracy for 15 minutes",
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
