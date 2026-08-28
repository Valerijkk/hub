package stalwart

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

const (
	// capabilityCore also governs standard JMAP Blob copy/upload/download.
	capabilityCore       = "urn:ietf:params:jmap:core"
	capabilityMail       = "urn:ietf:params:jmap:mail"
	capabilitySubmission = "urn:ietf:params:jmap:submission"
	capabilitySieve      = "urn:ietf:params:jmap:sieve"
	capabilityRegistry   = "urn:stalwart:jmap"

	maxSessionCapabilities = 16
	maxSessionAccounts     = 256
)

// JMAPRequest is a transport-private-safe JMAP request envelope. Arguments may
// contain mailbox data, so callers must not log this value. The transport
// validates its JSON depth, token count and duplicate keys before any I/O.
type JMAPRequest struct {
	Using       []string         `json:"using"`
	MethodCalls []JMAPMethodCall `json:"methodCalls"`
	CreatedIDs  map[string]string `json:"createdIds,omitempty"`
}

// RequestFitsTransport reports whether the exact serialized JMAP envelope fits
// the transport ceiling. Callers use this before mutation I/O so a locally
// oversized request can be rejected without an ambiguous provider outcome.
func RequestFitsTransport(request JMAPRequest) (bool, error) {
	payload, err := json.Marshal(request)
	if err != nil || validateJSON(payload) != nil {
		return false, newError(ErrorInvalidRequest)
	}
	return int64(len(payload)) <= maxJMAPBytes, nil
}

// JMAPMethodCall is the JMAP tuple [name, arguments, callId].
type JMAPMethodCall struct {
	Name      string
	Arguments json.RawMessage
	CallID    string
}

func (call JMAPMethodCall) MarshalJSON() ([]byte, error) {
	name, err := json.Marshal(call.Name)
	if err != nil {
		return nil, err
	}
	callID, err := json.Marshal(call.CallID)
	if err != nil {
		return nil, err
	}
	arguments := call.Arguments
	if len(arguments) == 0 {
		arguments = json.RawMessage("null")
	}
	return json.Marshal([]json.RawMessage{name, arguments, callID})
}

func (call *JMAPMethodCall) UnmarshalJSON(data []byte) error {
	var tuple []json.RawMessage
	if err := json.Unmarshal(data, &tuple); err != nil || len(tuple) != 3 {
		return newError(ErrorProtocol)
	}
	if err := json.Unmarshal(tuple[0], &call.Name); err != nil {
		return newError(ErrorProtocol)
	}
	if err := json.Unmarshal(tuple[2], &call.CallID); err != nil {
		return newError(ErrorProtocol)
	}
	call.Arguments = append(call.Arguments[:0], tuple[1]...)
	return nil
}

// JMAPResponse is intentionally limited to the standard envelope. Provider
// errors are mapped to Error before this value is returned to a caller.
type JMAPResponse struct {
	MethodResponses []JMAPMethodResponse `json:"methodResponses"`
	SessionState    string               `json:"sessionState"`
	CreatedIDs      map[string]string    `json:"createdIds,omitempty"`
}

// JMAPMethodResponse is the JMAP tuple [name, arguments, callId].
type JMAPMethodResponse struct {
	Name      string
	Arguments json.RawMessage
	CallID    string
}

func (response JMAPMethodResponse) MarshalJSON() ([]byte, error) {
	name, err := json.Marshal(response.Name)
	if err != nil {
		return nil, err
	}
	callID, err := json.Marshal(response.CallID)
	if err != nil {
		return nil, err
	}
	arguments := response.Arguments
	if len(arguments) == 0 {
		arguments = json.RawMessage("null")
	}
	return json.Marshal([]json.RawMessage{name, arguments, callID})
}

