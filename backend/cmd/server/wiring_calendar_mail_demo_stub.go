//go:build qa || release

package main

import (
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

func wireDevMailCalendarRuntime(
	_ *serverDeps,
	_ *mailmod.MailboxService,
	_ *mailmod.MailboxEntRepository,
	_ *calendarmod.EventEntRepository,
	_ *calendarmod.EventService,
	_ *calendarmod.EventReadService,
) bool {
	return false
}
