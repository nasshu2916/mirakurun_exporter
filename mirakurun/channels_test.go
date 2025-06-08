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

func TestGetChannels(t *testing.T) {
	testHelper := &util.TestHelper{}

	responseBody := testHelper.ReadFile(t, "../test/mirakurun/channels.json")

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

	channels, err := c.GetChannels(ctx, logger)
	if err != nil {
		t.Fatal(err)
	}

	want := &ChannelsResponse{
		{
			Type:      "GR",
			Channel:   "T27",
			Name:      "NHK総合・東京",
			TsmfRelTs: 0,
			Services: []ChannelService{
				{
					Id:                 3273601024,
					ServiceId:          1024,
					NetworkId:          32736,
					Name:               "ＮＨＫ総合１・東京",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 3273601025,
					ServiceId:          1025,
					NetworkId:          32736,
					Name:               "ＮＨＫ総合２・東京",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 3273601408,
					ServiceId:          1408,
					NetworkId:          32736,
					Name:               "ＮＨＫ携帯Ｇ・東京",
					Type:               192,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
			},
		},
		{
			Type:      "BS",
			Channel:   "BS15_0",
			Name:      "BS15/TS0",
			TsmfRelTs: 0,
			Services: []ChannelService{
				{
					Id:                 400101,
					ServiceId:          101,
					NetworkId:          4,
					Name:               "ＮＨＫ　ＢＳ",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 400102,
					ServiceId:          102,
					NetworkId:          4,
					Name:               "ＮＨＫ　ＢＳ",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 400700,
					ServiceId:          700,
					NetworkId:          4,
					Name:               "ＮＨＫデータ１",
					Type:               192,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 400701,
					ServiceId:          701,
					NetworkId:          4,
					Name:               "ＮＨＫデータ２",
					Type:               192,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 400707,
					ServiceId:          707,
					NetworkId:          4,
					Name:               "７０７チャンネル",
					Type:               192,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 400929,
					ServiceId:          929,
					NetworkId:          4,
					Name:               "Ｄｐａダウンロード",
					Type:               164,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
			},
		},
		{
			Type:      "CS",
			Channel:   "CS02",
			Name:      "ND02",
			TsmfRelTs: 0,
			Services: []ChannelService{
				{
					Id:                 600296,
					ServiceId:          296,
					NetworkId:          6,
					Name:               "ＴＢＳチャンネル１",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 600298,
					ServiceId:          298,
					NetworkId:          6,
					Name:               "テレ朝チャンネル１",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 600299,
					ServiceId:          299,
					NetworkId:          6,
					Name:               "テレ朝チャンネル２",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
				{
					Id:                 600339,
					ServiceId:          339,
					NetworkId:          6,
					Name:               "ディズニージュニア",
					Type:               1,
					LogoId:             0,
					HasLogoData:        false,
					RemoteControlKeyId: 0,
					EpgReady:           false,
					EpgUpdatedAt:       0,
					Channel:            struct{}{},
				},
			},
		},
	}

	if diff := cmp.Diff(want, channels); diff != "" {
		t.Fatalf("channels mismatch (-want +got):\n%s", diff)
	}
}
