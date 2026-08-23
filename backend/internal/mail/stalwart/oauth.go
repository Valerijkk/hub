package stalwart

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const oauthRequestDeadline = 10 * time.Second

// OAuthDiscovery contains only fixed, validated, query-free public endpoints.
// Revocation metadata is deliberately neither required nor exposed.
type OAuthDiscovery struct {
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
}

// OAuthPKCE contains one verifier and its S256 challenge.
type OAuthPKCE struct {
	Verifier  string
	Challenge string
}

// OAuthAuthorization contains the browser URL and correlation secrets that a
// caller must store privately until the exact callback is handled.
type OAuthAuthorization struct {
	URL          string
	State        string
	CodeVerifier string
}

// OAuthTokenSet contains only values required for mailbox authorization and
// refresh. Any id_token returned by a provider is discarded during parsing.
type OAuthTokenSet struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	Scope            string
	ExpiresInSeconds int64
}

// GenerateOAuthState produces 43 unpadded base64url characters from 32 random
// bytes.
func GenerateOAuthState(random io.Reader) (string, error) {
	return oauthRandomBase64URL(random, oauthStateBytes, oauthStateLength)
}

// GenerateOAuthPKCE produces an 86-character verifier from 64 random bytes and
// its 43-character S256 challenge.
func GenerateOAuthPKCE(random io.Reader) (OAuthPKCE, error) {
	verifier, err := oauthRandomBase64URL(random, oauthVerifierBytes, oauthVerifierLength)
	if err != nil {
		return OAuthPKCE{}, err
	}
	digest := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(digest[:])
	if len(challenge) != oauthChallengeLength {
		return OAuthPKCE{}, newOAuthError(OAuthErrorProtocol)
	}
	return OAuthPKCE{Verifier: verifier, Challenge: challenge}, nil
}

// Discover validates the pinned Stalwart v0.16.17 OAuth metadata. It performs
// one bounded request and never follows redirects or caches the result.
func (client *OAuthClient) Discover(ctx context.Context) (OAuthDiscovery, error) {
	if err := client.validateOperation(ctx); err != nil {
		return OAuthDiscovery{}, err
	}
	response, payload, err := client.doOAuthRequest(ctx, http.MethodGet, client.issuerOrigin+oauthDiscoveryPath, "", oauthMaxDiscoveryBytes, false)
	if err != nil {
		return OAuthDiscovery{}, err
	}
	defer clear(payload)
	if response.StatusCode != http.StatusOK {
		return OAuthDiscovery{}, oauthStatusError(response.StatusCode, response.Header.Get("Content-Type"), payload)
	}
	if !oauthJSONContentType(response.Header.Get("Content-Type")) || validateJSON(payload) != nil {
		return OAuthDiscovery{}, newOAuthError(OAuthErrorProtocol)
	}
	var metadata struct {
		Issuer                      string   `json:"issuer"`
		AuthorizationEndpoint       string   `json:"authorization_endpoint"`
		TokenEndpoint               string   `json:"token_endpoint"`
		ScopesSupported             []string `json:"scopes_supported"`
		ResponseTypesSupported      []string `json:"response_types_supported"`
		GrantTypesSupported         []string `json:"grant_types_supported"`
		CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
	}
	if err := json.Unmarshal(payload, &metadata); err != nil {
		return OAuthDiscovery{}, newOAuthError(OAuthErrorProtocol)
	}
	expectedAuthorization := client.issuerOrigin + oauthAuthorizationPath
	expectedToken := client.issuerOrigin + oauthTokenPath
	if metadata.Issuer != client.issuerOrigin || metadata.AuthorizationEndpoint != expectedAuthorization || metadata.TokenEndpoint != expectedToken || !oauthStringSetContains(metadata.ScopesSupported, "offline_access", "mail") || !oauthStringSetContains(metadata.ResponseTypesSupported, "code") || !oauthStringSetContains(metadata.GrantTypesSupported, "authorization_code", "refresh_token") || !oauthStringSetContains(metadata.CodeChallengeMethodsSupported, "S256") {
		return OAuthDiscovery{}, newOAuthError(OAuthErrorProtocol)
	}
	return OAuthDiscovery{Issuer: metadata.Issuer, AuthorizationEndpoint: metadata.AuthorizationEndpoint, TokenEndpoint: metadata.TokenEndpoint}, nil
}

