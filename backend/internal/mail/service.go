package mail

import (
	"context"
	"log/slog"
)

// SendRequest — входящий запрос на transactional alert (NATS/HTTP bridge, Phase 15.2+).
type SendRequest struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
	TextBody string `json:"text_body"`
	Template string `json:"template"`
}

// Service — отправка transactional email (Resend). Skeleton: логирует без API key.
type Service struct {
	cfg Config
}

// NewService создаёт mail service.
func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// SendAlert отправляет alert email. Без RESEND_API_KEY — dev no-op + log.
func (s *Service) SendAlert(ctx context.Context, req SendRequest) error {
	if req.To == "" || req.Subject == "" {
		return ErrInvalidRequest
	}
	if s.cfg.ResendAPIKey == "" {
		slog.InfoContext(ctx, "hub-mail dev mode: alert skipped",
			"to", req.To, "subject", req.Subject, "template", req.Template)
		return nil
	}
	// Phase 15.2+: Resend HTTP client (паттерн auth/resend)
	slog.InfoContext(ctx, "hub-mail send queued", "to", req.To, "template", req.Template)
	return nil
}

// Status возвращает статус сервиса для health/admin.
func (s *Service) Status() StatusResponse {
	return StatusResponse{
		Provider:     "resend",
		FromAlerts:   s.cfg.FromAlerts,
		ResendReady:  s.cfg.ResendAPIKey != "",
		DailyLimit:   s.cfg.DailyLimit,
		AlertsEnabled: true,
	}
}
