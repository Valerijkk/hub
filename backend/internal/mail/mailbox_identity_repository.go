package mail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	defaultMailboxIdentityLimit              = 50
	maxMailboxIdentityLimit                  = 100
	maxMailboxIdentityCursor                 = 512
	maxMailboxIdentityResponseBytes    int64 = 8 << 20
	mailboxIdentityCursorVersion             = 1
)

var (
	errMailboxIdentityRequestInvalid       = errors.New("mailbox identity request invalid")
	errMailboxIdentityResponseLimitExceeded = errors.New("mailbox identity response limit exceeded")
)

// MailIdentity is the public, server-authorized From identity projection.
// It intentionally excludes account and engine-private identifiers.
type MailIdentity struct {
	ID          uuid.UUID `json:"id"`
	Address     string    `json:"address"`
	DisplayName string    `json:"display_name"`
	Kind        string    `json:"kind"`
	SendAllowed bool      `json:"send_allowed"`
	IsDefault   bool      `json:"is_default"`
	Status      string    `json:"status"`

	normalizedAddress string
}

// MailboxIdentityListQuery contains a validated cursor page request.
type MailboxIdentityListQuery struct {
	Cursor *MailboxIdentityCursor
	Limit  int
}

// MailboxIdentityCursor is an opaque account-bound keyset position.
type MailboxIdentityCursor struct {
	accountID         uuid.UUID
	normalizedAddress string
	id                uuid.UUID
}

// MailboxIdentityList is the bounded public identity page.
type MailboxIdentityList struct {
	Items      []MailIdentity `json:"items"`
	NextCursor *string        `json:"next_cursor"`
}

// MailboxIdentityScope is the complete authenticated owner scope for one read.
type MailboxIdentityScope struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
}

// MailboxIdentityRepository reads a durable, owner-scoped identity page.
type MailboxIdentityRepository interface {
	ListMailboxIdentities(context.Context, MailboxIdentityScope, MailboxIdentityListQuery) (MailboxIdentityList, error)
}

type mailboxIdentityReader interface {
	ListMailboxIdentities(context.Context, uuid.UUID, MailboxIdentityListQuery) (MailboxIdentityList, error)
}

// MailboxIdentityService resolves the owner before its identity repository read.
type MailboxIdentityService struct {
	owners     mailboxOwnerResolver
	repository MailboxIdentityRepository
}

type mailboxIdentityCursorWire struct {
	Version           int    `json:"v"`
	AccountID         string `json:"a"`
	NormalizedAddress string `json:"n"`
	ID                string `json:"i"`
}

// NewMailboxIdentityService creates the owner-gated public identity read service.
func NewMailboxIdentityService(owners mailboxOwnerResolver, repository MailboxIdentityRepository) (*MailboxIdentityService, error) {
	if mailboxEntNil(owners) || mailboxEntNil(repository) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxIdentityService{owners: owners, repository: repository}, nil
}

// ListMailboxIdentities returns only active server-authorized From identities.
func (service *MailboxIdentityService) ListMailboxIdentities(ctx context.Context, actorID uuid.UUID, query MailboxIdentityListQuery) (MailboxIdentityList, error) {
	if service == nil || mailboxEntNil(service.owners) || mailboxEntNil(service.repository) || ctx == nil || actorID == uuid.Nil || !mailboxIdentityListQueryValid(query) {
		return MailboxIdentityList{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	query = mailboxIdentityListQueryCopy(query)

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxIdentityList{}, mailboxIdentityOwnerError(err)
	}
	scope, err := mailboxIdentityScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxIdentityList{}, err
	}
	query, err = mailboxIdentityQueryForScope(query, scope)
	if err != nil {
		return MailboxIdentityList{}, errMailboxIdentityRequestInvalid
	}

	page, err := service.repository.ListMailboxIdentities(ctx, scope, query)
	if err != nil {
		return MailboxIdentityList{}, mailboxIdentityRepositoryError(err)
	}
	if !mailboxIdentityListValid(page, query, scope.AccountID) {
		return MailboxIdentityList{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return mailboxIdentityListCopy(page), nil
}

func mailboxIdentityScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (MailboxIdentityScope, error) {
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil {
		return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || hasProjectionState {
			return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorNotFound)
	case MailboxOwnerDeleting:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorDeleting)
	case MailboxOwnerProvisioning, MailboxOwnerSuspended, MailboxOwnerError:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	case MailboxOwnerActive:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || projectionState != MailboxProjectionReady || !owner.hasEngineAccountID || !validMailboxOwnerOpaque(owner.engineAccountID) {
			return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorUnavailable)
		}
		return MailboxIdentityScope{
			ActorID:        actorID,
			OrganizationID: owner.OrganizationID(),
			AccountID:      accountID,
		}, nil
	default:
		return MailboxIdentityScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
}

func mailboxIdentityOwnerError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound, MailboxServiceErrorUnavailable, MailboxServiceErrorDeleting, MailboxServiceErrorInvalid:
		return mailboxServiceError(serviceError.Kind)
	default:
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
}

func mailboxIdentityRepositoryError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound, MailboxServiceErrorUnavailable, MailboxServiceErrorDeleting, MailboxServiceErrorInvalid:
		return mailboxServiceError(serviceError.Kind)
	default:
		return mailboxServiceError(MailboxServiceErrorUnavailable)
	}
}

func mailboxIdentityListQueryValid(query MailboxIdentityListQuery) bool {
	return query.Limit >= 1 && query.Limit <= maxMailboxIdentityLimit && (query.Cursor == nil || mailboxIdentityCursorValid(*query.Cursor))
}

