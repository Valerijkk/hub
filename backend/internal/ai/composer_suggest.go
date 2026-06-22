package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	composerSuggestMinChars   = 8
	composerSuggestMaxTokens  = 80
	composerSuggestSystemPrompt = `You complete short chat messages for a dev team messenger.
Return ONLY the continuation text in Russian — do not repeat the prefix, no quotes, no markdown.
One short sentence or phrase max.`
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
		return nil, fmt.Errorf("prefix must be at least %d characters", composerSuggestMinChars)
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
		return nil, fmt.Errorf("empty suggestion")
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
	if strings.HasPrefix(s, prefix) {
		s = strings.TrimSpace(strings.TrimPrefix(s, prefix))
	} else {
		lowerPrefix := strings.ToLower(strings.TrimSpace(prefix))
		if strings.HasPrefix(strings.ToLower(s), lowerPrefix) {
			s = strings.TrimSpace(s[len(prefix):])
		}
	}
	return strings.TrimSpace(s)
}
