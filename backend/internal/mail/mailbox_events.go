package mail

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const (
	maxMailEventEngineMessageIDBytes = 512
	maxMailEventRecipientItems       = 200
	maxMailEventFolderItems          = 100
	maxMailEventChangedFieldItems    = 30
	maxMailEventFlagItems            = 30
	maxMailEventLabelItems           = 100
)

// MailEventErrorKind classifies a content-free Mail event mapping failure.
type MailEventErrorKind string

const (
	// MailEventErrorInvalid denotes invalid or incomplete Mail event input.
	MailEventErrorInvalid MailEventErrorKind = "invalid"
	// MailEventErrorTooLarge denotes a Mail event input exceeding a fixed bound.
	MailEventErrorTooLarge MailEventErrorKind = "too_large"
	// MailEventErrorDuplicate denotes a duplicate address or identifier in a Mail event input.
	MailEventErrorDuplicate MailEventErrorKind = "duplicate"
	// MailEventErrorAccountMismatch denotes an identifier outside the Mailbox account scope.
	MailEventErrorAccountMismatch MailEventErrorKind = "account_mismatch"
)

var (
	// ErrMailEventInvalid is the stable sentinel for invalid Mail event input.
	ErrMailEventInvalid = errors.New("mail event invalid")
	// ErrMailEventTooLarge is the stable sentinel for a Mail event input exceeding bounds.
	ErrMailEventTooLarge = errors.New("mail event exceeds bounds")
	// ErrMailEventDuplicate is the stable sentinel for duplicate Mail event input.
	ErrMailEventDuplicate = errors.New("mail event duplicate")
	// ErrMailEventAccountMismatch is the stable sentinel for a cross-account Mail event identifier.
	ErrMailEventAccountMismatch = errors.New("mail event account mismatch")
)

// MailEventPolicyError provides a stable, content-free Mail event mapping error.
type MailEventPolicyError struct {
	Kind MailEventErrorKind
}

// Error returns a stable content-free error string.
func (errorValue *MailEventPolicyError) Error() string {
	if errorValue == nil {
		return "mail event policy error"
	}

	switch errorValue.Kind {
	case MailEventErrorInvalid:
		return ErrMailEventInvalid.Error()
	case MailEventErrorTooLarge:
		return ErrMailEventTooLarge.Error()
	case MailEventErrorDuplicate:
		return ErrMailEventDuplicate.Error()
	case MailEventErrorAccountMismatch:
		return ErrMailEventAccountMismatch.Error()
	default:
		return "mail event policy error"
	}
}

// Unwrap returns the stable sentinel for a known Mail event mapping error kind.
func (errorValue *MailEventPolicyError) Unwrap() error {
	if errorValue == nil {
		return nil
	}

	switch errorValue.Kind {
	case MailEventErrorInvalid:
		return ErrMailEventInvalid
	case MailEventErrorTooLarge:
		return ErrMailEventTooLarge
	case MailEventErrorDuplicate:
		return ErrMailEventDuplicate
	case MailEventErrorAccountMismatch:
		return ErrMailEventAccountMismatch
	default:
		return nil
	}
}

// MailboxEventID binds one Hub resource identifier to the owning Mailbox account.
type MailboxEventID struct {
	AccountID uuid.UUID
	ID        uuid.UUID
}

// MailboxEventOpaqueID binds one bounded opaque engine identifier to the owning Mailbox account.
type MailboxEventOpaqueID struct {
	AccountID uuid.UUID
	ID        string
}

// MailEventAddress is a normalized owner-safe Mail address projection.
type MailEventAddress struct {
	Email string
	Name  *string
}

// MailCalendarMethod is the nullable bounded iTIP method projection for a received message.
type MailCalendarMethod string

const (
	// MailCalendarMethodRequest represents an iTIP request method.
	MailCalendarMethodRequest MailCalendarMethod = "request"
	// MailCalendarMethodCancel represents an iTIP cancel method.
	MailCalendarMethodCancel MailCalendarMethod = "cancel"
	// MailCalendarMethodReply represents an iTIP reply method.
	MailCalendarMethodReply MailCalendarMethod = "reply"
)

// MailDeliveryClass is the bounded delivery classification for a confirmed Mail submission.
type MailDeliveryClass string

const (
	// MailDeliveryClassInternal represents delivery within a Hub organization.
	MailDeliveryClassInternal MailDeliveryClass = "internal"
	// MailDeliveryClassInstanceDomain represents delivery to the Hub instance domain.
	MailDeliveryClassInstanceDomain MailDeliveryClass = "instance_domain"
	// MailDeliveryClassExternal represents delivery outside the Hub instance domain.
	MailDeliveryClassExternal MailDeliveryClass = "external"
)