func (response *JMAPMethodResponse) UnmarshalJSON(data []byte) error {
	var tuple []json.RawMessage
	if err := json.Unmarshal(data, &tuple); err != nil || len(tuple) != 3 {
		return newError(ErrorProtocol)
	}
	if err := json.Unmarshal(tuple[0], &response.Name); err != nil {
		return newError(ErrorProtocol)
	}
	if err := json.Unmarshal(tuple[2], &response.CallID); err != nil {
		return newError(ErrorProtocol)
	}
	response.Arguments = append(response.Arguments[:0], tuple[1]...)
	return nil
}

// MailboxSession is a validated, same-origin session result. Its URLs are
// retained privately so callers cannot replace a trusted route with a session
// value from another origin.
type MailboxSession struct {
	apiURL       *url.URL
	uploadURL    *url.URL
	downloadURL  *url.URL
	capabilities map[string]struct{}
	accountIDs   map[string]struct{}
}

// HasAccount reports whether the exact opaque account identifier was present
// in the validated session. The retained set is deliberately private: callers
// can prove one expected binding but cannot enumerate provider account data.
func (session *MailboxSession) HasAccount(accountID string) bool {
	if session == nil || !validOpaque(accountID) {
		return false
	}
	_, present := session.accountIDs[accountID]
	return present
}

// SupportsSieve reports whether the optional standards-based Sieve capability
// was explicitly advertised by the validated session.
func (session *MailboxSession) SupportsSieve() bool {
	if session == nil {
		return false
	}
	_, present := session.capabilities[capabilitySieve]
	return present
}

type jmapContext uint8

const (
	jmapRegistry jmapContext = iota + 1
	jmapMailbox
)

var registryCapabilities = map[string]struct{}{
	capabilityCore:     {},
	capabilityRegistry: {},
}

var mailboxCapabilities = map[string]struct{}{
	capabilityCore:       {},
	capabilityMail:       {},
	capabilitySubmission: {},
	capabilitySieve:      {},
}

var registryMethods = map[string]struct{}{
	"x:Domain/get":   {},
	"x:Domain/query": {},
	"x:Domain/set":   {},
	"x:Account/get":   {},
	"x:Account/query": {},
	"x:Account/set":   {},
}

var mailboxMethods = map[string]struct{}{
	"Mailbox/get":             {},
	"Mailbox/set":             {},
	"Mailbox/changes":         {},
	"Mailbox/query":           {},
	"Mailbox/queryChanges":    {},
	"Email/get":               {},
	"Email/set":               {},
	"Email/changes":           {},
	"Email/query":             {},
	"Email/queryChanges":      {},
	"Email/import":            {},
	"Thread/get":              {},
	"Thread/changes":          {},
	"Identity/get":            {},
	"Identity/set":            {},
	"Identity/changes":        {},
	"EmailSubmission/get":     {},
	"EmailSubmission/set":     {},
	"EmailSubmission/changes": {},
	"EmailSubmission/query":   {},
	"EmailSubmission/queryChanges": {},
	"Blob/copy":                    {},
	"SieveScript/get":              {},
	"SieveScript/set":              {},
	"SieveScript/changes":          {},
}

