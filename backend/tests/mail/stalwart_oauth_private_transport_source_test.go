package mail_test

import (
	"os"
	"strings"
	"testing"
)

func stalwartOAuthPrivateTransportSourceSlice(t *testing.T, source, start, end string) string {
	t.Helper()

	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source is missing start marker %q", start)
	}
	endOffset := strings.Index(source[startIndex+len(start):], end)
	if endOffset < 0 {
		t.Fatalf("source is missing end marker %q after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endOffset]
}

func stalwartOAuthPrivateTransportCompact(source string) string {
	return strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(source)
}

func requireStalwartOAuthPrivateTransportContains(t *testing.T, source string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(source, fragment) {
			t.Fatalf("source is missing %q", fragment)
		}
	}
}

func TestStalwartOAuthPrivateTransportSourceContract(t *testing.T) {
	configBytes, err := os.ReadFile("../../internal/mail/stalwart/oauth_config.go")
	if err != nil {
		t.Fatalf("read OAuth config source: %v", err)
	}
	oauthBytes, err := os.ReadFile("../../internal/mail/stalwart/oauth.go")
	if err != nil {
		t.Fatalf("read OAuth source: %v", err)
	}
	configSource := string(configBytes)
	oauthSource := string(oauthBytes)
	compactConfig := stalwartOAuthPrivateTransportCompact(configSource)
	compactOAuth := stalwartOAuthPrivateTransportCompact(oauthSource)

	requireStalwartOAuthPrivateTransportContains(t, compactConfig,
		"TransportOriginstring",
		"transportOriginstring",
		"transportOrigin,ok:=canonicalOAuthOrigin(config.TransportOrigin)",
		"transportOrigin:transportOrigin,",
	)
	if strings.Contains(compactConfig, "transportOrigin=issuerOrigin") || strings.Contains(compactConfig, "TransportOrigin:config.IssuerOrigin") {
		t.Fatal("OAuth private transport must not fall back to the public issuer")
	}

	discover := stalwartOAuthPrivateTransportSourceSlice(t, oauthSource,
		"func (client *OAuthClient) Discover",
		"func (client *OAuthClient) BeginAuthorization",
	)
	exchange := stalwartOAuthPrivateTransportSourceSlice(t, oauthSource,
		"func (client *OAuthClient) exchangeToken",
		"func (client *OAuthClient) doOAuthRequest",
	)
	validateOperation := stalwartOAuthPrivateTransportSourceSlice(t, oauthSource,
		"func (client *OAuthClient) validateOperation",
		"func oauthRandomBase64URL",
	)

	requireStalwartOAuthPrivateTransportContains(t, discover,
		"client.transportOrigin+oauthDiscoveryPath",
		"expectedAuthorization := client.issuerOrigin + oauthAuthorizationPath",
		"expectedToken := client.issuerOrigin + oauthTokenPath",
		"metadata.Issuer != client.issuerOrigin",
		"metadata.AuthorizationEndpoint != expectedAuthorization",
		"metadata.TokenEndpoint != expectedToken",
	)
	if strings.Contains(stalwartOAuthPrivateTransportCompact(discover), "client.issuerOrigin+oauthDiscoveryPath") {
		t.Fatal("OAuth discovery must not send a protocol request to the public issuer")
	}

	requireStalwartOAuthPrivateTransportContains(t, exchange,
		"if _, err := client.Discover(ctx); err != nil",
		"client.transportOrigin+oauthTokenPath",
	)
	compactExchange := stalwartOAuthPrivateTransportCompact(exchange)
	if strings.Contains(compactExchange, "client.issuerOrigin+oauthTokenPath") || strings.Contains(compactExchange, "discovery.TokenEndpoint") {
		t.Fatal("OAuth token exchange must use only the private transport target after metadata validation")
	}
	if !strings.Contains(compactOAuth, "client.transportOrigin+oauthDiscoveryPath") || !strings.Contains(compactOAuth, "client.transportOrigin+oauthTokenPath") {
		t.Fatal("OAuth protocol requests must target the private transport origin")
	}
	requireStalwartOAuthPrivateTransportContains(t, validateOperation, "client.transportOrigin == \"\"")
}
