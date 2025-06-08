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

func TestGetVersion(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/version.json")

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

	version, err := c.GetVersion(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &VersionResponse{
		Current: "4.0.0-beta.18",
		Latest:  "4.0.0-beta.19",
	}

	if diff := cmp.Diff(want, version); diff != "" {
		t.Fatalf("version mismatch (-want +got):\n%s", diff)
	}
}
