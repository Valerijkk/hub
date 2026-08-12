// Пакет crypto — AES-256-GCM для секретов at-rest (HI-05).
// (токены интеграций, webhook secrets и т.д.).
//
// Формат Encrypt/Decrypt:
//
//	base64-std( nonce || ciphertext-with-gcm-tag )
//
// 12-байтный GCM nonce в начале; ciphertext-with-tag — результат
// (*cipher.AEAD).Seal с 16-байтным GCM auth tag в конце
// — одна строка удобна для TEXT-колонки в Postgres.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// nonceSize для AES-GCM по стандарту NIST.
const nonceSize = 12

// ErrCipherKeyNotConfigured — глобальный cipher не инициализирован;
// Encrypt/Decrypt вызывать нельзя — panic в prod недопустим.
var ErrCipherKeyNotConfigured = errors.New("integration encryption key not configured")

// Cipher оборачивает AES-256-GCM AEAD для повторных encrypt/decrypt.
// Создавать через NewCipherFromEnv/NewCipherFromKey; zero value невалиден.
type Cipher struct {
	aead cipher.AEAD
}

// NewCipherFromKey строит Cipher из 32-байтного ключа AES-256.
func NewCipherFromKey(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("AES-256-GCM key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new GCM: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// DecodeKeyMaterial декодирует 32-байтовый AES-ключ из hex или base64
// (std/url-safe, с padding или без).
func DecodeKeyMaterial(raw string) ([]byte, error) {
	if raw == "" {
		return nil, ErrCipherKeyNotConfigured
	}
	// Сначала hex — наиболее частый формат ключа в prod.
	if len(raw) == 64 {
		if key, err := hex.DecodeString(raw); err == nil && len(key) == 32 {
			return key, nil
		}
	}
	// Затем base64 (std, потом url) — для ключей из секрет-хранилищ.
	if key, err := base64.StdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := base64.URLEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	if key, err := base64.RawURLEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return key, nil
	}
	return nil, fmt.Errorf("integration encryption key must be 32 bytes encoded as hex or base64")
}

// NewCipherFromString принимает ключ как 64-char hex или base64
// (std/url-safe, с padding или без) и возвращает Cipher.
// Декодированный ключ — ровно 32 байта (AES-256).
func NewCipherFromString(raw string) (*Cipher, error) {
	key, err := DecodeKeyMaterial(raw)
	if err != nil {
		return nil, err
	}
	return NewCipherFromKey(key)
}

// DeriveDevKey — детерминированный ключ ТОЛЬКО для DevMode.
// SHA-256 от фиксированной строки — dev не падает при старте
// без INTEGRATION_ENCRYPTION_KEY; ключ ЗАПРЕЩЁН в
// production — см. config.validate / IsDevDerivedKey.
func DeriveDevKey() []byte {
	sum := sha256.Sum256([]byte("hub-dev-integration-encryption-key-v1"))
	return sum[:]
}

// IsDevDerivedKey — true, если raw декодируется в DeriveDevKey().
// W4-A4-009: config.validate отклоняет такой ключ при !DevMode.
func IsDevDerivedKey(raw string) bool {
	key, err := DecodeKeyMaterial(strings.TrimSpace(raw))
	if err != nil || len(key) != 32 {
		return false
	}
	return subtle.ConstantTimeCompare(key, DeriveDevKey()) == 1
}

// Encrypt шифрует с новым random nonce и возвращает одну
// base64-std строку: base64( nonce || ciphertext-with-tag ).
// AAD пустой — legacy-путь (до A2W4-4-001). Новые записи с привязкой
// к сущности — EncryptWithAAD(..., entityID).
func (c *Cipher) Encrypt(plaintext []byte) (string, error) {
	return c.EncryptWithAAD(plaintext, nil)
}

// EncryptWithAAD — AES-GCM Seal с additional authenticated data.
// A2W4-4-001 / P3-BE-U02: AAD=entityID (UUID bytes/string) привязывает
// ciphertext к строке БД — swap blob между rows падает на Open.
func (c *Cipher) EncryptWithAAD(plaintext, aad []byte) (string, error) {
	if c == nil || c.aead == nil {
		return "", ErrCipherKeyNotConfigured
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, aad)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// EncryptString — обёртка для string plaintext (legacy AAD=nil).
func (c *Cipher) EncryptString(s string) (string, error) {
	return c.Encrypt([]byte(s))
}

// EncryptStringWithAAD — string plaintext + AAD (обычно entity UUID).
func (c *Cipher) EncryptStringWithAAD(s string, aad []byte) (string, error) {
	return c.EncryptWithAAD([]byte(s), aad)
}

// Decrypt отменяет Encrypt (AAD=nil); ошибка при невалидном base64,
// длине < nonce или провале GCM auth (подмена ciphertext
// или legacy plaintext без шифрования — callers используют для
// обнаружения legacy plaintext; см. HI-05).
func (c *Cipher) Decrypt(input string) ([]byte, error) {
	return c.DecryptWithAAD(input, nil)
}

// DecryptWithAAD отменяет EncryptWithAAD с тем же AAD.
func (c *Cipher) DecryptWithAAD(input string, aad []byte) ([]byte, error) {
	if c == nil || c.aead == nil {
		return nil, ErrCipherKeyNotConfigured
	}
	raw, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	if len(raw) < nonceSize+c.aead.Overhead() {
		return nil, errors.New("ciphertext too short")
	}
	nonce := raw[:nonceSize]
	ct := raw[nonceSize:]
	pt, err := c.aead.Open(nil, nonce, ct, aad)
	if err != nil {
		return nil, fmt.Errorf("gcm open: %w", err)
	}
	return pt, nil
}

// DecryptString возвращает plaintext как string (legacy AAD=nil).
func (c *Cipher) DecryptString(input string) (string, error) {
	b, err := c.Decrypt(input)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecryptStringWithAAD — DecryptWithAAD → string.
func (c *Cipher) DecryptStringWithAAD(input string, aad []byte) (string, error) {
	b, err := c.DecryptWithAAD(input, aad)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MaybeDecryptString пробует decrypt с AAD=nil; при ошибке (напр.
// legacy plaintext до HI-05) возвращает
// input как есть и ok=false — caller решает,
// re-encrypt при save. ok=true — успешный decrypt.
//
// NOTE: намеренно мягко — ТОЛЬКО для migration window legacy ciphertext
// без AAD. Новые записи — EncryptWithAAD(entityID). После полной
// миграции callers должны читать через DecryptWithAAD и не полагаться
// на soft-passthrough (игнорировать ok запрещено).
func (c *Cipher) MaybeDecryptString(input string) (plaintext string, ok bool) {
	if input == "" {
		return "", false
	}
	s, err := c.DecryptString(input)
	if err != nil {
		return input, false
	}
	return s, true
}
