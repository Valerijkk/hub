//go:build qa && !release

package qamailcalendar

import (
	"crypto/subtle"
	"encoding/binary"
	"errors"
)

const (
	EnvQAMailCalendarToken = "HUB_QA_MAIL_CALENDAR_TOKEN"
	HeaderHubQAToken        = "X-Hub-QA-Token"
)

const (
	ScenarioInboundRequestV1         Scenario = "inbound-request-v1"
	ScenarioAccountMaterializationV1 Scenario = "account-materialization-v1"
)

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
	var configuredLength [8]byte
	var presentedLength [8]byte
	var expectedLength [8]byte
	copy(configured[:], c.token)
	copy(presented[:], candidate)

	binary.LittleEndian.PutUint64(configuredLength[:], uint64(len(c.token)))
	binary.LittleEndian.PutUint64(presentedLength[:], uint64(len(candidate)))
	binary.LittleEndian.PutUint64(expectedLength[:], uint64(len(configured)))
	configuredLengthOK := subtle.ConstantTimeCompare(configuredLength[:], expectedLength[:])
	presentedLengthOK := subtle.ConstantTimeCompare(presentedLength[:], expectedLength[:])
	equal := subtle.ConstantTimeCompare(configured[:], presented[:])

	return configuredLengthOK&presentedLengthOK&equal == 1
}

func ParseScenario(value string) (Scenario, error) {
	switch Scenario(value) {
	case ScenarioInboundRequestV1:
		return ScenarioInboundRequestV1, nil
	case ScenarioAccountMaterializationV1:
		return ScenarioAccountMaterializationV1, nil
	default:
		return "", errInvalidScenario
	}
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