func validateRequest(contextKind jmapContext, request JMAPRequest, session *MailboxSession) ([]byte, bool, error) {
	if len(request.MethodCalls) == 0 || len(request.MethodCalls) > maxJMAPMethodCalls || len(request.Using) == 0 || len(request.Using) > maxJMAPMethodCalls {
		return nil, false, newError(ErrorInvalidRequest)
	}
	if !validCapabilities(contextKind, request.Using, session) {
		return nil, false, newError(ErrorUnsupported)
	}

	seenCallIDs := make(map[string]struct{}, len(request.MethodCalls))
	mutationCount := 0
	for _, method := range request.MethodCalls {
		if !validCallID(method.CallID) {
			return nil, false, newError(ErrorInvalidRequest)
		}
		if _, duplicate := seenCallIDs[method.CallID]; duplicate {
			return nil, false, newError(ErrorInvalidRequest)
		}
		seenCallIDs[method.CallID] = struct{}{}
		if !methodAllowed(contextKind, method.Name, session) || !validJSONObject(method.Arguments) {
			return nil, false, newError(ErrorUnsupported)
		}
		if isMutation(method.Name) {
			mutationCount++
		}
	}
	// A JMAP batch cannot safely expose a retryable failure when one mutation
	// may already have applied. Until the service has atomic batch semantics,
	// accept only a single mutation by itself.
	if mutationCount > 1 || mutationCount == 1 && len(request.MethodCalls) != 1 {
		return nil, false, newError(ErrorInvalidRequest)
	}
	if len(request.CreatedIDs) > maxJMAPMethodCalls {
		return nil, false, newError(ErrorInvalidRequest)
	}
	for key, value := range request.CreatedIDs {
		if !validCallID(key) || !validOpaque(value) {
			return nil, false, newError(ErrorInvalidRequest)
		}
	}

	payload, err := json.Marshal(request)
	if err != nil || int64(len(payload)) > maxJMAPBytes || validateJSON(payload) != nil {
		return nil, false, newError(ErrorInvalidRequest)
	}
	return payload, mutationCount == 1, nil
}

func validCapabilities(contextKind jmapContext, capabilities []string, session *MailboxSession) bool {
	allowed := registryCapabilities
	required := map[string]struct{}{capabilityCore: {}, capabilityRegistry: {}}
	if contextKind == jmapMailbox {
		allowed = mailboxCapabilities
		required = map[string]struct{}{capabilityCore: {}, capabilityMail: {}, capabilitySubmission: {}}
		if session == nil || session.apiURL == nil {
			return false
		}
	}
	seen := make(map[string]struct{}, len(capabilities))
	for _, capability := range capabilities {
		if _, known := allowed[capability]; !known {
			return false
		}
		if _, duplicate := seen[capability]; duplicate {
			return false
		}
		if contextKind == jmapMailbox {
			if _, advertised := session.capabilities[capability]; !advertised {
				return false
			}
		}
		seen[capability] = struct{}{}
	}
	for capability := range required {
		if _, present := seen[capability]; !present {
			return false
		}
	}
	return true
}

func methodAllowed(contextKind jmapContext, method string, session *MailboxSession) bool {
	if strings.HasPrefix(method, "x:Sieve") {
		return false
	}
	if contextKind == jmapRegistry {
		_, allowed := registryMethods[method]
		return allowed
	}
	if _, allowed := mailboxMethods[method]; !allowed {
		return false
	}
	if strings.HasPrefix(method, "SieveScript/") {
		return session != nil && session.SupportsSieve()
	}
	return true
}

func isMutation(method string) bool {
	return strings.HasSuffix(method, "/set") || method == "Blob/copy" || method == "Email/import"
}

