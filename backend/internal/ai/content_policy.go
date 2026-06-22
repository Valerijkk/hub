package ai

import (
	"fmt"
	"regexp"
	"strings"
)

// ErrContentPolicyViolation — заблокированный prompt (SEC-R8 / AI-SPEC S7).
var ErrContentPolicyViolation = fmt.Errorf("message blocked by content policy")

var contentPolicyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions`),
	regexp.MustCompile(`(?i)disregard\s+(all\s+)?prior\s+instructions`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+DAN`),
	regexp.MustCompile(`(?i)jailbreak`),
	regexp.MustCompile(`(?i)reveal\s+(your\s+)?system\s+prompt`),
	regexp.MustCompile(`(?i)bypass\s+(security|safety|filter)`),
}

// CheckContentPolicy сканирует messages на injection/abuse (SEC-R8).
func CheckContentPolicy(messages []ChatMessage) error {
	for _, m := range messages {
		if m.Role != "user" {
			continue
		}
		text := strings.TrimSpace(m.Content)
		if text == "" {
			continue
		}
		for _, re := range contentPolicyPatterns {
			if re.MatchString(text) {
				return ErrContentPolicyViolation
			}
		}
	}
	return nil
}
