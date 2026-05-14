// Package crypto provides AES-256-GCM encryption helpers for at-rest secrets
// (integration access tokens, webhook secrets, etc.).
//
// Output format of Encrypt / input format of Decrypt:
//
//	base64-std( nonce || ciphertext-with-gcm-tag )
//
// The 12-byte GCM nonce is prepended; ciphertext-with-tag is what
// (*cipher.AEAD).Seal returns and contains the 16-byte GCM auth tag at the
// end. A single string is convenient for storing in a TEXT column.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// nonceSize for AES-GCM as required by the standard.
const nonceSize = 12

// ErrCipherKeyNotConfigured indicates that the global cipher has not been
// initialised — Encrypt and Decrypt must not be called.
var ErrCipherKeyNotConfigured = errors.New("integration encryption key not configured")

// Cipher wraps an AES-256-GCM AEAD for repeated encrypt/decrypt operations.
// Construct via NewCipherFromEnv / NewCipherFromKey. The zero value is invalid.
type Cipher struct {
	aead cipher.AEAD
}

// NewCipherFromKey builds a Cipher from a raw 32-byte key.
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

// NewCipherFromString accepts a key as either 64-char hex or base64
// (std or url-safe, with or without padding) and returns a Cipher.
// The decoded key must be exactly 32 bytes (AES-256).
func NewCipherFromString(raw string) (*Cipher, error) {
	if raw == "" {
		return nil, ErrCipherKeyNotConfigured
	}
	// Try hex first.
	if len(raw) == 64 {
		if key, err := hex.DecodeString(raw); err == nil && len(key) == 32 {
			return NewCipherFromKey(key)
		}
	}
	// Then base64 (std then url).
	if key, err := base64.StdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return NewCipherFromKey(key)
	}
	if key, err := base64.RawStdEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return NewCipherFromKey(key)
	}
	if key, err := base64.URLEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return NewCipherFromKey(key)
	}
	if key, err := base64.RawURLEncoding.DecodeString(raw); err == nil && len(key) == 32 {
		return NewCipherFromKey(key)
	}
	return nil, fmt.Errorf("integration encryption key must be 32 bytes encoded as hex or base64")
}

// DeriveDevKey returns a deterministic 32-byte key suitable ONLY for DevMode.
// It hashes a fixed string with SHA-256 so dev environments don't fail-start
// when INTEGRATION_ENCRYPTION_KEY is unset. The key MUST NOT be used in
// production — see config.validate.
func DeriveDevKey() []byte {
	sum := sha256.Sum256([]byte("hub-dev-integration-encryption-key-v1"))
	return sum[:]
}

// Encrypt encrypts plaintext with a fresh random nonce and returns a single
// base64-std string: base64( nonce || ciphertext-with-tag ).
func (c *Cipher) Encrypt(plaintext []byte) (string, error) {
	if c == nil || c.aead == nil {
		return "", ErrCipherKeyNotConfigured
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	sealed := c.aead.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(nonce)+len(sealed))
	out = append(out, nonce...)
	out = append(out, sealed...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// EncryptString is a convenience helper for string plaintexts.
func (c *Cipher) EncryptString(s string) (string, error) {
	return c.Encrypt([]byte(s))
}

// Decrypt reverses Encrypt. Returns an error if the input is not valid base64,
// is shorter than the nonce, or fails GCM authentication (tampered ciphertext
// OR legitimate plaintext that was never encrypted — callers can use that to
// detect legacy plaintext rows; see HI-05).
func (c *Cipher) Decrypt(input string) ([]byte, error) {
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
	pt, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, fmt.Errorf("gcm open: %w", err)
	}
	return pt, nil
}

// DecryptString returns the plaintext as a string.
func (c *Cipher) DecryptString(input string) (string, error) {
	b, err := c.Decrypt(input)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MaybeDecryptString attempts to decrypt input. If decryption fails (e.g. the
// stored value is legacy plaintext from before HI-05 was applied), it returns
// the input verbatim along with ok=false so the caller can decide whether to
// re-encrypt on next save. ok=true means a real decryption happened.
//
// NOTE: this is intentionally lenient and is intended ONLY for the migration
// window. New writes must always go through Encrypt. The check is a one-way
// signal — once a value passes GCM auth, it is, by definition, a previously
// encrypted blob.
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
