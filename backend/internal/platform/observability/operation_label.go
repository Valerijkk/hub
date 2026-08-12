package observability

import (
	"strings"
	"unicode"
)

// unknownOperationLabel — bucket для неизвестных имён Timed (W4-A3-008 / P3-BE-12).
const unknownOperationLabel = "other"

const maxOperationLabelLen = 64

// allowedOperationDomains — префиксы до первой точки (tracker.CreateIssue → tracker).
// Неизвестный / user-controlled → "other", иначе unbounded Prometheus cardinality.
var allowedOperationDomains = map[string]struct{}{
	"auth": {}, "audit": {}, "tracker": {}, "messenger": {}, "project": {},
	"tms": {}, "calls": {}, "notification": {}, "search": {}, "rbac": {},
	"storage": {}, "sprint": {}, "ws": {}, "ai": {}, "cron": {}, "outbox": {},
	"nats": {}, "health": {}, "platform": {}, "bases": {}, "crm": {}, "hr": {},
	"calendar": {}, "intranet": {}, "accounting": {}, "marketplace": {},
	"github": {}, "gitlab": {}, "apigateway": {}, "mail": {}, "imex": {},
	"backup": {}, "dashboard": {}, "dev": {},
	// hubtest / unit Timed names
	"test": {}, "otel": {}, "prom": {}, "timed": {}, "void": {},
}

// sanitizeOperationLabel — allowlist по domain-префиксу; иначе "other".
func sanitizeOperationLabel(operation string) string {
	op := strings.TrimSpace(operation)
	if op == "" {
		return ""
	}
	if len(op) > maxOperationLabelLen {
		return unknownOperationLabel
	}
	dot := strings.IndexByte(op, '.')
	if dot <= 0 || dot == len(op)-1 {
		return unknownOperationLabel
	}
	domain := op[:dot]
	if _, ok := allowedOperationDomains[domain]; !ok {
		return unknownOperationLabel
	}
	rest := op[dot+1:]
	// A3W4-3-002: сегменты [A-Za-z][A-Za-z0-9_]* — без hyphen-hex UUID (cardinality).
	for _, seg := range strings.Split(rest, ".") {
		if !isValidOperationSegment(seg) {
			return unknownOperationLabel
		}
	}
	return op
}

// isValidOperationSegment — имя операции без UUID/hyphen-runs (tracker.<uuid> → other).
func isValidOperationSegment(seg string) bool {
	if seg == "" {
		return false
	}
	rs := []rune(seg)
	if !unicode.IsLetter(rs[0]) {
		return false
	}
	for _, r := range rs[1:] {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			continue
		}
		return false
	}
	return true
}
