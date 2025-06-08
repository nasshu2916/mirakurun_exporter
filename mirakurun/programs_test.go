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

func TestGetPrograms(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/programs.json")

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

	programs, err := c.GetPrograms(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &ProgramsResponse{
		{
			ID:        327360102403001,
			ServiceID: 1024,
			NetworkID: 32736,
			Name:      "ニュース",
		},
		{
			ID:        327360102403023,
			ServiceID: 1024,
			NetworkID: 32736,
			Name:      "ニュース・気象情報🈑🈐",
		},
	}

	if diff := cmp.Diff(want, programs); diff != "" {
		t.Fatalf("programs mismatch (-want +got):\n%s", diff)
	}
}
