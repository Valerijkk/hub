package mail

import "errors"

var (
	// ErrInvalidRequest — некорректный запрос отправки.
	ErrInvalidRequest = errors.New("invalid mail send request")
	// ErrAlertsDisabled — admin/config выключил transactional alerts (W7-A1-003).
	ErrAlertsDisabled = errors.New("mail alerts disabled")
	// ErrDailyLimitExceeded — превышен daily_limit на получателя (W7-A1-003).
	ErrDailyLimitExceeded = errors.New("mail daily limit exceeded")
)

// StatusResponse — публичный статус hub-mail (без секретов).
type StatusResponse struct {
	Provider      string `json:"provider"`
	FromAlerts    string `json:"from_alerts"`
	ResendReady   bool   `json:"resend_ready"`
	DailyLimit    int    `json:"daily_limit"`
	AlertsEnabled bool   `json:"alerts_enabled"`
}

// SettingsResponse — admin view настроек почты (Phase 15.2).
type SettingsResponse struct {
	HubMailURL    string `json:"hub_mail_url"`
	FromAlerts    string `json:"from_alerts"`
	AlertsEnabled bool   `json:"alerts_enabled"`
	DailyLimit    int    `json:"daily_limit"`
	ResendReady   bool   `json:"resend_ready"`
}

// UpdateSettingsRequest — обновление admin настроек.
type UpdateSettingsRequest struct {
	HubMailURL    string `json:"hub_mail_url"`
	FromAlerts    string `json:"from_alerts"`
	AlertsEnabled *bool  `json:"alerts_enabled"`
	DailyLimit    *int   `json:"daily_limit"`
}

const (
	settingHubMailURL    = "mail.hub_url"
	settingFromAlerts    = "mail.from_alerts"
	settingAlertsEnabled = "mail.alerts_enabled"
	settingDailyLimit    = "mail.daily_limit"
)
