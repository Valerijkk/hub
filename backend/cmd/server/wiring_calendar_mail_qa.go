//go:build qa && !release

package main

import (
	"os"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
	"github.com/valerii/hub/backend/internal/qamailcalendar"
)

func wireQAMailCalendarRuntime(d *serverDeps, mailboxSvc *mailmod.MailboxService, mailboxRepo *mailmod.MailboxEntRepository, eventRepo *calendarmod.EventEntRepository, eventSvc *calendarmod.EventService) bool {
	if _, err := qamailcalendar.ParseConfig(os.LookupEnv); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}

	mailboxEngine := mailmod.NewFakeMailEngine(mailmod.FakeMailEngineConfig{
		Now: serverUTCClock{}.Now,
	})
	wireMailboxMessageDetail(d, mailboxSvc, mailboxRepo, mailboxEngine)
	wireCalendarInboundRuntime(d, mailboxRepo, mailboxEngine, eventRepo, eventSvc)
	return true
}
