package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
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
	// store — опционально admin SystemSetting (Hub server); nil → только cfg.
	store SettingsStore

	mu       sync.Mutex
	dayKey   string // UTC YYYY-MM-DD
	perEmail map[string]int
}

// NewService создаёт mail service.
// AlertsEnabled по умолчанию true (литерал Config{} в тестах); LoadConfig задаёт явно.
func NewService(cfg Config) *Service {
	if cfg.DailyLimit < 1 {
		cfg.DailyLimit = 20
	}
	if !cfg.alertsSet {
		cfg.AlertsEnabled = true
	}
	return &Service{
		cfg:      cfg,
		client:   &http.Client{Timeout: 10 * time.Second},
		perEmail: make(map[string]int),
	}
}

// SetSettingsStore подключает admin settings (alerts_enabled / daily_limit) к SendAlert.
// W7-A1-003: без store настройки админки декоративны.
func (s *Service) SetSettingsStore(store SettingsStore) {
	if s == nil {
		return
	}
	s.store = store
}

// SetAlertsEnabled явно включает/выключает alerts на уровне cfg (тесты / override).
func (s *Service) SetAlertsEnabled(enabled bool) {
	if s == nil {
		return
	}
	s.cfg.AlertsEnabled = enabled
	s.cfg.alertsSet = true
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
// W7-A1-003: enforcement alerts_enabled + daily_limit (cfg и/или admin store).
func (s *Service) SendAlert(ctx context.Context, req SendRequest) error {
	if req.To == "" || req.Subject == "" {
		return ErrInvalidRequest
	}
	enabled, limit := s.effectivePolicy(ctx)
	if !enabled {
		return ErrAlertsDisabled
	}
	if err := s.reserveDailySlot(req.To, limit); err != nil {
		return err
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
		s.releaseDailySlot(req.To)
		return fmt.Errorf("hub-mail: marshal: %w", err)
	}
	// Detach from Fiber RequestCtx: fasthttp переиспользует ctx после ответа/timeout теста
	// (race с dial). Outbound HTTP живёт на своём deadline.
	httpCtx, httpCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer httpCancel()
	httpReq, err := http.NewRequestWithContext(httpCtx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		s.releaseDailySlot(req.To)
		return fmt.Errorf("hub-mail: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.ResendAPIKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		s.releaseDailySlot(req.To)
		return fmt.Errorf("hub-mail: send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		s.releaseDailySlot(req.To)
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
	enabled, limit := s.effectivePolicy(context.Background())
	return StatusResponse{
		Provider:      "resend",
		FromAlerts:    s.cfg.FromAlerts,
		ResendReady:   s.cfg.ResendAPIKey != "",
		DailyLimit:    limit,
		AlertsEnabled: enabled,
	}
}

func (s *Service) effectivePolicy(ctx context.Context) (enabled bool, limit int) {
	enabled = s.cfg.AlertsEnabled
	limit = s.cfg.DailyLimit
	if limit < 1 {
		limit = 20
	}
	if s.store == nil {
		return enabled, limit
	}
	if v, err := s.store.Get(ctx, settingAlertsEnabled); err == nil && v != "" {
		enabled = v != "false"
	}
	if v, err := s.store.Get(ctx, settingDailyLimit); err == nil && v != "" {
		if n, parseErr := parsePositiveInt(v); parseErr == nil {
			limit = n
		}
	}
	return enabled, limit
}

func (s *Service) reserveDailySlot(email string, limit int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	today := time.Now().UTC().Format("2006-01-02")
	if s.dayKey != today {
		s.dayKey = today
		s.perEmail = make(map[string]int)
	}
	if s.perEmail[email] >= limit {
		return ErrDailyLimitExceeded
	}
	s.perEmail[email]++
	return nil
}

func (s *Service) releaseDailySlot(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.perEmail[email] > 0 {
		s.perEmail[email]--
	}
}
