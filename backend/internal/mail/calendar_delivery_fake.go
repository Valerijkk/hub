package mail

import (
	"context"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CalendarMailDeliveryFakeClaim configures one deterministic repository claim
// for an isolated unit. It has no content field: the fake materializer owns
// its fixed request-lifetime output.
type CalendarMailDeliveryFakeClaim struct {
	DeliveryID                 CalendarMailDeliveryID
	InviteID                   CalendarInviteDeliveryID
	ReplyID                    CalendarReplyDeliveryID
	Kind                       CalendarMailDeliveryKind
	Method                     CalendarMailDeliveryMethod
	EventID                    uuid.UUID
	AttendeeID                 uuid.UUID
	OrganizerID                uuid.UUID
	OrganizationID             uuid.UUID
	AccountID                  uuid.UUID
	SenderActorID              uuid.UUID
	FromIdentityID             uuid.UUID
	Recipient                  EngineAddress
	EventSequence              int
	AttendeeVersion            int
	RecurrenceID               string
	ClaimGeneration            CalendarMailDeliveryClaimGeneration
	LeaseUntil                 time.Time
}

// CalendarMailDeliveryRepositoryFake is a deterministic in-memory durable
// boundary for units. Every returned slice, claim, and recorded request is
// defensively copied.
type CalendarMailDeliveryRepositoryFake struct {
	mu         sync.Mutex
	now        time.Time
	due        []CalendarMailDeliveryID
	claims     map[CalendarMailDeliveryID]CalendarMailDeliveryClaim
	bindings   map[CalendarMailDeliveryID]uuid.UUID
	inFlight   map[CalendarMailDeliveryID]struct{}
	admissions []CalendarInviteDeliveryAdmissionRequest
	wakes      []CalendarReplyDeliveryWakeRequest
	resolves   []CalendarMailDeliveryResolveRequest
}

var _ CalendarMailDeliveryRepository = (*CalendarMailDeliveryRepositoryFake)(nil)

// NewCalendarMailDeliveryRepositoryFake creates an empty deterministic fake.
func NewCalendarMailDeliveryRepositoryFake() *CalendarMailDeliveryRepositoryFake {
	return &CalendarMailDeliveryRepositoryFake{
		now:      time.Unix(0, 0).UTC(),
		due:      []CalendarMailDeliveryID{},
		claims:   make(map[CalendarMailDeliveryID]CalendarMailDeliveryClaim),
		bindings: make(map[CalendarMailDeliveryID]uuid.UUID),
		inFlight: make(map[CalendarMailDeliveryID]struct{}),
	}
}

// SetNow sets the fake's deterministic canonical clock. It is synchronized
// with every repository operation, so tests can advance it without a race.
func (fake *CalendarMailDeliveryRepositoryFake) SetNow(now time.Time) error {
	if fake == nil || !validCalendarMailDeliveryLeaseUntil(now) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	fake.now = now
	fake.mu.Unlock()
	return nil
}

// SetDue replaces the fake's stable due order after validating every opaque ID.
func (fake *CalendarMailDeliveryRepositoryFake) SetDue(deliveryIDs []CalendarMailDeliveryID) error {
	if fake == nil || deliveryIDs == nil || len(deliveryIDs) > calendarMailDeliveryMaximumBatchSize {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	result := CalendarMailDeliveryDueScanResult{DeliveryIDs: append([]CalendarMailDeliveryID(nil), deliveryIDs...)}
	if !validCalendarMailDeliveryDueScanResult(CalendarMailDeliveryDueScanRequest{Limit: calendarMailDeliveryMaximumBatchSize}, result) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	fake.due = result.DeliveryIDs
	fake.mu.Unlock()
	return nil
}

// SetClaim installs one sealed fake claim. Its account, actor, identity, and
// recipient values remain repository-side configuration, never service input.
func (fake *CalendarMailDeliveryRepositoryFake) SetClaim(input CalendarMailDeliveryFakeClaim) error {
	if fake == nil {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	claim, err := calendarMailDeliveryFakeClaim(input)
	if err != nil {
		return err
	}
	fake.mu.Lock()
	fake.claims[claim.deliveryID] = calendarMailDeliveryClaimCopy(claim)
	delete(fake.bindings, claim.deliveryID)
	delete(fake.inFlight, claim.deliveryID)
	fake.mu.Unlock()
	return nil
}

// AdmitCalendarInviteDeliveries records one validated immutable admission.
func (fake *CalendarMailDeliveryRepositoryFake) AdmitCalendarInviteDeliveries(_ context.Context, request CalendarInviteDeliveryAdmissionRequest) error {
	if fake == nil || !validCalendarInviteDeliveryAdmissionRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	fake.admissions = append(fake.admissions, calendarMailDeliveryAdmissionRequestCopy(request))
	fake.mu.Unlock()
	return nil
}

// WakeCalendarReplyDelivery records one reply wake without exposing row data.
func (fake *CalendarMailDeliveryRepositoryFake) WakeCalendarReplyDelivery(_ context.Context, request CalendarReplyDeliveryWakeRequest) error {
	if fake == nil || !validCalendarReplyDeliveryWakeRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	fake.wakes = append(fake.wakes, request)
	fake.mu.Unlock()
	return nil
}

// ScanDueCalendarMailDeliveries returns the configured repository order and
// rejects a caller that asks beyond the fixed boundary.
func (fake *CalendarMailDeliveryRepositoryFake) ScanDueCalendarMailDeliveries(_ context.Context, request CalendarMailDeliveryDueScanRequest) (CalendarMailDeliveryDueScanResult, error) {
	if fake == nil || !validCalendarMailDeliveryDueScanRequest(request) {
		return CalendarMailDeliveryDueScanResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if len(fake.due) > request.Limit {
		return CalendarMailDeliveryDueScanResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorUnavailable)
	}
	return calendarMailDeliveryDueScanResultCopy(CalendarMailDeliveryDueScanResult{DeliveryIDs: fake.due}), nil
}

// ClaimCalendarMailDelivery returns a copied unbound configured claim exactly
// once while it is in flight. A bound delivery is observed locally and never
// returned as an external claim, matching the durable repository contract.
func (fake *CalendarMailDeliveryRepositoryFake) ClaimCalendarMailDelivery(_ context.Context, deliveryID CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error) {
	if fake == nil || !validCalendarMailDeliveryID(deliveryID) {
		return CalendarMailDeliveryClaimResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	claim, found := fake.claims[deliveryID]
	if !found {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
	}
	if _, bound := fake.bindings[deliveryID]; bound {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeObserved}, nil
	}
	if !claim.leaseUntil.After(fake.now) {
		reclaimed, ok := calendarMailDeliveryFakeReclaim(claim, fake.now)
		if !ok {
			return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeNoop}, nil
		}
		fake.claims[deliveryID] = calendarMailDeliveryClaimCopy(reclaimed)
		fake.inFlight[deliveryID] = struct{}{}
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeClaimed, claim: calendarMailDeliveryClaimCopy(reclaimed)}, nil
	}
	if _, claimed := fake.inFlight[deliveryID]; claimed {
		return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeBusy}, nil
	}
	fake.inFlight[deliveryID] = struct{}{}
	return CalendarMailDeliveryClaimResult{outcome: CalendarMailDeliveryClaimOutcomeClaimed, claim: calendarMailDeliveryClaimCopy(claim)}, nil
}