func mailboxIdentityListQueryCopy(query MailboxIdentityListQuery) MailboxIdentityListQuery {
	result := MailboxIdentityListQuery{Limit: query.Limit}
	if query.Cursor != nil {
		cursor := *query.Cursor
		result.Cursor = &cursor
	}
	return result
}

func mailboxIdentityQueryForScope(query MailboxIdentityListQuery, scope MailboxIdentityScope) (MailboxIdentityListQuery, error) {
	if scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || !mailboxIdentityListQueryValid(query) {
		return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
	}
	result := mailboxIdentityListQueryCopy(query)
	if result.Cursor != nil && result.Cursor.accountID != scope.AccountID {
		return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
	}
	return result, nil
}

func mailboxIdentityCursorValid(cursor MailboxIdentityCursor) bool {
	return cursor.accountID != uuid.Nil && cursor.id != uuid.Nil && validMailboxIdentityAddress(cursor.normalizedAddress)
}

func mailboxIdentityDecodeCursor(value string) (*MailboxIdentityCursor, error) {
	if value == "" || len(value) > maxMailboxIdentityCursor {
		return nil, errMailboxIdentityRequestInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != value {
		return nil, errMailboxIdentityRequestInvalid
	}

	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	var wire mailboxIdentityCursorWire
	if err := decoder.Decode(&wire); err != nil {
		return nil, errMailboxIdentityRequestInvalid
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, errMailboxIdentityRequestInvalid
	}
	accountID, err := uuid.Parse(wire.AccountID)
	if err != nil || accountID == uuid.Nil || wire.AccountID != accountID.String() {
		return nil, errMailboxIdentityRequestInvalid
	}
	id, err := uuid.Parse(wire.ID)
	if err != nil || id == uuid.Nil || wire.ID != id.String() {
		return nil, errMailboxIdentityRequestInvalid
	}
	cursor := MailboxIdentityCursor{accountID: accountID, normalizedAddress: wire.NormalizedAddress, id: id}
	if wire.Version != mailboxIdentityCursorVersion || !mailboxIdentityCursorValid(cursor) {
		return nil, errMailboxIdentityRequestInvalid
	}
	encoded, err := mailboxIdentityEncodeCursor(cursor)
	if err != nil || encoded != value {
		return nil, errMailboxIdentityRequestInvalid
	}
	return &cursor, nil
}

func mailboxIdentityEncodeCursor(cursor MailboxIdentityCursor) (string, error) {
	if !mailboxIdentityCursorValid(cursor) {
		return "", errMailboxIdentityRequestInvalid
	}
	encoded, err := json.Marshal(mailboxIdentityCursorWire{
		Version:           mailboxIdentityCursorVersion,
		AccountID:         cursor.accountID.String(),
		NormalizedAddress: cursor.normalizedAddress,
		ID:                cursor.id.String(),
	})
	if err != nil {
		return "", errMailboxIdentityRequestInvalid
	}
	value := base64.RawURLEncoding.EncodeToString(encoded)
	if len(value) == 0 || len(value) > maxMailboxIdentityCursor {
		return "", errMailboxIdentityRequestInvalid
	}
	return value, nil
}

func mailboxIdentityListCopy(page MailboxIdentityList) MailboxIdentityList {
	result := MailboxIdentityList{
		Items:      make([]MailIdentity, len(page.Items)),
		NextCursor: mailboxEntStringPointer(page.NextCursor),
	}
	copy(result.Items, page.Items)
	return result
}

func mailboxIdentityListValid(page MailboxIdentityList, query MailboxIdentityListQuery, accountID uuid.UUID) bool {
	if accountID == uuid.Nil || !mailboxIdentityListQueryValid(query) || len(page.Items) > query.Limit {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(page.Items))
	for index, item := range page.Items {
		if !validMailboxIdentity(item) {
			return false
		}
		if _, duplicate := seen[item.ID]; duplicate {
			return false
		}
		seen[item.ID] = struct{}{}
		if index > 0 && !mailboxIdentityBefore(page.Items[index-1], item) {
			return false
		}
	}
	if page.NextCursor == nil {
		return true
	}
	if len(page.Items) != query.Limit {
		return false
	}
	cursor, err := mailboxIdentityDecodeCursor(*page.NextCursor)
	if err != nil || cursor.accountID != accountID {
		return false
	}
	last := page.Items[len(page.Items)-1]
	return cursor.normalizedAddress == last.normalizedAddress && cursor.id == last.ID
}

func mailboxIdentityBefore(left, right MailIdentity) bool {
	if left.normalizedAddress != right.normalizedAddress {
		return left.normalizedAddress < right.normalizedAddress
	}
	return left.ID.String() < right.ID.String()
}

func validMailboxIdentity(item MailIdentity) bool {
	if item.ID == uuid.Nil || item.Status != "active" || !item.SendAllowed || item.Kind == "group" || !validMailboxIdentityKind(item.Kind) || !validMailboxIdentityAddress(item.Address) || item.normalizedAddress != item.Address {
		return false
	}
	return validMailboxIdentityDisplayName(item.DisplayName)
}

func validMailboxIdentityKind(value string) bool {
	switch value {
	case "primary", "alias":
		return true
	default:
		return false
	}
}

func validMailboxIdentityAddress(value string) bool {
	if len(value) == 0 || len(value) > 320 || !utf8.ValidString(value) {
		return false
	}
	normalized, err := NormalizeAddress(EngineAddress{Email: value})
	return err == nil && normalized.Email == value
}

func validMailboxIdentityDisplayName(value string) bool {
	return utf8.ValidString(value) && len(value) <= 640 && utf8.RuneCountInString(value) <= 160
}