func parseJMAPResponse(contextKind jmapContext, request JMAPRequest, data []byte, session *MailboxSession) (JMAPResponse, error) {
	if int64(len(data)) > maxJMAPBytes || validateJSON(data) != nil {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	var response JMAPResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	implicitSubmissionUpdate := len(request.MethodCalls) == 1 && request.MethodCalls[0].Name == "EmailSubmission/set" && len(response.MethodResponses) == 2
	if len(response.MethodResponses) != len(request.MethodCalls) && !implicitSubmissionUpdate {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	if response.SessionState != "" && !validOpaque(response.SessionState) {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	if len(response.CreatedIDs) > maxJMAPMethodCalls {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	for creationID, objectID := range response.CreatedIDs {
		if !validCallID(creationID) || !validOpaque(objectID) {
			return JMAPResponse{}, newError(ErrorProtocol)
		}
	}
	if implicitSubmissionUpdate {
		if err := validateImplicitSubmissionUpdate(request.MethodCalls[0], response.MethodResponses); err != nil {
			return JMAPResponse{}, err
		}
		return response, nil
	}
	requestNames := make(map[string]string, len(request.MethodCalls))
	for _, requestMethod := range request.MethodCalls {
		requestNames[requestMethod.CallID] = requestMethod.Name
	}
	seen := make(map[string]struct{}, len(response.MethodResponses))
	for _, responseMethod := range response.MethodResponses {
		requestName, requested := requestNames[responseMethod.CallID]
		if !requested || !validCallID(responseMethod.CallID) {
			return JMAPResponse{}, newError(ErrorProtocol)
		}
		if _, duplicate := seen[responseMethod.CallID]; duplicate || validateJSON(responseMethod.Arguments) != nil {
			return JMAPResponse{}, newError(ErrorProtocol)
		}
		seen[responseMethod.CallID] = struct{}{}
		if responseMethod.Name != "error" && (responseMethod.Name != requestName || !methodAllowed(contextKind, responseMethod.Name, session)) {
			return JMAPResponse{}, newError(ErrorProtocol)
		}
	}
	if len(seen) != len(requestNames) {
		return JMAPResponse{}, newError(ErrorProtocol)
	}
	// Correlate the complete provider batch before mapping a call-level error.
	// This prevents a partial/mixed result from looking safely retryable.
	for _, responseMethod := range response.MethodResponses {
		if responseMethod.Name == "error" {
			return JMAPResponse{}, responseError(responseMethod.Arguments)
		}
	}
	return response, nil
}

func validateImplicitSubmissionUpdate(request JMAPMethodCall, responses []JMAPMethodResponse) error {
	if request.Name != "EmailSubmission/set" || len(responses) != 2 {
		return newError(ErrorProtocol)
	}
	first, second := responses[0], responses[1]
	if first.CallID != request.CallID || second.CallID != request.CallID || !validCallID(first.CallID) || validateJSON(first.Arguments) != nil || validateJSON(second.Arguments) != nil {
		return newError(ErrorProtocol)
	}
	if first.Name == "error" {
		return newError(ErrorIndeterminate)
	}
	if first.Name != "EmailSubmission/set" || !methodAllowed(jmapMailbox, first.Name, nil) {
		return newError(ErrorProtocol)
	}
	if second.Name == "error" {
		return newError(ErrorIndeterminate)
	}
	if second.Name != "Email/set" || !methodAllowed(jmapMailbox, second.Name, nil) {
		return newError(ErrorProtocol)
	}
	return nil
}

func responseError(arguments json.RawMessage) error {
	var payload struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(arguments, &payload); err != nil || payload.Type == "" || len(payload.Type) > maxCallIDBytes {
		return newError(ErrorProtocol)
	}
	return newError(jmapErrorKind(payload.Type))
}

func parseMailboxSession(client *Client, data []byte) (*MailboxSession, error) {
	if client == nil || int64(len(data)) > maxSessionBytes || validateJSON(data) != nil {
		return nil, newError(ErrorProtocol)
	}
	var payload struct {
		Capabilities map[string]json.RawMessage `json:"capabilities"`
		Accounts     map[string]json.RawMessage `json:"accounts"`
		APIURL       string                     `json:"apiUrl"`
		UploadURL    string                     `json:"uploadUrl"`
		DownloadURL  string                     `json:"downloadUrl"`
	}
	if err := json.Unmarshal(data, &payload); err != nil || len(payload.Capabilities) == 0 || len(payload.Capabilities) > maxSessionCapabilities || len(payload.Accounts) == 0 || len(payload.Accounts) > maxSessionAccounts {
		return nil, newError(ErrorProtocol)
	}
	capabilities := make(map[string]struct{}, len(payload.Capabilities))
	for capability, detail := range payload.Capabilities {
		if _, allowed := mailboxCapabilities[capability]; !allowed || validateJSON(detail) != nil {
			return nil, newError(ErrorUnsupported)
		}
		capabilities[capability] = struct{}{}
	}
	for _, required := range []string{capabilityCore, capabilityMail, capabilitySubmission} {
		if _, present := capabilities[required]; !present {
			return nil, newError(ErrorUnsupported)
		}
	}
	accountIDs := make(map[string]struct{}, len(payload.Accounts))
	for accountID, detail := range payload.Accounts {
		if !validOpaque(accountID) || int64(len(detail)) > maxSessionBytes || !validJSONObject(detail) {
			return nil, newError(ErrorProtocol)
		}
		accountIDs[accountID] = struct{}{}
	}
	apiURL, ok := parseSessionURL(payload.APIURL, client, "/jmap/", "")
	if !ok {
		return nil, newError(ErrorProtocol)
	}
	uploadURL, ok := parseSessionURL(payload.UploadURL, client, "/jmap/upload/{accountId}/", "")
	if !ok {
		return nil, newError(ErrorProtocol)
	}
	downloadURL, ok := parseSessionURL(payload.DownloadURL, client, "/jmap/download/{accountId}/{blobId}/{name}", "type={type}")
	if !ok {
		return nil, newError(ErrorProtocol)
	}
	return &MailboxSession{apiURL: apiURL, uploadURL: uploadURL, downloadURL: downloadURL, capabilities: capabilities, accountIDs: accountIDs}, nil
}

func parseSessionURL(value string, client *Client, path, rawQuery string) (*url.URL, bool) {
	if value == "" || len(value) > maxOriginBytes || !utf8Valid(value) {
		return nil, false
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil || parsed.Opaque != "" || parsed.User != nil || parsed.Fragment != "" || parsed.ForceQuery || (parsed.RawPath != "" && parsed.RawPath != path) || parsed.Path != path || parsed.RawQuery != rawQuery || !client.sameOrigin(parsed) {
		return nil, false
	}
	return &url.URL{Scheme: client.origin.Scheme, Host: client.origin.Host, Path: path, RawQuery: rawQuery}, true
}

func validJSONObject(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	return len(trimmed) > 1 && trimmed[0] == '{' && validateJSON(trimmed) == nil
}

type jsonBudget struct {
	tokens int
}

func validateJSON(data []byte) error {
	if len(data) == 0 {
		return newError(ErrorProtocol)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	budget := jsonBudget{}
	if err := walkJSON(decoder, 1, &budget); err != nil {
		return newError(ErrorProtocol)
	}
	if _, err := decoder.Token(); err != io.EOF {
		return newError(ErrorProtocol)
	}
	return nil
}

func walkJSON(decoder *json.Decoder, depth int, budget *jsonBudget) error {
	if depth > maxJSONDepth {
		return newError(ErrorProtocol)
	}
	token, err := nextJSONToken(decoder, budget)
	if err != nil {
		return err
	}
	switch delimiter := token.(type) {
	case json.Delim:
		switch delimiter {
		case '{':
			keys := make(map[string]struct{})
			for decoder.More() {
				keyToken, err := nextJSONToken(decoder, budget)
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok {
					return newError(ErrorProtocol)
				}
				if _, duplicate := keys[key]; duplicate {
					return newError(ErrorProtocol)
				}
				keys[key] = struct{}{}
				if err := walkJSON(decoder, depth+1, budget); err != nil {
					return err
				}
			}
			closing, err := nextJSONToken(decoder, budget)
			if err != nil || closing != json.Delim('}') {
				return newError(ErrorProtocol)
			}
		case '[':
			for decoder.More() {
				if err := walkJSON(decoder, depth+1, budget); err != nil {
					return err
				}
			}
			closing, err := nextJSONToken(decoder, budget)
			if err != nil || closing != json.Delim(']') {
				return newError(ErrorProtocol)
			}
		default:
			return newError(ErrorProtocol)
		}
	}
	return nil
}

func nextJSONToken(decoder *json.Decoder, budget *jsonBudget) (json.Token, error) {
	if budget == nil {
		return nil, newError(ErrorProtocol)
	}
	budget.tokens++
	if budget.tokens > maxJSONTokens {
		return nil, newError(ErrorProtocol)
	}
	return decoder.Token()
}

func utf8Valid(value string) bool {
	return strings.ToValidUTF8(value, "") == value
}
