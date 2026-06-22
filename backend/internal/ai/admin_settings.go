package ai

import (
	"context"
	"strings"
)

// DefaultBotSystemPrompt — значение по умолчанию M3-014 (переопределяется админом).
const DefaultBotSystemPrompt = "Ты AI-ассистент команды разработки Hub. Помогаешь с кодом, тестами, документацией. Отвечай на русском, если пользователь пишет по-русски. Форматируй ответы в Markdown."

// BotSystemPrompt — admin-configured или default bot prompt.
func (s *Service) BotSystemPrompt(ctx context.Context) string {
	if s.settings == nil {
		return DefaultBotSystemPrompt
	}
	v, err := s.settings.Get(ctx, settingBotSystemPrompt)
	if err != nil || strings.TrimSpace(v) == "" {
		return DefaultBotSystemPrompt
	}
	return v
}

// OfflineOnlyEnabled — флаг AI_OFFLINE_ONLY.
func (s *Service) OfflineOnlyEnabled() bool {
	return s.offlineOnly
}

// GetAdminSettings GET /admin/ai/settings.
func (s *Service) GetAdminSettings(ctx context.Context) (*AdminSettingsResponse, error) {
	return &AdminSettingsResponse{
		SystemPrompt: s.BotSystemPrompt(ctx),
		OfflineOnly:  s.offlineOnly,
	}, nil
}

// UpdateAdminSettings PATCH /admin/ai/settings.
func (s *Service) UpdateAdminSettings(ctx context.Context, req UpdateAdminSettingsRequest) (*AdminSettingsResponse, error) {
	if req.SystemPrompt != nil {
		prompt := strings.TrimSpace(*req.SystemPrompt)
		if prompt == "" {
			prompt = DefaultBotSystemPrompt
		}
		if len(prompt) > 8000 {
			return nil, ErrSettingsTooLong
		}
		if s.settings == nil {
			return nil, ErrSettingsNotConfigured
		}
		if err := s.settings.Upsert(ctx, settingBotSystemPrompt, prompt, "@hub-ai system prompt (M3-014)"); err != nil {
			return nil, err
		}
	}
	return s.GetAdminSettings(ctx)
}
