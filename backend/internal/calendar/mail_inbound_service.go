package calendar

import (
	"context"
	"errors"
	"reflect"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/ical"
	"github.com/valerii/hub/backend/internal/mail"
)

// CalendarInboundMessageService receives one sanitized Calendar delivery and
// completes its durable receipt in the same transaction as any Calendar work.
type CalendarInboundMessageService struct {
	repository   CalendarInboundMessageRepository
	source       mail.CalendarPartSource
	eventService *EventService
}

var _ mail.CalendarInboundSink = (*CalendarInboundMessageService)(nil)

// NewCalendarInboundMessageService creates the receipt-first inbound Calendar
// orchestrator. The EventService repository is the only transaction owner.
func NewCalendarInboundMessageService(
	repository CalendarInboundMessageRepository,
	source mail.CalendarPartSource,
	eventService *EventService,
) (*CalendarInboundMessageService, error) {
	if nilCalendarInboundMessageRepository(repository) ||
		calendarInboundMessageServiceDependencyNil(source) ||
		eventService == nil ||
		calendarInboundMessageServiceDependencyNil(eventService.repository) {
		return nil, ErrEventInvalid
	}
	return &CalendarInboundMessageService{
		repository:   repository,
		source:       source,
		eventService: eventService,
	}, nil
}

// HandleCalendarInbound registers and claims exactly one durable receipt before
// processing a Calendar part. A durable no-op intentionally does no further I/O.
func (service *CalendarInboundMessageService) HandleCalendarInbound(
	ctx context.Context,
	delivery mail.CalendarInboundDelivery,
) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrEventInvalid
		}
	}()
	if ctx == nil || !service.ready() {
		return ErrEventInvalid
	}

	registration, err := NewCalendarInboundMessageRegistration(
		delivery.EnvelopeID(),
		delivery.OrganizationID(),
		delivery.MailboxID(),
		delivery.MessageID(),
		delivery.LocatorHash(),
	)
	if err != nil {
		return ErrEventInvalid
	}
	claim, err := service.repository.RegisterAndClaimCalendarInboundMessage(ctx, registration)
	if err != nil {
		return calendarInboundMessageServiceSafeError(err)
	}

	switch claim.Outcome() {
	case CalendarInboundMessageClaimOutcomeDurableNoop:
		return nil
	case CalendarInboundMessageClaimOutcomeClaimed:
		work, claimed := claim.Work()
		if !claimed {
			return ErrEventInvalid
		}
		return service.ProcessCalendarInboundMessage(ctx, work)
	default:
		return ErrEventInvalid
	}
}

// ProcessCalendarInboundMessage performs one sealed Mail read before opening
// one Event transaction whose final operation completes the exact receipt.
func (service *CalendarInboundMessageService) ProcessCalendarInboundMessage(
	ctx context.Context,
	work CalendarInboundMessageWork,
) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrEventInvalid
		}
	}()
	if ctx == nil || !service.ready() || !work.valid() {
		return ErrEventInvalid
	}

	prepared, err := service.prepareCalendarInboundMessage(ctx, work)
	if err != nil {
		return calendarInboundMessageServiceSafeError(err)
	}

	err = service.eventService.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		if transaction == nil {
			return ErrEventInvalid
		}
		scope, locked, lockErr := transaction.LockCalendarInboundMessageReceipt(ctx, work.Claim())
		if lockErr != nil {
			return lockErr
		}
		if !locked {
			return nil
		}

		completion, completionErr := service.completePreparedCalendarInboundMessage(ctx, transaction, scope, prepared)
		if completionErr != nil {
			return completionErr
		}
		if !completion.valid() {
			return ErrEventInvalid
		}
		return transaction.CompleteCalendarInboundMessageReceipt(ctx, scope, completion)
	})
	if err != nil {
		return calendarInboundMessageServiceSafeError(err)
	}
	return nil
}

type calendarInboundMessagePreparation struct {
	completion     CalendarInboundMessageCompletion
	failed         bool
	ownerID        uuid.UUID
	sender         string
	message        ITIPMessage
	authentication CalendarInboundAuthenticationClass
}