// MailChangedField is one allowed public Mail message projection change.
type MailChangedField string

const (
	// MailChangedFieldFolderIDs denotes a folder membership change.
	MailChangedFieldFolderIDs MailChangedField = "folder_ids"
	// MailChangedFieldSeen denotes a seen-state change.
	MailChangedFieldSeen MailChangedField = "seen"
	// MailChangedFieldAnswered denotes an answered-state change.
	MailChangedFieldAnswered MailChangedField = "answered"
	// MailChangedFieldFlagged denotes a flagged-state change.
	MailChangedFieldFlagged MailChangedField = "flagged"
	// MailChangedFieldJunk denotes a junk-state change.
	MailChangedFieldJunk MailChangedField = "junk"
	// MailChangedFieldLabels denotes a Hub label membership change.
	MailChangedFieldLabels MailChangedField = "labels"
	// MailChangedFieldDeleted denotes a deleted message with no live engine state.
	MailChangedFieldDeleted MailChangedField = "deleted"
)

// MailMessageFlag is one allowed current Mail message flag projection.
type MailMessageFlag string

const (
	// MailMessageFlagSeen denotes a seen message flag.
	MailMessageFlagSeen MailMessageFlag = "seen"
	// MailMessageFlagAnswered denotes an answered message flag.
	MailMessageFlagAnswered MailMessageFlag = "answered"
	// MailMessageFlagFlagged denotes a flagged message flag.
	MailMessageFlagFlagged MailMessageFlag = "flagged"
	// MailMessageFlagJunk denotes a junk message flag.
	MailMessageFlagJunk MailMessageFlag = "junk"
	// MailMessageFlagDraft denotes a draft message flag.
	MailMessageFlagDraft MailMessageFlag = "draft"
)

// MailReceivedEventPayload is the bounded source projection for mail.received.v1.
type MailReceivedEventPayload struct {
	Mailbox        MailboxEventID
	Message        MailboxEventID
	EngineMessage  MailboxEventOpaqueID
	Thread         MailboxEventID
	From           MailEventAddress
	To             []MailEventAddress
	ReceivedAt     time.Time
	FolderIDs      []MailboxEventID
	HasAttachments bool
	CalendarMethod *MailCalendarMethod
	AuthenticationClass      MailAuthenticationClass
	CalendarPartLocatorHash  []byte
}

// MailSentEventPayload is the bounded source projection for mail.sent.v1.
type MailSentEventPayload struct {
	Mailbox       MailboxEventID
	Submission    MailboxEventID
	Message       MailboxEventID
	Thread        MailboxEventID
	From          MailEventAddress
	To            []MailEventAddress
	SentAt        time.Time
	DeliveryClass MailDeliveryClass
}

// MailMessageChangedEventPayload is the bounded source projection for mail.message.changed.v1.
type MailMessageChangedEventPayload struct {
	Mailbox       MailboxEventID
	Message       MailboxEventID
	ChangedFields []MailChangedField
	FolderIDs     []MailboxEventID
	Flags         []MailMessageFlag
	LabelIDs      []MailboxEventID
}

type mailEventAddressJSON struct {
	Email string  `json:"email"`
	Name  *string `json:"name,omitempty"`
}

type mailReceivedEventJSON struct {
	MailboxID      uuid.UUID               `json:"mailbox_id"`
	MessageID      uuid.UUID               `json:"message_id"`
	EngineMessageID string                  `json:"engine_message_id"`
	ThreadID       uuid.UUID               `json:"thread_id"`
	From           mailEventAddressJSON   `json:"from"`
	To             []mailEventAddressJSON `json:"to"`
	ReceivedAt     time.Time               `json:"received_at"`
	FolderIDs      []uuid.UUID             `json:"folder_ids"`
	HasAttachments bool                    `json:"has_attachments"`
	CalendarMethod *MailCalendarMethod     `json:"calendar_method"`
	AuthenticationClass      MailAuthenticationClass `json:"authentication_class"`
	CalendarPartLocatorHash  *string                 `json:"calendar_part_locator_sha256"`
}

type mailSentEventJSON struct {
	MailboxID     uuid.UUID               `json:"mailbox_id"`
	SubmissionID  uuid.UUID               `json:"submission_id"`
	MessageID     uuid.UUID               `json:"message_id"`
	ThreadID      uuid.UUID               `json:"thread_id"`
	From          mailEventAddressJSON   `json:"from"`
	To            []mailEventAddressJSON `json:"to"`
	SentAt        time.Time               `json:"sent_at"`
	DeliveryClass MailDeliveryClass       `json:"delivery_class"`
}

