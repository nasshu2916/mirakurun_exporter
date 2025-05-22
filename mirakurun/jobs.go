package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type JobsResponse []struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Id           string `json:"id"`
	Status       string `json:"status"`
	RetryCount   int    `json:"retryCount"`
	IsRerunnable bool   `json:"isRerunnable"`
	RetryOnAbort bool   `json:"retryOnAbort"`
	RetryOnFail  bool   `json:"retryOnFail"`
	RetryMax     int    `json:"retryMax"`
	RetryDelay   int    `json:"retryDelay"`
	IsAborting   bool   `json:"isAborting"`
	HasAborted   bool   `json:"hasAborted"`
	HasSkipped   bool   `json:"hasSkipped"`
	HasFailed    bool   `json:"hasFailed"`
	Error        string `json:"error"`
	CreatedAt    int    `json:"createdAt"`
	UpdatedAt    int    `json:"updatedAt"`
	StartedAt    int    `json:"startedAt"`
	FinishedAt   int    `json:"finishedAt"`
	Duration     int    `json:"duration"`
}

func (c *Client) GetJobs(ctx context.Context, logger *slog.Logger) (*JobsResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/jobs", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var programs JobsResponse
	if err := decodeBody(resp, &programs); err != nil {
		return nil, err
	}

	return &programs, nil
}
