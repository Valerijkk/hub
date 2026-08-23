package mail

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
)

const maxMailboxFolderProjectionItems = 100

// MailboxFolderProjectionScope is the complete authenticated ownership scope for one folder projection read.
type MailboxFolderProjectionScope struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
}

// MailboxFolderProjectionRepository reads an owner-scoped durable folder projection.
type MailboxFolderProjectionRepository interface {
	ListMailboxFolderProjection(context.Context, MailboxFolderProjectionScope) (MailboxFolderProjection, error)
}

// MailboxFolderProjection is the safe, owner-scoped folder read model.
type MailboxFolderProjection struct {
	Items           []MailboxFolderProjectionItem
	ProjectionState MailboxProjectionState
	ProjectionAsOf  *time.Time
}

// MailboxFolderProjectionItem is the public MailFolder DTO and carries no account or engine identifiers.
type MailboxFolderProjectionItem struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Role        *string    `json:"role"`
	Name        string     `json:"name"`
	Color       *string    `json:"color"`
	SortOrder   int32      `json:"sort_order"`
	UnreadCount int32      `json:"unread_count"`
	TotalCount  int32      `json:"total_count"`
	State       string     `json:"state"`
	Version     int64      `json:"version"`

	normalizedName string
}

func mailboxFolderProjectionCopy(projection MailboxFolderProjection) MailboxFolderProjection {
	result := MailboxFolderProjection{
		Items:           make([]MailboxFolderProjectionItem, len(projection.Items)),
		ProjectionState: projection.ProjectionState,
		ProjectionAsOf:  mailboxFolderProjectionTimePointer(projection.ProjectionAsOf),
	}
	for index, item := range projection.Items {
		result.Items[index] = mailboxFolderProjectionItemCopy(item)
	}
	return result
}

func mailboxFolderProjectionItemCopy(item MailboxFolderProjectionItem) MailboxFolderProjectionItem {
	return MailboxFolderProjectionItem{
		ID:             item.ID,
		ParentID:       mailboxFolderProjectionUUIDPointer(item.ParentID),
		Role:           mailboxFolderProjectionStringPointer(item.Role),
		Name:           item.Name,
		Color:          mailboxFolderProjectionStringPointer(item.Color),
		SortOrder:      item.SortOrder,
		UnreadCount:    item.UnreadCount,
		TotalCount:     item.TotalCount,
		State:          item.State,
		Version:        item.Version,
		normalizedName: item.normalizedName,
	}
}

func mailboxFolderProjectionUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxFolderProjectionStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxFolderProjectionTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func sortMailboxFolderProjectionItems(items []MailboxFolderProjectionItem) {
	sort.Slice(items, func(leftIndex, rightIndex int) bool {
		left := items[leftIndex]
		right := items[rightIndex]
		leftRank := mailboxFolderProjectionRoleRank(left.Role)
		rightRank := mailboxFolderProjectionRoleRank(right.Role)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if left.SortOrder != right.SortOrder {
			return left.SortOrder < right.SortOrder
		}
		if left.normalizedName != right.normalizedName {
			return left.normalizedName < right.normalizedName
		}
		return left.ID.String() < right.ID.String()
	})
}

func mailboxFolderProjectionRoleRank(role *string) int {
	if role == nil {
		return 6
	}
	switch *role {
	case "inbox":
		return 0
	case "sent":
		return 1
	case "drafts":
		return 2
	case "archive":
		return 3
	case "trash":
		return 4
	case "junk":
		return 5
	default:
		return 7
	}
}