func (service *CalendarInboundMessageService) prepareCalendarInboundMessage(
	ctx context.Context,
	work CalendarInboundMessageWork,
) (calendarInboundMessagePreparation, error) {
	registration := work.Registration()
	request, err := mail.NewCalendarPartReadRequest(
		registration.OrganizationID(),
		registration.MailboxID(),
		registration.MailMessageID(),
		registration.LocatorHash(),
		mail.MaxCalendarPartBytes,
	)
	if err != nil {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorInternal),
		)
	}

	result, err := calendarInboundMessageReadPart(ctx, service.source, request)
	if err != nil {
		return calendarInboundMessageFailurePreparation(calendarInboundMessageSourceFailureCompletion(err))
	}
	return calendarInboundMessagePreparationFromSource(result)
}

func calendarInboundMessagePreparationFromSource(
	result mail.CalendarPartReadResult,
) (calendarInboundMessagePreparation, error) {
	authentication, admitted := calendarInboundMessageAuthentication(result.AuthenticationClass())
	if !admitted {
		if result.AuthenticationClass() == mail.MailAuthenticationClassUntrusted {
			return calendarInboundMessageFailurePreparation(
				NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorUntrustedSender),
			)
		}
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorInternal),
		)
	}

	method, supported := calendarInboundMessageITIPMethod(result.Method())
	if !supported {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorUnsupportedMethod),
		)
	}
	if result.OwnerID() == uuid.Nil {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorInternal),
		)
	}

	sender, err := ical.NormalizeITIPEmail(result.Sender())
	if err != nil || sender != result.Sender() {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorMalformedITIP),
		)
	}
	payload := result.ITIP()
	if len(payload) > mail.MaxCalendarPartBytes {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundDeadLetterCompletion(CalendarInboundMessageErrorResourceLimit),
		)
	}
	if len(payload) == 0 {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorInternal),
		)
	}

	message, err := ParseITIP(payload, DefaultITIPParseLimits())
	if err != nil {
		if errors.Is(err, ErrITIPLimitExceeded) {
			return calendarInboundMessageFailurePreparation(
				NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorResourceLimit),
			)
		}
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorMalformedITIP),
		)
	}
	if message.Method != method {
		return calendarInboundMessageFailurePreparation(
			NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorUnsupportedMethod),
		)
	}
	if method == ITIPMethodRequest || method == ITIPMethodCancel {
		organizer, organizerErr := ical.NormalizeITIPEmail(message.Organizer.Email)
		if organizerErr != nil || organizer != message.Organizer.Email {
			return calendarInboundMessageFailurePreparation(
				NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorMalformedITIP),
			)
		}
		if sender != organizer {
			return calendarInboundMessageFailurePreparation(
				NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorOrganizerMismatch),
			)
		}
	}

	return calendarInboundMessagePreparation{
		ownerID:        result.OwnerID(),
		sender:         sender,
		message:        message,
		authentication: authentication,
	}, nil
}

func (service *CalendarInboundMessageService) completePreparedCalendarInboundMessage(
	ctx context.Context,
	transaction *EventTransaction,
	scope CalendarInboundMessageScope,
	prepared calendarInboundMessagePreparation,
) (CalendarInboundMessageCompletion, error) {
	if prepared.failed {
		return cloneCalendarInboundMessageCompletion(prepared.completion), nil
	}

	switch prepared.message.Method {
	case ITIPMethodRequest, ITIPMethodCancel:
		request, err := NewCalendarInboundInviteRequest(prepared.message, prepared.authentication)
		if err != nil {
			return CalendarInboundMessageCompletion{}, ErrEventInvalid
		}
		result, err := service.eventService.ApplyInboundITIPInvite(ctx, transaction, scope, request)
		if err != nil {
			return CalendarInboundMessageCompletion{}, err
		}
		completion := result.Completion()
		if !completion.valid() {
			return CalendarInboundMessageCompletion{}, ErrEventInvalid
		}
		return completion, nil
	case ITIPMethodReply:
		result, err := service.eventService.applyInboundITIPReplyInTransaction(ctx, transaction, CalendarInboundReplyInput{
			OwnerID: prepared.ownerID,
			Sender:  prepared.sender,
			Message: prepared.message,
		})
		if err != nil {
			return CalendarInboundMessageCompletion{}, err
		}
		return result.calendarInboundMessageCompletion()
	default:
		return NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorUnsupportedMethod)
	}
}

