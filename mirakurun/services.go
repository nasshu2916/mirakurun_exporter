package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type ServicesResponse []struct {
	ID                 int64          `json:"id"`
	ServiceID          int            `json:"serviceId"`
	NetworkID          int            `json:"networkId"`
	Name               string         `json:"name"`
	Type               int            `json:"type"`
	LogoID             int            `json:"logoId"`
	HasLogoData        bool           `json:"hasLogoData"`
	RemoteControlKeyID int            `json:"remoteControlKeyId"`
	EpgReady           bool           `json:"epgReady"`
	EpgUpdatedAt       int64          `json:"epgUpdatedAt"`
	Channel            ServiceChannel `json:"channel"`
}

type ServiceChannel struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Name      string                 `json:"name"`
	TsmfRelTs int                    `json:"tsmfRelTs"`
	Services  map[string]interface{} `json:"services"`
}

func (c *Client) GetServices(ctx context.Context, logger *slog.Logger) (*ServicesResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/services", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var services ServicesResponse
	if err := decodeBody(resp, &services); err != nil {
		return nil, err
	}

	return &services, nil
}
