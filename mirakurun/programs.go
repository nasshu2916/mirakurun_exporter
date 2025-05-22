package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type ProgramsResponse []struct {
	ID                 int64  `json:"id"`
	ServiceID          int    `json:"serviceId"`
	NetworkID          int    `json:"networkId"`
	Name               string `json:"name"`
	Type               int    `json:"type"`
	LogoID             *int   `json:"logoId"`
	HasLogoData        *bool  `json:"hasLogoData"`
	RemoteControlKeyID *int   `json:"remoteControlKeyId"`
	EpgReady           *bool  `json:"epgReady"`
	EpgUpdatedAt       *int64 `json:"epgUpdatedAt"`
	Channel            *struct {
		Type      string                 `json:"type"`
		Channel   string                 `json:"channel"`
		Name      string                 `json:"name"`
		TsmfRelTs *int                   `json:"tsmfRelTs"`
		Services  map[string]interface{} `json:"services"`
	} `json:"channel"`
}

func (c *Client) GetPrograms(ctx context.Context, logger *slog.Logger) (*ProgramsResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/programs", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var programs ProgramsResponse
	if err := decodeBody(resp, &programs); err != nil {
		return nil, err
	}

	return &programs, nil
}
