package stalwart

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	registryAccountResponseMaximumBytes = 16 << 10
	registryAccountLocalPartMaximumBytes = 64
	registryAccountEmailMaximumBytes     = 320
	registryAccountDomainMaximumBytes    = 253

	registryAccountQueryCallID = "registry-account-query"
	registryAccountGetCallID   = "registry-account-get"
	registryAccountSetCallID   = "registry-account-set"
	registryAccountCreationID  = "hub-local-user"
)

// LocalUserAccountSpec contains the already-authorized local identity needed
// to ensure one Stalwart User account. DomainID is an opaque Registry object
// identifier. DomainName is optional and is used only to validate Stalwart's
// server-set address when composition has supplied a safe canonical name.
type LocalUserAccountSpec struct {
	EmailLocalPart string
	DomainID       string
	DomainName     string
}

// LocalUserAccount is the narrow server-confirmed account binding retained by
// the caller. It never exposes credentials or Registry response data.
type LocalUserAccount struct {
	AccountID    string
	EmailAddress string
}

// RegistryAccountClient owns the narrow Registry account ensure protocol. It
// uses Client only through the Registry context, so mailbox credentials and
// mailbox JMAP methods cannot enter this operation.
type RegistryAccountClient struct {
	client *Client
}

// NewRegistryAccountClient constructs a Registry-only account codec. Client
// construction already validates the trusted origin and both authorizer slots.
func NewRegistryAccountClient(client *Client) (*RegistryAccountClient, error) {
	if client == nil {
		return nil, newError(ErrorInvalidConfiguration)
	}
	return &RegistryAccountClient{client: client}, nil
}

// EnsureLocalUserAccount returns the exact existing local User account or
// creates the documented credential-free User shape once. It does not retry a
// mutation: an uncertain create result remains indeterminate for the caller.
// A known write-conflict is resolved by one read-only re-query of the winner.
func (client *RegistryAccountClient) EnsureLocalUserAccount(ctx context.Context, spec LocalUserAccountSpec) (LocalUserAccount, error) {
	if client == nil || client.client == nil {
		return LocalUserAccount{}, newError(ErrorInvalidConfiguration)
	}
	if err := registryAccountContextError(ctx); err != nil {
		return LocalUserAccount{}, err
	}
	if !validRegistryAccountSpec(spec) {
		return LocalUserAccount{}, newError(ErrorInvalidRequest)
	}

	accountID, found, err := client.queryLocalUserAccount(ctx, spec)
	if err != nil {
		return LocalUserAccount{}, err
	}
	if found {
		return client.getLocalUserAccount(ctx, accountID, spec)
	}

	createdID, err := client.createLocalUserAccount(ctx, spec)
	if err != nil {
		if !errors.Is(err, ErrStateMismatch) {
			return LocalUserAccount{}, err
		}
		recoveredID, recovered, recoveryErr := client.queryLocalUserAccount(ctx, spec)
		if recoveryErr != nil {
			return LocalUserAccount{}, recoveryErr
		}
		if !recovered {
			return LocalUserAccount{}, err
		}
		return client.getLocalUserAccount(ctx, recoveredID, spec)
	}
	return client.getLocalUserAccount(ctx, createdID, spec)
}

func (client *RegistryAccountClient) queryLocalUserAccount(ctx context.Context, spec LocalUserAccountSpec) (string, bool, error) {
	if err := registryAccountContextError(ctx); err != nil {
		return "", false, err
	}
	arguments, err := json.Marshal(struct {
		Filter struct {
			Operator   string `json:"operator"`
			Conditions []any  `json:"conditions"`
		} `json:"filter"`
		Limit int `json:"limit"`
		Sort  []struct {
			Property    string `json:"property"`
			IsAscending bool   `json:"isAscending"`
		} `json:"sort"`
	}{
		Filter: struct {
			Operator   string `json:"operator"`
			Conditions []any  `json:"conditions"`
		}{
			Operator: "AND",
			Conditions: []any{
				struct {
					Name string `json:"name"`
				}{Name: spec.EmailLocalPart},
				struct {
					DomainID string `json:"domainId"`
				}{DomainID: spec.DomainID},
			},
		},
		Limit: 2,
		Sort: []struct {
			Property    string `json:"property"`
			IsAscending bool   `json:"isAscending"`
		}{{Property: "id", IsAscending: true}},
	})
	if err != nil {
		return "", false, newError(ErrorInvalidRequest)
	}
	response, err := client.client.CallRegistry(ctx, JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name:      "x:Account/query",
			Arguments: arguments,
			CallID:    registryAccountQueryCallID,
		}},
	})
	if err != nil {
		return "", false, err
	}
	payload, err := registryAccountMethodPayload(response, "x:Account/query", registryAccountQueryCallID)
	if err != nil {
		return "", false, err
	}
	var query struct {
		AccountID  json.RawMessage `json:"accountId"`
		QueryState string          `json:"queryState"`
		IDs        json.RawMessage `json:"ids"`
	}
	var ids []string
	if err := json.Unmarshal(payload, &query); err != nil || !validRegistryAccountResponseAccountID(query.AccountID) || !validOpaque(query.QueryState) || len(query.IDs) == 0 || !strings.HasPrefix(strings.TrimSpace(string(query.IDs)), "[") || json.Unmarshal(query.IDs, &ids) != nil || len(ids) > 1 {
		return "", false, newError(ErrorProtocol)
	}
	if len(ids) == 0 {
		return "", false, nil
	}
	if !validOpaque(ids[0]) {
		return "", false, newError(ErrorProtocol)
	}
	return ids[0], true, nil
}

