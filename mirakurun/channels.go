package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type ChannelsResponse []struct {
	Type      string           `json:"type"`
	Channel   string           `json:"channel"`
	Name      string           `json:"name"`
	TsmfRelTs int              `json:"tsmfRelTs"`
	Services  []ChannelService `json:"services"`
}

type ChannelService struct {
	Id                 int64  `json:"id"`
	ServiceId          int    `json:"serviceId"`
	NetworkId          int    `json:"networkId"`
	Name               string `json:"name"`
	Type               int    `json:"type"`
	LogoId             int    `json:"logoId"`
	HasLogoData        bool   `json:"hasLogoData"`
	RemoteControlKeyId int    `json:"remoteControlKeyId"`
	EpgReady           bool   `json:"epgReady"`
	EpgUpdatedAt       int    `json:"epgUpdatedAt"`
	Channel            struct {
	} `json:"channel"`
}

func (c *Client) GetChannels(ctx context.Context, logger *slog.Logger) (*ChannelsResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/channels", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var channels ChannelsResponse
	if err := decodeBody(resp, &channels); err != nil {
		return nil, err
	}

	return &channels, nil
}