func calendarInboundMessageFailurePreparation(
	completion CalendarInboundMessageCompletion,
	err error,
) (calendarInboundMessagePreparation, error) {
	if err != nil || !completion.valid() {
		return calendarInboundMessagePreparation{}, ErrEventInvalid
	}
	return calendarInboundMessagePreparation{
		completion: cloneCalendarInboundMessageCompletion(completion),
		failed:     true,
	}, nil
}

func calendarInboundMessageSourceFailureCompletion(err error) (CalendarInboundMessageCompletion, error) {
	kind, known := mail.CalendarPartSourceErrorKindOf(err)
	if !known {
		return NewCalendarInboundRetryCompletion(CalendarInboundMessageErrorInternal)
	}
	switch kind {
	case mail.CalendarPartSourceErrorUntrusted:
		return NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorUntrustedSender)
	case mail.CalendarPartSourceErrorMessageUnavailable, mail.CalendarPartSourceErrorDeleting:
		return NewCalendarInboundDeadLetterCompletion(CalendarInboundMessageErrorMessageUnavailable)
	case mail.CalendarPartSourceErrorPartUnavailable:
		return NewCalendarInboundDeadLetterCompletion(CalendarInboundMessageErrorPartUnavailable)
	case mail.CalendarPartSourceErrorPayloadTooLarge:
		return NewCalendarInboundDeadLetterCompletion(CalendarInboundMessageErrorResourceLimit)
	case mail.CalendarPartSourceErrorRepositoryUnavailable:
		return NewCalendarInboundRetryCompletion(CalendarInboundMessageErrorDatabaseUnavailable)
	case mail.CalendarPartSourceErrorEngineUnavailable:
		return NewCalendarInboundRetryCompletion(CalendarInboundMessageErrorEngineUnavailable)
	case mail.CalendarPartSourceErrorInvalid:
		return NewCalendarInboundQuarantinedCompletion(CalendarInboundMessageErrorInternal)
	default:
		return NewCalendarInboundRetryCompletion(CalendarInboundMessageErrorInternal)
	}
}

func calendarInboundMessageAuthentication(
	value mail.MailAuthenticationClass,
) (CalendarInboundAuthenticationClass, bool) {
	switch value {
	case mail.MailAuthenticationClassTrustedInternal:
		return CalendarInboundAuthenticationTrustedInternal, true
	case mail.MailAuthenticationClassDMARCAligned:
		return CalendarInboundAuthenticationDMARCAligned, true
	default:
		return "", false
	}
}

func calendarInboundMessageITIPMethod(value mail.MailCalendarMethod) (ITIPMethod, bool) {
	switch value {
	case mail.MailCalendarMethodRequest:
		return ITIPMethodRequest, true
	case mail.MailCalendarMethodCancel:
		return ITIPMethodCancel, true
	case mail.MailCalendarMethodReply:
		return ITIPMethodReply, true
	default:
		return "", false
	}
}

func calendarInboundMessageReadPart(
	ctx context.Context,
	source mail.CalendarPartSource,
	request mail.CalendarPartReadRequest,
) (result mail.CalendarPartReadResult, err error) {
	defer func() {
		if recover() != nil {
			result = mail.CalendarPartReadResult{}
			err = ErrEventInternal
		}
	}()
	return source.ReadCalendarPart(ctx, request)
}

func (service *CalendarInboundMessageService) ready() bool {
	return service != nil &&
		!nilCalendarInboundMessageRepository(service.repository) &&
		!calendarInboundMessageServiceDependencyNil(service.source) &&
		service.eventService != nil &&
		!calendarInboundMessageServiceDependencyNil(service.eventService.repository)
}

func calendarInboundMessageServiceDependencyNil(value any) bool {
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

func calendarInboundMessageServiceSafeError(error) error { return ErrEventInvalid }
