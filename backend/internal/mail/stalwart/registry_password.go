package stalwart

import (
	"bytes"
	"context"
	"encoding/json"
	"unicode"
	"unicode/utf8"
)

const (
	registryPasswordMethod = "x:Account/set"
	registryPasswordCallID = "set-user-password"
	registryPasswordSlot   = "0"

	minimumRegistryPasswordBytes = 12
	maximumRegistryPasswordBytes = 256
)

// RegistryPasswordClient narrows a constructed Registry client to the single
// account-password mutation required by the Hub provisioning boundary. It has
// no account discovery, creation, deletion, role, permission, or API-key API.
type RegistryPasswordClient struct {
	registry *Client
}

// NewRegistryPasswordClient binds password mutation to an already constructed
// Registry transport. Client.CallRegistry always uses the fixed POST /jmap
// Registry endpoint and its separate RegistryAuthorizer.
func NewRegistryPasswordClient(client *Client) (*RegistryPasswordClient, error) {
	if client == nil {
		return nil, newError(ErrorInvalidConfiguration)
	}
	return &RegistryPasswordClient{registry: client}, nil
}

// SetUserPassword replaces an already authorized account's credentials with
// exactly one Stalwart Password credential. The caller retains its password
// buffer; this method clears every secret or request JSON buffer it owns before
// returning. Errors are always the existing detail-free transport errors.
func (client *RegistryPasswordClient) SetUserPassword(ctx context.Context, accountID string, secret []byte) error {
	if client == nil || client.registry == nil || ctx == nil || !validOpaque(accountID) || !validRegistryPassword(secret) {
		return newError(ErrorInvalidRequest)
	}
	if ctx.Err() != nil {
		return newError(ErrorIndeterminate)
	}

	ownedSecret := append([]byte(nil), secret...)
	defer clearRegistryPasswordBytes(ownedSecret)

	arguments := makeRegistryPasswordArguments(accountID, ownedSecret)
	defer clearRegistryPasswordBytes(arguments)

	response, err := client.registry.CallRegistry(ctx, JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name:      registryPasswordMethod,
			Arguments: json.RawMessage(arguments),
			CallID:    registryPasswordCallID,
		}},
	})
	if err != nil {
		return err
	}
	return requireRegistryPasswordUpdated(response, accountID)
}

func validRegistryPassword(secret []byte) bool {
	if len(secret) < minimumRegistryPasswordBytes || len(secret) > maximumRegistryPasswordBytes || !utf8.Valid(secret) {
		return false
	}
	for len(secret) > 0 {
		character, size := utf8.DecodeRune(secret)
		if size == 0 || character == utf8.RuneError && size == 1 || unicode.IsControl(character) || (!unicode.IsGraphic(character) && !unicode.IsSpace(character)) {
			return false
		}
		secret = secret[size:]
	}
	return true
}

func makeRegistryPasswordArguments(accountID string, secret []byte) []byte {
	arguments := make([]byte, 0, len(accountID)+len(secret)+96)
	arguments = append(arguments, `{"update":{`...)
	arguments = appendRegistryJSONString(arguments, []byte(accountID))
	arguments = append(arguments, `:{"credentials":{"`...)
	arguments = append(arguments, registryPasswordSlot...)
	arguments = append(arguments, `":{"@type":"Password","secret":`...)
	arguments = appendRegistryJSONString(arguments, secret)
	arguments = append(arguments, `}}}}`...)
	return arguments
}

func appendRegistryJSONString(destination, value []byte) []byte {
	destination = append(destination, '"')
	for _, character := range value {
		switch character {
		case '"':
			destination = append(destination, '\\', '"')
		case '\\':
			destination = append(destination, '\\', '\\')
		default:
			destination = append(destination, character)
		}
	}
	return append(destination, '"')
}

func requireRegistryPasswordUpdated(response JMAPResponse, accountID string) error {
	if len(response.MethodResponses) != 1 {
		return newError(ErrorProtocol)
	}
	method := response.MethodResponses[0]
	if method.Name != registryPasswordMethod || method.CallID != registryPasswordCallID || !validJSONObject(method.Arguments) {
		return newError(ErrorProtocol)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(method.Arguments, &fields); err != nil {
		return newError(ErrorProtocol)
	}
	updated, present := fields["updated"]
	if !present || len(fields["notUpdated"]) != 0 {
		return newError(ErrorProtocol)
	}

	var updatedIDs map[string]json.RawMessage
	if err := json.Unmarshal(updated, &updatedIDs); err != nil || len(updatedIDs) != 1 {
		return newError(ErrorProtocol)
	}
	result, present := updatedIDs[accountID]
	if !present || !bytes.Equal(bytes.TrimSpace(result), []byte("null")) {
		return newError(ErrorProtocol)
	}
	return nil
}

func clearRegistryPasswordBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