func (client *RegistryAccountClient) createLocalUserAccount(ctx context.Context, spec LocalUserAccountSpec) (string, error) {
	if err := registryAccountContextError(ctx); err != nil {
		return "", err
	}
	arguments, err := json.Marshal(struct {
		Create    map[string]registryLocalUserAccount `json:"create"`
	}{
		Create: map[string]registryLocalUserAccount{
			registryAccountCreationID: {
				Type:           "User",
				Name:           spec.EmailLocalPart,
				DomainID:       spec.DomainID,
				Credentials:    map[string]struct{}{},
				MemberGroupIDs: map[string]struct{}{},
				Roles:          registryAccountTypedValue{Type: "User"},
				Permissions:    registryAccountTypedValue{Type: "Inherit"},
				Quotas:         map[string]struct{}{},
				Aliases:        map[string]struct{}{},
				EncryptionAtRest: registryAccountTypedValue{
					Type: "Disabled",
				},
			},
		},
	})
	if err != nil {
		return "", newError(ErrorInvalidRequest)
	}
	response, err := client.client.CallRegistry(ctx, JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name:      "x:Account/set",
			Arguments: arguments,
			CallID:    registryAccountSetCallID,
		}},
	})
	if err != nil {
		return "", err
	}
	payload, err := registryAccountMethodPayload(response, "x:Account/set", registryAccountSetCallID)
	if err != nil {
		return "", newError(ErrorIndeterminate)
	}
	var result struct {
		AccountID  json.RawMessage            `json:"accountId"`
		Created    map[string]json.RawMessage `json:"created"`
		NotCreated map[string]json.RawMessage `json:"notCreated"`
	}
	if err := json.Unmarshal(payload, &result); err != nil || !validRegistryAccountResponseAccountID(result.AccountID) {
		return "", newError(ErrorIndeterminate)
	}
	if len(result.NotCreated) != 0 {
		if len(result.Created) != 0 || len(result.NotCreated) != 1 {
			return "", newError(ErrorIndeterminate)
		}
		notCreated, present := result.NotCreated[registryAccountCreationID]
		if !present || !validJSONObject(notCreated) {
			return "", newError(ErrorIndeterminate)
		}
		var failure struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(notCreated, &failure); err != nil || failure.Type != "stateMismatch" {
			return "", newError(ErrorIndeterminate)
		}
		return "", newError(ErrorStateMismatch)
	}
	if len(result.Created) != 1 {
		return "", newError(ErrorIndeterminate)
	}
	created, present := result.Created[registryAccountCreationID]
	if !present || len(created) == 0 || int64(len(created)) > registryAccountResponseMaximumBytes {
		return "", newError(ErrorIndeterminate)
	}
	var createdResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created, &createdResult); err != nil || !validOpaque(createdResult.ID) {
		return "", newError(ErrorIndeterminate)
	}
	return createdResult.ID, nil
}

func (client *RegistryAccountClient) getLocalUserAccount(ctx context.Context, accountID string, spec LocalUserAccountSpec) (LocalUserAccount, error) {
	if err := registryAccountContextError(ctx); err != nil {
		return LocalUserAccount{}, err
	}
	arguments, err := json.Marshal(struct {
		IDs        []string `json:"ids"`
		Properties []string `json:"properties"`
	}{
		IDs:        []string{accountID},
		Properties: []string{"id", "@type", "name", "domainId", "emailAddress"},
	})
	if err != nil {
		return LocalUserAccount{}, newError(ErrorInvalidRequest)
	}
	response, err := client.client.CallRegistry(ctx, JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name:      "x:Account/get",
			Arguments: arguments,
			CallID:    registryAccountGetCallID,
		}},
	})
	if err != nil {
		return LocalUserAccount{}, err
	}
	payload, err := registryAccountMethodPayload(response, "x:Account/get", registryAccountGetCallID)
	if err != nil {
		return LocalUserAccount{}, err
	}
	var result struct {
		AccountID json.RawMessage   `json:"accountId"`
		List      []json.RawMessage `json:"list"`
		NotFound  []string          `json:"notFound"`
	}
	if err := json.Unmarshal(payload, &result); err != nil || !validRegistryAccountResponseAccountID(result.AccountID) || len(result.List) != 1 || len(result.NotFound) != 0 || int64(len(result.List[0])) > registryAccountResponseMaximumBytes {
		return LocalUserAccount{}, newError(ErrorProtocol)
	}
	return registryAccountFromObject(result.List[0], accountID, spec)
}

