package mirakurun

import (
	"context"
	"github.com/nasshu2916/mirakurun_exporter/util"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/status.json")

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseBody))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, 1)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	ctx := context.Background()
	logger := slog.Default()

	status, err := c.GetStatus(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &StatusResponse{
		Time:    1748000000000,
		Version: "4.0.0-beta.18",
		Process: Process{
			Arch:     "x64",
			Platform: "linux",
			Versions: map[string]string{
				"node":             "22.14.0",
				"acorn":            "8.14.0",
				"ada":              "2.9.2",
				"amaro":            "0.3.0",
				"ares":             "1.34.4",
				"brotli":           "1.1.0",
				"cjs_module_lexer": "1.4.1",
				"cldr":             "46.0",
				"icu":              "76.1",
				"llhttp":           "9.2.1",
				"modules":          "127",
				"napi":             "10",
				"nbytes":           "0.1.1",
				"ncrypto":          "0.0.1",
				"nghttp2":          "1.64.0",
				"nghttp3":          "1.6.0",
				"ngtcp2":           "1.10.0",
				"openssl":          "3.0.15+quic",
				"simdjson":         "3.10.1",
				"simdutf":          "6.0.3",
				"sqlite":           "3.47.2",
				"tz":               "2024b",
				"undici":           "6.21.1",
				"unicode":          "16.0",
				"uv":               "1.49.2",
				"uvwasi":           "0.0.21",
				"v8":               "12.4.254.21-node.22",
				"zlib":             "1.3.0.1-motley-82a5fec",
			},
			Env: map[string]string{
				"PATH":                 "dummy-path",
				"DOCKER":               "YES",
				"NODE_ENV":             "production",
				"SERVER_CONFIG_PATH":   "/app-config/server.yml",
				"TUNERS_CONFIG_PATH":   "/app-config/tuners.yml",
				"CHANNELS_CONFIG_PATH": "/app-config/channels.yml",
				"SERVICES_DB_PATH":     "/app-data/services.json",
				"PROGRAMS_DB_PATH":     "/app-data/programs.json",
				"LOGO_DATA_DIR_PATH":   "/app-data/logo-data",
			},
			PID: 101,
			MemoryUsage: MemoryUsage{
				RSS:          1000,
				HeapTotal:    2000,
				HeapUsed:     3000,
				External:     4000,
				ArrayBuffers: 5000,
			},
		},
		EPG: EPG{
			GatheringNetworks: []int{},
			StoredEvents:      20000,
		},
		RPCCount: 1,
		StreamCount: StreamCount{
			TunerDevice: 1,
			TSFilter:    2,
			Decoder:     3,
		},
		ErrorCount: ErrorCount{
			UncaughtException:  0,
			UnhandledRejection: 0,
			BufferOverflow:     0,
			TunerDeviceRespawn: 0,
			DecoderRespawn:     0,
		},
		TimerAccuracy: TimerAccuracy{
			Last: 1234.5,
			M1:   TimerAccuracyValue{Avg: 927.0941833333334, Min: -326.604, Max: 1305.511},
			M5:   TimerAccuracyValue{Avg: 876.90855, Min: -859.261, Max: 1799.097},
			M15:  TimerAccuracyValue{Avg: 527.8147788888889, Min: -859.261, Max: 15287.381},
		},
	}

	if diff := cmp.Diff(want, status); diff != "" {
		t.Fatalf("device mismatch (-want +got):\n%s", diff)
	}

	assert.Equal(t,
		status.TimerAccuracy.M1.GetValue("avg"),
		927.0941833333334,
	)

	assert.Equal(t,
		status.TimerAccuracy.M1.GetValue("min"),
		-326.604,
	)

	assert.Equal(t,
		status.TimerAccuracy.M1.GetValue("max"),
		1305.511,
	)

	assert.Panics(t, func() { status.TimerAccuracy.M1.GetValue("Avg") })
}
