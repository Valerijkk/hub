package mail

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	calendarMailDeliveryDraftKeyPrefix      = "hub.calendar.delivery.draft.v1/"
	calendarMailDeliverySubmissionKeyPrefix = "hub.calendar.delivery.submission.v1/"
	calendarMailDeliveryMaximumScanInterval = 5 * time.Minute
	calendarMailDeliveryMinimumScanInterval = time.Second
	calendarMailDeliveryMaximumSubjectBytes = 998
	calendarMailDeliveryResolveTimeout       = 5 * time.Second
	calendarMailDeliveryLeaseSafetyMargin   = time.Second
)

// CalendarMailDeliveryDraftCreator is the only draft capability required by
// Calendar delivery.
type CalendarMailDeliveryDraftCreator interface {
	CreateMailboxDraft(context.Context, uuid.UUID, MailboxDraftCreateInput) (MailboxDraftView, error)
}

// CalendarMailDeliverySubmissionCreator is the only submission capability
// required by Calendar delivery.
type CalendarMailDeliverySubmissionCreator interface {
	CreateMailboxSubmission(context.Context, uuid.UUID, uuid.UUID, MailboxSubmissionCreateInput) (MailboxSubmissionView, error)
}

// CalendarMailMaterializer is pure: it transforms a sealed, content-free
// Calendar input into request-lifetime Mail content and performs no I/O.
type CalendarMailMaterializer interface {
	MaterializeCalendarMail(CalendarMailDeliveryMaterializationInput) (CalendarMailMaterialization, error)
}

// CalendarMailMaterialization is request-lifetime content produced after a
// claim. It is never persisted by the delivery repository.
type CalendarMailMaterialization struct {
	subject string
	content MailboxContentInput
}

