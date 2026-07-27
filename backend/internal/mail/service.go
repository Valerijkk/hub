package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const maxResendResponseBodyBytes = 10 * 1024 * 1024

// SendRequest — входящий запрос на transactional alert (NATS/HTTP bridge, Phase 15.2+).
type SendRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
	Template string `json:"template"`
}

// Service — отправка transactional email (Resend).
type Service struct {
	cfg    Config
	client *http.Client
}

// NewService создаёт mail service.
func NewService(cfg Config) *Service {
	return &Service{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text,omitempty"`
}

// SendAlert отправляет alert email. Без RESEND_API_KEY — dev no-op + log.
// A7-005: при наличии ключа реально вызывает Resend (не stub «queued»).
func (s *Service) SendAlert(ctx context.Context, req SendRequest) error {
	if req.To == "" || req.Subject == "" {
		return ErrInvalidRequest
	}
	if s.cfg.ResendAPIKey == "" {
		slog.InfoContext(ctx, "hub-mail dev mode: alert skipped",
			"to", req.To, "subject", req.Subject, "template", req.Template)
		return nil
	}

	html := req.HTMLBody
	if html == "" {
		html = req.TextBody
	}
	if html == "" {
		html = req.Subject
	}
	body := resendPayload{
		From:    s.cfg.FromAlerts,
		To:      []string{req.To},
		Subject: req.Subject,
		HTML:    html,
		Text:    req.TextBody,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("hub-mail: marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("hub-mail: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.ResendAPIKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("hub-mail: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResendResponseBodyBytes))
		if readErr != nil {
			return fmt.Errorf("hub-mail: API error (status %d): %w", resp.StatusCode, readErr)
		}
		return fmt.Errorf("hub-mail: API error (status %d): %s", resp.StatusCode, string(respBody))
	}
	slog.InfoContext(ctx, "hub-mail alert sent", "to", req.To, "template", req.Template)
	return nil
}

// Status возвращает статус сервиса для health/admin.
func (s *Service) Status() StatusResponse {
	return StatusResponse{
		Provider:      "resend",
		FromAlerts:    s.cfg.FromAlerts,
		ResendReady:   s.cfg.ResendAPIKey != "",
		DailyLimit:    s.cfg.DailyLimit,
		AlertsEnabled: true,
	}
}
