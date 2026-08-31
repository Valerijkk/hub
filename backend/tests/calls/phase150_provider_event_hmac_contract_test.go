//go:build hubtest

package calls_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	calls "github.com/valerii/hub/backend/internal/calls"
)

func TestPhase150VerifyProviderEventEnvelopeHMACSHA256Boundaries(t *testing.T) {
	secret := phase150SizedBytes(32, 0x31)

	t.Run("rejects nil secret", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertInvalidProviderEventEnvelope(t, nil, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("rejects secret shorter than 32 bytes", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertInvalidProviderEventEnvelope(t, phase150SizedBytes(31, 0x31), raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("accepts exactly 32-byte secret", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertVerifiedProviderEventEnvelope(t, secret, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("accepts secret longer than 32 bytes", func(t *testing.T) {
		longSecret := phase150SizedBytes(33, 0x31)
		raw := []byte("provider-event")
		phase150AssertVerifiedProviderEventEnvelope(t, longSecret, raw, phase150ProviderEventMAC(longSecret, raw))
	})

	t.Run("rejects empty raw envelope", func(t *testing.T) {
		raw := []byte{}
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("rejects nil raw envelope", func(t *testing.T) {
		phase150AssertInvalidProviderEventEnvelope(t, secret, nil, phase150ProviderEventMAC(secret, nil))
	})

	t.Run("accepts one-byte raw envelope", func(t *testing.T) {
		raw := []byte{0xff}
		phase150AssertVerifiedProviderEventEnvelope(t, secret, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("accepts 65536-byte raw envelope", func(t *testing.T) {
		raw := phase150SizedBytes(65536, 0x41)
		phase150AssertVerifiedProviderEventEnvelope(t, secret, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("rejects 65537-byte raw envelope", func(t *testing.T) {
		raw := phase150SizedBytes(65537, 0x41)
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("rejects MAC shorter than 32 bytes", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, phase150SizedBytes(31, 0x44))
	})

	t.Run("rejects nil MAC", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, nil)
	})

	t.Run("accepts matching 32-byte decoded digest", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertVerifiedProviderEventEnvelope(t, secret, raw, phase150ProviderEventMAC(secret, raw))
	})

	t.Run("rejects MAC longer than 32 bytes", func(t *testing.T) {
		raw := []byte("provider-event")
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, phase150SizedBytes(33, 0x44))
	})
}

func TestPhase150VerifyProviderEventEnvelopeHMACSHA256RejectsEncodedAndMutatedMACs(t *testing.T) {
	secret := phase150SizedBytes(32, 0x55)
	raw := []byte("provider-event")
	validMAC := phase150ProviderEventMAC(secret, raw)

	t.Run("rejects ASCII hex instead of decoded digest", func(t *testing.T) {
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, []byte(hex.EncodeToString(validMAC)))
	})

	t.Run("rejects ASCII base64 instead of decoded digest", func(t *testing.T) {
		phase150AssertInvalidProviderEventEnvelope(t, secret, raw, []byte(base64.StdEncoding.EncodeToString(validMAC)))
	})

	for _, mutation := range []struct {
		name  string
		apply func([]byte)
	}{
		{
			name: "one-byte mutation",
			apply: func(receivedMAC []byte) {
				receivedMAC[len(receivedMAC)-1] ^= 0x01
			},
		},
		{
			name: "whitespace mutation",
			apply: func(receivedMAC []byte) {
				if receivedMAC[0] == ' ' {
					receivedMAC[0] = '\t'
					return
				}
				receivedMAC[0] = ' '
			},
		},
		{
			name: "newline mutation",
			apply: func(receivedMAC []byte) {
				if receivedMAC[0] == '\n' {
					receivedMAC[0] = '\r'
					return
				}
				receivedMAC[0] = '\n'
			},
		},
		{
			name: "NUL mutation",
			apply: func(receivedMAC []byte) {
				if receivedMAC[0] == 0 {
					receivedMAC[0] = 0x01
					return
				}
				receivedMAC[0] = 0
			},
		},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			receivedMAC := append([]byte(nil), validMAC...)
			mutation.apply(receivedMAC)
			phase150AssertInvalidProviderEventEnvelope(t, secret, raw, receivedMAC)
		})
	}

	for _, mutation := range []struct {
		name  string
		apply func([]byte)
	}{
		{
			name: "raw one-byte mutation",
			apply: func(mutatedRaw []byte) {
				mutatedRaw[len(mutatedRaw)-1] ^= 0x01
			},
		},
		{
			name: "raw whitespace mutation",
			apply: func(mutatedRaw []byte) {
				mutatedRaw[0] = ' '
			},
		},
		{
			name: "raw newline mutation",
			apply: func(mutatedRaw []byte) {
				mutatedRaw[0] = '\n'
			},
		},
		{
			name: "raw NUL mutation",
			apply: func(mutatedRaw []byte) {
				mutatedRaw[0] = 0
			},
		},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			mutatedRaw := append([]byte(nil), raw...)
			mutation.apply(mutatedRaw)
			phase150AssertInvalidProviderEventEnvelope(t, secret, mutatedRaw, validMAC)
		})
	}
}

func TestPhase150VerifyProviderEventEnvelopeHMACSHA256TreatsEnvelopeAsOpaqueBytes(t *testing.T) {
	secret := phase150SizedBytes(32, 0x72)
	for _, testCase := range []struct {
		name string
		raw  []byte
	}{
		{name: "leading and trailing whitespace with NUL", raw: []byte{' ', '\t', 0x00, 0xff, 0x01, '{', 0x80, '\r', '\n'}},
		{name: "JSON-looking invalid bytes", raw: []byte("{not-json}")},
		{name: "gzip-looking opaque bytes", raw: []byte{0x1f, 0x8b, 0x00, 0xff}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			phase150AssertVerifiedProviderEventEnvelope(t, secret, testCase.raw, phase150ProviderEventMAC(secret, testCase.raw))
		})
	}
}

func TestPhase150VerifiedProviderEventEnvelopeKeepsBytesPrivate(t *testing.T) {
	var value calls.Phase150VerifiedProviderEventEnvelope
	valueType := reflect.TypeOf(value)
	if valueType.Kind() != reflect.Struct {
		t.Fatal("verified value is not a struct")
	}
	for index := 0; index < valueType.NumField(); index++ {
		if valueType.Field(index).IsExported() {
			t.Fatal("verified value exposes an exported field")
		}
	}

	for _, candidate := range []any{value, &value} {
		if _, ok := candidate.(fmt.Stringer); ok {
			t.Fatal("verified value unexpectedly implements fmt.Stringer")
		}
		if _, ok := candidate.(encoding.TextMarshaler); ok {
			t.Fatal("verified value unexpectedly implements encoding.TextMarshaler")
		}
		if _, ok := candidate.(json.Marshaler); ok {
			t.Fatal("verified value unexpectedly implements json.Marshaler")
		}
	}
}

func phase150AssertInvalidProviderEventEnvelope(t *testing.T, secret, rawEnvelope, receivedMAC []byte) {
	t.Helper()

	beforeSecret := append([]byte(nil), secret...)
	beforeRawEnvelope := append([]byte(nil), rawEnvelope...)
	beforeReceivedMAC := append([]byte(nil), receivedMAC...)

	result, err := calls.ExportPhase150VerifyProviderEventEnvelopeHMACSHA256(secret, rawEnvelope, receivedMAC)
	if err == nil {
		t.Fatal("invalid envelope returned nil error")
	}
	if err != calls.ErrPhase150ProviderEventEnvelopeInvalid {
		t.Fatal("invalid envelope did not return the exact non-diagnostic sentinel")
	}

	var zero calls.Phase150VerifiedProviderEventEnvelope
	if !reflect.DeepEqual(result, zero) {
		t.Fatal("invalid envelope returned a non-zero verified value")
	}
	if bytes := result.Bytes(); bytes != nil {
		t.Fatal("invalid envelope exposed verified bytes")
	}

	phase150AssertProviderEventInputsUnchanged(t, secret, beforeSecret, rawEnvelope, beforeRawEnvelope, receivedMAC, beforeReceivedMAC)
}

func phase150AssertVerifiedProviderEventEnvelope(t *testing.T, secret, rawEnvelope, receivedMAC []byte) {
	t.Helper()

	beforeSecret := append([]byte(nil), secret...)
	beforeRawEnvelope := append([]byte(nil), rawEnvelope...)
	beforeReceivedMAC := append([]byte(nil), receivedMAC...)

	result, err := calls.ExportPhase150VerifyProviderEventEnvelopeHMACSHA256(secret, rawEnvelope, receivedMAC)
	if err != nil {
		t.Fatal("valid envelope returned an error")
	}
	phase150AssertProviderEventInputsUnchanged(t, secret, beforeSecret, rawEnvelope, beforeRawEnvelope, receivedMAC, beforeReceivedMAC)

	first := result.Bytes()
	if !bytes.Equal(first, beforeRawEnvelope) {
		t.Fatal("verified bytes differ from the exact authenticated envelope")
	}

	rawEnvelope[0] ^= 0xff
	if got := result.Bytes(); !bytes.Equal(got, beforeRawEnvelope) {
		t.Fatal("verified bytes alias the caller input")
	}

	second := result.Bytes()
	first[0] ^= 0xff
	if bytes.Equal(first, second) {
		t.Fatalf("result.Bytes() calls share mutable storage")
	}
	if got := result.Bytes(); !bytes.Equal(got, beforeRawEnvelope) {
		t.Fatal("successive Bytes calls share caller-mutable storage")
	}
}

func phase150AssertProviderEventInputsUnchanged(t *testing.T, secret, beforeSecret, rawEnvelope, beforeRawEnvelope, receivedMAC, beforeReceivedMAC []byte) {
	t.Helper()

	if !bytes.Equal(secret, beforeSecret) {
		t.Fatal("secret input was mutated")
	}
	if !bytes.Equal(rawEnvelope, beforeRawEnvelope) {
		t.Fatal("raw envelope input was mutated")
	}
	if !bytes.Equal(receivedMAC, beforeReceivedMAC) {
		t.Fatal("received MAC input was mutated")
	}
}

func phase150ProviderEventMAC(secret, rawEnvelope []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(rawEnvelope)
	return mac.Sum(nil)
}

func phase150SizedBytes(length int, seed byte) []byte {
	bytes := make([]byte, length)
	for index := range bytes {
		bytes[index] = seed + byte(index%23)
	}
	return bytes
}
