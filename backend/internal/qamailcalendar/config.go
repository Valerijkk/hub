//go:build qa && !release

package qamailcalendar

import (
	"crypto/subtle"
	"errors"
)

const (
	EnvQAMailCalendarToken = "HUB_QA_MAIL_CALENDAR_TOKEN"
	HeaderHubQAToken        = "X-Hub-QA-Token"
)

const ScenarioInboundRequestV1 Scenario = "inbound-request-v1"

var (
	errNilLookup       = errors.New("qa mail calendar environment lookup is nil")
	errInvalidToken    = errors.New("qa mail calendar control token is missing or invalid")
	errInvalidScenario = errors.New("qa mail calendar scenario is invalid")
)

type LookupEnv func(string) (string, bool)

type Config struct {
	token []byte
}

type Scenario string

func ParseConfig(lookup LookupEnv) (Config, error) {
	if lookup == nil {
		return Config{}, errNilLookup
	}

	value, found := lookup(EnvQAMailCalendarToken)
	if !found || !validToken(value) {
		return Config{}, errInvalidToken
	}

	token := make([]byte, len(value))
	copy(token, value)

	return Config{token: token}, nil
}

func (c Config) Authorized(candidate string) bool {
	var configured [64]byte
	var presented [64]byte
	copy(configured[:], c.token)
	copy(presented[:], candidate)

	configuredLengthOK := subtle.ConstantTimeEq(int32(len(c.token)), int32(len(configured)))
	presentedLengthOK := subtle.ConstantTimeEq(int32(len(candidate)), int32(len(presented)))
	equal := subtle.ConstantTimeCompare(configured[:], presented[:])

	return configuredLengthOK&presentedLengthOK&equal == 1
}

func ParseScenario(value string) (Scenario, error) {
	if value != string(ScenarioInboundRequestV1) {
		return "", errInvalidScenario
	}

	return ScenarioInboundRequestV1, nil
}

func ValidNonce(value string) bool {
	return validLowerHex(value, 32)
}

func validToken(value string) bool {
	return validLowerHex(value, 64)
}

func validLowerHex(value string, length int) bool {
	if len(value) != length {
		return false
	}

	for i := range value {
		character := value[i]
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}

	return true
}
