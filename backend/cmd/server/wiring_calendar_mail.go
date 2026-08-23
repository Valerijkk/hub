package main

import (
	"time"

	"github.com/google/uuid"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

var (
	_ calendarmod.EventClock         = serverUTCClock{}
	_ mailmod.MailboxRepositoryClock = serverUTCClock{}
	_ calendarmod.EventIDGenerator   = serverEventIDs{}
	_ calendarmod.EventSerializer    = serverITIPSerializer{}
)

type serverUTCClock struct{}

func (serverUTCClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

type serverEventIDs struct{}

func (serverEventIDs) NewEventID() uuid.UUID {
	return uuid.New()
}

type serverITIPSerializer struct{}

func (serverITIPSerializer) Serialize(message calendarmod.ITIPMessage) ([]byte, error) {
	return calendarmod.SerializeITIP(message)
}

func wireCalendarMail(d *serverDeps) {
	calendarRepo := calendarmod.NewEntRepository(d.db)
	calendarSvc := calendarmod.NewService(calendarRepo)
	calendarHandler := calendarmod.NewHandler(calendarSvc)
	calendarHandler.RegisterRoutes(d.api, d.requireAuth)

	eventRepo := calendarmod.NewEventEntRepository(d.db)
	eventSvc := calendarmod.NewEventService(eventRepo, serverUTCClock{}, serverEventIDs{}, serverITIPSerializer{})
	eventHandler := calendarmod.NewEventHandler(eventSvc)
	eventHandler.RegisterRoutes(d.api, d.requireAuth)

	// Legacy ScheduledCall rows remain the bounded availability source until the
	// CalendarEvent-to-call bridge has completed its production cutover.
	availabilityRepo := calendarmod.NewEntAvailabilityRepository(d.db, true)
	availabilitySvc := calendarmod.NewAvailabilityService(availabilityRepo)
	availabilityHandler := calendarmod.NewAvailabilityHandler(availabilitySvc)
	availabilityHandler.RegisterRoutes(d.api, d.requireAuth)

	mailboxRepo, err := mailmod.NewMailboxEntRepository(d.db, serverUTCClock{})
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxSvc, err := mailmod.NewMailboxService(mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxHandler, err := mailmod.NewMailboxHandler(mailboxSvc)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxProjectionService, err := mailmod.NewMailboxFolderProjectionService(mailboxSvc, mailboxRepo)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	mailboxProjectionHandler, err := mailmod.NewMailboxFolderProjectionHandler(mailboxProjectionService)
	if err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
	if err := mailboxProjectionHandler.RegisterRoutes(d.api, d.requireAuth); err != nil {
		d.fatal("failed to wire calendar and mail routes", err)
	}
}
