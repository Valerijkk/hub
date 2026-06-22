package onec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	hrmod "github.com/valerii/hub/backend/internal/hr"
	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// UserLookup загружает поля заявителя для маппинга в 1С.
type UserLookup interface {
	GetUserEmailPhone(ctx context.Context, userID uuid.UUID) (email, phone, displayName string, err error)
}

// Service — коннектор webhook 1С (MOD-07).
type Service struct {
	store  SettingsStore
	users  UserLookup
	client *http.Client
}

// NewService создаёт сервис коннектора 1С.
func NewService(store SettingsStore, users UserLookup) *Service {
	return &Service{
		store:  store,
		users:  users,
		client: safehttp.NewSafeHTTPClient(15 * time.Second),
	}
}

func (s *Service) loadConfig(ctx context.Context) (Config, error) {
	raw, err := s.store.Get(ctx, settingsKey)
	if err != nil {
		return Config{}, fmt.Errorf("load 1c config: %w", err)
	}
	return decodeConfig(raw)
}

func (s *Service) saveConfig(ctx context.Context, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}
	encoded, err := encodeConfig(cfg)
	if err != nil {
		return err
	}
	return s.store.Upsert(ctx, settingsKey, encoded, "1C leave webhook connector (MOD-07)")
}

// GetSettings возвращает настройки для админа (токен замаскирован).
func (s *Service) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	resp := toSettingsResponse(cfg)
	return &resp, nil
}

// UpdateSettings сохраняет настройки админа.
func (s *Service) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) (*SettingsResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.WebhookURL != nil {
		cfg.WebhookURL = *req.WebhookURL
		if cfg.WebhookURL != "" {
			if err := safehttp.ValidateOutboundURL(cfg.WebhookURL); err != nil {
				return nil, ErrInvalidConfig
			}
		}
	}
	if req.BearerToken != nil {
		cfg.BearerToken = *req.BearerToken
	}
	if req.EmployeeIDField != nil {
		cfg.EmployeeIDField = *req.EmployeeIDField
	}
	if err := s.saveConfig(ctx, cfg); err != nil {
		return nil, err
	}
	resp := toSettingsResponse(cfg)
	return &resp, nil
}

// TestWebhook отправляет ping-payload на настроенный endpoint.
func (s *Service) TestWebhook(ctx context.Context) (*TestWebhookResponse, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.WebhookURL == "" {
		return nil, ErrNotConfigured
	}
	body, _ := json.Marshal(map[string]string{
		"event":   "ping",
		"source":  "hub",
		"message": "1c connector test",
	})
	code, err := s.post(ctx, cfg, body)
	if err != nil {
		return &TestWebhookResponse{OK: false, Message: err.Error()}, nil
	}
	ok := code >= 200 && code < 300
	msg := fmt.Sprintf("HTTP %d", code)
	if !ok {
		msg = "unexpected status: " + msg
	}
	return &TestWebhookResponse{OK: ok, StatusCode: code, Message: msg}, nil
}

// ExportApprovedLeave реализует hr.LeaveExporter — POST в 1С при одобрении.
func (s *Service) ExportApprovedLeave(ctx context.Context, req hrmod.LeaveRequestResponse, reviewerID uuid.UUID) error {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return err
	}
	if !cfg.Enabled || cfg.WebhookURL == "" {
		return nil
	}

	email, phone, displayName, err := s.users.GetUserEmailPhone(ctx, req.RequesterID)
	if err != nil {
		slog.WarnContext(ctx, "1c export: user lookup failed", "user_id", req.RequesterID, "error", err)
	}

	payload := buildLeavePayload(req, reviewerID, email, phone, displayName, cfg.EmployeeIDField)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal 1c payload: %w", err)
	}

	code, err := s.post(ctx, cfg, body)
	if err != nil {
		slog.ErrorContext(ctx, "1c leave export failed", "leave_id", req.ID, "error", err)
		return err
	}
	if code < 200 || code >= 300 {
		slog.ErrorContext(ctx, "1c leave export bad status", "leave_id", req.ID, "status", code)
		return ErrWebhookFailed
	}
	slog.InfoContext(ctx, "1c leave exported", "leave_id", req.ID, "status", code)
	return nil
}

func buildLeavePayload(req hrmod.LeaveRequestResponse, reviewerID uuid.UUID, email, phone, displayName string, field EmployeeIDField) LeaveWebhookPayload {
	emp := EmployeePayload{
		HubUserID:   req.RequesterID.String(),
		Email:       email,
		Phone:       phone,
		DisplayName: displayName,
	}
	if displayName == "" {
		emp.DisplayName = req.RequesterName
	}
	switch field {
	case EmployeeIDPhone:
		emp.ExternalID = phone
	case EmployeeIDHubUUID:
		emp.ExternalID = req.RequesterID.String()
	default:
		emp.ExternalID = email
	}

	var approvedAt *string
	if req.ReviewedAt != nil {
		s := req.ReviewedAt.UTC().Format(time.RFC3339)
		approvedAt = &s
	}

	return LeaveWebhookPayload{
		Event:      "leave.approved",
		LeaveID:    req.ID.String(),
		Employee:   emp,
		LeaveType:  string(req.LeaveType),
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Reason:     req.Reason,
		ReviewNote: req.ReviewNote,
		ApprovedAt: approvedAt,
		ReviewerID: reviewerID.String(),
	}
}

func (s *Service) post(ctx context.Context, cfg Config, body []byte) (int, error) {
	if err := safehttp.ValidateOutboundURL(cfg.WebhookURL); err != nil {
		return 0, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Hub-1C-Connector/1.0")
	if cfg.BearerToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cfg.BearerToken)
	}
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	return resp.StatusCode, nil
}