// BeginAuthorization discovers exact endpoints and builds one deterministic
// authorization query around fresh state and PKCE values.
func (client *OAuthClient) BeginAuthorization(ctx context.Context) (OAuthAuthorization, error) {
	if err := client.validateOperation(ctx); err != nil {
		return OAuthAuthorization{}, err
	}
	state, err := GenerateOAuthState(client.random)
	if err != nil {
		return OAuthAuthorization{}, err
	}
	pkce, err := GenerateOAuthPKCE(client.random)
	if err != nil {
		return OAuthAuthorization{}, err
	}
	discovery, err := client.Discover(ctx)
	if err != nil {
		return OAuthAuthorization{}, err
	}
	query := url.Values{}
	query.Set("client_id", client.clientID)
	query.Set("code_challenge", pkce.Challenge)
	query.Set("code_challenge_method", "S256")
	query.Set("redirect_uri", client.redirectURI)
	query.Set("resource", client.resourceOrigin)
	query.Set("response_type", "code")
	query.Set("scope", oauthScope)
	query.Set("state", state)
	authorizationURL := discovery.AuthorizationEndpoint + "?" + query.Encode()
	if len(authorizationURL) > oauthMaxAuthorizationBytes {
		return OAuthAuthorization{}, newOAuthError(OAuthErrorPayloadTooLarge)
	}
	return OAuthAuthorization{URL: authorizationURL, State: state, CodeVerifier: pkce.Verifier}, nil
}

// ExchangeAuthorizationCode performs one authorization_code exchange using
// HTTP Basic client authentication and the exact configured redirect/resource.
func (client *OAuthClient) ExchangeAuthorizationCode(ctx context.Context, code, codeVerifier string) (OAuthTokenSet, error) {
	if err := client.validateOperation(ctx); err != nil {
		return OAuthTokenSet{}, err
	}
	if !validOAuthOpaque(code, oauthMaxCodeBytes) || !validOAuthBase64URL(codeVerifier, oauthVerifierLength) {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorInvalidRequest)
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", client.redirectURI)
	form.Set("resource", client.resourceOrigin)
	return client.exchangeToken(ctx, form, "", true)
}

// RefreshAccessToken performs one refresh exchange. A successful response may
// rotate the refresh token; omission preserves the exact input token.
func (client *OAuthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (OAuthTokenSet, error) {
	if err := client.validateOperation(ctx); err != nil {
		return OAuthTokenSet{}, err
	}
	if !validOAuthOpaque(refreshToken, oauthMaxTokenBytes) {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorInvalidRequest)
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("resource", client.resourceOrigin)
	form.Set("scope", oauthScope)
	return client.exchangeToken(ctx, form, refreshToken, false)
}

func (client *OAuthClient) exchangeToken(ctx context.Context, form url.Values, previousRefreshToken string, requireRefreshToken bool) (OAuthTokenSet, error) {
	if err := client.validateOperation(ctx); err != nil || form == nil {
		if err != nil {
			return OAuthTokenSet{}, err
		}
		return OAuthTokenSet{}, newOAuthError(OAuthErrorInvalidRequest)
	}
	discovery, err := client.Discover(ctx)
	if err != nil {
		return OAuthTokenSet{}, err
	}
	response, payload, err := client.doOAuthRequest(ctx, http.MethodPost, discovery.TokenEndpoint, form.Encode(), oauthMaxTokenResponseBytes, true)
	if err != nil {
		return OAuthTokenSet{}, err
	}
	defer clear(payload)
	if response.StatusCode != http.StatusOK {
		return OAuthTokenSet{}, oauthStatusError(response.StatusCode, response.Header.Get("Content-Type"), payload)
	}
	if !oauthJSONContentType(response.Header.Get("Content-Type")) || validateJSON(payload) != nil {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorProtocol)
	}
	var tokenResponse struct {
		AccessToken  string          `json:"access_token"`
		TokenType    string          `json:"token_type"`
		ExpiresIn    json.Number     `json:"expires_in"`
		RefreshToken *string         `json:"refresh_token"`
		Scope        *string         `json:"scope"`
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&tokenResponse); err != nil {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorProtocol)
	}
	expiresIn, err := strconv.ParseInt(tokenResponse.ExpiresIn.String(), 10, 64)
	if err != nil || expiresIn < 1 || expiresIn > oauthMaxExpiresSeconds || !strings.EqualFold(tokenResponse.TokenType, "Bearer") || !validOAuthBearerToken(tokenResponse.AccessToken) {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorProtocol)
	}
	refreshToken := previousRefreshToken
	if tokenResponse.RefreshToken != nil {
		if !validOAuthOpaque(*tokenResponse.RefreshToken, oauthMaxTokenBytes) {
			return OAuthTokenSet{}, newOAuthError(OAuthErrorProtocol)
		}
		refreshToken = *tokenResponse.RefreshToken
	}
	if requireRefreshToken && tokenResponse.RefreshToken == nil || !validOAuthOpaque(refreshToken, oauthMaxTokenBytes) {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorProtocol)
	}
	if tokenResponse.Scope != nil && !validOAuthTokenScope(*tokenResponse.Scope) {
		return OAuthTokenSet{}, newOAuthError(OAuthErrorProtocol)
	}
	return OAuthTokenSet{AccessToken: tokenResponse.AccessToken, RefreshToken: refreshToken, TokenType: "Bearer", Scope: oauthScope, ExpiresInSeconds: expiresIn}, nil
}

