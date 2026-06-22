package schema

import (
	"regexp"
	"time"
)

// keyRegexp — ключ проекта: заглавная + 1-9 заглавных/цифр.
var keyRegexp = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

// nowUTC — текущее время UTC; канонический default/update
// hook для timestamp в Ent — Postgres всегда получает
// UTC независимо от локального TZ сервера.
func nowUTC() time.Time { return time.Now().UTC() }
