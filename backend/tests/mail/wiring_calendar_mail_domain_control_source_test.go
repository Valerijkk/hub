package mail_test

import (
	"os"
	"strings"
	"testing"
)

func mailDomainControlWiringSource(t *testing.T) string {
	t.Helper()

	source, err := os.ReadFile("../../cmd/server/wiring_calendar_mail.go")
	if err != nil {
		t.Fatalf("read calendar/mail production wiring: %v", err)
	}
	return string(source)
}

func requireMailDomainControlWiringOrdered(t *testing.T, source string, fragments ...string) {
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

func TestMailDomainControlAdminWiringSourceContract(t *testing.T) {
	source := mailDomainControlWiringSource(t)

	for _, required := range []string{
		"github.com/gofiber/fiber/v3",
		"mailmod.NewMailDomainControl(mailboxRepo)",
		"mailmod.NewMailDomainControlHandler(mailDomainControl, func(ctx fiber.Ctx, userID uuid.UUID) (bool, error) {",
		"return d.auditSvc.IsSystemAdmin(ctx.Context(), userID)",
		"mailDomainControlHandler.RegisterRoutes(d.api, d.requireAuth)",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("production wiring is missing %q", required)
		}
	}

	requireMailDomainControlWiringOrdered(t, source,
		"mailboxRepo, err := mailmod.NewMailboxEntRepository(d.db, serverUTCClock{})",
		"mailmod.NewMailDomainControl(mailboxRepo)",
		"mailmod.NewMailDomainControlHandler(mailDomainControl, func(ctx fiber.Ctx, userID uuid.UUID) (bool, error) {",
		"mailDomainControlHandler.RegisterRoutes(d.api, d.requireAuth)",
		"if wireQAMailCalendarRuntime(d, mailboxSvc, mailboxRepo, eventRepo, eventSvc) {",
		"mailboxEngineConfig, err := mailmod.ParseMailboxEngineConfig(os.LookupEnv)",
	)
}
