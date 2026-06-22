package onec

import (
	"encoding/json"
	"errors"
)

const settingsKey = "integrations.onec.config"

var (
	ErrNotConfigured = errors.New("1c connector not configured")
	ErrInvalidConfig = errors.New("invalid 1c connector config")
	ErrWebhookFailed = errors.New("1c webhook request failed")
)

// EmployeeIDField — поле сотрудника в payload для 1C/ZUP mapping.
type EmployeeIDField string

const (
	EmployeeIDEmail    EmployeeIDField = "email"
	EmployeeIDHubUUID  EmployeeIDField = "hub_user_id"
	EmployeeIDPhone    EmployeeIDField = "phone"
)

// Config — сохранённые настройки коннектора (MOD-07).
type Config struct {
	Enabled         bool            `json:"enabled"`
	WebhookURL      string          `json:"webhook_url"`
	BearerToken     string          `json:"bearer_token,omitempty"`
	EmployeeIDField EmployeeIDField `json:"employee_id_field"`
}

// SettingsResponse — ответ admin GET (токен замаскирован).
type SettingsResponse struct {
	Enabled         bool            `json:"enabled"`
	WebhookURL      string          `json:"webhook_url"`
	BearerTokenSet  bool            `json:"bearer_token_set"`
	EmployeeIDField EmployeeIDField `json:"employee_id_field"`
}

// UpdateSettingsRequest — запрос admin PUT/PATCH.
type UpdateSettingsRequest struct {
	Enabled         *bool            `json:"enabled,omitempty"`
	WebhookURL      *string          `json:"webhook_url,omitempty"`
	BearerToken     *string          `json:"bearer_token,omitempty"`
	EmployeeIDField *EmployeeIDField `json:"employee_id_field,omitempty"`
}

// TestWebhookResponse — результат POST /admin/integrations/onec/test.
type TestWebhookResponse struct {
	OK         bool   `json:"ok"`
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
}

// LeaveWebhookPayload — JSON-тело, отправляемое в 1С при одобрении отпуска.
type LeaveWebhookPayload struct {
	Event     string         `json:"event"`
	LeaveID   string         `json:"leave_id"`
	Employee  EmployeePayload `json:"employee"`
	LeaveType string         `json:"leave_type"`
	StartDate string         `json:"start_date"`
	EndDate   string         `json:"end_date"`
	Reason    string         `json:"reason,omitempty"`
	ReviewNote string        `json:"review_note,omitempty"`
	ApprovedAt *string      `json:"approved_at,omitempty"`
	ReviewerID string        `json:"reviewer_id,omitempty"`
}

// EmployeePayload — идентификаторы сотрудника для маппинга в 1С.
type EmployeePayload struct {
	HubUserID   string `json:"hub_user_id"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	ExternalID  string `json:"external_id"`
}

func (c Config) validate() error {
	if !c.Enabled {
		return nil
	}
	if c.WebhookURL == "" {
		return ErrInvalidConfig
	}
	switch c.EmployeeIDField {
	case EmployeeIDEmail, EmployeeIDHubUUID, EmployeeIDPhone, "":
	default:
		return ErrInvalidConfig
	}
	return nil
}

func defaultConfig() Config {
	return Config{
		Enabled:         false,
		EmployeeIDField: EmployeeIDEmail,
	}
}

func encodeConfig(c Config) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeConfig(raw string) (Config, error) {
	if raw == "" {
		return defaultConfig(), nil
	}
	var c Config
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return Config{}, err
	}
	if c.EmployeeIDField == "" {
		c.EmployeeIDField = EmployeeIDEmail
	}
	return c, nil
}

func toSettingsResponse(c Config) SettingsResponse {
	return SettingsResponse{
		Enabled:         c.Enabled,
		WebhookURL:      c.WebhookURL,
		BearerTokenSet:  c.BearerToken != "",
		EmployeeIDField: c.EmployeeIDField,
	}
}
