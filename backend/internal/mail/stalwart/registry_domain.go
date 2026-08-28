package stalwart

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	registryDomainMaximumBytes                    = 253
	registryDomainMaximumResponseBytes            = 16 << 10
	registryDomainMaximumDNSZoneFileResponseBytes = 64 << 10
	registryDomainMaximumDNSZoneFileBytes         = 64 << 10
	registryDomainQueryCallID                     = "domain-query"
	registryDomainGetCallID                       = "domain-get"
	registryDomainDNSZoneFileGetCallID            = "domain-zone-get"
	registryDomainSetCallID                       = "domain-set"
	registryDomainCreationID                      = "create-local-domain"
)

// RegistryDomainController owns the narrow Registry-only domain admission
// flow. Its Client must be a management-context client; mailbox clients use a
// separate authorizer and Client instance even though both use the same
// transport implementation.
type RegistryDomainController struct {
	client *Client
}

// LocalDomainDNSZoneFile is one bounded Registry read result. Its raw zone
// remains private to the Stalwart adapter: callers may use the stable domain
// identity but cannot log, persist, or otherwise project the provider value.
type LocalDomainDNSZoneFile struct {
	domainID   string
	domainName string
	zoneFile   string
}

// LocalDomainDNSObservation distinguishes a confirmed absent local domain
// from an authoritative DNS zone without exposing that provider zone.
type LocalDomainDNSObservation struct {
	present  bool
	zoneFile LocalDomainDNSZoneFile
}

// IsPresent reports whether the observation has an internally valid
// authoritative local-domain DNS zone.
func (observation LocalDomainDNSObservation) IsPresent() bool {
	return observation.present && validLocalDomainDNSZoneFile(observation.zoneFile)
}

// DNSZoneFile returns an independent private zone value only for an internally
// valid present observation. Absent or invalid observations return zero, false.
func (observation LocalDomainDNSObservation) DNSZoneFile() (LocalDomainDNSZoneFile, bool) {
	if !observation.IsPresent() {
		return LocalDomainDNSZoneFile{}, false
	}
	return LocalDomainDNSZoneFile{
		domainID:   strings.Clone(observation.zoneFile.domainID),
		domainName: strings.Clone(observation.zoneFile.domainName),
		zoneFile:   strings.Clone(observation.zoneFile.zoneFile),
	}, true
}

// DomainID returns the validated opaque Registry domain identifier.
func (result LocalDomainDNSZoneFile) DomainID() string { return result.domainID }

// DomainName returns the canonical validated local domain name.
func (result LocalDomainDNSZoneFile) DomainName() string { return result.domainName }

// NewRegistryDomainController binds an already constructed management Client
// without performing I/O or obtaining credentials.
func NewRegistryDomainController(client *Client) (*RegistryDomainController, error) {
	if client == nil {
		return nil, ErrInvalidConfiguration
	}
	return &RegistryDomainController{client: client}, nil
}

// EnsureLocalDomain returns the one enabled, non-relaying local x:Domain for
// name. A missing domain is created with manual certificate and DNS
// management plus automatic DKIM management; this operation neither verifies
// DNS nor enables delivery.
func (controller *RegistryDomainController) EnsureLocalDomain(ctx context.Context, name string) (string, error) {
	if controller == nil || controller.client == nil {
		return "", ErrInvalidConfiguration
	}
	if err := registryDomainContextError(ctx); err != nil {
		return "", err
	}
	domainName, ok := canonicalRegistryDomainName(name)
	if !ok {
		return "", ErrInvalidRequest
	}

	response, err := controller.client.CallRegistry(ctx, registryDomainQueryRequest(domainName))
	if err != nil {
		return "", err
	}
	ids, err := parseRegistryDomainQuery(response)
	if err != nil {
		return "", err
	}
	switch len(ids) {
	case 0:
		return controller.createAndVerifyLocalDomain(ctx, domainName)
	case 1:
		return controller.getAndValidateLocalDomain(ctx, ids[0], domainName)
	default:
		return "", ErrStateMismatch
	}
}

