package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type StatusResponse struct {
	Time          int64         `json:"time"`
	Version       string        `json:"version"`
	Process       Process       `json:"process"`
	EPG           EPG           `json:"EPG"`
	RPCCount      int           `json:"rpcCount"`
	StreamCount   StreamCount   `json:"streamCount"`
	ErrorCount    ErrorCount    `json:"errorCount"`
	TimerAccuracy TimerAccuracy `json:"timerAccuracy"`
}

type Process struct {
	Arch        string            `json:"arch"`
	Platform    string            `json:"platform"`
	Versions    map[string]string `json:"versions"`
	Env         map[string]string `json:"env"`
	PID         int               `json:"pid"`
	MemoryUsage MemoryUsage       `json:"memoryUsage"`
}

type MemoryUsage struct {
	RSS          int64 `json:"rss"`
	HeapTotal    int64 `json:"heapTotal"`
	HeapUsed     int64 `json:"heapUsed"`
	External     int64 `json:"external"`
	ArrayBuffers int64 `json:"arrayBuffers"`
}

type EPG struct {
	GatheringNetworks []int `json:"gatheringNetworks"`
	StoredEvents      int64 `json:"storedEvents"`
}

type StreamCount struct {
	TunerDevice int `json:"tunerDevice"`
	TSFilter    int `json:"tsFilter"`
	Decoder     int `json:"decoder"`
}

type ErrorCount struct {
	UncaughtException  int `json:"uncaughtException"`
	UnhandledRejection int `json:"unhandledRejection"`
	BufferOverflow     int `json:"bufferOverflow"`
	TunerDeviceRespawn int `json:"tunerDeviceRespawn"`
	DecoderRespawn     int `json:"decoderRespawn"`
}

type TimerAccuracy struct {
	Last float64            `json:"last"`
	M1   TimerAccuracyValue `json:"m1"`
	M5   TimerAccuracyValue `json:"m5"`
	M15  TimerAccuracyValue `json:"m15"`
}

type TimerAccuracyValue struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

func (c *Client) GetStatus(ctx context.Context, logger *slog.Logger) (*StatusResponse, error) {
	resp, err := c.request(ctx, "GET", "/api/status", nil, logger)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	var status StatusResponse
	if err := decodeBody(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

func (timerAccuracyValue *TimerAccuracyValue) GetValue(field string) float64 {
	switch field {
	case "avg":
		return timerAccuracyValue.Avg
	case "min":
		return timerAccuracyValue.Min
	case "max":
		return timerAccuracyValue.Max
	default:
		panic("unknown timer accuracy value")
	}
}