type mailMessageChangedEventJSON struct {
	MailboxID     uuid.UUID         `json:"mailbox_id"`
	MessageID     uuid.UUID         `json:"message_id"`
	ChangedFields []MailChangedField `json:"changed_fields"`
	FolderIDs     []uuid.UUID       `json:"folder_ids"`
	Flags         []MailMessageFlag `json:"flags"`
	LabelIDs      []uuid.UUID       `json:"label_ids"`
}

// NewMailReceivedEnvelopeV1 maps one validated Mail receipt projection into mail.received.v1.
func NewMailReceivedEnvelopeV1(
	eventID uuid.UUID,
	occurredAt time.Time,
	organizationID uuid.UUID,
	actorID *uuid.UUID,
	correlationID string,
	payload MailReceivedEventPayload,
) (events.EnvelopeV1, error) {
	accountID, err := validateMailboxEventID(payload.Mailbox)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	if err := validateMailboxEventIDForAccount(accountID, payload.Message); err != nil {
		return events.EnvelopeV1{}, err
	}
	if err := validateMailboxEventIDForAccount(accountID, payload.Thread); err != nil {
		return events.EnvelopeV1{}, err
	}
	if err := validateMailboxEventOpaqueIDForAccount(accountID, payload.EngineMessage); err != nil {
		return events.EnvelopeV1{}, err
	}
	from, err := validateMailEventAddress(payload.From)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	to, err := validateMailEventAddresses(payload.To, true)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	folderIDs, err := validateMailboxEventIDsForAccount(accountID, payload.FolderIDs, 1, maxMailEventFolderItems)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	if !validMailEventTime(payload.ReceivedAt) || !validMailCalendarMethod(payload.CalendarMethod) || !validMailAuthenticationClass(payload.AuthenticationClass) || (payload.CalendarMethod == nil) != (payload.CalendarPartLocatorHash == nil) || (payload.CalendarPartLocatorHash != nil && len(payload.CalendarPartLocatorHash) != sha256.Size) {
		return events.EnvelopeV1{}, mailEventPolicyError(MailEventErrorInvalid)
	}
	calendarPartLocatorHash := mailCalendarPartLocatorHash(payload.CalendarPartLocatorHash)

	return events.NewEnvelopeV1("mail.received.v1", eventID, occurredAt, organizationID, actorID, correlationID, mailReceivedEventJSON{
		MailboxID:       payload.Mailbox.ID,
		MessageID:       payload.Message.ID,
		EngineMessageID: payload.EngineMessage.ID,
		ThreadID:        payload.Thread.ID,
		From:            from,
		To:              to,
		ReceivedAt:      payload.ReceivedAt,
		FolderIDs:       folderIDs,
		HasAttachments:  payload.HasAttachments,
		CalendarMethod:  copyMailCalendarMethod(payload.CalendarMethod),
		AuthenticationClass:      payload.AuthenticationClass,
		CalendarPartLocatorHash:  calendarPartLocatorHash,
	})
}

// NewMailSentEnvelopeV1 maps one validated confirmed submission into mail.sent.v1.
func NewMailSentEnvelopeV1(
	eventID uuid.UUID,
	occurredAt time.Time,
	organizationID uuid.UUID,
	actorID *uuid.UUID,
	correlationID string,
	payload MailSentEventPayload,
) (events.EnvelopeV1, error) {
	accountID, err := validateMailboxEventID(payload.Mailbox)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	for _, value := range []MailboxEventID{payload.Submission, payload.Message, payload.Thread} {
		if err := validateMailboxEventIDForAccount(accountID, value); err != nil {
			return events.EnvelopeV1{}, err
		}
	}
	from, err := validateMailEventAddress(payload.From)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	to, err := validateMailEventAddresses(payload.To, true)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	if !validMailEventTime(payload.SentAt) || !validMailDeliveryClass(payload.DeliveryClass) {
		return events.EnvelopeV1{}, mailEventPolicyError(MailEventErrorInvalid)
	}

	return events.NewEnvelopeV1("mail.sent.v1", eventID, occurredAt, organizationID, actorID, correlationID, mailSentEventJSON{
		MailboxID:     payload.Mailbox.ID,
		SubmissionID:  payload.Submission.ID,
		MessageID:     payload.Message.ID,
		ThreadID:      payload.Thread.ID,
		From:          from,
		To:            to,
		SentAt:        payload.SentAt,
		DeliveryClass: payload.DeliveryClass,
	})
}