// ObserveLocalDomainDNS distinguishes an absent local domain from one
// authoritative enabled, non-relaying domain with automatic DKIM and manual
// DNS management. It neither parses nor verifies public DNS.
func (controller *RegistryDomainController) ObserveLocalDomainDNS(ctx context.Context, name string) (LocalDomainDNSObservation, error) {
	if controller == nil || controller.client == nil {
		return LocalDomainDNSObservation{}, ErrInvalidConfiguration
	}
	if err := registryDomainContextError(ctx); err != nil {
		return LocalDomainDNSObservation{}, err
	}
	domainName, ok := canonicalRegistryDomainName(name)
	if !ok {
		return LocalDomainDNSObservation{}, ErrInvalidRequest
	}

	response, err := controller.client.CallRegistry(ctx, registryDomainQueryRequest(domainName))
	if err != nil {
		return LocalDomainDNSObservation{}, err
	}
	ids, err := parseRegistryDomainQuery(response)
	if err != nil {
		return LocalDomainDNSObservation{}, err
	}
	switch len(ids) {
	case 0:
		return LocalDomainDNSObservation{}, nil
	case 1:
		zoneFile, err := controller.getAndReadLocalDomainDNSZoneFile(ctx, ids[0], domainName)
		if err != nil {
			return LocalDomainDNSObservation{}, err
		}
		observation := LocalDomainDNSObservation{present: true, zoneFile: zoneFile}
		if !observation.IsPresent() {
			return LocalDomainDNSObservation{}, ErrProtocol
		}
		return observation, nil
	default:
		return LocalDomainDNSObservation{}, ErrStateMismatch
	}
}

// ReadLocalDomainDNSZoneFile reads only the provider's expected manual-DNS
// zone for one already-present enabled non-relaying local domain. It neither
// parses nor verifies DNS and never promotes any local state.
func (controller *RegistryDomainController) ReadLocalDomainDNSZoneFile(ctx context.Context, name string) (LocalDomainDNSZoneFile, error) {
	observation, err := controller.ObserveLocalDomainDNS(ctx, name)
	if err != nil {
		return LocalDomainDNSZoneFile{}, err
	}
	zoneFile, present := observation.DNSZoneFile()
	if !present {
		return LocalDomainDNSZoneFile{}, ErrStateMismatch
	}
	return zoneFile, nil
}

func (controller *RegistryDomainController) createAndVerifyLocalDomain(ctx context.Context, name string) (string, error) {
	if err := registryDomainContextError(ctx); err != nil {
		return "", err
	}
	response, err := controller.client.CallRegistry(ctx, registryDomainCreateRequest(name))
	if err != nil {
		return "", err
	}
	id, err := parseRegistryDomainCreate(response)
	if err != nil {
		return "", err
	}
	return controller.getAndValidateLocalDomain(ctx, id, name)
}

func (controller *RegistryDomainController) getAndValidateLocalDomain(ctx context.Context, id, name string) (string, error) {
	if !validOpaque(id) {
		return "", ErrProtocol
	}
	if err := registryDomainContextError(ctx); err != nil {
		return "", err
	}
	response, err := controller.client.CallRegistry(ctx, registryDomainGetRequest(id))
	if err != nil {
		return "", err
	}
	return parseRegistryLocalDomain(response, id, name)
}

func (controller *RegistryDomainController) getAndReadLocalDomainDNSZoneFile(ctx context.Context, id, name string) (LocalDomainDNSZoneFile, error) {
	if !validOpaque(id) {
		return LocalDomainDNSZoneFile{}, ErrProtocol
	}
	if err := registryDomainContextError(ctx); err != nil {
		return LocalDomainDNSZoneFile{}, err
	}
	response, err := controller.client.CallRegistry(ctx, registryDomainDNSZoneFileGetRequest(id))
	if err != nil {
		return LocalDomainDNSZoneFile{}, err
	}
	return parseRegistryLocalDomainDNSZoneFile(response, id, name)
}