func (client *OAuthClient) doOAuthRequest(ctx context.Context, method, target, encodedForm string, maximum int64, authorizeClient bool) (*http.Response, []byte, error) {
	if err := client.validateOperation(ctx); err != nil || maximum < 1 || target == "" || (method != http.MethodGet && method != http.MethodPost) || method == http.MethodGet && encodedForm != "" || method == http.MethodPost && encodedForm == "" {
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, newOAuthError(OAuthErrorInvalidRequest)
	}
	requestContext, cancel := context.WithTimeout(ctx, oauthRequestDeadline)
	defer cancel()
	var body io.Reader
	if method == http.MethodPost {
		body = strings.NewReader(encodedForm)
	}
	request, err := http.NewRequestWithContext(requestContext, method, target, body)
	if err != nil {
		return nil, nil, newOAuthError(OAuthErrorInvalidRequest)
	}
	request.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if authorizeClient {
		secret, err := client.clientSecrets.OAuthClientSecret(requestContext, client.clientID)
		if err != nil {
			return nil, nil, safeOAuthSecretProviderError(requestContext, err)
		}
		if requestContext.Err() != nil {
			clear(secret)
			return nil, nil, oauthContextOrUnavailable(requestContext)
		}
		if len(secret) < 1 || len(secret) > oauthMaxClientSecretBytes {
			clear(secret)
			return nil, nil, newOAuthError(OAuthErrorInvalidClient)
		}
		encodedClientID := url.QueryEscape(client.clientID)
		encodedSecret := url.QueryEscape(string(secret))
		clear(secret)
		request.SetBasicAuth(encodedClientID, encodedSecret)
		defer request.Header.Del("Authorization")
	}
	httpClient := *client.httpClient
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, nil, oauthContextOrUnavailable(requestContext)
	}
	payload, readErr := readOAuthBody(response.Body, maximum)
	response.Body.Close()
	if readErr != nil {
		if requestContext.Err() != nil && !errors.Is(readErr, ErrOAuthPayloadTooLarge) {
			return nil, nil, oauthContextOrUnavailable(requestContext)
		}
		return nil, nil, readErr
	}
	return response, payload, nil
}

func (client *OAuthClient) validateOperation(ctx context.Context) error {
	if client == nil || client.httpClient == nil || client.clientSecrets == nil || client.random == nil || client.issuerOrigin == "" || client.resourceOrigin == "" || client.clientID == "" || client.redirectURI == "" {
		return newOAuthError(OAuthErrorInvalidConfiguration)
	}
	if ctx == nil {
		return newOAuthError(OAuthErrorInvalidRequest)
	}
	if ctx.Err() != nil {
		return oauthContextOrUnavailable(ctx)
	}
	return nil
}

