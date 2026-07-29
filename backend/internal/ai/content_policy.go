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
// A3W5-2-003: расширены синонимы; второй слой — homoglyph fold + compact inserts.
var contentPolicyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore\s+(all\s+)?previous\s+instructions`),
	regexp.MustCompile(`(?i)disregard\s+(all\s+)?prior\s+instructions`),
	regexp.MustCompile(`(?i)forget\s+(all\s+)?(previous|prior)\s+instructions`),
	regexp.MustCompile(`(?i)override\s+(all\s+)?(previous|prior)\s+instructions`),
	regexp.MustCompile(`(?i)you\s+are\s+now\s+dan`),
	regexp.MustCompile(`(?i)jail\s*br[ea]{1,2}k`),
	regexp.MustCompile(`(?i)reveal\s+(your\s+)?system\s+prompt`),
	regexp.MustCompile(`(?i)bypass\s+(security|safety|filter)`),
	regexp.MustCompile(`(?i)игнорируй\s+(все\s+)?(предыдущие|прежние)\s+инструкции`),
	regexp.MustCompile(`(?i)раскрой\s+(свой\s+)?системн\w*\s+промпт`),
}

// policyHomoglyphFold — Cyrillic/Greek lookalikes → Latin (SEC-R8 residual A3W5-2-003).
// Слой поверх NFKC+ZW: «іgnоrе» → «ignore» для EN denylist; RU-паттерны бьют по оригиналу.
var policyHomoglyphFold = map[rune]rune{
	'а': 'a', 'А': 'A',
	'е': 'e', 'Е': 'E',
	'ё': 'e', 'Ё': 'E',
	'о': 'o', 'О': 'O',
	'р': 'p', 'Р': 'P',
	'с': 'c', 'С': 'C',
	'у': 'y', 'У': 'Y',
	'х': 'x', 'Х': 'X',
	'і': 'i', 'І': 'I',
	'ї': 'i', 'Ї': 'I',
	'α': 'a', 'Α': 'A',
	'ε': 'e', 'Ε': 'E',
	'ο': 'o', 'Ο': 'O',
	'ρ': 'p', 'Ρ': 'P',
	'υ': 'y', 'Υ': 'Y',
	'χ': 'x', 'Χ': 'X',
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

// foldPolicyHomoglyphs — второй слой: lookalike → Latin для EN-паттернов.
func foldPolicyHomoglyphs(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if f, ok := policyHomoglyphFold[r]; ok {
			b.WriteRune(f)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// compactPolicyInserts — снимает вставки (пунктуация/символы между буквами),
// сохраняя пробелы между словами: «i.g.n.o.r.e all previous» → «ignore all previous».
func compactPolicyInserts(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevSpace = false
			continue
		}
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// policyScanForms — кандидаты для denylist: оригинал (RU+EN), fold, fold+compact / compact.
func policyScanForms(text string) []string {
	seen := map[string]struct{}{text: {}}
	forms := []string{text}

	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		forms = append(forms, s)
	}

	folded := foldPolicyHomoglyphs(text)
	add(folded)
	add(compactPolicyInserts(folded))
	add(compactPolicyInserts(text))
	return forms
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
		for _, form := range policyScanForms(text) {
			for _, re := range contentPolicyPatterns {
				if re.MatchString(form) {
					return ErrContentPolicyViolation
				}
			}
		}
	}
	return nil
}