func registryDomainQueryRequest(name string) JMAPRequest {
	return JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name: "x:Domain/query",
			Arguments: mustRegistryDomainJSON(struct {
				Filter struct {
					Name string `json:"name"`
				} `json:"filter"`
				Position int `json:"position"`
				Limit    int `json:"limit"`
			}{
				Filter: struct {
					Name string `json:"name"`
				}{Name: name},
				Position: 0,
				Limit:    2,
			}),
			CallID: registryDomainQueryCallID,
		}},
	}
}

func registryDomainGetRequest(id string) JMAPRequest {
	return JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name: "x:Domain/get",
			Arguments: mustRegistryDomainJSON(struct {
				IDs        []string `json:"ids"`
				Properties []string `json:"properties"`
			}{
				IDs:        []string{id},
				Properties: []string{"name", "isEnabled", "allowRelaying", "dkimManagement", "dnsManagement"},
			}),
			CallID: registryDomainGetCallID,
		}},
	}
}

func registryDomainDNSZoneFileGetRequest(id string) JMAPRequest {
	return JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name: "x:Domain/get",
			Arguments: mustRegistryDomainJSON(struct {
				IDs        []string `json:"ids"`
				Properties []string `json:"properties"`
			}{
				IDs:        []string{id},
				Properties: []string{"name", "isEnabled", "allowRelaying", "dkimManagement", "dnsManagement", "dnsZoneFile"},
			}),
			CallID: registryDomainDNSZoneFileGetCallID,
		}},
	}
}

func registryDomainCreateRequest(name string) JMAPRequest {
	type managementMode struct {
		Type string `json:"@type"`
	}
	type createdDomain struct {
		Name                  string          `json:"name"`
		Aliases               map[string]bool `json:"aliases"`
		IsEnabled             bool            `json:"isEnabled"`
		AllowRelaying         bool            `json:"allowRelaying"`
		CertificateManagement managementMode  `json:"certificateManagement"`
		DKIMManagement        managementMode  `json:"dkimManagement"`
		DNSManagement         managementMode  `json:"dnsManagement"`
		SubAddressing         managementMode  `json:"subAddressing"`
	}
	return JMAPRequest{
		Using: []string{capabilityCore, capabilityRegistry},
		MethodCalls: []JMAPMethodCall{{
			Name: "x:Domain/set",
			Arguments: mustRegistryDomainJSON(struct {
				Create map[string]createdDomain `json:"create"`
			}{
				Create: map[string]createdDomain{
					registryDomainCreationID: {
						Name:          name,
						Aliases:       map[string]bool{},
						IsEnabled:     true,
						AllowRelaying: false,
						CertificateManagement: managementMode{
							Type: "Manual",
						},
						DKIMManagement: managementMode{
							Type: "Automatic",
						},
						DNSManagement: managementMode{
							Type: "Manual",
						},
						SubAddressing: managementMode{
							Type: "Enabled",
						},
					},
				},
			}),
			CallID: registryDomainSetCallID,
		}},
	}
}

func mustRegistryDomainJSON(value any) json.RawMessage {
	payload, err := json.Marshal(value)
	if err != nil || len(payload) > registryDomainMaximumResponseBytes {
		return json.RawMessage("{}")
	}
	return payload
}

func parseRegistryDomainQuery(response JMAPResponse) ([]string, error) {
	arguments, err := registryDomainResponseArguments(response, "x:Domain/query", registryDomainQueryCallID)
	if err != nil {
		return nil, err
	}
	var payload struct {
		AccountID           string   `json:"accountId"`
		QueryState          string   `json:"queryState"`
		CanCalculateChanges *bool    `json:"canCalculateChanges"`
		Position            *int     `json:"position"`
		IDs                 []string `json:"ids"`
	}
	if !decodeRegistryDomainArguments(arguments, &payload) || !validOpaque(payload.AccountID) || !validOpaque(payload.QueryState) || payload.CanCalculateChanges == nil || payload.Position == nil || *payload.Position != 0 || payload.IDs == nil || len(payload.IDs) > 2 {
		return nil, ErrProtocol
	}
	seen := make(map[string]struct{}, len(payload.IDs))
	for _, id := range payload.IDs {
		if !validOpaque(id) {
			return nil, ErrProtocol
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, ErrStateMismatch
		}
		seen[id] = struct{}{}
	}
	return payload.IDs, nil
}

