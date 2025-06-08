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

type mockVersionGetter struct {
	version *mirakurun.VersionResponse
	err     error
}

func (m *mockVersionGetter) GetVersion(ctx context.Context, logger *slog.Logger) (*mirakurun.VersionResponse, error) {
	return m.version, m.err
}

func TestVersionCollector_Collect(t *testing.T) {
	tests := []struct {
		name    string
		version *mirakurun.VersionResponse
		wantErr bool
		checks  func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			version: &mirakurun.VersionResponse{
				Current: "1.0.0",
				Latest:  "1.1.0",
			},
			wantErr: false,
			checks: func(t *testing.T, metrics []prometheus.Metric) {
				// メトリクスの数を確認
				assert.Equal(t, 1, len(metrics))

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

					if fqName == "mirakurun_version_mirakurun_version" {
						metricMap[fqName] = append(metricMap[fqName], info)
					}
				}

				// バージョンメトリクスの検証
				versionMetrics := metricMap["mirakurun_version_mirakurun_version"]
				require.Len(t, versionMetrics, 1)

				// バージョン情報を検証
				versionInfo := versionMetrics[0].Labels
				assert.Equal(t, "1.0.0", versionInfo["current"])
				assert.Equal(t, "1.1.0", versionInfo["latest"])
			},
		},
		{
			name:    "エラー系",
			version: nil,
			wantErr: true,
			checks:  func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockVersionGetter{
				version: tt.version,
				err:     nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newVersionCollector(context.Background(), nil, slog.Default())
			collector.(*versionCollector).versionGetter = mock

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

func TestVersionCollector_Describe(t *testing.T) {
	collector := newVersionCollector(context.Background(), nil, slog.Default())
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
		"mirakurun_version_mirakurun_version": "Mirakurun version",
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
