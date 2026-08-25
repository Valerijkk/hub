package mail

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	calendarMailDeliveryLease               = 60 * time.Second
	calendarMailDeliveryDefaultScanInterval = 30 * time.Second
	calendarMailDeliveryDefaultBatchSize    = 50
	calendarMailDeliveryMaximumBatchSize    = 50
	calendarMailDeliveryMaximumAttempts     = 8
	calendarMailDeliveryRetryMinimum        = 30 * time.Second
	calendarMailDeliveryRetryMaximum        = 5 * time.Minute
	calendarMailDeliveryMaximumRecurrenceID = 255
)

// CalendarMailDeliveryID is the opaque durable identity shared by invite and
// reply delivery rows.
type CalendarMailDeliveryID uuid.UUID

// NewCalendarMailDeliveryID validates one durable delivery identity.
func NewCalendarMailDeliveryID(value uuid.UUID) (CalendarMailDeliveryID, error) {
	if value == uuid.Nil {
		return CalendarMailDeliveryID{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return CalendarMailDeliveryID(value), nil
}

// UUID returns the opaque UUID value, or uuid.Nil for an invalid identity.
func (id CalendarMailDeliveryID) UUID() uuid.UUID {
	if !validCalendarMailDeliveryID(id) {
		return uuid.Nil
	}
	return uuid.UUID(id)
}

// CalendarInviteDeliveryID is the immutable identity of one authoritative
// invite transport intent.
type CalendarInviteDeliveryID uuid.UUID

// NewCalendarInviteDeliveryID validates one invite delivery identity.
func NewCalendarInviteDeliveryID(value uuid.UUID) (CalendarInviteDeliveryID, error) {
	if value == uuid.Nil {
		return CalendarInviteDeliveryID{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return CalendarInviteDeliveryID(value), nil
}

// UUID returns the opaque UUID value, or uuid.Nil for an invalid identity.
func (id CalendarInviteDeliveryID) UUID() uuid.UUID {
	if !validCalendarInviteDeliveryID(id) {
		return uuid.Nil
	}
	return uuid.UUID(id)
}

// CalendarReplyDeliveryID is the immutable identity of one authoritative RSVP
// transport intent.
type CalendarReplyDeliveryID uuid.UUID

// NewCalendarReplyDeliveryID validates one reply delivery identity.
func NewCalendarReplyDeliveryID(value uuid.UUID) (CalendarReplyDeliveryID, error) {
	if value == uuid.Nil {
		return CalendarReplyDeliveryID{}, newCalendarMailDeliveryRepositoryError(CalendarMailDeliveryRepositoryErrorInvalid)
	}
	return CalendarReplyDeliveryID(value), nil
}

// UUID returns the opaque UUID value, or uuid.Nil for an invalid identity.
func (id CalendarReplyDeliveryID) UUID() uuid.UUID {
	if !validCalendarReplyDeliveryID(id) {
		return uuid.Nil
	}
	return uuid.UUID(id)
}

// CalendarMailDeliveryKind identifies which immutable calendar intent owns a
// generic delivery row.
type CalendarMailDeliveryKind string

const (
	CalendarMailDeliveryKindInvite CalendarMailDeliveryKind = "invite"
	CalendarMailDeliveryKindReply  CalendarMailDeliveryKind = "reply"
)

// CalendarMailDeliveryMethod is the closed iTIP method family accepted by the
// delivery boundary.
type CalendarMailDeliveryMethod string

const (
	CalendarMailDeliveryMethodRequest CalendarMailDeliveryMethod = "REQUEST"
	CalendarMailDeliveryMethodCancel  CalendarMailDeliveryMethod = "CANCEL"
	CalendarMailDeliveryMethodReply   CalendarMailDeliveryMethod = "REPLY"
)

// CalendarMailDeliveryState is the durable delivery lifecycle. The repository
// owns every state transition and retry deadline.
type CalendarMailDeliveryState string

const (
	CalendarMailDeliveryStatePending         CalendarMailDeliveryState = "pending"
	CalendarMailDeliveryStateSubmitting      CalendarMailDeliveryState = "submitting"
	CalendarMailDeliveryStateUnknown         CalendarMailDeliveryState = "unknown"
	CalendarMailDeliveryStateSent            CalendarMailDeliveryState = "sent"
	CalendarMailDeliveryStateSuperseded      CalendarMailDeliveryState = "superseded"
	CalendarMailDeliveryStateFailedPermanent CalendarMailDeliveryState = "failed_permanent"
)

// CalendarMailDeliveryClaimToken is the repository-minted short lease token.
type CalendarMailDeliveryClaimToken uuid.UUID

// CalendarMailDeliveryClaimGeneration is the monotonic durable claim fence.
type CalendarMailDeliveryClaimGeneration int64

// CalendarMailDeliveryMaterializationInput contains only authoritative
// calendar identity and revision facts. It deliberately carries no body,
// address, account, identity, or draft data.
type CalendarMailDeliveryMaterializationInput struct {
	deliveryID       CalendarMailDeliveryID
	kind             CalendarMailDeliveryKind
	method           CalendarMailDeliveryMethod
	eventID          uuid.UUID
	attendeeID       uuid.UUID
	organizerID      uuid.UUID
	eventSequence    int
	attendeeVersion  int
	recurrenceID     string
}

// CalendarInviteDeliveryAdmission is the content-free Calendar request for
// durable invite intents. Mail authority is resolved only by the repository.
type CalendarInviteDeliveryAdmission struct {
	InviteID      CalendarInviteDeliveryID
	EventID       uuid.UUID
	AttendeeID    uuid.UUID
	OrganizerID   uuid.UUID
	Method        CalendarMailDeliveryMethod
	EventSequence int
	RecurrenceID  string
}

// CalendarInviteDeliveryAdmissionRequest bounds one atomic admission batch.
type CalendarInviteDeliveryAdmissionRequest struct {
	Invites []CalendarInviteDeliveryAdmission
}

// CalendarReplyDeliveryWakeRequest contains only a durable RSVP delivery ID.
// The repository resolves all Mail authority from the immutable intent.
type CalendarReplyDeliveryWakeRequest struct {
	ReplyID CalendarReplyDeliveryID
}

// CalendarMailDeliveryDueScanRequest bounds a repository-owned stable due
// ordering. The caller cannot choose the order or a retry timestamp.
type CalendarMailDeliveryDueScanRequest struct {
	Limit int
}

// CalendarMailDeliveryDueScanResult contains copied opaque IDs only.
type CalendarMailDeliveryDueScanResult struct {
	DeliveryIDs []CalendarMailDeliveryID
}

// CalendarMailDeliveryClaimOutcome is the closed result of a durable claim.
type CalendarMailDeliveryClaimOutcome string

const (
	CalendarMailDeliveryClaimOutcomeNoop        CalendarMailDeliveryClaimOutcome = "noop"
	CalendarMailDeliveryClaimOutcomeBusy        CalendarMailDeliveryClaimOutcome = "busy"
	CalendarMailDeliveryClaimOutcomeSuperseded  CalendarMailDeliveryClaimOutcome = "superseded"
	// Observed means a previously bound delivery was handled entirely by the
	// repository against its same MailSubmission. It grants no external work.
	CalendarMailDeliveryClaimOutcomeObserved    CalendarMailDeliveryClaimOutcome = "observed"
	CalendarMailDeliveryClaimOutcomeClaimed     CalendarMailDeliveryClaimOutcome = "claimed"
)

// CalendarMailDeliveryClaim is a sealed immutable repository capability for
// one unbound durable submitting row. The service accepts only its delivery
// ID; account, actor, identity, recipient, materialization input, lease token,
// and sequence facts are never caller inputs. A bound row is never exposed as
// a claim because its MailSubmission is observed only inside the repository.
type CalendarMailDeliveryClaim struct {
	deliveryID      CalendarMailDeliveryID
	inviteID        CalendarInviteDeliveryID
	replyID         CalendarReplyDeliveryID
	kind            CalendarMailDeliveryKind
	method          CalendarMailDeliveryMethod
	state           CalendarMailDeliveryState
	claimToken      CalendarMailDeliveryClaimToken
	claimGeneration CalendarMailDeliveryClaimGeneration
	organizationID  uuid.UUID
	leaseUntil      time.Time
	eventID         uuid.UUID
	attendeeID      uuid.UUID
	organizerID     uuid.UUID
	accountID       uuid.UUID
	senderActorID   uuid.UUID
	fromIdentityID  uuid.UUID
	recipient       EngineAddress
	materialization CalendarMailDeliveryMaterializationInput
}

// DeliveryID returns the durable identity held by a valid sealed claim.
func (claim CalendarMailDeliveryClaim) DeliveryID() CalendarMailDeliveryID {
	if !validCalendarMailDeliveryClaim(claim) {
		return CalendarMailDeliveryID{}
	}
	return claim.deliveryID
}

// SenderActorID returns the distinct internal Mail owner authorized to create
// the hidden draft and submission. It is not the organizer identity.
func (claim CalendarMailDeliveryClaim) SenderActorID() uuid.UUID {
	if !validCalendarMailDeliveryClaim(claim) {
		return uuid.Nil
	}
	return claim.senderActorID
}

// CalendarMailDeliveryClaimResult carries exactly one closed claim outcome.
type CalendarMailDeliveryClaimResult struct {
	outcome CalendarMailDeliveryClaimOutcome
	claim   CalendarMailDeliveryClaim
}

// Outcome returns a closed durable result, or the zero value for malformed
// repository output.
func (result CalendarMailDeliveryClaimResult) Outcome() CalendarMailDeliveryClaimOutcome {
	if !validCalendarMailDeliveryClaimResult(result) {
		return ""
	}
	return result.outcome
}

// Claim returns a defensive sealed capability only for a valid claimed result.
func (result CalendarMailDeliveryClaimResult) Claim() (CalendarMailDeliveryClaim, bool) {
	if !validCalendarMailDeliveryClaimResult(result) || result.outcome != CalendarMailDeliveryClaimOutcomeClaimed {
		return CalendarMailDeliveryClaim{}, false
	}
	return calendarMailDeliveryClaimCopy(result.claim), true
}

// CalendarMailDeliverySubmissionBindRequest links one known durable Mail
// submission to its exact sealed claim. The repository treats the same UUID as
// idempotent and a different UUID as a conflict.
type CalendarMailDeliverySubmissionBindRequest struct {
	Claim        CalendarMailDeliveryClaim
	SubmissionID uuid.UUID
}

// CalendarMailDeliverySubmissionBindOutcome is the closed result of a bind.
type CalendarMailDeliverySubmissionBindOutcome string

const (
	CalendarMailDeliverySubmissionBindOutcomeBound    CalendarMailDeliverySubmissionBindOutcome = "bound"
	CalendarMailDeliverySubmissionBindOutcomeReplay   CalendarMailDeliverySubmissionBindOutcome = "replay"
	CalendarMailDeliverySubmissionBindOutcomeConflict CalendarMailDeliverySubmissionBindOutcome = "conflict"
)

// CalendarMailDeliverySubmissionBindResult contains no submission detail
// beyond the durable identifier supplied by the service.
type CalendarMailDeliverySubmissionBindResult struct {
	outcome CalendarMailDeliverySubmissionBindOutcome
}

// Outcome returns the closed bind outcome, or the zero value for malformed
// repository output.
func (result CalendarMailDeliverySubmissionBindResult) Outcome() CalendarMailDeliverySubmissionBindOutcome {
	if !validCalendarMailDeliverySubmissionBindResult(result) {
		return ""
	}
	return result.outcome
}

// CalendarMailDeliveryResolveCode is the content-free, pre-bind durable
// conclusion of one unbound claim. Retry timing and terminal state remain
// repository-owned. Submission outcomes are intentionally absent: once a
// submission has an ID, binding is the durable boundary and the repository
// owns all later observation of that same submission.
type CalendarMailDeliveryResolveCode string

const (
	CalendarMailDeliveryResolveCodeMailboxUnavailable CalendarMailDeliveryResolveCode = "mailbox_unavailable"
	CalendarMailDeliveryResolveCodeRejected           CalendarMailDeliveryResolveCode = "rejected"
	CalendarMailDeliveryResolveCodePolicyDenied       CalendarMailDeliveryResolveCode = "policy_denied"
	CalendarMailDeliveryResolveCodeInvalid            CalendarMailDeliveryResolveCode = "invalid"
	CalendarMailDeliveryResolveCodeInternal           CalendarMailDeliveryResolveCode = "internal"
)

// CalendarMailDeliveryResolveRequest releases or resolves exactly one valid
// unbound sealed claim with a safe pre-bind code and never carries a raw
// failure.
type CalendarMailDeliveryResolveRequest struct {
	Claim CalendarMailDeliveryClaim
	Code  CalendarMailDeliveryResolveCode
}

// CalendarMailDeliveryRepository owns durable invite admission, reply wake,
// stable due selection, claim fencing, submission binding, and resolution. It
// performs no Mail service or external I/O. ClaimCalendarMailDelivery emits a
// claimed capability only for an unbound submitting row. For a due bound
// unknown row it performs DB-local observation, rescheduling, or
// terminalization against that exact MailSubmission and returns Noop or
// Observed; it never emits an external claim for a bound row.
type CalendarMailDeliveryRepository interface {
	AdmitCalendarInviteDeliveries(context.Context, CalendarInviteDeliveryAdmissionRequest) error
	WakeCalendarReplyDelivery(context.Context, CalendarReplyDeliveryWakeRequest) error
	ScanDueCalendarMailDeliveries(context.Context, CalendarMailDeliveryDueScanRequest) (CalendarMailDeliveryDueScanResult, error)
	ClaimCalendarMailDelivery(context.Context, CalendarMailDeliveryID) (CalendarMailDeliveryClaimResult, error)
	BindCalendarMailDeliverySubmission(context.Context, CalendarMailDeliverySubmissionBindRequest) (CalendarMailDeliverySubmissionBindResult, error)
	ResolveCalendarMailDelivery(context.Context, CalendarMailDeliveryResolveRequest) error
}

// CalendarMailDeliveryRepositoryErrorKind is the closed repository failure
// family for the Calendar-to-Mail durable boundary.
type CalendarMailDeliveryRepositoryErrorKind string

const (
	CalendarMailDeliveryRepositoryErrorInvalid     CalendarMailDeliveryRepositoryErrorKind = "invalid"
	CalendarMailDeliveryRepositoryErrorUnavailable CalendarMailDeliveryRepositoryErrorKind = "unavailable"
	CalendarMailDeliveryRepositoryErrorConflict    CalendarMailDeliveryRepositoryErrorKind = "conflict"
)

const calendarMailDeliveryRepositoryErrorText = "calendar mail delivery repository error"

// CalendarMailDeliveryRepositoryError contains only a closed safe error kind.
type CalendarMailDeliveryRepositoryError struct {
	kind CalendarMailDeliveryRepositoryErrorKind
}

func (err *CalendarMailDeliveryRepositoryError) Error() string {
	return calendarMailDeliveryRepositoryErrorText
}

// Kind returns a validated repository error kind.
func (err *CalendarMailDeliveryRepositoryError) Kind() CalendarMailDeliveryRepositoryErrorKind {
	if err == nil || !validCalendarMailDeliveryRepositoryErrorKind(err.kind) {
		return CalendarMailDeliveryRepositoryErrorInvalid
	}
	return err.kind
}

func newCalendarMailDeliveryRepositoryError(kind CalendarMailDeliveryRepositoryErrorKind) *CalendarMailDeliveryRepositoryError {
	if !validCalendarMailDeliveryRepositoryErrorKind(kind) {
		kind = CalendarMailDeliveryRepositoryErrorInvalid
	}
	return &CalendarMailDeliveryRepositoryError{kind: kind}
}

func validCalendarMailDeliveryID(value CalendarMailDeliveryID) bool {
	return uuid.UUID(value) != uuid.Nil
}

func validCalendarInviteDeliveryID(value CalendarInviteDeliveryID) bool {
	return uuid.UUID(value) != uuid.Nil
}

func validCalendarReplyDeliveryID(value CalendarReplyDeliveryID) bool {
	return uuid.UUID(value) != uuid.Nil
}

func validCalendarMailDeliveryKind(value CalendarMailDeliveryKind) bool {
	return value == CalendarMailDeliveryKindInvite || value == CalendarMailDeliveryKindReply
}

func validCalendarMailDeliveryMethod(value CalendarMailDeliveryMethod) bool {
	switch value {
	case CalendarMailDeliveryMethodRequest, CalendarMailDeliveryMethodCancel, CalendarMailDeliveryMethodReply:
		return true
	default:
		return false
	}
}

func validCalendarMailDeliveryState(value CalendarMailDeliveryState) bool {
	switch value {
	case CalendarMailDeliveryStatePending,
		CalendarMailDeliveryStateSubmitting,
		CalendarMailDeliveryStateUnknown,
		CalendarMailDeliveryStateSent,
		CalendarMailDeliveryStateSuperseded,
		CalendarMailDeliveryStateFailedPermanent:
		return true
	default:
		return false
	}
}

func validCalendarMailDeliveryLeaseUntil(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}

func validCalendarMailDeliveryPreBindResolveCode(value CalendarMailDeliveryResolveCode) bool {
	switch value {
	case CalendarMailDeliveryResolveCodeMailboxUnavailable,
		CalendarMailDeliveryResolveCodeRejected,
		CalendarMailDeliveryResolveCodePolicyDenied,
		CalendarMailDeliveryResolveCodeInvalid,
		CalendarMailDeliveryResolveCodeInternal:
		return true
	default:
		return false
	}
}

func validCalendarMailDeliveryRepositoryErrorKind(value CalendarMailDeliveryRepositoryErrorKind) bool {
	switch value {
	case CalendarMailDeliveryRepositoryErrorInvalid,
		CalendarMailDeliveryRepositoryErrorUnavailable,
		CalendarMailDeliveryRepositoryErrorConflict:
		return true
	default:
		return false
	}
}

func validCalendarInviteDeliveryAdmissionRequest(request CalendarInviteDeliveryAdmissionRequest) bool {
	if request.Invites == nil || len(request.Invites) == 0 || len(request.Invites) > calendarMailDeliveryMaximumBatchSize {
		return false
	}
	seen := make(map[CalendarInviteDeliveryID]struct{}, len(request.Invites))
	for _, invite := range request.Invites {
		if !validCalendarInviteDeliveryAdmission(invite) {
			return false
		}
		if _, duplicate := seen[invite.InviteID]; duplicate {
			return false
		}
		seen[invite.InviteID] = struct{}{}
	}
	return true
}

func validCalendarInviteDeliveryAdmission(value CalendarInviteDeliveryAdmission) bool {
	return validCalendarInviteDeliveryID(value.InviteID) && value.EventID != uuid.Nil && value.AttendeeID != uuid.Nil && value.OrganizerID != uuid.Nil &&
		(value.Method == CalendarMailDeliveryMethodRequest || value.Method == CalendarMailDeliveryMethodCancel) && value.EventSequence >= 1 && validCalendarMailDeliveryRecurrenceID(value.RecurrenceID)
}

func validCalendarReplyDeliveryWakeRequest(request CalendarReplyDeliveryWakeRequest) bool {
	return validCalendarReplyDeliveryID(request.ReplyID)
}

func validCalendarMailDeliveryDueScanRequest(request CalendarMailDeliveryDueScanRequest) bool {
	return request.Limit >= 1 && request.Limit <= calendarMailDeliveryMaximumBatchSize
}

func validCalendarMailDeliveryDueScanResult(request CalendarMailDeliveryDueScanRequest, result CalendarMailDeliveryDueScanResult) bool {
	if !validCalendarMailDeliveryDueScanRequest(request) || result.DeliveryIDs == nil || len(result.DeliveryIDs) > request.Limit {
		return false
	}
	seen := make(map[CalendarMailDeliveryID]struct{}, len(result.DeliveryIDs))
	for _, deliveryID := range result.DeliveryIDs {
		if !validCalendarMailDeliveryID(deliveryID) {
			return false
		}
		if _, duplicate := seen[deliveryID]; duplicate {
			return false
		}
		seen[deliveryID] = struct{}{}
	}
	return true
}

func calendarMailDeliveryDueScanResultCopy(result CalendarMailDeliveryDueScanResult) CalendarMailDeliveryDueScanResult {
	result.DeliveryIDs = append([]CalendarMailDeliveryID(nil), result.DeliveryIDs...)
	return result
}

func validCalendarMailDeliveryClaimResult(result CalendarMailDeliveryClaimResult) bool {
	switch result.outcome {
	case CalendarMailDeliveryClaimOutcomeNoop,
		CalendarMailDeliveryClaimOutcomeBusy,
		CalendarMailDeliveryClaimOutcomeSuperseded,
		CalendarMailDeliveryClaimOutcomeObserved:
		return result.claim == (CalendarMailDeliveryClaim{})
	case CalendarMailDeliveryClaimOutcomeClaimed:
		return validCalendarMailDeliveryClaim(result.claim)
	default:
		return false
	}
}

func validCalendarMailDeliveryClaim(claim CalendarMailDeliveryClaim) bool {
	if !validCalendarMailDeliveryID(claim.deliveryID) || !validCalendarMailDeliveryKind(claim.kind) || !validCalendarMailDeliveryMethod(claim.method) ||
		claim.state != CalendarMailDeliveryStateSubmitting || uuid.UUID(claim.claimToken) == uuid.Nil || claim.claimGeneration < 1 ||
		claim.organizationID == uuid.Nil || !validCalendarMailDeliveryLeaseUntil(claim.leaseUntil) ||
		claim.eventID == uuid.Nil || claim.attendeeID == uuid.Nil || claim.organizerID == uuid.Nil || claim.accountID == uuid.Nil ||
		claim.senderActorID == uuid.Nil || claim.fromIdentityID == uuid.Nil || !validCalendarMailDeliveryMaterializationInput(claim.materialization) {
		return false
	}
	if claim.materialization.deliveryID != claim.deliveryID || claim.materialization.kind != claim.kind || claim.materialization.method != claim.method ||
		claim.materialization.eventID != claim.eventID || claim.materialization.attendeeID != claim.attendeeID || claim.materialization.organizerID != claim.organizerID {
		return false
	}
	recipient, ok := calendarMailDeliveryRecipientCopy(claim.recipient)
	if !ok || recipient != claim.recipient {
		return false
	}
	switch claim.kind {
	case CalendarMailDeliveryKindInvite:
		if !validCalendarInviteDeliveryID(claim.inviteID) || validCalendarReplyDeliveryID(claim.replyID) || claim.method == CalendarMailDeliveryMethodReply || claim.materialization.attendeeVersion != 0 {
			return false
		}
	case CalendarMailDeliveryKindReply:
		if validCalendarInviteDeliveryID(claim.inviteID) || !validCalendarReplyDeliveryID(claim.replyID) || claim.method != CalendarMailDeliveryMethodReply || claim.materialization.attendeeVersion < 1 {
			return false
		}
	default:
		return false
	}
	return true
}

func validCalendarMailDeliveryMaterializationInput(input CalendarMailDeliveryMaterializationInput) bool {
	if !validCalendarMailDeliveryID(input.deliveryID) || !validCalendarMailDeliveryKind(input.kind) || !validCalendarMailDeliveryMethod(input.method) ||
		input.eventID == uuid.Nil || input.attendeeID == uuid.Nil || input.organizerID == uuid.Nil || !validCalendarMailDeliveryRecurrenceID(input.recurrenceID) {
		return false
	}
	if input.kind == CalendarMailDeliveryKindInvite {
		return (input.method == CalendarMailDeliveryMethodRequest || input.method == CalendarMailDeliveryMethodCancel) && input.eventSequence >= 1 && input.attendeeVersion == 0
	}
	return input.kind == CalendarMailDeliveryKindReply && input.method == CalendarMailDeliveryMethodReply && input.eventSequence >= 0 && input.attendeeVersion >= 1
}

func validCalendarMailDeliveryRecurrenceID(value string) bool {
	if len(value) > calendarMailDeliveryMaximumRecurrenceID || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if character < 0x20 || (character >= 0x7f && character <= 0x9f) {
			return false
		}
	}
	return strings.TrimSpace(value) == value
}

func calendarMailDeliveryRecipientCopy(value EngineAddress) (EngineAddress, bool) {
	normalized, err := NormalizeAddress(value)
	if err != nil || normalized.Email == "" {
		return EngineAddress{}, false
	}
	return normalized, true
}

func calendarMailDeliveryClaimCopy(claim CalendarMailDeliveryClaim) CalendarMailDeliveryClaim {
	return claim
}

func validCalendarMailDeliverySubmissionBindRequest(request CalendarMailDeliverySubmissionBindRequest) bool {
	return validCalendarMailDeliveryClaim(request.Claim) && request.SubmissionID != uuid.Nil
}

func validCalendarMailDeliverySubmissionBindResult(result CalendarMailDeliverySubmissionBindResult) bool {
	switch result.outcome {
	case CalendarMailDeliverySubmissionBindOutcomeBound,
		CalendarMailDeliverySubmissionBindOutcomeReplay,
		CalendarMailDeliverySubmissionBindOutcomeConflict:
		return true
	default:
		return false
	}
}

func validCalendarMailDeliveryResolveRequest(request CalendarMailDeliveryResolveRequest) bool {
	return validCalendarMailDeliveryClaim(request.Claim) && validCalendarMailDeliveryPreBindResolveCode(request.Code)
}

func calendarMailDeliveryRepositoryErrorKindOf(err error) CalendarMailDeliveryRepositoryErrorKind {
	var repositoryError *CalendarMailDeliveryRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil {
		return repositoryError.Kind()
	}
	return CalendarMailDeliveryRepositoryErrorUnavailable
}
