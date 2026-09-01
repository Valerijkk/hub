//go:build !qa && !release

package main

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"

	appent "github.com/valerii/hub/backend/ent"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const (
	devMailDemoDatasetVersion = 1
	devMailDemoObserveTimeout = 2 * time.Minute
	devMailDemoObservePoll    = 5 * time.Second
)

var (
	errDevMailDemoInvalid     = errors.New("dev mail demo adapter input is invalid")
	errDevMailDemoUnavailable = errors.New("dev mail demo adapter is unavailable")
	errDevMailDemoPending     = errors.New("dev mail demo projection is pending")
)

var _ devdemo.MailDemoPort = (*devMailDemoAdapter)(nil)

// devMailDemoPrepared is private immutable first-load state. It contains no
// caller-owned scope map or provider/message payload.
type devMailDemoPrepared struct {
	organizationID uuid.UUID
	pmUserID       uuid.UUID
	dev1UserID     uuid.UUID
	control        devMailControlPlaneResult
	source         devMailSourceResult
}

// devMailDemoAdapter is the dev-only bridge between the seeded Mail source
// and the production-owned Mail/Calendar projection workers. It only observes
// public read services after first-load; it never advances a worker itself.
type devMailDemoAdapter struct {
	client         *appent.Client
	engine         *mailmod.FakeMailEngine
	mailboxSvc     *mailmod.MailboxService
	mailboxRepo    *mailmod.MailboxEntRepository
	eventReadSvc   *calendarmod.EventReadService
	projectionSvc  *mailmod.MailboxFolderProjectionService
	now            func() time.Time

	guard             sync.Mutex
	prepared          *devMailDemoPrepared
	observedPMEventID uuid.UUID
}

// newDevMailDemoAdapter constructs the server-only Mail demo port. The
// projection service is intentionally built here so the observer uses the
// same owner-gated public read boundary as HTTP.
func newDevMailDemoAdapter(
	client *appent.Client,
	engine *mailmod.FakeMailEngine,
	mailboxSvc *mailmod.MailboxService,
	mailboxRepo *mailmod.MailboxEntRepository,
	eventReadSvc *calendarmod.EventReadService,
	now func() time.Time,
) (*devMailDemoAdapter, error) {
	if devMailDemoNil(client) || devMailDemoNil(engine) || devMailDemoNil(mailboxSvc) || devMailDemoNil(mailboxRepo) || devMailDemoNil(eventReadSvc) || now == nil || !validDevMailControlNow(now()) {
		return nil, errDevMailDemoInvalid
	}
	projectionSvc, err := mailmod.NewMailboxFolderProjectionService(mailboxSvc, mailboxRepo)
	if err != nil || devMailDemoNil(projectionSvc) {
		return nil, errDevMailDemoInvalid
	}
	return &devMailDemoAdapter{
		client:        client,
		engine:        engine,
		mailboxSvc:    mailboxSvc,
		mailboxRepo:   mailboxRepo,
		eventReadSvc:  eventReadSvc,
		projectionSvc: projectionSvc,
		now:           now,
	}, nil
}

// PrepareMailDemo provisions the provider and immutable static dataset once.
// Later calls must use the same organization and Mail demo identities; they do
// not reload the source and therefore preserve interactive mutations.
func (adapter *devMailDemoAdapter) PrepareMailDemo(ctx context.Context, input devdemo.MailDemoPrepareInput) error {
	if adapter == nil || !adapter.ready() || !validDevMailControlContext(ctx) || input.OrganizationID == uuid.Nil {
		return errDevMailDemoInvalid
	}
	pmUserID, dev1UserID, ok := devMailDemoScopeUsers(input.Scope.UserIDs)
	if !ok {
		return errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return errDevMailDemoUnavailable
	}

	adapter.guard.Lock()
	defer adapter.guard.Unlock()

	if adapter.prepared != nil {
		if adapter.prepared.matches(input.OrganizationID, pmUserID, dev1UserID) {
			return nil
		}
		return errDevMailDemoInvalid
	}
	now := adapter.now()
	if !validDevMailControlNow(now) {
		return errDevMailDemoInvalid
	}
	users := map[string]uuid.UUID{
		devMailSourcePMEmail:   pmUserID,
		devMailSourceDev1Email: dev1UserID,
	}
	control, err := reconcileDevMailControlPlane(ctx, adapter.client, adapter.engine, input.OrganizationID, users, now)
	if err != nil || !validDevMailControlResult(control) {
		return errDevMailDemoUnavailable
	}
	source, err := seedDevMailSource(ctx, adapter.engine, control)
	if err != nil || !validDevMailSourceResult(source) {
		return errDevMailDemoUnavailable
	}
	adapter.prepared = &devMailDemoPrepared{
		organizationID: input.OrganizationID,
		pmUserID:       pmUserID,
		dev1UserID:     dev1UserID,
		control:        cloneDevMailControlResult(control),
		source:         cloneDevMailSourceResult(source),
	}
	return nil
}

