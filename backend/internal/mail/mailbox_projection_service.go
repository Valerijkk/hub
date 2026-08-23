package mail

import (
	"context"
	"errors"
	"reflect"
	"unicode/utf8"

	"github.com/google/uuid"
)

// MailboxFolderProjectionService resolves the authenticated owner before reading its durable folder projection.
type MailboxFolderProjectionService struct {
	owners     mailboxOwnerResolver
	repository MailboxFolderProjectionRepository
}

// NewMailboxFolderProjectionService creates the owner-gated folder projection read service.
func NewMailboxFolderProjectionService(owners mailboxOwnerResolver, repository MailboxFolderProjectionRepository) (*MailboxFolderProjectionService, error) {
	if mailboxFolderProjectionDependencyNil(owners) || mailboxFolderProjectionDependencyNil(repository) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxFolderProjectionService{owners: owners, repository: repository}, nil
}

// ListMailboxFolders returns the persisted folder projection for one authenticated owner.
func (service *MailboxFolderProjectionService) ListMailboxFolders(ctx context.Context, actorID uuid.UUID) (MailboxFolderProjection, error) {
	if service == nil || mailboxFolderProjectionDependencyNil(service.owners) || mailboxFolderProjectionDependencyNil(service.repository) || ctx == nil || actorID == uuid.Nil {
		return MailboxFolderProjection{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxFolderProjection{}, mailboxFolderProjectionOwnerError(err)
	}
	scope, err := mailboxFolderProjectionScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxFolderProjection{}, err
	}

	projection, err := service.repository.ListMailboxFolderProjection(ctx, scope)
	if err != nil {
		return MailboxFolderProjection{}, mailboxFolderProjectionRepositoryError(err)
	}
	if err := validateMailboxFolderProjection(projection); err != nil {
		return MailboxFolderProjection{}, err
	}
	projection = mailboxFolderProjectionCopy(projection)
	sortMailboxFolderProjectionItems(projection.Items)
	return projection, nil
}

func mailboxFolderProjectionScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (MailboxFolderProjectionScope, error) {
	if actorID == uuid.Nil || owner.OrganizationID() == uuid.Nil {
		return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || hasProjectionState {
			return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorNotFound)
	case MailboxOwnerDeleting:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorDeleting)
	case MailboxOwnerProvisioning, MailboxOwnerSuspended, MailboxOwnerError:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorUnavailable)
	case MailboxOwnerActive:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return MailboxFolderProjectionScope{
			ActorID:        actorID,
			OrganizationID: owner.OrganizationID(),
			AccountID:      accountID,
		}, nil
	default:
		return MailboxFolderProjectionScope{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
}

func mailboxFolderProjectionOwnerError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound, MailboxServiceErrorUnavailable, MailboxServiceErrorDeleting, MailboxServiceErrorInvalid:
		return mailboxServiceError(serviceError.Kind)
	default:
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
}

func mailboxFolderProjectionRepositoryError(err error) error {
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

func validateMailboxFolderProjection(projection MailboxFolderProjection) error {
	if len(projection.Items) > maxMailboxFolderProjectionItems || !validMailboxProjectionState(projection.ProjectionState) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionState == MailboxProjectionReady && projection.ProjectionAsOf == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionAsOf != nil && !mailboxEntValidTime(*projection.ProjectionAsOf) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	folders := make([]FolderTreeFolder, len(projection.Items))
	for index, item := range projection.Items {
		if !validMailboxFolderProjectionItem(item) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		parentID := uuid.Nil
		if item.ParentID != nil {
			parentID = *item.ParentID
		}
		role := FolderTreeRole("")
		if item.Role != nil {
			role = FolderTreeRole(*item.Role)
		}
		folders[index] = FolderTreeFolder{
			AccountID: mailboxFolderProjectionValidationAccountID,
			ID:        item.ID,
			ParentID:  parentID,
			Role:      role,
			Name:      item.Name,
			SortOrder: item.SortOrder,
		}
	}
	if _, err := NormalizeFolderTree(FolderTreeSnapshot{AccountID: mailboxFolderProjectionValidationAccountID, Folders: folders}); err != nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return nil
}

func validMailboxFolderProjectionItem(item MailboxFolderProjectionItem) bool {
	if item.ID == uuid.Nil || item.UnreadCount < 0 || item.TotalCount < 0 || item.UnreadCount > item.TotalCount || item.Version < 1 || !validMailboxFolderProjectionState(item.State) || !validMailboxFolderProjectionColor(item.Color) {
		return false
	}
	if item.ParentID != nil && (*item.ParentID == uuid.Nil || *item.ParentID == item.ID) {
		return false
	}
	normalizedName, err := normalizeFolderName(item.Name)
	if err != nil || normalizedName != item.normalizedName {
		return false
	}
	if item.State != "active" && (item.UnreadCount != 0 || item.TotalCount != 0) {
		return false
	}
	return mailboxFolderProjectionRoleRank(item.Role) != 7
}

var mailboxFolderProjectionValidationAccountID = uuid.MustParse("c8f06746-6a9d-4a71-8493-a66d9441c8f5")

func validMailboxFolderProjectionColor(value *string) bool {
	if value == nil {
		return true
	}
	return utf8.ValidString(*value) && len(*value) >= 1 && len(*value) <= 128 && utf8.RuneCountInString(*value) <= 32
}

func validMailboxFolderProjectionState(state string) bool {
	switch state {
	case "creating", "active", "unknown", "deleting", "error":
		return true
	default:
		return false
	}
}

func mailboxFolderProjectionDependencyNil(dependency any) bool {
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