// BindCalendarMailDeliverySubmission preserves an existing same submission and
// reports a different durable submission as a closed conflict.
func (fake *CalendarMailDeliveryRepositoryFake) BindCalendarMailDeliverySubmission(_ context.Context, request CalendarMailDeliverySubmissionBindRequest) (CalendarMailDeliverySubmissionBindResult, error) {
	if fake == nil || !validCalendarMailDeliverySubmissionBindRequest(request) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	claim, found := fake.claims[request.Claim.deliveryID]
	if !found || claim.claimToken != request.Claim.claimToken || claim.claimGeneration != request.Claim.claimGeneration ||
		claim.organizationID != request.Claim.organizationID || !claim.leaseUntil.Equal(request.Claim.leaseUntil) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if boundSubmissionID, bound := fake.bindings[request.Claim.deliveryID]; bound {
		if boundSubmissionID == request.SubmissionID {
			return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeReplay}, nil
		}
		return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeConflict}, nil
	}
	if !claim.leaseUntil.After(fake.now) {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if _, claimed := fake.inFlight[request.Claim.deliveryID]; !claimed {
		return CalendarMailDeliverySubmissionBindResult{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.bindings[request.Claim.deliveryID] = request.SubmissionID
	delete(fake.inFlight, request.Claim.deliveryID)
	return CalendarMailDeliverySubmissionBindResult{outcome: CalendarMailDeliverySubmissionBindOutcomeBound}, nil
}

// ResolveCalendarMailDelivery records one copied safe resolution.
func (fake *CalendarMailDeliveryRepositoryFake) ResolveCalendarMailDelivery(_ context.Context, request CalendarMailDeliveryResolveRequest) error {
	if fake == nil || !validCalendarMailDeliveryResolveRequest(request) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	claim, found := fake.claims[request.Claim.deliveryID]
	if !found || claim.claimToken != request.Claim.claimToken || claim.claimGeneration != request.Claim.claimGeneration ||
		claim.organizationID != request.Claim.organizationID || !claim.leaseUntil.Equal(request.Claim.leaseUntil) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if _, bound := fake.bindings[request.Claim.deliveryID]; bound {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if !claim.leaseUntil.After(fake.now) {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	if _, claimed := fake.inFlight[request.Claim.deliveryID]; !claimed {
		return newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	fake.resolves = append(fake.resolves, CalendarMailDeliveryResolveRequest{Claim: calendarMailDeliveryClaimCopy(request.Claim), Code: request.Code})
	delete(fake.inFlight, request.Claim.deliveryID)
	return nil
}

// Admissions returns defensive copies of recorded durable invite admissions.
func (fake *CalendarMailDeliveryRepositoryFake) Admissions() []CalendarInviteDeliveryAdmissionRequest {
	if fake == nil {
		return nil
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	result := make([]CalendarInviteDeliveryAdmissionRequest, len(fake.admissions))
	for index, request := range fake.admissions {
		result[index] = calendarMailDeliveryAdmissionRequestCopy(request)
	}
	return result
}

// Wakes returns copied reply wake requests.
func (fake *CalendarMailDeliveryRepositoryFake) Wakes() []CalendarReplyDeliveryWakeRequest {
	if fake == nil {
		return nil
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return append([]CalendarReplyDeliveryWakeRequest(nil), fake.wakes...)
}

// Resolutions returns copied closed delivery conclusions.
func (fake *CalendarMailDeliveryRepositoryFake) Resolutions() []CalendarMailDeliveryResolveRequest {
	if fake == nil {
		return nil
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	result := make([]CalendarMailDeliveryResolveRequest, len(fake.resolves))
	for index, request := range fake.resolves {
		result[index] = CalendarMailDeliveryResolveRequest{Claim: calendarMailDeliveryClaimCopy(request.Claim), Code: request.Code}
	}
	return result
}

// CalendarMailMaterializerFake is a deterministic pure materializer for units.
type CalendarMailMaterializerFake struct{}

var _ CalendarMailMaterializer = CalendarMailMaterializerFake{}

// NewCalendarMailMaterializerFake creates the fixed pure materializer.
func NewCalendarMailMaterializerFake() CalendarMailMaterializer {
	return CalendarMailMaterializerFake{}
}

// MaterializeCalendarMail creates a fixed bounded Mail representation only
// after the content-free sealed input validates.
func (CalendarMailMaterializerFake) MaterializeCalendarMail(input CalendarMailDeliveryMaterializationInput) (CalendarMailMaterialization, error) {
	if !validCalendarMailDeliveryMaterializationInput(input) {
		return CalendarMailMaterialization{}, newCalendarMailDeliveryServiceError(CalendarMailDeliveryServiceErrorInvalid)
	}
	return CalendarMailMaterialization{
		subject: "Calendar update",
		content: MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: "Calendar delivery"},
	}, nil
}

func calendarMailDeliveryAdmissionRequestCopy(request CalendarInviteDeliveryAdmissionRequest) CalendarInviteDeliveryAdmissionRequest {
	request.Invites = append([]CalendarInviteDeliveryAdmission(nil), request.Invites...)
	return request
}

func calendarMailDeliveryFakeClaimToken(deliveryID CalendarMailDeliveryID, generation CalendarMailDeliveryClaimGeneration) CalendarMailDeliveryClaimToken {
	if !validCalendarMailDeliveryID(deliveryID) || generation < 1 {
		return CalendarMailDeliveryClaimToken{}
	}
	return CalendarMailDeliveryClaimToken(uuid.NewSHA1(
		uuid.NameSpaceURL,
		[]byte("hub.calendar.delivery.claim.v1/"+deliveryID.UUID().String()+"/"+strconv.FormatInt(int64(generation), 10)),
	))
}

func calendarMailDeliveryFakeReclaim(claim CalendarMailDeliveryClaim, now time.Time) (CalendarMailDeliveryClaim, bool) {
	if !validCalendarMailDeliveryClaim(claim) || !validCalendarMailDeliveryLeaseUntil(now) ||
		claim.claimGeneration == CalendarMailDeliveryClaimGeneration(math.MaxInt64) {
		return CalendarMailDeliveryClaim{}, false
	}
	leaseUntil := now.Add(calendarMailDeliveryLease)
	if !validCalendarMailDeliveryLeaseUntil(leaseUntil) || !leaseUntil.After(now) {
		return CalendarMailDeliveryClaim{}, false
	}
	claim.claimGeneration++
	claim.claimToken = calendarMailDeliveryFakeClaimToken(claim.deliveryID, claim.claimGeneration)
	claim.leaseUntil = leaseUntil
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryClaim{}, false
	}
	return calendarMailDeliveryClaimCopy(claim), true
}

func calendarMailDeliveryFakeClaim(input CalendarMailDeliveryFakeClaim) (CalendarMailDeliveryClaim, error) {
	claimGeneration := input.ClaimGeneration
	if claimGeneration == 0 {
		claimGeneration = 1
	}
	claimToken := calendarMailDeliveryFakeClaimToken(input.DeliveryID, claimGeneration)
	claim := CalendarMailDeliveryClaim{
		deliveryID:      input.DeliveryID,
		inviteID:        input.InviteID,
		replyID:         input.ReplyID,
		kind:            input.Kind,
		method:          input.Method,
		state:           CalendarMailDeliveryStateSubmitting,
		claimToken:      claimToken,
		claimGeneration: claimGeneration,
		organizationID:  input.OrganizationID,
		leaseUntil:      input.LeaseUntil,
		eventID:         input.EventID,
		attendeeID:      input.AttendeeID,
		organizerID:     input.OrganizerID,
		accountID:       input.AccountID,
		senderActorID:   input.SenderActorID,
		fromIdentityID:  input.FromIdentityID,
		recipient:       input.Recipient,
		materialization: CalendarMailDeliveryMaterializationInput{deliveryID: input.DeliveryID, kind: input.Kind, method: input.Method, eventID: input.EventID, attendeeID: input.AttendeeID, organizerID: input.OrganizerID, eventSequence: input.EventSequence, attendeeVersion: input.AttendeeVersion, recurrenceID: input.RecurrenceID},
	}
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryClaim{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return calendarMailDeliveryClaimCopy(claim), nil
}
