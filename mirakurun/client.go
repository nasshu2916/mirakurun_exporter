package mirakurun

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	URL string

	httpClient *http.Client
}

func NewClient(mirakurunUrl string) (*Client, error) {
	if mirakurunUrl == "" {
		return nil, fmt.Errorf("mirakurun url is empty")
	}

	if _, err := url.ParseRequestURI(mirakurunUrl); err != nil {
		return nil, fmt.Errorf("mirakurun url is invalid: %w", err)
	}

	client := &Client{
		URL:        mirakurunUrl,
		httpClient: &http.Client{},
	}

	return client, nil
}

func (c *Client) request(ctx context.Context, method string, path string, body io.Reader, logger *slog.Logger) (*http.Response, error) {
	begin := time.Now()
	req, err := http.NewRequestWithContext(ctx, method, c.URL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}

	duration := time.Since(begin)
	logger.Debug("mirakurun request", "method", method, "path", path, "duration_seconds", duration.Seconds())

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func decodeBody(resp *http.Response, v interface{}) error {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to decode response body: %w", err)
	}

	return nil
}
