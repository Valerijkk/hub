package apigateway

import (
	"encoding/hex"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignPayload_DeterministicHMAC проверяет, что один и тот же (payload, secret)
// всегда даёт одну и ту же подпись. Это базовое свойство HMAC, нарушение которого
// сделало бы невозможной верификацию подписи получателем.
func TestSignPayload_DeterministicHMAC(t *testing.T) {
	cases := []struct {
		name    string
		payload []byte
		secret  string
	}{
		{"simple json", []byte(`{"event":"ping"}`), "secret-1"},
		{"empty body", []byte(``), "secret-1"},
		{"binary-ish body", []byte{0x00, 0xff, 0x10, 0x80}, "another-secret"},
		{"long body", []byte(`{"data":"` + string(make([]byte, 2048)) + `"}`), "k"},
		{"unicode body", []byte(`{"msg":"привет 🚀"}`), "ключ"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sig1 := signPayload(tc.payload, tc.secret)
			sig2 := signPayload(tc.payload, tc.secret)
			sig3 := signPayload(tc.payload, tc.secret)
			assert.Equal(t, sig1, sig2, "signature must be deterministic across calls")
			assert.Equal(t, sig2, sig3, "signature must be deterministic across calls")
			require.NotEmpty(t, sig1, "signature must not be empty")
		})
	}
}

// TestSignPayload_DifferentSecretChangesSig проверяет, что подпись зависит от
// секрета. Регресс этого свойства означал бы, что злоумышленник со старым
// (отозванным) секретом может подделать любую подпись.
func TestSignPayload_DifferentSecretChangesSig(t *testing.T) {
	payload := []byte(`{"event":"issue.created","id":42}`)

	sigA := signPayload(payload, "secret-A")
	sigB := signPayload(payload, "secret-B")
	sigC := signPayload(payload, "Secret-A") // case-sensitive

	assert.NotEqual(t, sigA, sigB, "different secrets must yield different signatures")
	assert.NotEqual(t, sigA, sigC, "secret comparison must be case-sensitive")
	assert.NotEqual(t, sigB, sigC)

	// Пустой секрет тоже даёт детерминированную, но другую подпись.
	sigEmpty := signPayload(payload, "")
	assert.NotEqual(t, sigA, sigEmpty)
	assert.NotEqual(t, sigB, sigEmpty)
}

// TestSignPayload_DifferentBodyChangesSig проверяет, что любая модификация
// тела меняет подпись (целостность). Это ключ к защите от tampering.
func TestSignPayload_DifferentBodyChangesSig(t *testing.T) {
	secret := "shared-secret"

	sigOrig := signPayload([]byte(`{"event":"a","payload":1}`), secret)
	// Изменение одного символа.
	sigEdit := signPayload([]byte(`{"event":"a","payload":2}`), secret)
	// Изменение пробела/форматирования (важно — подпись по байтам, не по семантике JSON).
	sigSpace := signPayload([]byte(`{"event":"a", "payload":1}`), secret)
	// Префиксное расширение.
	sigPrefix := signPayload([]byte(` {"event":"a","payload":1}`), secret)

	assert.NotEqual(t, sigOrig, sigEdit, "changing payload value must change signature")
	assert.NotEqual(t, sigOrig, sigSpace, "byte-level change must change signature")
	assert.NotEqual(t, sigOrig, sigPrefix, "leading whitespace must change signature")
}

// TestSignPayload_HexLowercase проверяет, что выход signPayload — это
// строго [0-9a-f]{64}: 32-байтовый HMAC-SHA256 в lowercase hex. Это контракт,
// на который полагается верифицирующая сторона (нельзя сравнивать со строкой
// в другом регистре без явного нижнего-регистра).
func TestSignPayload_HexLowercase(t *testing.T) {
	sig := signPayload([]byte(`{"event":"test"}`), "k")

	require.Len(t, sig, 64, "HMAC-SHA256 hex must be exactly 64 characters")

	re := regexp.MustCompile(`^[0-9a-f]{64}$`)
	assert.Regexp(t, re, sig, "signature must contain only [0-9a-f]")

	// Дополнительный sanity-check: hex.DecodeString должен пройти, и результат — 32 байта.
	raw, err := hex.DecodeString(sig)
	require.NoError(t, err)
	assert.Len(t, raw, 32, "decoded HMAC-SHA256 must be 32 bytes")
}
