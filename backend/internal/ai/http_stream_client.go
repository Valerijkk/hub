package ai

import (
	"net/http"
	"time"

	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// newSafeStreamHTTPClient — HTTP client для SSE без Client.Timeout на весь запрос.
// W5-A2-008 / TASK-085: NewSafeHTTPClient(90s) рвёт длинный stream; cancel — только через ctx.
func newSafeStreamHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext:           safehttp.SafeDialContext(),
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			IdleConnTimeout:       90 * time.Second,
		},
	}
}
