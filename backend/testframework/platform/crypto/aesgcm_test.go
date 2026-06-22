package crypto_test
import (
	"github.com/valerii/hub/backend/internal/platform/crypto"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := crypto.DeriveDevKey()
	c, err := crypto.NewCipherFromKey(key)
	if err != nil {
		t.Fatalf("crypto.NewCipherFromKey: %v", err)
	}

	cases := []string{
		"",
		"a",
		"ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		strings.Repeat("x", 4096),
	}
	for _, tc := range cases {
		ct, err := c.EncryptString(tc)
		if err != nil {
			t.Fatalf("EncryptString(%q): %v", tc, err)
		}
		pt, err := c.DecryptString(ct)
		if err != nil {
			t.Fatalf("DecryptString(%q): %v", tc, err)
		}
		if pt != tc {
			t.Fatalf("roundtrip mismatch: got %q, want %q", pt, tc)
		}
	}
}

func TestNonceUniqueness(t *testing.T) {
	c, _ := crypto.NewCipherFromKey(crypto.DeriveDevKey())
	ct1, _ := c.EncryptString("same-plaintext")
	ct2, _ := c.EncryptString("same-plaintext")
	if ct1 == ct2 {
		t.Fatal("identical ciphertexts for same plaintext — nonce reuse")
	}
}

func TestDecryptRejectsTamperedCiphertext(t *testing.T) {
	c, _ := crypto.NewCipherFromKey(crypto.DeriveDevKey())
	ct, _ := c.EncryptString("secret")
	raw, _ := base64.StdEncoding.DecodeString(ct)
	raw[len(raw)-1] ^= 0xff // flip a tag byte
	tampered := base64.StdEncoding.EncodeToString(raw)
	if _, err := c.DecryptString(tampered); err == nil {
		t.Fatal("expected error on tampered ciphertext, got nil")
	}
}

func TestNewCipherFromString_AcceptsHexAndBase64(t *testing.T) {
	raw := crypto.DeriveDevKey()
	cases := []string{
		hex.EncodeToString(raw),
		base64.StdEncoding.EncodeToString(raw),
		base64.RawStdEncoding.EncodeToString(raw),
		base64.URLEncoding.EncodeToString(raw),
		base64.RawURLEncoding.EncodeToString(raw),
	}
	for _, k := range cases {
		if _, err := crypto.NewCipherFromString(k); err != nil {
			t.Fatalf("crypto.NewCipherFromString(%q): %v", k, err)
		}
	}
}

func TestNewCipherFromString_RejectsEmpty(t *testing.T) {
	if _, err := crypto.NewCipherFromString(""); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestMaybeDecryptString_LegacyPlaintextPassthrough(t *testing.T) {
	c, _ := crypto.NewCipherFromKey(crypto.DeriveDevKey())
	const legacy = "ghp_legacyTokenStoredAsPlaintext"
	got, ok := c.MaybeDecryptString(legacy)
	if ok {
		t.Fatalf("expected ok=false for non-ciphertext, got ok=true")
	}
	if got != legacy {
		t.Fatalf("expected passthrough, got %q", got)
	}

	encrypted, _ := c.EncryptString(legacy)
	got2, ok := c.MaybeDecryptString(encrypted)
	if !ok {
		t.Fatalf("expected ok=true for valid ciphertext")
	}
	if got2 != legacy {
		t.Fatalf("expected roundtrip, got %q", got2)
	}
}
