package mirakurun

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nasshu2916/mirakurun_exporter/util"

	"github.com/google/go-cmp/cmp"
)

func TestGetJobs(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/jobs.json")

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(responseBody)); err != nil {
				t.Errorf("failed to write response: %v", err)
			}
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, 1)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	ctx := context.Background()
	logger := slog.Default()

	jobs, err := c.GetJobs(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &JobsResponse{
		{
			Key:        "EPG.Gatherer",
			Name:       "EPG Gatherer",
			Id:         "MBGOCIW0.3001",
			Status:     "finished",
			RetryCount: 0,
			IsAborting: false,
			CreatedAt:  1749380000000,
			UpdatedAt:  1749380000001,
			Duration:   1,
			StartedAt:  1749380000000,
			HasAborted: false,
			HasSkipped: false,
			HasFailed:  false,
			FinishedAt: 1749381000000,
		},
		{
			Key:        "Program.GC",
			Name:       "Program GC",
			Id:         "MBGOCIW0.3002",
			Status:     "finished",
			RetryCount: 0,
			IsAborting: false,
			CreatedAt:  1749375000000,
			UpdatedAt:  1749375000001,
			Duration:   2,
			StartedAt:  1749375000000,
			HasAborted: false,
			HasSkipped: false,
			HasFailed:  false,
			FinishedAt: 1749376000001,
		},
	}

	if diff := cmp.Diff(want, jobs); diff != "" {
		t.Fatalf("jobs mismatch (-want +got):\n%s", diff)
	}
}
