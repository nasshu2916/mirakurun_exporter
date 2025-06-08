package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type TunersResponse []struct {
	Index       int         `json:"index"`
	Name        string      `json:"name"`
	Types       []string    `json:"types"`
	Command     string      `json:"command"`
	PID         int         `json:"pid"`
	Users       []TunerUser `json:"users"`
	IsAvailable bool        `json:"isAvailable"`
	IsRemote    bool        `json:"isRemote"`
	IsFree      bool        `json:"isFree"`
	IsUsing     bool        `json:"isUsing"`
	IsFault     bool        `json:"isFault"`
}

type TunerUser struct {
	ID             string                  `json:"id"`
	Priority       int                     `json:"priority"`
	Agent          string                  `json:"agent"`
	URL            string                  `json:"url"`
	DisableDecoder bool                    `json:"disableDecoder"`
	StreamSetting  TunerStreamSetting      `json:"streamSetting"`
	StreamInfo     map[int]TunerStreamInfo `json:"streamInfo"`
}

type TunerStreamSetting struct {
	Channel   TunerStreamSettingChannel `json:"channel"`
	NetworkID int                       `json:"networkId"`
	ServiceID int                       `json:"serviceId"`
	EventID   int                       `json:"eventId"`
	NoProvide bool                      `json:"noProvide"`
	ParseEIT  bool                      `json:"parseEIT"`
	ParseSDT  bool                      `json:"parseSDT"`
	ParseNIT  bool                      `json:"parseNIT"`
}

type TunerStreamSettingChannel struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Channel     string                 `json:"channel"`
	ServiceID   int                    `json:"serviceId"`
	TsmfRelTs   int                    `json:"tsmfRelTs"`
	CommandVars map[string]interface{} `json:"commandVars"`
}

type TunerStreamInfo struct {
	Packet int64 `json:"packet"`
	Drop   int64 `json:"drop"`
}

func (c *Client) GetTuners(ctx context.Context, logger *slog.Logger) (*TunersResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/tuners", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var tunersResponse TunersResponse
	if err := decodeBody(resp, &tunersResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &tunersResponse, nil
}