// ReconcileMailDemo observes the production projection path until the seeded
// Calendar invitation has an outbound PM message and an independent inbound
// dev1 receipt. It deliberately does not call worker step/run APIs.
func (adapter *devMailDemoAdapter) ReconcileMailDemo(ctx context.Context, input devdemo.MailDemoInput) (devdemo.MailDemoResult, error) {
	if adapter == nil || !adapter.ready() || !validDevMailControlContext(ctx) || input.CalendarEvent.Key != devdemo.MailInvitationEventKey || input.CalendarEvent.EventID == uuid.Nil {
		return devdemo.MailDemoResult{}, errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	pmUserID, dev1UserID, ok := devMailDemoScopeUsers(input.Scope.UserIDs)
	if !ok {
		return devdemo.MailDemoResult{}, errDevMailDemoInvalid
	}

	adapter.guard.Lock()
	defer adapter.guard.Unlock()
	if adapter.prepared == nil || !adapter.prepared.matches(adapter.prepared.organizationID, pmUserID, dev1UserID) {
		return devdemo.MailDemoResult{}, errDevMailDemoInvalid
	}
	if err := adapter.observe(ctx, *adapter.prepared, input.CalendarEvent.EventID); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	replayed := 0
	if adapter.observedPMEventID == input.CalendarEvent.EventID {
		replayed = 1
	} else {
		adapter.observedPMEventID = input.CalendarEvent.EventID
	}
	return devMailDemoResult(*adapter.prepared, replayed)
}

func (adapter *devMailDemoAdapter) observe(ctx context.Context, prepared devMailDemoPrepared, pmEventID uuid.UUID) error {
	if ctx == nil {
		return errDevMailDemoUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, devMailDemoObserveTimeout)
	defer cancel()
	for {
		err := adapter.observeOnce(deadlineCtx, prepared, pmEventID)
		if err == nil {
			return nil
		}
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if deadlineErr := deadlineCtx.Err(); deadlineErr != nil {
			return deadlineErr
		}
		if !errors.Is(err, errDevMailDemoPending) {
			return errDevMailDemoUnavailable
		}
		timer := time.NewTimer(devMailDemoObservePoll)
		select {
		case <-deadlineCtx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return deadlineCtx.Err()
		case <-timer.C:
		}
	}
}

func devMailDemoResult(prepared devMailDemoPrepared, replayed int) (devdemo.MailDemoResult, error) {
	source := prepared.source
	if replayed < 0 || replayed > 1 || !validDevMailSourceResult(source) ||
		source.accountsTotal != 2 || source.foldersTotal != 12 || source.threadsTotal != 3 ||
		source.messagesTotal != 4 || source.attachmentsTotal != 1 {
		return devdemo.MailDemoResult{}, errDevMailDemoUnavailable
	}
	return devdemo.MailDemoResult{
		DatasetVersion:       devMailDemoDatasetVersion,
		AccountsTotal:        source.accountsTotal,
		FoldersTotal:         source.foldersTotal,
		ThreadsTotal:         source.threadsTotal,
		MessagesTotal:        source.messagesTotal,
		AttachmentsTotal:     source.attachmentsTotal,
		InvitationsTotal:     1,
		InvitationsReplayed: replayed,
	}, nil
}

func devMailDemoSourceResultEqual(left, right devMailSourceResult) bool {
	return reflect.DeepEqual(left, right)
}

func (adapter *devMailDemoAdapter) ready() bool {
	return adapter != nil && !devMailDemoNil(adapter.client) && !devMailDemoNil(adapter.engine) &&
		!devMailDemoNil(adapter.mailboxSvc) && !devMailDemoNil(adapter.mailboxRepo) &&
		!devMailDemoNil(adapter.eventReadSvc) && !devMailDemoNil(adapter.projectionSvc) &&
		adapter.now != nil
}

func (prepared devMailDemoPrepared) matches(organizationID, pmUserID, dev1UserID uuid.UUID) bool {
	return organizationID != uuid.Nil && pmUserID != uuid.Nil && dev1UserID != uuid.Nil &&
		prepared.organizationID == organizationID && prepared.pmUserID == pmUserID && prepared.dev1UserID == dev1UserID
}

func devMailDemoScopeUsers(userIDs map[string]uuid.UUID) (uuid.UUID, uuid.UUID, bool) {
	if userIDs == nil {
		return uuid.Nil, uuid.Nil, false
	}
	pmUserID, pmOK := userIDs[devMailSourcePMEmail]
	dev1UserID, dev1OK := userIDs[devMailSourceDev1Email]
	if !pmOK || !dev1OK || pmUserID == uuid.Nil || dev1UserID == uuid.Nil || pmUserID == dev1UserID {
		return uuid.Nil, uuid.Nil, false
	}
	return pmUserID, dev1UserID, true
}

func devMailDemoNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
