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

type mockProgramsGetter struct {
	programs *mirakurun.ProgramsResponse
	err      error
}

func (m *mockProgramsGetter) GetPrograms(ctx context.Context, logger *slog.Logger) (*mirakurun.ProgramsResponse, error) {
	return m.programs, m.err
}

func TestProgramsCollector_Collect(t *testing.T) {
	tests := []struct {
		name     string
		programs *mirakurun.ProgramsResponse
		wantErr  bool
		checks   func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			programs: &mirakurun.ProgramsResponse{
				{
					ServiceID: 1,
				},
				{
					ServiceID: 1,
				},
				{
					ServiceID: 2,
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

					if fqName == "mirakurun_programs_count" {
						metricMap[fqName] = append(metricMap[fqName], info)
					}
				}

				// プログラム数メトリクスの検証
				countMetrics := metricMap["mirakurun_programs_count"]
				require.Len(t, countMetrics, 2)

				// サービスIDごとのプログラム数を検証
				countValues := make(map[string]float64)
				for _, m := range countMetrics {
					countValues[m.Labels["service_id"]] = m.Value
				}
				assert.Equal(t, 2.0, countValues["1"])
				assert.Equal(t, 1.0, countValues["2"])
			},
		},
		{
			name:     "エラー系",
			programs: nil,
			wantErr:  true,
			checks:   func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockProgramsGetter{
				programs: tt.programs,
				err:      nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newProgramsCollector(context.Background(), nil, slog.Default())
			collector.(*programsCollector).programsGetter = mock

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

func TestProgramsCollector_Describe(t *testing.T) {
	collector := newProgramsCollector(context.Background(), nil, slog.Default())
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
		"mirakurun_programs_count": "Count of programs by service",
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