// NewMailMessageChangedEnvelopeV1 maps one validated projection delta into mail.message.changed.v1.
func NewMailMessageChangedEnvelopeV1(
	eventID uuid.UUID,
	occurredAt time.Time,
	organizationID uuid.UUID,
	actorID *uuid.UUID,
	correlationID string,
	payload MailMessageChangedEventPayload,
) (events.EnvelopeV1, error) {
	accountID, err := validateMailboxEventID(payload.Mailbox)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	if err := validateMailboxEventIDForAccount(accountID, payload.Message); err != nil {
		return events.EnvelopeV1{}, err
	}
	if err := validateMailChangedFields(payload.ChangedFields); err != nil {
		return events.EnvelopeV1{}, err
	}
	folderIDs, err := validateMailboxEventIDsForAccount(accountID, payload.FolderIDs, 0, maxMailEventFolderItems)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	flags, err := validateMailMessageFlags(payload.Flags)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	labelIDs, err := validateMailboxEventIDsForAccount(accountID, payload.LabelIDs, 0, maxMailEventLabelItems)
	if err != nil {
		return events.EnvelopeV1{}, err
	}
	if hasMailChangedField(payload.ChangedFields, MailChangedFieldDeleted) && (len(payload.ChangedFields) != 1 || len(folderIDs) != 0 || len(flags) != 0) {
		return events.EnvelopeV1{}, mailEventPolicyError(MailEventErrorInvalid)
	}

	return events.NewEnvelopeV1("mail.message.changed.v1", eventID, occurredAt, organizationID, actorID, correlationID, mailMessageChangedEventJSON{
		MailboxID:     payload.Mailbox.ID,
		MessageID:     payload.Message.ID,
		ChangedFields: append([]MailChangedField(nil), payload.ChangedFields...),
		FolderIDs:     folderIDs,
		Flags:         flags,
		LabelIDs:      labelIDs,
	})
}

func validateMailboxEventID(value MailboxEventID) (uuid.UUID, error) {
	if value.AccountID == uuid.Nil || value.ID == uuid.Nil {
		return uuid.Nil, mailEventPolicyError(MailEventErrorInvalid)
	}
	return value.AccountID, nil
}

func validateMailboxEventIDForAccount(accountID uuid.UUID, value MailboxEventID) error {
	if value.AccountID == uuid.Nil || value.ID == uuid.Nil {
		return mailEventPolicyError(MailEventErrorInvalid)
	}
	if value.AccountID != accountID {
		return mailEventPolicyError(MailEventErrorAccountMismatch)
	}
	return nil
}

func validateMailboxEventOpaqueIDForAccount(accountID uuid.UUID, value MailboxEventOpaqueID) error {
	if value.AccountID == uuid.Nil || !utf8.ValidString(value.ID) || value.ID == "" {
		return mailEventPolicyError(MailEventErrorInvalid)
	}
	if len(value.ID) > maxMailEventEngineMessageIDBytes {
		return mailEventPolicyError(MailEventErrorTooLarge)
	}
	if value.AccountID != accountID {
		return mailEventPolicyError(MailEventErrorAccountMismatch)
	}
	return nil
}

func validateMailboxEventIDsForAccount(accountID uuid.UUID, values []MailboxEventID, minimum, maximum int) ([]uuid.UUID, error) {
	if len(values) < minimum {
		return nil, mailEventPolicyError(MailEventErrorInvalid)
	}
	if len(values) > maximum {
		return nil, mailEventPolicyError(MailEventErrorTooLarge)
	}

	ids := make([]uuid.UUID, len(values))
	seen := make(map[uuid.UUID]struct{}, len(values))
	for index, value := range values {
		if err := validateMailboxEventIDForAccount(accountID, value); err != nil {
			return nil, err
		}
		if _, exists := seen[value.ID]; exists {
			return nil, mailEventPolicyError(MailEventErrorDuplicate)
		}
		seen[value.ID] = struct{}{}
		ids[index] = value.ID
	}
	return ids, nil
}

func validateMailEventAddress(value MailEventAddress) (mailEventAddressJSON, error) {
	name := ""
	if value.Name != nil {
		name = *value.Name
	}
	normalized, err := NormalizeAddress(EngineAddress{Email: value.Email, Name: name})
	if err != nil {
		return mailEventAddressJSON{}, mailEventPolicyError(mailEventAddressErrorKind(err))
	}
	if normalized.Email != value.Email {
		return mailEventAddressJSON{}, mailEventPolicyError(MailEventErrorInvalid)
	}
	return mailEventAddressJSON{Email: value.Email, Name: copyMailEventName(value.Name)}, nil
}

