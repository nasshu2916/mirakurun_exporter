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

func TestGetTuners(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/tuners.json")

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

	tuners, err := c.GetTuners(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &TunersResponse{
		{
			Index:       0,
			Name:        "Tuner (Terrestrial) #1",
			Types:       []string{"GR"},
			Command:     "",
			PID:         0,
			Users:       []TunerUser{},
			IsAvailable: true,
			IsRemote:    false,
			IsFree:      true,
			IsUsing:     false,
			IsFault:     false,
		},
		{
			Index:   1,
			Name:    "Tuner (Terrestrial) #2",
			Types:   []string{"GR"},
			Command: "recisdb tune --device /dev/pt3video3 --channel T27 -",
			PID:     5978,
			Users: []TunerUser{
				{
					ID:             "192.168.1.10:53833",
					Priority:       0,
					URL:            "/api/channels/GR/T27/stream?decode=1",
					DisableDecoder: false,
					StreamSetting: TunerStreamSetting{
						Channel: TunerStreamSettingChannel{
							Name:    "NHK総合・東京",
							Type:    "GR",
							Channel: "T27",
							CommandVars: map[string]interface{}{
								"satellite": " ",
							},
						},
						NetworkID: 32736,
						ParseEIT:  true,
					},
					StreamInfo: map[int]TunerStreamInfo{
						0: {
							Packet: 10,
							Drop:   0,
						},
						16: {
							Packet: 20,
							Drop:   0,
						},
					},
				},
			},
			IsAvailable: true,
			IsRemote:    false,
			IsFree:      false,
			IsUsing:     true,
			IsFault:     false,
		},
		{
			Index:       2,
			Name:        "Tuner (Satellite) #1",
			Types:       []string{"BS", "CS"},
			Command:     "",
			PID:         0,
			Users:       []TunerUser{},
			IsAvailable: true,
			IsRemote:    false,
			IsFree:      true,
			IsUsing:     false,
			IsFault:     false,
		},
		{
			Index:       3,
			Name:        "Tuner (Satellite) #2",
			Types:       []string{"BS", "CS"},
			Command:     "",
			PID:         0,
			Users:       []TunerUser{},
			IsAvailable: true,
			IsRemote:    false,
			IsFree:      true,
			IsUsing:     false,
			IsFault:     false,
		},
	}

	if diff := cmp.Diff(want, tuners); diff != "" {
		t.Fatalf("tuners mismatch (-want +got):\n%s", diff)
	}
}