func parseRegistryDomainCreate(response JMAPResponse) (string, error) {
	arguments, err := registryDomainResponseArguments(response, "x:Domain/set", registryDomainSetCallID)
	if err != nil {
		return "", err
	}
	var payload struct {
		AccountID string `json:"accountId"`
		OldState  string `json:"oldState"`
		NewState  string `json:"newState"`
		Created   map[string]struct {
			ID string `json:"id"`
		} `json:"created"`
		NotCreated map[string]json.RawMessage `json:"notCreated"`
	}
	if !decodeRegistryDomainArguments(arguments, &payload) || !validOpaque(payload.AccountID) {
		return "", ErrProtocol
	}
	if !validOpaque(payload.OldState) || !validOpaque(payload.NewState) || payload.OldState == payload.NewState {
		return "", ErrStateMismatch
	}
	if len(payload.NotCreated) != 0 {
		return "", ErrIndeterminate
	}
	if len(payload.Created) != 1 {
		return "", ErrProtocol
	}
	created, ok := payload.Created[registryDomainCreationID]
	if !ok || !validOpaque(created.ID) {
		return "", ErrProtocol
	}
	return created.ID, nil
}

func parseRegistryLocalDomain(response JMAPResponse, id, name string) (string, error) {
	arguments, err := registryDomainResponseArguments(response, "x:Domain/get", registryDomainGetCallID)
	if err != nil {
		return "", err
	}
	var payload struct {
		AccountID string `json:"accountId"`
		State     string `json:"state"`
		List      []struct {
			ID             string          `json:"id"`
			Name           string          `json:"name"`
			IsEnabled      *bool           `json:"isEnabled"`
			AllowRelaying  *bool           `json:"allowRelaying"`
			DKIMManagement json.RawMessage `json:"dkimManagement"`
			DNSManagement  json.RawMessage `json:"dnsManagement"`
		} `json:"list"`
		NotFound []string `json:"notFound"`
	}
	if !decodeRegistryDomainArguments(arguments, &payload) || !validOpaque(payload.AccountID) || !validOpaque(payload.State) || payload.List == nil || len(payload.List) != 1 || len(payload.NotFound) != 0 {
		return "", ErrProtocol
	}
	domain := payload.List[0]
	canonicalName, validName := canonicalRegistryDomainName(domain.Name)
	if !validName || !validOpaque(domain.ID) {
		return "", ErrProtocol
	}
	if domain.ID != id || canonicalName != name || domain.IsEnabled == nil || !*domain.IsEnabled || domain.AllowRelaying == nil || *domain.AllowRelaying || !validRegistryDomainAutomaticDKIMManagement(domain.DKIMManagement) || !validRegistryDomainManualDNSManagement(domain.DNSManagement) {
		return "", ErrStateMismatch
	}
	return id, nil
}

