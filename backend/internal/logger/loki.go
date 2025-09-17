package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

type lokiPayload struct {
	Streams []lokiStream `json:"streams"`
}

// LokiWriter batches logs and pushes them to Loki periodically.
type LokiWriter struct {
	url       string
	labels    map[string]string
	client    *http.Client
	batch     [][2]string
	mu        sync.Mutex
	flushTick *time.Ticker
	stopCh    chan struct{}
}

func NewLokiWriter(url string, labels map[string]string, flushInterval time.Duration) *LokiWriter {
	lw := &LokiWriter{
		url:       url,
		labels:    labels,
		client:    &http.Client{Timeout: 5 * time.Second},
		batch:     make([][2]string, 0, 100),
		flushTick: time.NewTicker(flushInterval),
		stopCh:    make(chan struct{}),
	}

	go lw.run()
	return lw
}

func (lw *LokiWriter) Write(p []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	lw.batch = append(lw.batch, [2]string{
		fmt.Sprintf("%d", time.Now().UnixNano()),
		string(bytes.TrimSpace(p)),
	})
	return len(p), nil
}

func (lw *LokiWriter) run() {
	for {
		select {
		case <-lw.flushTick.C:
			lw.flush()
		case <-lw.stopCh:
			lw.flush()
			return
		}
	}
}

func (lw *LokiWriter) flush() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if len(lw.batch) == 0 {
		return
	}

	payload := lokiPayload{
		Streams: []lokiStream{{
			Stream: lw.labels,
			Values: lw.batch,
		}},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		GetLogger().Error().Err(err).Msg("loki marshal error")
		lw.batch = lw.batch[:0]
		return
	}

	resp, err := lw.client.Post(lw.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		GetLogger().Error().Err(err).Msg("loki push error")
		lw.batch = lw.batch[:0]
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		GetLogger().Error().Msgf("loki push failed: %s", resp.Status)
	}

	// Clear batch
	lw.batch = lw.batch[:0]
}

func (lw *LokiWriter) Close() {
	close(lw.stopCh)
}