func validRegistryAccountResponseAccountID(payload json.RawMessage) bool {
	var accountID string
	return len(payload) != 0 && json.Unmarshal(payload, &accountID) == nil && validOpaque(accountID)
}

func registryAccountMethodPayload(response JMAPResponse, method, callID string) (json.RawMessage, error) {
	if len(response.MethodResponses) != 1 {
		return nil, newError(ErrorProtocol)
	}
	methodResponse := response.MethodResponses[0]
	if methodResponse.Name != method || methodResponse.CallID != callID || len(methodResponse.Arguments) == 0 || int64(len(methodResponse.Arguments)) > registryAccountResponseMaximumBytes || !validJSONObject(methodResponse.Arguments) {
		return nil, newError(ErrorProtocol)
	}
	return append(json.RawMessage(nil), methodResponse.Arguments...), nil
}

func registryAccountFromObject(payload json.RawMessage, expectedID string, spec LocalUserAccountSpec) (LocalUserAccount, error) {
	if len(payload) == 0 || int64(len(payload)) > registryAccountResponseMaximumBytes || !validJSONObject(payload) {
		return LocalUserAccount{}, newError(ErrorProtocol)
	}
	var account struct {
		ID           string `json:"id"`
		Type         string `json:"@type"`
		Name         string `json:"name"`
		DomainID     string `json:"domainId"`
		EmailAddress string `json:"emailAddress"`
	}
	if err := json.Unmarshal(payload, &account); err != nil || account.ID != expectedID || !validOpaque(account.ID) || account.Type != "User" || account.Name != spec.EmailLocalPart || account.DomainID != spec.DomainID || !validOpaque(account.DomainID) || !validRegistryAccountEmailAddress(account.EmailAddress) {
		return LocalUserAccount{}, newError(ErrorStateMismatch)
	}
	if spec.DomainName != "" && account.EmailAddress != spec.EmailLocalPart+"@"+spec.DomainName {
		return LocalUserAccount{}, newError(ErrorStateMismatch)
	}
	return LocalUserAccount{AccountID: account.ID, EmailAddress: account.EmailAddress}, nil
}

func validRegistryAccountSpec(spec LocalUserAccountSpec) bool {
	return validRegistryAccountLocalPart(spec.EmailLocalPart) && validOpaque(spec.DomainID) && (spec.DomainName == "" || validRegistryAccountDomainName(spec.DomainName))
}

func validRegistryAccountLocalPart(value string) bool {
	if value == "" || len(value) > registryAccountLocalPartMaximumBytes || !utf8.ValidString(value) {
		return false
	}
	for index := range value {
		character := value[index]
		if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '.' || character == '-' || character == '_') {
			return false
		}
		if (index == 0 || index == len(value)-1) && (character == '.' || character == '-' || character == '_') {
			return false
		}
	}
	return true
}

func validRegistryAccountDomainName(value string) bool {
	if value == "" || len(value) > registryAccountDomainMaximumBytes || !utf8.ValidString(value) || strings.ToLower(value) != value {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 {
			return false
		}
		for index := range label {
			character := label[index]
			if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '-') {
				return false
			}
			if (index == 0 || index == len(label)-1) && character == '-' {
				return false
			}
		}
	}
	return true
}

func validRegistryAccountEmailAddress(value string) bool {
	if value == "" || len(value) > registryAccountEmailMaximumBytes || !utf8.ValidString(value) || strings.Count(value, "@") != 1 || strings.ContainsAny(value, "\r\n\t ") {
		return false
	}
	localPart, domain, ok := strings.Cut(value, "@")
	return ok && localPart != "" && len(localPart) <= registryAccountLocalPartMaximumBytes && domain != "" && len(domain) <= registryAccountDomainMaximumBytes
}

func registryAccountContextError(ctx context.Context) error {
	if ctx == nil {
		return newError(ErrorInvalidRequest)
	}
	if err := ctx.Err(); err != nil {
		if err == context.Canceled {
			return newError(ErrorCancelled)
		}
		return newError(ErrorUnavailable)
	}
	return nil
}

type registryAccountTypedValue struct {
	Type string `json:"@type"`
}

type registryLocalUserAccount struct {
	Type              string                     `json:"@type"`
	Name              string                     `json:"name"`
	DomainID          string                     `json:"domainId"`
	Credentials       map[string]struct{}        `json:"credentials"`
	MemberGroupIDs    map[string]struct{}        `json:"memberGroupIds"`
	Roles             registryAccountTypedValue  `json:"roles"`
	Permissions       registryAccountTypedValue  `json:"permissions"`
	Quotas            map[string]struct{}        `json:"quotas"`
	Aliases           map[string]struct{}        `json:"aliases"`
	EncryptionAtRest  registryAccountTypedValue  `json:"encryptionAtRest"`
}
