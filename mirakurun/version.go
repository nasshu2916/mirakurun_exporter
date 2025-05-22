package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type VersionResponse struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
}

func (c *Client) GetVersion(ctx context.Context, logger *slog.Logger) (*VersionResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/version", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var version VersionResponse
	if err := decodeBody(resp, &version); err != nil {
		return nil, err
	}

	return &version, nil
}
