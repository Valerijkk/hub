//go:build !qa && !release

package main

import (
	"errors"
	"os"

	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

func wireDevMailCalendarRuntime(
	d *serverDeps,
	_ *mailmod.MailboxService,
	_ *mailmod.MailboxEntRepository,
	_ *calendarmod.EventEntRepository,
	eventSvc *calendarmod.EventService,
	eventReadSvc *calendarmod.EventReadService,
) bool {
	config, err := devdemo.ParseMailDemoRuntimeConfig(os.LookupEnv, d.cfg.Server.DevMode)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
		return true
	}

	if d.cfg.Server.DevMode {
		calendarDemoAdapter, err := newCalendarDemoAdapter(eventSvc, eventReadSvc)
		if err != nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", err)
			return true
		}
		if calendarDemoAdapter == nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", errors.New("calendar demo adapter is unavailable"))
			return true
		}

		mailCalendarDemoSeeder, err := devdemo.NewMailCalendarDemoSeeder(d.db, d.rawDB, d.postgresAdvisoryLocks, calendarDemoAdapter)
		if err != nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", err)
			return true
		}
		if mailCalendarDemoSeeder == nil {
			d.fatal("failed to wire Mail/Calendar demo seeder", errors.New("demo seeder is unavailable"))
			return true
		}
		d.mailCalendarDemoSeeder = mailCalendarDemoSeeder
	}

	if config.Enabled {
		d.fatal("failed to wire calendar and mail routes", errors.New("demo runtime is not wired"))
		return true
	}

	return false
}
