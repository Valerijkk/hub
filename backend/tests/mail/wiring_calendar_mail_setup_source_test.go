package mail_test

import (
	"os"
	"strings"
	"testing"
)

func mailboxCalendarMailWiringSource(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile("../../cmd/server/wiring_calendar_mail.go")
	if err != nil {
		t.Fatalf("read calendar/mail production wiring: %v", err)
	}
	return string(source)
}

func requireMailboxCalendarMailWiringOrdered(t *testing.T, source string, fragments ...string) {
	t.Helper()

	previous := -1
	for _, fragment := range fragments {
		position := strings.Index(source, fragment)
		if position < 0 {
			t.Fatalf("production wiring is missing %q", fragment)
		}
		if position <= previous {
			t.Fatalf("production wiring puts %q before its required predecessor", fragment)
		}
		previous = position
	}
}

func TestCalendarMailProductionSetupWiringSourceContract(t *testing.T) {
	source := mailboxCalendarMailWiringSource(t)

	for _, required := range []string{
		"!publicURLPresent || stalwartPublicURL == \"\"",
		"IssuerOrigin:         stalwartPublicURL",
		"ResourceOrigin:       stalwartPublicURL",
		"mailmod.NewMailboxOAuthService(mailboxRepo, oauthStates, oauthClient, oauthAccountProof, d.integrationCipher, serverUTCClock{}, stalwartPublicURL)",
		"mailmod.NewStalwartMailboxAuthorizerSource(stalwartPublicURL, mailboxRepo, mailboxOAuthService, d.integrationCipher, serverUTCClock{})",
		"MAILBOX_ENGINE_REGISTRY_API_KEY",
		"mailmod.NewStalwartRegistryAPIKeyAuthorizer(mailboxEngineConfig.Origin, func(ctx context.Context) ([]byte, error) {",
		"if ctx == nil {",
		"if err := ctx.Err(); err != nil {",
		"registryAPIKey, registryAPIKeyPresent := os.LookupEnv(",
		"if !registryAPIKeyPresent || registryAPIKey == ",
		"return append([]byte(nil), registryAPIKey...), nil",
		"mailmod.NewStalwartRegistryClient(mailboxEngineConfig.Origin, mailboxHTTPClient, managementRegistryAuthorizer)",
		"stalwartmod.NewRegistryPasswordClient(managementRegistryClient)",
		"mailmod.NewMailboxPasswordEnrollmentService(mailboxRepo, managementPasswordClient)",
		"mailmod.NewStalwartMailboxRegistryAuthorizer()",
		"mailmod.NewStalwartMailboxOAuthAccountProof(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer)",
		"mailmod.NewStalwartMailboxClientFactory(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer, mailboxAuthorizers)",
		"mailmod.NewMailboxBootstrapOAuthRedisStateStore(d.rdb)",
		"mailmod.NewStalwartMailboxBootstrapIdentityProof(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer)",
		"mailmod.NewMailboxBootstrapOAuthService(mailboxRepo, bootstrapOAuthStates, oauthClient, bootstrapIdentityProof, d.integrationCipher, serverUTCClock{}, stalwartPublicURL)",
		"mailmod.NewMailboxSetupHandler(managementPasswordService, mailboxBootstrapOAuthService, mailboxSvc, mailmod.MailboxSetupHandlerConfig{",
		"BindingCookieName:      \"hub_mail_setup_oauth_binding\"",
		"BindingCookieSecure:    !d.cfg.Server.DevMode",
		"BindingCookieMaxAgeSec: 600",
		"MaximumJSONBytes:       9216",
		"MaximumQueryBytes:      2048",
		"MaximumPasswordBytes:   256",
		"mailboxSetupHandler.RegisterRoutes(d.api, d.requireAuth)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("production setup wiring is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"stalwartPublicURL != mailboxEngineConfig.Origin",
		"mailmod.NewMailboxDomainProvisioningWorker",
		"mailmod.NewMailboxAccountProvisioningWorker",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("production setup wiring contains forbidden %q", forbidden)
		}
	}

	requireMailboxCalendarMailWiringOrdered(t, source,
		"managementRegistryAuthorizer, err := mailmod.NewStalwartRegistryAPIKeyAuthorizer",
		"managementRegistryClient, err := mailmod.NewStalwartRegistryClient",
		"managementPasswordClient, err := stalwartmod.NewRegistryPasswordClient",
		"managementPasswordService, err := mailmod.NewMailboxPasswordEnrollmentService",
		"mailboxRegistryAuthorizer := mailmod.NewStalwartMailboxRegistryAuthorizer()",
		"mailmod.NewStalwartMailboxOAuthAccountProof(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer)",
		"mailmod.NewStalwartMailboxClientFactory(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer, mailboxAuthorizers)",
		"bootstrapOAuthStates, err := mailmod.NewMailboxBootstrapOAuthRedisStateStore(d.rdb)",
		"bootstrapIdentityProof, err := mailmod.NewStalwartMailboxBootstrapIdentityProof(mailboxEngineConfig.Origin, mailboxHTTPClient, mailboxRegistryAuthorizer)",
		"mailboxBootstrapOAuthService, err := mailmod.NewMailboxBootstrapOAuthService",
		"mailboxSetupHandler, err := mailmod.NewMailboxSetupHandler",
		"mailboxSetupHandler.RegisterRoutes(d.api, d.requireAuth)",
	)
	for _, retained := range []string{
		"mailboxOAuthHandler.RegisterRoutes(d.api, d.requireAuth)",
		"NewMailboxDraftService",
		"NewMailboxReplyDraftService",
		"NewMailboxReplyAllDraftService",
		"NewMailboxForwardDraftService",
		"mailboxAttachmentStagingHandler.RegisterRoutes(d.api, d.requireAuth)",
		"NewMailboxSyncWorker",
		"NewMailboxSubmissionWorker",
		"wireCalendarInboundRuntime(d, mailboxRepo, mailboxEngine, eventRepo, eventSvc)",
	} {
		if !strings.Contains(source, retained) {
			t.Fatalf("production wiring no longer retains %q", retained)
		}
	}
}