func validateMailEventAddresses(values []MailEventAddress, required bool) ([]mailEventAddressJSON, error) {
	if required && len(values) == 0 {
		return nil, mailEventPolicyError(MailEventErrorInvalid)
	}
	if len(values) > maxMailEventRecipientItems {
		return nil, mailEventPolicyError(MailEventErrorTooLarge)
	}

	addresses := make([]mailEventAddressJSON, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		address, err := validateMailEventAddress(value)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[address.Email]; exists {
			return nil, mailEventPolicyError(MailEventErrorDuplicate)
		}
		seen[address.Email] = struct{}{}
		addresses[index] = address
	}
	return addresses, nil
}

func validateMailChangedFields(values []MailChangedField) error {
	if len(values) == 0 {
		return mailEventPolicyError(MailEventErrorInvalid)
	}
	if len(values) > maxMailEventChangedFieldItems {
		return mailEventPolicyError(MailEventErrorTooLarge)
	}
	seen := make(map[MailChangedField]struct{}, len(values))
	for _, value := range values {
		if !validMailChangedField(value) {
			return mailEventPolicyError(MailEventErrorInvalid)
		}
		if _, exists := seen[value]; exists {
			return mailEventPolicyError(MailEventErrorDuplicate)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateMailMessageFlags(values []MailMessageFlag) ([]MailMessageFlag, error) {
	if len(values) > maxMailEventFlagItems {
		return nil, mailEventPolicyError(MailEventErrorTooLarge)
	}
	flags := make([]MailMessageFlag, len(values))
	seen := make(map[MailMessageFlag]struct{}, len(values))
	for index, value := range values {
		if !validMailMessageFlag(value) {
			return nil, mailEventPolicyError(MailEventErrorInvalid)
		}
		if _, exists := seen[value]; exists {
			return nil, mailEventPolicyError(MailEventErrorDuplicate)
		}
		seen[value] = struct{}{}
		flags[index] = value
	}
	return flags, nil
}

func validMailCalendarMethod(value *MailCalendarMethod) bool {
	return value == nil || *value == MailCalendarMethodRequest || *value == MailCalendarMethodCancel || *value == MailCalendarMethodReply
}

func validMailAuthenticationClass(value MailAuthenticationClass) bool {
	return value == MailAuthenticationClassTrustedInternal || value == MailAuthenticationClassDMARCAligned || value == MailAuthenticationClassUntrusted
}

func mailCalendarPartLocatorHash(value []byte) *string {
	if value == nil {
		return nil
	}
	encoded := hex.EncodeToString(value)
	return &encoded
}

func decodeMailCalendarPartLocatorHash(value *string) ([]byte, bool) {
	if value == nil {
		return nil, true
	}
	if len(*value) != hex.EncodedLen(sha256.Size) {
		return nil, false
	}
	decoded, err := hex.DecodeString(*value)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != *value {
		return nil, false
	}
	return decoded, true
}

func validMailDeliveryClass(value MailDeliveryClass) bool {
	return value == MailDeliveryClassInternal || value == MailDeliveryClassInstanceDomain || value == MailDeliveryClassExternal
}

func validMailChangedField(value MailChangedField) bool {
	switch value {
	case MailChangedFieldFolderIDs, MailChangedFieldSeen, MailChangedFieldAnswered, MailChangedFieldFlagged, MailChangedFieldJunk, MailChangedFieldLabels, MailChangedFieldDeleted:
		return true
	default:
		return false
	}
}

func validMailMessageFlag(value MailMessageFlag) bool {
	switch value {
	case MailMessageFlagSeen, MailMessageFlagAnswered, MailMessageFlagFlagged, MailMessageFlagJunk, MailMessageFlagDraft:
		return true
	default:
		return false
	}
}

func validMailEventTime(value time.Time) bool {
	if value.IsZero() {
		return false
	}
	_, err := value.MarshalJSON()
	return err == nil
}

func hasMailChangedField(values []MailChangedField, candidate MailChangedField) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func copyMailEventName(value *string) *string {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func copyMailCalendarMethod(value *MailCalendarMethod) *MailCalendarMethod {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}

func mailEventAddressErrorKind(err error) MailEventErrorKind {
	if errors.Is(err, ErrAddressTooLarge) {
		return MailEventErrorTooLarge
	}
	if errors.Is(err, ErrAddressDuplicate) {
		return MailEventErrorDuplicate
	}
	return MailEventErrorInvalid
}

func mailEventPolicyError(kind MailEventErrorKind) error {
	return &MailEventPolicyError{Kind: kind}
}
