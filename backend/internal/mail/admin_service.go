package mail

import (
	"context"
)

// SettingsStore — persistence для mail admin settings (SystemSetting).
type SettingsStore interface {
	Get(ctx context.Context, key string) (string, error)
	Upsert(ctx context.Context, key, value, description string) error
}

// AdminService — бизнес-логика admin mail settings в основном Hub API.
type AdminService struct {
	store      SettingsStore
	defaultURL string
	fromAlerts string
	resendReady bool
}

// NewAdminService создаёт admin mail service.
func NewAdminService(store SettingsStore, hubMailURL, fromAlerts string, resendReady bool) *AdminService {
	return &AdminService{
		store:       store,
		defaultURL:  hubMailURL,
		fromAlerts:  fromAlerts,
		resendReady: resendReady,
	}
}

// GetSettings возвращает текущие настройки почтового сервиса.
func (s *AdminService) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	url := s.defaultURL
	if v, err := s.store.Get(ctx, settingHubMailURL); err == nil && v != "" {
		url = v
	}
	from := s.fromAlerts
	if v, err := s.store.Get(ctx, settingFromAlerts); err == nil && v != "" {
		from = v
	}
	enabled := true
	if v, err := s.store.Get(ctx, settingAlertsEnabled); err == nil && v == "false" {
		enabled = false
	}
	limit := 20
	if v, err := s.store.Get(ctx, settingDailyLimit); err == nil && v != "" {
		if n, parseErr := parsePositiveInt(v); parseErr == nil {
			limit = n
		}
	}
	return &SettingsResponse{
		HubMailURL:    url,
		FromAlerts:    from,
		AlertsEnabled: enabled,
		DailyLimit:    limit,
		ResendReady:   s.resendReady,
	}, nil
}

// UpdateSettings сохраняет admin настройки.
func (s *AdminService) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) (*SettingsResponse, error) {
	if req.HubMailURL != "" {
		if err := s.store.Upsert(ctx, settingHubMailURL, req.HubMailURL, "hub-mail service URL"); err != nil {
			return nil, err
		}
	}
	if req.FromAlerts != "" {
		if err := s.store.Upsert(ctx, settingFromAlerts, req.FromAlerts, "transactional alerts from address"); err != nil {
			return nil, err
		}
	}
	if req.AlertsEnabled != nil {
		val := "true"
		if !*req.AlertsEnabled {
			val = "false"
		}
		if err := s.store.Upsert(ctx, settingAlertsEnabled, val, "transactional email alerts enabled"); err != nil {
			return nil, err
		}
	}
	if req.DailyLimit != nil && *req.DailyLimit > 0 {
		if err := s.store.Upsert(ctx, settingDailyLimit, itoa(*req.DailyLimit), "max alert emails per user per day"); err != nil {
			return nil, err
		}
	}
	return s.GetSettings(ctx)
}

func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, ErrInvalidRequest
		}
		n = n*10 + int(c-'0')
	}
	if n < 1 {
		return 0, ErrInvalidRequest
	}
	return n, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
