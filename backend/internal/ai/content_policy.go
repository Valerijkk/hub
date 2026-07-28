package ai

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// ErrContentPolicyViolation — заблокированный prompt (SEC-R8 / AI-SPEC S7).
var ErrContentPolicyViolation = fmt.Errorf("message blocked by content policy")

// contentPolicyPatterns — EN+RU denylist после NFKC/ZW-нормализации (A2W5-A2-006).
var contentPolicyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions`),
	regexp.MustCompile(`(?i)disregard\s+(all\s+)?prior\s+instructions`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+dan`),
	regexp.MustCompile(`(?i)jail\s*br[ea]{1,2}k`),
	regexp.MustCompile(`(?i)reveal\s+(your\s+)?system\s+prompt`),
	regexp.MustCompile(`(?i)bypass\s+(security|safety|filter)`),
	regexp.MustCompile(`(?i)игнорируй\s+(все\s+)?(предыдущие|прежние)\s+инструкции`),
	regexp.MustCompile(`(?i)раскрой\s+(свой\s+)?системн\w*\s+промпт`),
}

// normalizePolicyText — NFKC + снятие zero-width / format chars (обход denylist).
func normalizePolicyText(s string) string {
	s = norm.NFKC.String(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.Is(unicode.Cf, r) { // format incl. ZWSP/ZWNJ/ZWJ/BOM
			continue
		}
		switch r {
		case '\u200b', '\u200c', '\u200d', '\ufeff', '\u2060', '\u180e':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// CheckContentPolicy сканирует messages на injection/abuse (SEC-R8).
// P1-BE-29: проверяются все роли allowlist (user|assistant|system), не только user —
// иначе jailbreak через role:"system" обходил фильтр.
func CheckContentPolicy(messages []ChatMessage) error {
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		switch role {
		case "user", "assistant", "system":
		default:
			return fmt.Errorf("%w: invalid role %q", ErrContentPolicyViolation, m.Role)
		}
		text := strings.TrimSpace(normalizePolicyText(m.Content))
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