func parseRegistryLocalDomainDNSZoneFile(response JMAPResponse, id, name string) (LocalDomainDNSZoneFile, error) {
	arguments, err := registryDomainDNSZoneFileResponseArguments(response, "x:Domain/get", registryDomainDNSZoneFileGetCallID)
	if err != nil {
		return LocalDomainDNSZoneFile{}, err
	}
	var payload struct {
		AccountID string `json:"accountId"`
		State     string `json:"state"`
		List      []struct {
			ID             string           `json:"id"`
			Name           string           `json:"name"`
			IsEnabled      *bool            `json:"isEnabled"`
			AllowRelaying  *bool            `json:"allowRelaying"`
			DKIMManagement json.RawMessage  `json:"dkimManagement"`
			DNSManagement  json.RawMessage  `json:"dnsManagement"`
			DNSZoneFile    *json.RawMessage `json:"dnsZoneFile"`
		} `json:"list"`
		NotFound []string `json:"notFound"`
	}
	if !decodeRegistryDomainDNSZoneFileArguments(arguments, &payload) || !validOpaque(payload.AccountID) || !validOpaque(payload.State) || payload.List == nil || len(payload.List) != 1 || len(payload.NotFound) != 0 {
		return LocalDomainDNSZoneFile{}, ErrProtocol
	}
	domain := payload.List[0]
	canonicalName, validName := canonicalRegistryDomainName(domain.Name)
	if !validName || !validOpaque(domain.ID) || domain.DNSZoneFile == nil {
		return LocalDomainDNSZoneFile{}, ErrProtocol
	}
	if domain.ID != id || canonicalName != name || domain.IsEnabled == nil || !*domain.IsEnabled || domain.AllowRelaying == nil || *domain.AllowRelaying || !validRegistryDomainAutomaticDKIMManagement(domain.DKIMManagement) || !validRegistryDomainManualDNSManagement(domain.DNSManagement) {
		return LocalDomainDNSZoneFile{}, ErrStateMismatch
	}
	zoneFile, err := parseRegistryDomainDNSZoneFile(*domain.DNSZoneFile)
	if err != nil {
		return LocalDomainDNSZoneFile{}, err
	}
	return LocalDomainDNSZoneFile{domainID: domain.ID, domainName: canonicalName, zoneFile: zoneFile}, nil
}

const (
	registryDomainManagementMaximumBytes                 = 4 << 10
	registryDomainManagementMaximumSelectorTemplateBytes = 1024
	registryDomainManagementMaximumDurationMillis        = int64(36_500 * 24 * 60 * 60 * 1000)
)

func validRegistryDomainAutomaticDKIMManagement(raw json.RawMessage) bool {
	var management struct {
		Type             *string         `json:"@type"`
		Algorithms       json.RawMessage `json:"algorithms"`
		SelectorTemplate json.RawMessage `json:"selectorTemplate"`
		RotateAfter      json.RawMessage `json:"rotateAfter"`
		RetireAfter      json.RawMessage `json:"retireAfter"`
		DeleteAfter      json.RawMessage `json:"deleteAfter"`
	}
	if !decodeRegistryDomainManagementObject(raw, &management) || management.Type == nil || *management.Type != "Automatic" {
		return false
	}
	return validRegistryDomainDKIMAlgorithms(management.Algorithms) &&
		validRegistryDomainManagementText(management.SelectorTemplate) &&
		validRegistryDomainManagementDuration(management.RotateAfter) &&
		validRegistryDomainManagementDuration(management.RetireAfter) &&
		validRegistryDomainManagementDuration(management.DeleteAfter)
}

func validRegistryDomainManualDNSManagement(raw json.RawMessage) bool {
	var management struct {
		Type *string `json:"@type"`
	}
	return decodeRegistryDomainManagementObject(raw, &management) && management.Type != nil && *management.Type == "Manual"
}

func decodeRegistryDomainManagementObject(raw json.RawMessage, target any) bool {
	if len(raw) == 0 || len(raw) > registryDomainManagementMaximumBytes || target == nil {
		return false
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) < 2 || trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return false
	}
	return decodeRegistryDomainManagementValue(trimmed, target)
}

func decodeRegistryDomainManagementValue(raw json.RawMessage, target any) bool {
	if len(raw) == 0 || len(raw) > registryDomainManagementMaximumBytes || target == nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}

func validRegistryDomainDKIMAlgorithms(raw json.RawMessage) bool {
	if raw == nil {
		return true
	}
	var algorithms map[string]bool
	if !decodeRegistryDomainManagementObject(raw, &algorithms) || len(algorithms) == 0 {
		return false
	}
	for algorithm, enabled := range algorithms {
		if !enabled || !validRegistryDomainDKIMAlgorithm(algorithm) {
			return false
		}
	}
	return true
}

func validRegistryDomainDKIMAlgorithm(value string) bool {
	switch value {
	case "Dkim1Ed25519Sha256", "Dkim1RsaSha256", "Dkim2Ed25519Sha256", "Dkim2RsaSha256":
		return true
	default:
		return false
	}
}

