package mirakurun

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type StatusResponse struct {
	Time    int64  `json:"time"`
	Version string `json:"version"`
	Process struct {
		Arch        string            `json:"arch"`
		Platform    string            `json:"platform"`
		Versions    map[string]string `json:"versions"`
		Env         map[string]string `json:"env"`
		PID         int               `json:"pid"`
		MemoryUsage struct {
			RSS          int64 `json:"rss"`
			HeapTotal    int64 `json:"heapTotal"`
			HeapUsed     int64 `json:"heapUsed"`
			External     int64 `json:"external"`
			ArrayBuffers int64 `json:"arrayBuffers"`
		} `json:"memoryUsage"`
	} `json:"process"`
	EPG struct {
		GatheringNetworks []int `json:"gatheringNetworks"`
		StoredEvents      int64 `json:"storedEvents"`
	} `json:"EPG"`
	RPCCount    *int `json:"rpcCount"`
	StreamCount struct {
		TunerDevice int `json:"tunerDevice"`
		TSFilter    int `json:"tsFilter"`
		Decoder     int `json:"decoder"`
	} `json:"streamCount"`
	ErrorCount struct {
		UncaughtException  int `json:"uncaughtException"`
		UnhandledRejection int `json:"unhandledRejection"`
		BufferOverflow     int `json:"bufferOverflow"`
		TunerDeviceRespawn int `json:"tunerDeviceRespawn"`
		DecoderRespawn     int `json:"decoderRespawn"`
	} `json:"errorCount"`
	TimerAccuracy struct {
		Last float64       `json:"last"`
		M1   TimerAccuracy `json:"m1"`
		M5   TimerAccuracy `json:"m5"`
		M15  TimerAccuracy `json:"m15"`
	} `json:"timerAccuracy"`
}

type TimerAccuracy struct {
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

func (timerAccuracy *TimerAccuracy) GetValue(field string) float64 {
	switch field {
	case "avg":
		return timerAccuracy.Avg
	case "min":
		return timerAccuracy.Min
	case "max":
		return timerAccuracy.Max
	default:
		panic("unknown timer accuracy value")
	}
}
