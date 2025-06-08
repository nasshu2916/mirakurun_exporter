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

func TestGetServices(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/services.json")

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

	services, err := c.GetServices(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &ServicesResponse{
		{
			ID:                 3273601024,
			ServiceID:          1024,
			NetworkID:          32736,
			Name:               "ＮＨＫ総合１・東京",
			Type:               1,
			LogoID:             0,
			HasLogoData:        true,
			RemoteControlKeyID: 1,
			EpgReady:           true,
			EpgUpdatedAt:       1749000000000,
			Channel: ServiceChannel{
				Type:    "GR",
				Channel: "T27",
			},
		},
		{
			ID:                 400101,
			ServiceID:          101,
			NetworkID:          4,
			Name:               "ＮＨＫ　ＢＳ",
			Type:               1,
			LogoID:             -1,
			HasLogoData:        false,
			RemoteControlKeyID: 0,
			EpgReady:           true,
			EpgUpdatedAt:       1749000000000,
			Channel: ServiceChannel{
				Type:    "BS",
				Channel: "BS15_0",
			},
		},
		{
			ID:                 400102,
			ServiceID:          102,
			NetworkID:          4,
			Name:               "ＮＨＫ　ＢＳ",
			Type:               1,
			LogoID:             -1,
			HasLogoData:        false,
			RemoteControlKeyID: 0,
			EpgReady:           true,
			EpgUpdatedAt:       1749000000000,
			Channel: ServiceChannel{
				Type:    "BS",
				Channel: "BS15_0",
			},
		},
	}

	if diff := cmp.Diff(want, services); diff != "" {
		t.Fatalf("services mismatch (-want +got):\n%s", diff)
	}
}