func oauthRandomBase64URL(random io.Reader, byteCount, expectedLength int) (string, error) {
	if random == nil || byteCount < 1 || expectedLength < 1 {
		return "", newOAuthError(OAuthErrorInvalidConfiguration)
	}
	raw := make([]byte, byteCount)
	defer clear(raw)
	if _, err := io.ReadFull(random, raw); err != nil {
		return "", newOAuthError(OAuthErrorUnavailable)
	}
	value := base64.RawURLEncoding.EncodeToString(raw)
	if len(value) != expectedLength {
		return "", newOAuthError(OAuthErrorProtocol)
	}
	return value, nil
}

func readOAuthBody(body io.ReadCloser, maximum int64) ([]byte, error) {
	if body == nil || maximum < 1 {
		return nil, newOAuthError(OAuthErrorProtocol)
	}
	payload, err := io.ReadAll(io.LimitReader(body, maximum+1))
	if err != nil {
		return nil, newOAuthError(OAuthErrorUnavailable)
	}
	if int64(len(payload)) > maximum {
		clear(payload)
		return nil, newOAuthError(OAuthErrorPayloadTooLarge)
	}
	return payload, nil
}

func oauthStatusError(status int, contentType string, payload []byte) error {
	if status == http.StatusBadRequest || status == http.StatusUnauthorized {
		if oauthJSONContentType(contentType) && validateJSON(payload) == nil {
			var response struct {
				Error string `json:"error"`
			}
			if json.Unmarshal(payload, &response) == nil {
				switch response.Error {
				case "invalid_grant":
					return newOAuthError(OAuthErrorInvalidGrant)
				case "invalid_client":
					return newOAuthError(OAuthErrorInvalidClient)
				}
			}
		}
		return newOAuthError(OAuthErrorInvalidRequest)
	}
	if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
		return newOAuthError(OAuthErrorUnavailable)
	}
	return newOAuthError(OAuthErrorProtocol)
}

func safeOAuthSecretProviderError(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) {
		return newOAuthError(OAuthErrorCancelled)
	}
	if ctx != nil && ctx.Err() != nil {
		return oauthContextOrUnavailable(ctx)
	}
	if errors.Is(err, ErrOAuthInvalidClient) {
		return newOAuthError(OAuthErrorInvalidClient)
	}
	return newOAuthError(OAuthErrorUnavailable)
}

func oauthContextOrUnavailable(ctx context.Context) error {
	if ctx != nil && errors.Is(ctx.Err(), context.Canceled) {
		return newOAuthError(OAuthErrorCancelled)
	}
	return newOAuthError(OAuthErrorUnavailable)
}

func oauthJSONContentType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(value))
	return err == nil && strings.EqualFold(mediaType, "application/json")
}

func oauthStringSetContains(values []string, required ...string) bool {
	if len(values) == 0 || len(values) > 64 {
		return false
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validOAuthOpaque(value, 256) {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, present := seen[value]; !present {
			return false
		}
	}
	return true
}

func validOAuthTokenScope(value string) bool {
	parts := strings.Split(value, " ")
	if len(parts) != 2 {
		return false
	}
	return parts[0] == "offline_access" && parts[1] == "mail" || parts[0] == "mail" && parts[1] == "offline_access"
}

func validOAuthBase64URL(value string, exactLength int) bool {
	if len(value) != exactLength {
		return false
	}
	for _, character := range value {
		if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}

func validOAuthOpaque(value string, maximum int) bool {
	if value == "" || len(value) > maximum || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character <= 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func validOAuthBearerToken(value string) bool {
	if value == "" || len(value) > oauthMaxTokenBytes {
		return false
	}
	padding := false
	content := false
	for _, character := range value {
		if character == '=' {
			padding = true
			continue
		}
		if padding || !(character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' || strings.ContainsRune("-._~+/", character)) {
			return false
		}
		content = true
	}
	return content
}