func validRegistryDomainManagementText(raw json.RawMessage) bool {
	if raw == nil {
		return true
	}
	if !validRegistryDomainManagementNonNull(raw) {
		return false
	}
	var value string
	if !decodeRegistryDomainManagementValue(raw, &value) || len(value) == 0 || len(value) > registryDomainManagementMaximumSelectorTemplateBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validRegistryDomainManagementDuration(raw json.RawMessage) bool {
	if raw == nil {
		return true
	}
	if !validRegistryDomainManagementNonNull(raw) {
		return false
	}
	var value int64
	return decodeRegistryDomainManagementValue(raw, &value) && value > 0 && value <= registryDomainManagementMaximumDurationMillis
}

func validRegistryDomainManagementNonNull(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) != 0 && !bytes.Equal(trimmed, []byte("null"))
}

func registryDomainResponseArguments(response JMAPResponse, name, callID string) (json.RawMessage, error) {
	if len(response.MethodResponses) != 1 {
		return nil, ErrProtocol
	}
	method := response.MethodResponses[0]
	if method.Name != name || method.CallID != callID || len(method.Arguments) == 0 || len(method.Arguments) > registryDomainMaximumResponseBytes {
		return nil, ErrProtocol
	}
	return method.Arguments, nil
}

func registryDomainDNSZoneFileResponseArguments(response JMAPResponse, name, callID string) (json.RawMessage, error) {
	if len(response.MethodResponses) != 1 {
		return nil, ErrProtocol
	}
	method := response.MethodResponses[0]
	if method.Name != name || method.CallID != callID || len(method.Arguments) == 0 || len(method.Arguments) > registryDomainMaximumDNSZoneFileResponseBytes {
		return nil, ErrProtocol
	}
	return method.Arguments, nil
}

func decodeRegistryDomainArguments(arguments json.RawMessage, target any) bool {
	if len(arguments) == 0 || len(arguments) > registryDomainMaximumResponseBytes || target == nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}

func decodeRegistryDomainDNSZoneFileArguments(arguments json.RawMessage, target any) bool {
	if len(arguments) == 0 || len(arguments) > registryDomainMaximumDNSZoneFileResponseBytes || target == nil {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return false
	}
	var extra any
	return errors.Is(decoder.Decode(&extra), io.EOF)
}

func parseRegistryDomainDNSZoneFile(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || len(raw) > registryDomainMaximumDNSZoneFileResponseBytes || !utf8.Valid(raw) {
		return "", ErrProtocol
	}
	var zoneFile string
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&zoneFile); err != nil {
		return "", ErrProtocol
	}
	var extra any
	if !errors.Is(decoder.Decode(&extra), io.EOF) {
		return "", ErrProtocol
	}
	if zoneFile == "" {
		return "", ErrStateMismatch
	}
	if !validRegistryDomainDNSZoneFile(zoneFile) {
		return "", ErrProtocol
	}
	return strings.Clone(zoneFile), nil
}

func validRegistryDomainDNSZoneFile(value string) bool {
	if len(value) == 0 || len(value) > registryDomainMaximumDNSZoneFileBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		switch character {
		case '\t', '\n', '\r':
			continue
		default:
			if unicode.IsControl(character) {
				return false
			}
		}
	}
	return true
}

func canonicalRegistryDomainName(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) == 0 || len(value) > registryDomainMaximumBytes || strings.HasSuffix(value, ".") {
		return "", false
	}
	labels := strings.Split(value, ".")
	if len(labels) < 2 {
		return "", false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return "", false
		}
		for _, character := range []byte(label) {
			if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' {
				continue
			}
			return "", false
		}
	}
	return value, true
}

func registryDomainContextError(ctx context.Context) error {
	if ctx == nil {
		return ErrInvalidRequest
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return ErrCancelled
		}
		return ErrUnavailable
	}
	return nil
}

func validLocalDomainDNSZoneFile(zoneFile LocalDomainDNSZoneFile) bool {
	canonicalName, validName := canonicalRegistryDomainName(zoneFile.domainName)
	return validOpaque(zoneFile.domainID) &&
		validName && canonicalName == zoneFile.domainName &&
		validRegistryDomainDNSZoneFile(zoneFile.zoneFile)
}