// NewCalendarMailMaterialization validates one bounded pure materialization.
func NewCalendarMailMaterialization(subject string, content MailboxContentInput) (CalendarMailMaterialization, error) {
	result := CalendarMailMaterialization{subject: subject, content: content}
	if !validCalendarMailMaterialization(result) {
		return CalendarMailMaterialization{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	return result, nil
}

// CalendarMailDeliveryTimer supplies bounded scan waits without granting Mail
// or repository authority.
type CalendarMailDeliveryTimer interface {
	Wait(context.Context, time.Duration) error
}

// CalendarMailDeliveryServiceOptions bounds the guarded durable due loop.
type CalendarMailDeliveryServiceOptions struct {
	Enabled      bool
	ScanInterval time.Duration
	BatchSize    int
}

// DefaultCalendarMailDeliveryServiceOptions returns the fixed conservative
// scan and batch values. Retry timing is repository-owned.
func DefaultCalendarMailDeliveryServiceOptions() CalendarMailDeliveryServiceOptions {
	return CalendarMailDeliveryServiceOptions{
		Enabled:      true,
		ScanInterval: calendarMailDeliveryDefaultScanInterval,
		BatchSize:    calendarMailDeliveryDefaultBatchSize,
	}
}

// CalendarMailDeliveryRunResult is a bounded content-free due-run summary.
type CalendarMailDeliveryRunResult struct {
	Due       int
	Delivered int
	Failed    int
}

// CalendarMailDeliveryServiceErrorKind is the closed failure family exposed by
// the Calendar-to-Mail orchestration layer.
type CalendarMailDeliveryServiceErrorKind string

const (
	CalendarMailDeliveryServiceErrorInvalid  CalendarMailDeliveryServiceErrorKind = "invalid"
	CalendarMailDeliveryServiceErrorTooLarge CalendarMailDeliveryServiceErrorKind = "too_large"
	CalendarMailDeliveryServiceErrorUnavailable CalendarMailDeliveryServiceErrorKind = "unavailable"
	CalendarMailDeliveryServiceErrorPartial   CalendarMailDeliveryServiceErrorKind = "partial"
)

const calendarMailDeliveryServiceErrorText = "calendar mail delivery service error"

// CalendarMailDeliveryServiceError contains no delivery identity, recipient,
// content, or underlying error detail.
type CalendarMailDeliveryServiceError struct {
	kind CalendarMailDeliveryServiceErrorKind
}

func (err *CalendarMailDeliveryServiceError) Error() string {
	return calendarMailDeliveryServiceErrorText
}

// Kind returns the validated closed service error kind.
func (err *CalendarMailDeliveryServiceError) Kind() CalendarMailDeliveryServiceErrorKind {
	if err == nil || !validCalendarMailDeliveryServiceErrorKind(err.kind) {
		return CalendarMailDeliveryServiceErrorInvalid
	}
	return err.kind
}

// CalendarMailDeliveryService claims one durable delivery at a time and uses
// only ordinary draft and submission services. It has no direct engine,
// transport, or external capability.
type CalendarMailDeliveryService struct {
	repository  CalendarMailDeliveryRepository
	drafts      CalendarMailDeliveryDraftCreator
	submissions CalendarMailDeliverySubmissionCreator
	materializer CalendarMailMaterializer
	timer       CalendarMailDeliveryTimer
	options     CalendarMailDeliveryServiceOptions

	guard   sync.Mutex
	running bool
}

type calendarMailDeliveryRuntimeTimer struct{}

var _ CalendarMailDeliveryTimer = calendarMailDeliveryRuntimeTimer{}

// NewCalendarMailDeliveryService constructs a disabled zero service or one
// guarded bounded due worker. Construction performs no repository or Mail I/O.
func NewCalendarMailDeliveryService(repository CalendarMailDeliveryRepository, drafts CalendarMailDeliveryDraftCreator, submissions CalendarMailDeliverySubmissionCreator, materializer CalendarMailMaterializer, timer CalendarMailDeliveryTimer, options CalendarMailDeliveryServiceOptions) (*CalendarMailDeliveryService, error) {
	if !options.Enabled {
		if options != (CalendarMailDeliveryServiceOptions{}) || !calendarMailDeliveryDependencyNil(repository) || !calendarMailDeliveryDependencyNil(drafts) || !calendarMailDeliveryDependencyNil(submissions) || !calendarMailDeliveryDependencyNil(materializer) || !calendarMailDeliveryDependencyNil(timer) {
			return nil, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
		}
		return &CalendarMailDeliveryService{}, nil
	}
	if calendarMailDeliveryDependencyNil(repository) || calendarMailDeliveryDependencyNil(drafts) || calendarMailDeliveryDependencyNil(submissions) || calendarMailDeliveryDependencyNil(materializer) || calendarMailDeliveryDependencyNil(timer) {
		return nil, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	configured, err := calendarMailDeliveryServiceOptionsConfigured(options)
	if err != nil {
		return nil, err
	}
	return &CalendarMailDeliveryService{repository: repository, drafts: drafts, submissions: submissions, materializer: materializer, timer: timer, options: configured}, nil
}

// NewCalendarMailDeliveryTimer returns the production bounded wait timer.
func NewCalendarMailDeliveryTimer() CalendarMailDeliveryTimer {
	return calendarMailDeliveryRuntimeTimer{}
}

func (calendarMailDeliveryRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if calendarMailDeliveryDependencyNil(ctx) {
		return context.Canceled
	}
	timer := time.NewTimer(duration)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Deliver claims and processes one opaque delivery identity. A stale claim is
// superseded by the repository before materialization or any Mail service call.
func (service *CalendarMailDeliveryService) Deliver(ctx context.Context, deliveryID CalendarMailDeliveryID) error {
	if !calendarMailDeliveryServiceValid(service) || calendarMailDeliveryDependencyNil(ctx) || !validCalendarMailDeliveryID(deliveryID) {
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if !service.options.Enabled || ctx.Err() != nil {
		return nil
	}
	return service.deliver(ctx, deliveryID)
}

// RunOnce scans exactly one repository-owned stable due batch and processes
// each retained opaque ID once. It does not retry within this invocation.
func (service *CalendarMailDeliveryService) RunOnce(ctx context.Context) (CalendarMailDeliveryRunResult, error) {
	if !calendarMailDeliveryServiceValid(service) || calendarMailDeliveryDependencyNil(ctx) {
		return CalendarMailDeliveryRunResult{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if !service.enter() {
		return CalendarMailDeliveryRunResult{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	defer service.leave()
	return service.runOnce(ctx)
}

// Run keeps the guarded due loop active until the parent context is cancelled.
func (service *CalendarMailDeliveryService) Run(ctx context.Context) (resultErr error) {
	if !calendarMailDeliveryServiceValid(service) || calendarMailDeliveryDependencyNil(ctx) {
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if !service.enter() {
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	defer func() {
		service.leave()
		if recover() != nil {
			resultErr = newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
		}
	}()
	if !service.options.Enabled {
		return nil
	}
	for {
		if ctx.Err() != nil {
			return nil
		}
		_, err := service.runOnce(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if kind := calendarMailDeliveryServiceErrorKindOf(err); kind == CalendarMailDeliveryServiceErrorInvalid || kind == CalendarMailDeliveryServiceErrorTooLarge {
			return err
		}
		if waitErr := service.wait(ctx); waitErr != nil {
			return waitErr
		}
	}
}

func (service *CalendarMailDeliveryService) runOnce(ctx context.Context) (CalendarMailDeliveryRunResult, error) {
	if ctx.Err() != nil || !service.options.Enabled {
		return CalendarMailDeliveryRunResult{}, nil
	}
	scan, err := service.safeScan(ctx)
	if ctx.Err() != nil {
		return CalendarMailDeliveryRunResult{}, nil
	}
	if err != nil {
		return CalendarMailDeliveryRunResult{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
	}
	if !validCalendarMailDeliveryDueScanResult(CalendarMailDeliveryDueScanRequest{Limit: service.options.BatchSize}, scan) {
		return CalendarMailDeliveryRunResult{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	result := CalendarMailDeliveryRunResult{Due: len(scan.DeliveryIDs)}
	for _, deliveryID := range scan.DeliveryIDs {
		if ctx.Err() != nil {
			return result, nil
		}
		if err := service.deliver(ctx, deliveryID); err != nil {
			result.Failed++
			continue
		}
		result.Delivered++
	}
	if result.Failed != 0 {
		return result, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorPartial)
	}
	return result, nil
}

func (service *CalendarMailDeliveryService) deliver(ctx context.Context, deliveryID CalendarMailDeliveryID) error {
	claimResult, err := service.safeClaim(ctx, deliveryID)
	if err != nil {
		return calendarMailDeliveryRepositoryServiceError(err)
	}
	switch claimResult.Outcome() {
	case CalendarMailDeliveryClaimOutcomeNoop,
		CalendarMailDeliveryClaimOutcomeBusy,
		CalendarMailDeliveryClaimOutcomeSuperseded,
		CalendarMailDeliveryClaimOutcomeObserved:
		return nil
	case CalendarMailDeliveryClaimOutcomeClaimed:
		// The sealed claim is extracted below.
	default:
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	claim, claimed := claimResult.Claim()
	if !claimed {
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeMailboxUnavailable)
	}
	deliveryContext, cancel, valid := calendarMailDeliveryExternalContext(ctx, claim.leaseUntil)
	if !valid {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeMailboxUnavailable)
	}
	defer cancel()
	if deliveryContext.Err() != nil {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeMailboxUnavailable)
	}

	materialized, materializeErr := service.safeMaterialize(claim.materialization)
	if materializeErr != nil || !validCalendarMailMaterialization(materialized) {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeInternal)
	}
	if deliveryContext.Err() != nil {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeMailboxUnavailable)
	}

	draftInput, draftInputErr := calendarMailDeliveryDraftCreateInput(claim, materialized)
	if draftInputErr != nil {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeInvalid)
	}
	draft, draftErr := service.safeCreateMailboxDraft(deliveryContext, claim.SenderActorID(), draftInput)
	if draftErr != nil {
		return service.resolve(ctx, claim, calendarMailDeliveryDraftResolveCode(draftErr))
	}
	if !validCalendarMailDeliveryDraftView(draft, claim.deliveryID) {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeInternal)
	}
	if deliveryContext.Err() != nil {
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeMailboxUnavailable)
	}

	submission, submissionErr := service.safeCreateMailboxSubmission(deliveryContext, claim.SenderActorID(), draft.ID, MailboxSubmissionCreateInput{
		ExpectedRevision: draft.Revision,
		IdempotencyKey:   calendarMailDeliverySubmissionKey(claim.deliveryID),
	})
	if submission.ID == uuid.Nil {
		if submissionErr != nil {
			// No submission ID crossed the durable boundary. The stable key permits
			// repository-owned transient retry while this unbound claim is released.
			return service.resolve(ctx, claim, calendarMailDeliverySubmissionPreBindResolveCode(submissionErr))
		}
		return service.resolve(ctx, claim, CalendarMailDeliveryResolveCodeInternal)
	}

	bindResult, bindErr := service.safeBind(ctx, CalendarMailDeliverySubmissionBindRequest{Claim: claim, SubmissionID: submission.ID})
	if bindErr != nil {
		// The submission ID exists, but bind may have committed. Do not resolve an
		// unbound claim: stale recovery can replay the stable key or observe the
		// committed same-ID binding without inventing an unknown submission.
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
	}
	switch bindResult.Outcome() {
	case CalendarMailDeliverySubmissionBindOutcomeBound, CalendarMailDeliverySubmissionBindOutcomeReplay:
		// Binding is the durable unbound-to-bound boundary. The repository owns
		// every later submission observation, so this invocation never resolves.
		return nil
	case CalendarMailDeliverySubmissionBindOutcomeConflict:
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
	default:
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
	}
}

func calendarMailDeliveryExternalContext(ctx context.Context, leaseUntil time.Time) (context.Context, context.CancelFunc, bool) {
	if calendarMailDeliveryDependencyNil(ctx) || !validCalendarMailDeliveryLeaseUntil(leaseUntil) {
		return nil, nil, false
	}
	deadline := leaseUntil.Add(-calendarMailDeliveryLeaseSafetyMargin)
	if !deadline.After(time.Now().UTC()) {
		return nil, nil, false
	}
	deliveryContext, cancel := context.WithDeadline(ctx, deadline)
	if deliveryContext.Err() != nil {
		cancel()
		return nil, nil, false
	}
	return deliveryContext, cancel, true
}

func (service *CalendarMailDeliveryService) resolve(ctx context.Context, claim CalendarMailDeliveryClaim, code CalendarMailDeliveryResolveCode) error {
	request := CalendarMailDeliveryResolveRequest{Claim: claim, Code: code}
	if calendarMailDeliveryDependencyNil(ctx) || !validCalendarMailDeliveryResolveRequest(request) {
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	resolveContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), calendarMailDeliveryResolveTimeout)
	defer cancel()
	if err := service.safeResolve(resolveContext, request); err != nil {
		return calendarMailDeliveryRepositoryServiceError(err)
	}
	return nil
}

func (service *CalendarMailDeliveryService) safeScan(ctx context.Context) (result CalendarMailDeliveryDueScanResult, resultErr error) {
	defer func() {
		if recover() != nil {
			result = CalendarMailDeliveryDueScanResult{}
			resultErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	}()
	return service.repository.ScanDueCalendarMailDeliveries(ctx, CalendarMailDeliveryDueScanRequest{Limit: service.options.BatchSize})
}

func (service *CalendarMailDeliveryService) safeClaim(ctx context.Context, deliveryID CalendarMailDeliveryID) (result CalendarMailDeliveryClaimResult, resultErr error) {
	defer func() {
		if recover() != nil {
			result = CalendarMailDeliveryClaimResult{}
			resultErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	}()
	return service.repository.ClaimCalendarMailDelivery(ctx, deliveryID)
}

func (service *CalendarMailDeliveryService) safeMaterialize(input CalendarMailDeliveryMaterializationInput) (result CalendarMailMaterialization, resultErr error) {
	defer func() {
		if recover() != nil {
			result = CalendarMailMaterialization{}
			resultErr = newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
		}
	}()
	return service.materializer.MaterializeCalendarMail(input)
}

func (service *CalendarMailDeliveryService) safeCreateMailboxDraft(ctx context.Context, actorID uuid.UUID, input MailboxDraftCreateInput) (result MailboxDraftView, resultErr error) {
	defer func() {
		if recover() != nil {
			result = MailboxDraftView{}
			resultErr = mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
		}
	}()
	return service.drafts.CreateMailboxDraft(ctx, actorID, input)
}

func (service *CalendarMailDeliveryService) safeCreateMailboxSubmission(ctx context.Context, actorID, draftID uuid.UUID, input MailboxSubmissionCreateInput) (result MailboxSubmissionView, resultErr error) {
	defer func() {
		if recover() != nil {
			result = MailboxSubmissionView{}
			resultErr = mailboxSubmissionServiceError(MailboxSubmissionServiceErrorUnavailable)
		}
	}()
	return service.submissions.CreateMailboxSubmission(ctx, actorID, draftID, input)
}

func (service *CalendarMailDeliveryService) safeBind(ctx context.Context, request CalendarMailDeliverySubmissionBindRequest) (result CalendarMailDeliverySubmissionBindResult, resultErr error) {
	defer func() {
		if recover() != nil {
			result = CalendarMailDeliverySubmissionBindResult{}
			resultErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	}()
	return service.repository.BindCalendarMailDeliverySubmission(ctx, request)
}

func (service *CalendarMailDeliveryService) safeResolve(ctx context.Context, request CalendarMailDeliveryResolveRequest) (resultErr error) {
	defer func() {
		if recover() != nil {
			resultErr = newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
		}
	}()
	return service.repository.ResolveCalendarMailDelivery(ctx, request)
}

func (service *CalendarMailDeliveryService) wait(ctx context.Context) (resultErr error) {
	defer func() {
		if recover() != nil {
			resultErr = newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
		}
	}()
	if err := service.timer.Wait(ctx, service.options.ScanInterval); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
	}
	return nil
}

func (service *CalendarMailDeliveryService) enter() bool {
	service.guard.Lock()
	defer service.guard.Unlock()
	if service.running {
		return false
	}
	service.running = true
	return true
}

func (service *CalendarMailDeliveryService) leave() {
	service.guard.Lock()
	service.running = false
	service.guard.Unlock()
}

func calendarMailDeliveryDraftCreateInput(
	claim CalendarMailDeliveryClaim,
	materialized CalendarMailMaterialization,
) (MailboxDraftCreateInput, error) {
	if !validCalendarMailDeliveryClaim(claim) ||
		!validCalendarMailMaterialization(materialized) {
		return MailboxDraftCreateInput{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	part, err := newCalendarMIMEPart(
		claim.deliveryID,
		claim.organizationID,
		claim.recipient,
		claim.materialization.message,
	)
	if err != nil {
		return MailboxDraftCreateInput{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	serialized := part.calendarBytes()
	part.serialized = serialized
	if !part.valid() ||
		part.deliveryID() != claim.deliveryID ||
		part.organization() != claim.organizationID ||
		part.mimeMethod() != string(claim.method) {
		return MailboxDraftCreateInput{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	partRecipient, err := NormalizeAddress(part.recipientAddress())
	if err != nil {
		return MailboxDraftCreateInput{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	claimRecipient, err := NormalizeAddress(claim.recipient)
	if err != nil || partRecipient != claimRecipient {
		return MailboxDraftCreateInput{},
			newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}

	return MailboxDraftCreateInput{
		MailboxDraftWriteInput: MailboxDraftWriteInput{
			DraftClientID:  calendarMailDeliveryDraftClientID(claim.deliveryID),
			FromIdentityID: claim.fromIdentityID,
			To:             []EngineAddress{claim.recipient},
			Subject:        materialized.subject,
			Content:        materialized.content,
		},
		IdempotencyKey: calendarMailDeliveryDraftKey(claim.deliveryID),
		calendar:       &part,
	}, nil
}

func calendarMailDeliveryDraftClientID(deliveryID CalendarMailDeliveryID) uuid.UUID {
	if !validCalendarMailDeliveryID(deliveryID) {
		return uuid.Nil
	}
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(calendarMailDeliveryDraftKey(deliveryID)))
}

func calendarMailDeliveryDraftKey(deliveryID CalendarMailDeliveryID) string {
	if !validCalendarMailDeliveryID(deliveryID) {
		return ""
	}
	return calendarMailDeliveryDraftKeyPrefix + deliveryID.UUID().String()
}

func calendarMailDeliverySubmissionKey(deliveryID CalendarMailDeliveryID) string {
	if !validCalendarMailDeliveryID(deliveryID) {
		return ""
	}
	return calendarMailDeliverySubmissionKeyPrefix + deliveryID.UUID().String()
}

func validCalendarMailDeliveryDraftView(view MailboxDraftView, deliveryID CalendarMailDeliveryID) bool {
	return view.ID != uuid.Nil && view.DraftClientID == calendarMailDeliveryDraftClientID(deliveryID) && view.Revision >= 1
}

func validCalendarMailMaterialization(value CalendarMailMaterialization) bool {
	if len(value.subject) == 0 || len(value.subject) > calendarMailDeliveryMaximumSubjectBytes || !utf8.ValidString(value.subject) || !validCalendarMailDeliverySubject(value.subject) {
		return false
	}
	if value.content.SourceFormat != MailboxContentFormatPlain && value.content.SourceFormat != MailboxContentFormatRich {
		return false
	}
	return len(value.content.Body) <= MailboxContentMaxInputBytes && utf8.ValidString(value.content.Body)
}

func validCalendarMailDeliverySubject(value string) bool {
	for _, character := range value {
		if character < 0x20 || (character >= 0x7f && character <= 0x9f) {
			return false
		}
	}
	return true
}

func calendarMailDeliveryDraftResolveCode(err error) CalendarMailDeliveryResolveCode {
	var draftError *MailboxDraftServiceError
	if !errors.As(err, &draftError) || draftError == nil {
		return CalendarMailDeliveryResolveCodeMailboxUnavailable
	}
	switch draftError.Kind() {
	case MailboxDraftServiceErrorUnavailable, MailboxDraftServiceErrorBusy, MailboxDraftServiceErrorOutcomeUnknown, MailboxDraftServiceErrorReconciliationRequired:
		return CalendarMailDeliveryResolveCodeMailboxUnavailable
	case MailboxDraftServiceErrorInvalid:
		return CalendarMailDeliveryResolveCodeInvalid
	default:
		return CalendarMailDeliveryResolveCodeRejected
	}
}

func calendarMailDeliverySubmissionPreBindResolveCode(err error) CalendarMailDeliveryResolveCode {
	var submissionError *MailboxSubmissionServiceError
	if !errors.As(err, &submissionError) || submissionError == nil {
		return CalendarMailDeliveryResolveCodeMailboxUnavailable
	}
	switch submissionError.Kind() {
	case MailboxSubmissionServiceErrorInvalid:
		return CalendarMailDeliveryResolveCodeInvalid
	case MailboxSubmissionServiceErrorDraftNotSendable, MailboxSubmissionServiceErrorIdempotencyConflict:
		return CalendarMailDeliveryResolveCodeRejected
	default:
		return CalendarMailDeliveryResolveCodeMailboxUnavailable
	}
}

func calendarMailDeliveryRepositoryServiceError(err error) error {
	switch calendarMailDeliveryRepositoryErrorKindOf(err) {
	case CalendarMailDeliveryRepositoryErrorInvalid:
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	default:
		return newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorUnavailable)
	}
}

func calendarMailDeliveryServiceOptionsConfigured(options CalendarMailDeliveryServiceOptions) (CalendarMailDeliveryServiceOptions, error) {
	defaults := DefaultCalendarMailDeliveryServiceOptions()
	if options.ScanInterval == 0 {
		options.ScanInterval = defaults.ScanInterval
	}
	if options.BatchSize == 0 {
		options.BatchSize = defaults.BatchSize
	}
	if options.ScanInterval < calendarMailDeliveryMinimumScanInterval || options.BatchSize < 1 {
		return CalendarMailDeliveryServiceOptions{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	if options.ScanInterval > calendarMailDeliveryMaximumScanInterval || options.BatchSize > calendarMailDeliveryMaximumBatchSize {
		return CalendarMailDeliveryServiceOptions{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorTooLarge)
	}
	return options, nil
}

func calendarMailDeliveryServiceValid(service *CalendarMailDeliveryService) bool {
	return service != nil && (service.options.Enabled && !calendarMailDeliveryDependencyNil(service.repository) && !calendarMailDeliveryDependencyNil(service.drafts) && !calendarMailDeliveryDependencyNil(service.submissions) && !calendarMailDeliveryDependencyNil(service.materializer) && !calendarMailDeliveryDependencyNil(service.timer) || !service.options.Enabled && calendarMailDeliveryDependencyNil(service.repository) && calendarMailDeliveryDependencyNil(service.drafts) && calendarMailDeliveryDependencyNil(service.submissions) && calendarMailDeliveryDependencyNil(service.materializer) && calendarMailDeliveryDependencyNil(service.timer))
}

func calendarMailDeliveryDependencyNil(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func newCalendarMailDeliveryServiceError(kind CalendarMailDeliveryServiceErrorKind) *CalendarMailDeliveryServiceError {
	if !validCalendarMailDeliveryServiceErrorKind(kind) {
		kind = CalendarMailDeliveryServiceErrorInvalid
	}
	return &CalendarMailDeliveryServiceError{kind: kind}
}

func validCalendarMailDeliveryServiceErrorKind(kind CalendarMailDeliveryServiceErrorKind) bool {
	switch kind {
	case CalendarMailDeliveryServiceErrorInvalid,
		CalendarMailDeliveryServiceErrorTooLarge,
		CalendarMailDeliveryServiceErrorUnavailable,
		CalendarMailDeliveryServiceErrorPartial:
		return true
	default:
		return false
	}
}

func calendarMailDeliveryServiceErrorKindOf(err error) CalendarMailDeliveryServiceErrorKind {
	var serviceError *CalendarMailDeliveryServiceError
	if errors.As(err, &serviceError) && serviceError != nil {
		return serviceError.Kind()
	}
	return CalendarMailDeliveryServiceErrorUnavailable
}
