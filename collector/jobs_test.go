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

type mockJobsGetter struct {
	jobs *mirakurun.JobsResponse
	err  error
}

func (m *mockJobsGetter) GetJobs(ctx context.Context, logger *slog.Logger) (*mirakurun.JobsResponse, error) {
	return m.jobs, m.err
}

func TestJobsCollector_Collect(t *testing.T) {
	tests := []struct {
		name    string
		jobs    *mirakurun.JobsResponse
		wantErr bool
		checks  func(t *testing.T, metrics []prometheus.Metric)
	}{
		{
			name: "正常系",
			jobs: &mirakurun.JobsResponse{
				{
					Status:     "finished",
					RetryCount: 0,
					IsAborting: false,
					HasSkipped: false,
					HasFailed:  false,
					Duration:   150,
				},
				{
					Status:     "running",
					RetryCount: 1,
					IsAborting: false,
					HasSkipped: false,
					HasFailed:  false,
					Duration:   50,
				},
				{
					Status:     "finished",
					RetryCount: 2,
					IsAborting: true,
					HasSkipped: true,
					HasFailed:  true,
					Duration:   200,
				},
			},
			wantErr: false,
			checks: func(t *testing.T, metrics []prometheus.Metric) {
				// メトリクスの数を確認
				assert.Equal(t, 7, len(metrics))

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

				// ジョブ数メトリクスの検証
				countMetrics := metricMap["mirakurun_jobs_count"]
				require.Len(t, countMetrics, 2)
				countValues := make(map[string]float64)
				for _, m := range countMetrics {
					countValues[m.Labels["status"]] = m.Value
				}
				assert.Equal(t, 2.0, countValues["finished"])
				assert.Equal(t, 1.0, countValues["running"])

				// リトライ数メトリクスの検証
				retryMetrics := metricMap["mirakurun_jobs_retry_count"]
				require.Len(t, retryMetrics, 1)
				assert.Equal(t, 3.0, retryMetrics[0].Value)

				// 中断数メトリクスの検証
				abortMetrics := metricMap["mirakurun_jobs_abort_count"]
				require.Len(t, abortMetrics, 1)
				assert.Equal(t, 1.0, abortMetrics[0].Value)

				// スキップ数メトリクスの検証
				skippedMetrics := metricMap["mirakurun_jobs_skipped_count"]
				require.Len(t, skippedMetrics, 1)
				assert.Equal(t, 1.0, skippedMetrics[0].Value)

				// 失敗数メトリクスの検証
				failedMetrics := metricMap["mirakurun_jobs_failed_count"]
				require.Len(t, failedMetrics, 1)
				assert.Equal(t, 1.0, failedMetrics[0].Value)

				// 平均実行時間メトリクスの検証
				durationMetrics := metricMap["mirakurun_jobs_duration_avg"]
				require.Len(t, durationMetrics, 1)
				assert.Equal(t, 150.0, durationMetrics[0].Value)
			},
		},
		{
			name:    "エラー系",
			jobs:    nil,
			wantErr: true,
			checks:  func(t *testing.T, metrics []prometheus.Metric) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockJobsGetter{
				jobs: tt.jobs,
				err:  nil,
			}
			if tt.wantErr {
				mock.err = assert.AnError
			}

			collector := newJobsCollector(context.Background(), nil, slog.Default())
			collector.(*jobsCollector).jobsGetter = mock

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

func TestJobsCollector_Describe(t *testing.T) {
	collector := newJobsCollector(context.Background(), nil, slog.Default())
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
	expectedDescs := 6
	assert.Equal(t, expectedDescs, len(descs))

	expectedDescsMap := map[string]string{
		"mirakurun_jobs_count":         "Count of jobs",
		"mirakurun_jobs_retry_count":   "Count of retried jobs",
		"mirakurun_jobs_abort_count":   "Count of aborted jobs",
		"mirakurun_jobs_skipped_count": "Count of skipped jobs",
		"mirakurun_jobs_failed_count":  "Count of failed jobs",
		"mirakurun_jobs_duration_avg":  "Average duration of jobs",
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
