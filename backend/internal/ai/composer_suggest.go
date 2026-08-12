package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	composerSuggestMinChars     = 8
	composerSuggestMaxTokens    = 80
	composerSuggestSystemPrompt = `You complete short chat messages for a dev team messenger.
Return ONLY the continuation text in Russian — do not repeat the prefix, no quotes, no markdown.
One short sentence or phrase max.`
)

// Sentinels для SuggestComposer — handler мапит через errors.Is (P3-BE-08 / W5-A1-008).
var (
	ErrInvalidPrefix   = fmt.Errorf("prefix must be at least %d characters", composerSuggestMinChars)
	ErrEmptySuggestion = fmt.Errorf("empty suggestion")
)

// ComposerSuggestRequest — частичный text для Tab completion (M3-024).
type ComposerSuggestRequest struct {
	Prefix  string `json:"prefix"`
	Model   string `json:"model,omitempty"`
	Channel string `json:"channel_id,omitempty"`
}

// ComposerSuggestResponse — сниппет AI-продолжения.
type ComposerSuggestResponse struct {
	Suggestion string `json:"suggestion"`
	Model      string `json:"model"`
}

// SuggestComposer дополняет partial message через провайдера.
func (s *Service) SuggestComposer(ctx context.Context, userID uuid.UUID, req ComposerSuggestRequest, ip, userAgent string) (*ComposerSuggestResponse, error) {
	prefix := strings.TrimSpace(req.Prefix)
	if len([]rune(prefix)) < composerSuggestMinChars {
		return nil, ErrInvalidPrefix
	}
	if err := CheckContentPolicy([]ChatMessage{{Role: "user", Content: prefix}}); err != nil {
		return nil, err
	}
	if err := s.enforceQuota(ctx, userID); err != nil {
		return nil, err
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		var err error
		model, err = s.DefaultModel(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	maxTok := composerSuggestMaxTokens
	temp := 0.6
	chatReq := ChatCompletionRequest{
		Model: model,
		Messages: []ChatMessage{
			{Role: "system", Content: composerSuggestSystemPrompt},
			{Role: "user", Content: prefix},
		},
		MaxTokens:   &maxTok,
		Temperature: &temp,
	}

	resp, used, latency, err := s.chatCompletionWithFallback(ctx, userID, model, chatReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty provider response")
	}
	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	suggestion := normalizeComposerCompletion(prefix, raw)
	if suggestion == "" {
		return nil, ErrEmptySuggestion
	}

	s.recordUsage(ctx, userID, used, model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, latency, ip, userAgent)
	return &ComposerSuggestResponse{Suggestion: suggestion, Model: model}, nil
}

func normalizeComposerCompletion(prefix, raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, `"'`)
	if s == "" {
		return ""
	}
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return strings.TrimSpace(s)
	}
	// A3W5-3-001: case/UTF-8-safe trim — не s[len(prefix):] после ToLower (байты ≠ руны).
	if end := equalFoldPrefixLen(s, prefix); end > 0 {
		s = strings.TrimSpace(s[end:])
	}
	return strings.TrimSpace(s)
}

// equalFoldPrefixLen возвращает байтовый индекс конца префикса в s при EqualFold-совпадении,
// иначе 0. Сравнение по рунам — безопасно для кириллицы/mixed-case.
func equalFoldPrefixLen(s, prefix string) int {
	pr := []rune(prefix)
	sr := []rune(s)
	if len(pr) == 0 || len(sr) < len(pr) {
		return 0
	}
	for i := range pr {
		if !strings.EqualFold(string(pr[i]), string(sr[i])) {
			return 0
		}
	}
	return len(string(sr[:len(pr)]))
}
