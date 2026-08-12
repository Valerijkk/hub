package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

const (
	gigaChatOAuthURL = "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	gigaChatScope    = "GIGACHAT_API_PERS"
)

type gigaTokenCache struct {
	token     string
	expiresAt time.Time
}

// GigaChatAdapter — Sber GigaChat с OAuth-учётными данными (AI-2).
type GigaChatAdapter struct {
	client *http.Client
	cache  sync.Map // clientID -> gigaTokenCache
}

// NewGigaChatAdapter — GigaChat с SSRF-safe HTTP client.
func NewGigaChatAdapter() *GigaChatAdapter {
	return &GigaChatAdapter{client: safehttp.NewSafeHTTPClient(90 * time.Second)}
}

func parseGigaCredentials(apiKey string) (clientID, clientSecret string, ok bool) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", "", false
	}
	parts := strings.SplitN(apiKey, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (a *GigaChatAdapter) accessToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	if v, ok := a.cache.Load(clientID); ok {
		entry := v.(gigaTokenCache)
		if time.Now().Before(entry.expiresAt) {
			return entry.token, nil
		}
	}
	if err := safehttp.ValidateOutboundURL(gigaChatOAuthURL); err != nil {
		return "", fmt.Errorf("gigachat oauth url: %w", err)
	}
	body := strings.NewReader("scope=" + gigaChatScope)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gigaChatOAuthURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", uuid.New().String())
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gigachat oauth: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		// W5-A2-010 / P3-BE-01: не глотать ошибку чтения OAuth body.
		return "", fmt.Errorf("gigachat oauth read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("gigachat oauth HTTP %d: %s", resp.StatusCode, truncate(string(raw), 256))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresAt   int64  `json:"expires_at"`
	}
	if err := json.Unmarshal(raw, &tok); err != nil {
		return "", fmt.Errorf("gigachat oauth parse: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("gigachat oauth: empty access_token")
	}
	exp := time.Now().Add(25 * time.Minute)
	if tok.ExpiresAt > 0 {
		exp = time.Unix(tok.ExpiresAt, 0).Add(-2 * time.Minute)
	}
	a.cache.Store(clientID, gigaTokenCache{token: tok.AccessToken, expiresAt: exp})
	return tok.AccessToken, nil
}

// ChatCompletion — обращение к GigaChat completions API.
func (a *GigaChatAdapter) ChatCompletion(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	clientID, clientSecret, ok := parseGigaCredentials(apiKey)
	if !ok {
		return nil, fmt.Errorf("gigachat: api_key must be client_id:client_secret")
	}
	token, err := a.accessToken(ctx, clientID, clientSecret)
	if err != nil {
		return nil, err
	}
	base := strings.TrimRight(baseURL, "/")
	if err := safehttp.ValidateOutboundURL(base); err != nil {
		return nil, fmt.Errorf("invalid provider URL: %w", err)
	}
	req.Model = model
	req.Stream = false
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gigachat request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("gigachat HTTP %d: %s", resp.StatusCode, truncate(string(raw), 512))
	}
	var out ChatCompletionResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("gigachat decode: %w", err)
	}
	return &out, nil
}

// ChatCompletionStream — GigaChat non-streaming fallback одним delta.
func (a *GigaChatAdapter) ChatCompletionStream(ctx context.Context, baseURL, apiKey, model string, req ChatCompletionRequest, onDelta func(StreamDelta) error) error {
	resp, err := a.ChatCompletion(ctx, baseURL, apiKey, model, req)
	if err != nil {
		return err
	}
	text := extractAssistantText(resp)
	if text != "" {
		if err := onDelta(StreamDelta{Content: text}); err != nil {
			return err
		}
	}
	return onDelta(StreamDelta{Done: true})
}
