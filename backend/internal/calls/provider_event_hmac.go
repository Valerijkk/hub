package calls

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

// ErrPhase150ProviderEventEnvelopeInvalid marks an invalid provider-event envelope.
var ErrPhase150ProviderEventEnvelopeInvalid = errors.New("invalid provider event envelope")

// Phase150VerifiedProviderEventEnvelope holds authenticated opaque provider-event bytes.
type Phase150VerifiedProviderEventEnvelope struct {
	rawEnvelope []byte
}

// Bytes returns a copy of the authenticated opaque provider-event bytes.
func (value Phase150VerifiedProviderEventEnvelope) Bytes() []byte {
	return append([]byte(nil), value.rawEnvelope...)
}

// ExportPhase150VerifyProviderEventEnvelopeHMACSHA256 verifies an HMAC-SHA256
// signature for an opaque provider-event envelope.
func ExportPhase150VerifyProviderEventEnvelopeHMACSHA256(
	secret []byte,
	rawEnvelope []byte,
	receivedMAC []byte,
) (Phase150VerifiedProviderEventEnvelope, error) {
	if len(secret) < 32 {
		return Phase150VerifiedProviderEventEnvelope{}, ErrPhase150ProviderEventEnvelopeInvalid
	}
	if len(rawEnvelope) < 1 || len(rawEnvelope) > 65536 {
		return Phase150VerifiedProviderEventEnvelope{}, ErrPhase150ProviderEventEnvelopeInvalid
	}
	if len(receivedMAC) != sha256.Size {
		return Phase150VerifiedProviderEventEnvelope{}, ErrPhase150ProviderEventEnvelopeInvalid
	}

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(rawEnvelope)
	if !hmac.Equal(mac.Sum(nil), receivedMAC) {
		return Phase150VerifiedProviderEventEnvelope{}, ErrPhase150ProviderEventEnvelopeInvalid
	}

	return Phase150VerifiedProviderEventEnvelope{
		rawEnvelope: append([]byte(nil), rawEnvelope...),
	}, nil
}
