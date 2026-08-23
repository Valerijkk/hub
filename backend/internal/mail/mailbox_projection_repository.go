package mail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const maxMailboxFolderProjectionItems = 100

const (
	defaultMailboxThreadProjectionLimit = 50
	maxMailboxThreadProjectionLimit     = 100
	maxMailboxThreadProjectionCursor    = 512
	maxMailboxThreadProjectionSearch    = 128
	mailboxThreadProjectionCursorV1     = 1
	mailboxThreadProjectionSort         = "pinned_desc,last_message_at_desc,id_desc"
	mailboxThreadDetailProjectionSort   = "created_at_asc,id_asc"
)

var errMailboxThreadProjectionRequestInvalid = errors.New("mailbox thread projection request invalid")

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

// MailboxThreadProjectionRepository reads an owner-scoped, cursor-paged durable thread projection.
type MailboxThreadProjectionRepository interface {
	ListMailboxThreadProjection(context.Context, MailboxFolderProjectionScope, MailboxThreadProjectionQuery) (MailboxThreadProjection, error)
}

// MailboxThreadDetailProjectionRepository reads one owner-scoped thread with a bounded message-summary page.
type MailboxThreadDetailProjectionRepository interface {
	GetMailboxThreadDetailProjection(context.Context, MailboxFolderProjectionScope, MailboxThreadDetailProjectionQuery) (MailboxThreadDetailProjection, error)
}

// MailboxThreadProjectionQuery contains only validated, owner-scoped thread-list filters.
type MailboxThreadProjectionQuery struct {
	Cursor         *MailboxThreadProjectionCursor
	Limit          int
	FolderID       *uuid.UUID
	LabelID        *uuid.UUID
	Unread         *bool
	HasAttachments *bool
	Starred        *bool
	Search         *string
}

// MailboxThreadProjectionCursor is an opaque continuation position bound to one mailbox account.
type MailboxThreadProjectionCursor struct {
	accountID     uuid.UUID
	isPinned      bool
	lastMessageAt time.Time
	id            uuid.UUID
	filterSignature string
}

// MailboxThreadProjection is the safe, owner-scoped MailThreadPage read model.
type MailboxThreadProjection struct {
	Items           []MailboxThreadProjectionItem
	NextCursor      *string
	ProjectionState MailboxProjectionState
	ProjectionAsOf  *time.Time
}

// MailboxThreadProjectionItem is the public MailThreadSummary DTO and carries no body, BCC, or internal IDs.
type MailboxThreadProjectionItem struct {
	ID             uuid.UUID                            `json:"id"`
	Subject        string                               `json:"subject"`
	Participants   []MailboxThreadProjectionParticipant `json:"participants"`
	MessageCount   int32                                `json:"message_count"`
	UnreadCount    int32                                `json:"unread_count"`
	LastMessageAt  time.Time                            `json:"last_message_at"`
	Snippet        string                               `json:"snippet"`
	IsPinned       bool                                 `json:"is_pinned"`
	HasAttachments bool                                 `json:"has_attachments"`
	Labels         []MailboxThreadProjectionLabel       `json:"labels"`
}

// MailboxThreadDetailProjectionQuery contains the owner-scoped path identity and message-page continuation.
type MailboxThreadDetailProjectionQuery struct {
	ThreadID uuid.UUID
	Cursor   *MailboxThreadDetailProjectionCursor
	Limit    int
}

// MailboxThreadDetailProjectionCursor is an opaque continuation position bound to one account and thread.
type MailboxThreadDetailProjectionCursor struct {
	accountID       uuid.UUID
	threadID        uuid.UUID
	createdAt       time.Time
	id              uuid.UUID
	filterSignature string
}

// MailboxThreadDetailProjection is the safe, owner-scoped MailThreadDetail read model.
type MailboxThreadDetailProjection struct {
	Thread          MailboxThreadProjectionItem
	Messages        []MailboxThreadDetailProjectionMessage
	NextCursor      *string
	ProjectionState MailboxProjectionState
	ProjectionAsOf  *time.Time
}

// MailboxThreadDetailProjectionMessage is the closed MailMessageSummary projection for one thread page.
type MailboxThreadDetailProjectionMessage struct {
	ID              uuid.UUID                              `json:"id"`
	ThreadID        uuid.UUID                              `json:"thread_id"`
	From            MailboxThreadDetailProjectionAddress   `json:"from"`
	To              []MailboxThreadDetailProjectionAddress `json:"to"`
	CC              []MailboxThreadDetailProjectionAddress `json:"cc"`
	Subject         string                                 `json:"subject"`
	Snippet         string                                 `json:"snippet"`
	SentAt          *time.Time                             `json:"sent_at"`
	ReceivedAt      *time.Time                             `json:"received_at"`
	Priority        string                                 `json:"priority"`
	SizeBytes       int64                                  `json:"size_bytes"`
	HasAttachments  bool                                   `json:"has_attachments"`
	IsDraft         bool                                   `json:"is_draft"`
	FolderIDs       []uuid.UUID                            `json:"folder_ids"`
	Seen            bool                                   `json:"seen"`
	Flagged         bool                                   `json:"flagged"`
	Answered        bool                                   `json:"answered"`
	Junk            bool                                   `json:"junk"`
	CalendarMethod  *MailCalendarMethod                    `json:"calendar_method"`
	CalendarEventID *uuid.UUID                             `json:"calendar_event_id"`

	createdAt time.Time
}

// MailboxThreadDetailProjectionAddress is the public Address subset backed by a stored mail address.
type MailboxThreadDetailProjectionAddress struct {
	Email string  `json:"email"`
	Name  *string `json:"name"`
}

// MailboxThreadProjectionParticipant is the public participant address projection.
type MailboxThreadProjectionParticipant struct {
	Email string `json:"email"`
}

// MailboxThreadProjectionLabel is the public MailLabel projection for one thread.
type MailboxThreadProjectionLabel struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Color   string    `json:"color"`
	Version int64     `json:"version"`
}

type mailboxThreadProjectionCursorWire struct {
	Version       int    `json:"v"`
	AccountID     string `json:"a"`
	IsPinned      bool   `json:"p"`
	LastMessageAt string `json:"t"`
	ID            string `json:"i"`
	FilterSignature string `json:"f"`
}

type mailboxThreadDetailProjectionCursorWire struct {
	Version         int    `json:"v"`
	AccountID       string `json:"a"`
	ThreadID        string `json:"t"`
	CreatedAt       string `json:"c"`
	ID              string `json:"i"`
	FilterSignature string `json:"f"`
}

type mailboxThreadProjectionFilterWire struct {
	Version        int    `json:"v"`
	Sort           string `json:"s"`
	FolderID       string `json:"f"`
	LabelID        string `json:"l"`
	Unread         string `json:"u"`
	HasAttachments string `json:"a"`
	Starred        string `json:"r"`
	Search         string `json:"q"`
}

type mailboxThreadDetailProjectionFilterWire struct {
	Version int    `json:"v"`
	Sort    string `json:"s"`
}

func mailboxThreadProjectionQueryCopy(query MailboxThreadProjectionQuery) MailboxThreadProjectionQuery {
	result := MailboxThreadProjectionQuery{
		Limit:          query.Limit,
		FolderID:       mailboxFolderProjectionUUIDPointer(query.FolderID),
		LabelID:        mailboxFolderProjectionUUIDPointer(query.LabelID),
		Unread:         mailboxThreadProjectionBoolPointer(query.Unread),
		HasAttachments: mailboxThreadProjectionBoolPointer(query.HasAttachments),
		Starred:        mailboxThreadProjectionBoolPointer(query.Starred),
		Search:         mailboxFolderProjectionStringPointer(query.Search),
	}
	if query.Cursor != nil {
		cursor := *query.Cursor
		cursor.lastMessageAt = cursor.lastMessageAt.UTC()
		result.Cursor = &cursor
	}
	return result
}

func mailboxThreadProjectionBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func mailboxThreadProjectionCopy(projection MailboxThreadProjection) MailboxThreadProjection {
	result := MailboxThreadProjection{
		Items:           make([]MailboxThreadProjectionItem, len(projection.Items)),
		NextCursor:      mailboxFolderProjectionStringPointer(projection.NextCursor),
		ProjectionState: projection.ProjectionState,
		ProjectionAsOf:  mailboxFolderProjectionTimePointer(projection.ProjectionAsOf),
	}
	for index, item := range projection.Items {
		result.Items[index] = mailboxThreadProjectionItemCopy(item)
	}
	return result
}

func mailboxThreadProjectionItemCopy(item MailboxThreadProjectionItem) MailboxThreadProjectionItem {
	result := MailboxThreadProjectionItem{
		ID:             item.ID,
		Subject:        item.Subject,
		Participants:   make([]MailboxThreadProjectionParticipant, len(item.Participants)),
		MessageCount:   item.MessageCount,
		UnreadCount:    item.UnreadCount,
		LastMessageAt:  item.LastMessageAt.UTC(),
		Snippet:        item.Snippet,
		IsPinned:       item.IsPinned,
		HasAttachments: item.HasAttachments,
		Labels:         make([]MailboxThreadProjectionLabel, len(item.Labels)),
	}
	copy(result.Participants, item.Participants)
	copy(result.Labels, item.Labels)
	return result
}

func mailboxThreadProjectionQueryValid(query MailboxThreadProjectionQuery) bool {
	if query.Limit < 1 || query.Limit > maxMailboxThreadProjectionLimit {
		return false
	}
	if query.FolderID != nil && *query.FolderID == uuid.Nil || query.LabelID != nil && *query.LabelID == uuid.Nil {
		return false
	}
	if query.Search != nil {
		normalized, err := mailboxThreadProjectionNormalizeSearch(*query.Search)
		if err != nil || normalized != *query.Search {
			return false
		}
	}
	return query.Cursor == nil || mailboxThreadProjectionCursorValid(*query.Cursor)
}

func mailboxThreadProjectionQueryForScope(query MailboxThreadProjectionQuery, scope MailboxFolderProjectionScope) (MailboxThreadProjectionQuery, error) {
	if scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || !mailboxThreadProjectionQueryValid(query) {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	result := mailboxThreadProjectionQueryCopy(query)
	if result.Cursor != nil && (result.Cursor.accountID != scope.AccountID || result.Cursor.filterSignature != mailboxThreadProjectionQuerySignature(result)) {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	return result, nil
}

func mailboxThreadProjectionCursorValid(cursor MailboxThreadProjectionCursor) bool {
	return cursor.accountID != uuid.Nil && cursor.id != uuid.Nil && syncValidTime(cursor.lastMessageAt) && mailboxThreadProjectionFilterSignatureValid(cursor.filterSignature)
}

func mailboxThreadProjectionFilterSignatureValid(value string) bool {
	if len(value) != base64.RawURLEncoding.EncodedLen(sha256.Size) {
		return false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && base64.RawURLEncoding.EncodeToString(decoded) == value
}

func mailboxThreadProjectionQuerySignature(query MailboxThreadProjectionQuery) string {
	encoded, err := json.Marshal(mailboxThreadProjectionFilterWire{
		Version:        mailboxThreadProjectionCursorV1,
		Sort:           mailboxThreadProjectionSort,
		FolderID:       mailboxThreadProjectionUUIDSignature(query.FolderID),
		LabelID:        mailboxThreadProjectionUUIDSignature(query.LabelID),
		Unread:         mailboxThreadProjectionBoolSignature(query.Unread),
		HasAttachments: mailboxThreadProjectionBoolSignature(query.HasAttachments),
		Starred:        mailboxThreadProjectionBoolSignature(query.Starred),
		Search:         mailboxThreadProjectionSearchSignature(query.Search),
	})
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(encoded)
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func mailboxThreadProjectionUUIDSignature(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func mailboxThreadProjectionBoolSignature(value *bool) string {
	if value == nil {
		return "any"
	}
	if *value {
		return "true"
	}
	return "false"
}

func mailboxThreadProjectionSearchSignature(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mailboxThreadProjectionNormalizeSearch(value string) (string, error) {
	if !syncValidSafeText(value, maxMailboxThreadProjectionSearch, false) {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	normalized := norm.NFC.String(value)
	normalized = cases.Fold().String(normalized)
	normalized = norm.NFC.String(normalized)
	normalized = collapseThreadSubjectSpaces(normalized)
	if !syncValidSafeText(normalized, maxMailboxThreadProjectionSearch, false) || strings.ContainsAny(normalized, "%_\\") {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	return normalized, nil
}

func mailboxThreadProjectionDecodeCursor(value string) (*MailboxThreadProjectionCursor, error) {
	if value == "" || len(value) > maxMailboxThreadProjectionCursor {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 || len(decoded) > maxMailboxThreadProjectionCursor {
		return nil, errMailboxThreadProjectionRequestInvalid
	}

	var wire mailboxThreadProjectionCursorWire
	if err := json.Unmarshal(decoded, &wire); err != nil || wire.Version != mailboxThreadProjectionCursorV1 {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	accountID, err := uuid.Parse(wire.AccountID)
	if err != nil || accountID == uuid.Nil {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	id, err := uuid.Parse(wire.ID)
	if err != nil || id == uuid.Nil {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	lastMessageAt, err := time.Parse(time.RFC3339Nano, wire.LastMessageAt)
	if err != nil || lastMessageAt.Location() != time.UTC || wire.LastMessageAt != lastMessageAt.UTC().Format(time.RFC3339Nano) || !strings.HasSuffix(wire.LastMessageAt, "Z") {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	cursor := &MailboxThreadProjectionCursor{accountID: accountID, isPinned: wire.IsPinned, lastMessageAt: lastMessageAt.UTC(), id: id, filterSignature: wire.FilterSignature}
	if !mailboxThreadProjectionCursorValid(*cursor) {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	canonical, err := mailboxThreadProjectionEncodeCursor(*cursor)
	if err != nil || canonical != value {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	return cursor, nil
}

func mailboxThreadProjectionEncodeCursor(cursor MailboxThreadProjectionCursor) (string, error) {
	if !mailboxThreadProjectionCursorValid(cursor) {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	wire := mailboxThreadProjectionCursorWire{
		Version:       mailboxThreadProjectionCursorV1,
		AccountID:     cursor.accountID.String(),
		IsPinned:      cursor.isPinned,
		LastMessageAt: cursor.lastMessageAt.UTC().Format(time.RFC3339Nano),
		ID:            cursor.id.String(),
		FilterSignature: cursor.filterSignature,
	}
	encoded, err := json.Marshal(wire)
	if err != nil || len(encoded) == 0 || len(encoded) > maxMailboxThreadProjectionCursor {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	result := base64.RawURLEncoding.EncodeToString(encoded)
	if len(result) == 0 || len(result) > maxMailboxThreadProjectionCursor {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	return result, nil
}

func validateMailboxThreadProjection(projection MailboxThreadProjection, query MailboxThreadProjectionQuery, accountID uuid.UUID) error {
	if accountID == uuid.Nil || !mailboxThreadProjectionQueryValid(query) || len(projection.Items) > query.Limit || !validMailboxProjectionState(projection.ProjectionState) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionState == MailboxProjectionReady && projection.ProjectionAsOf == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionAsOf != nil && !mailboxEntValidTime(*projection.ProjectionAsOf) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	seen := make(map[uuid.UUID]struct{}, len(projection.Items))
	for index, item := range projection.Items {
		if !validMailboxThreadProjectionItem(item) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := seen[item.ID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[item.ID] = struct{}{}
		if index > 0 && !mailboxThreadProjectionItemBefore(projection.Items[index-1], item) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	if projection.NextCursor == nil {
		return nil
	}
	if len(projection.Items) != query.Limit {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	cursor, err := mailboxThreadProjectionDecodeCursor(*projection.NextCursor)
	if err != nil || cursor.accountID != accountID || cursor.filterSignature != mailboxThreadProjectionQuerySignature(query) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	last := projection.Items[len(projection.Items)-1]
	if cursor.isPinned != last.IsPinned || !cursor.lastMessageAt.Equal(last.LastMessageAt) || cursor.id != last.ID {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return nil
}

func validMailboxThreadProjectionItem(item MailboxThreadProjectionItem) bool {
	if item.ID == uuid.Nil || item.MessageCount < 1 || item.UnreadCount < 0 || item.UnreadCount > item.MessageCount || !syncValidSafeText(item.Subject, 998, true) || !syncValidSafeText(item.Snippet, MaxEngineSnippetBytes, true) || !syncValidTime(item.LastMessageAt) || len(item.Participants) < 1 || len(item.Participants) > 401 || len(item.Labels) > maxMailboxThreadProjectionLimit {
		return false
	}
	participants := make(map[string]struct{}, len(item.Participants))
	for _, participant := range item.Participants {
		if !syncValidSafeText(participant.Email, 320, false) {
			return false
		}
		if _, duplicate := participants[participant.Email]; duplicate {
			return false
		}
		participants[participant.Email] = struct{}{}
	}
	labels := make(map[uuid.UUID]struct{}, len(item.Labels))
	for _, label := range item.Labels {
		if label.ID == uuid.Nil || !syncValidSafeText(label.Name, 240, false) || utf8.RuneCountInString(label.Name) > 60 || !syncValidSafeText(label.Color, 128, false) || utf8.RuneCountInString(label.Color) > 32 || label.Version < 1 {
			return false
		}
		if _, duplicate := labels[label.ID]; duplicate {
			return false
		}
		labels[label.ID] = struct{}{}
	}
	return true
}

func mailboxThreadProjectionItemBefore(left, right MailboxThreadProjectionItem) bool {
	if left.IsPinned != right.IsPinned {
		return left.IsPinned
	}
	if !left.LastMessageAt.Equal(right.LastMessageAt) {
		return left.LastMessageAt.After(right.LastMessageAt)
	}
	return syncLessUUID(right.ID, left.ID)
}

func mailboxThreadDetailProjectionQueryCopy(query MailboxThreadDetailProjectionQuery) MailboxThreadDetailProjectionQuery {
	result := MailboxThreadDetailProjectionQuery{ThreadID: query.ThreadID, Limit: query.Limit}
	if query.Cursor != nil {
		cursor := *query.Cursor
		cursor.createdAt = cursor.createdAt.UTC()
		result.Cursor = &cursor
	}
	return result
}

func mailboxThreadDetailProjectionCopy(projection MailboxThreadDetailProjection) MailboxThreadDetailProjection {
	result := MailboxThreadDetailProjection{
		Thread:          mailboxThreadProjectionItemCopy(projection.Thread),
		Messages:        make([]MailboxThreadDetailProjectionMessage, len(projection.Messages)),
		NextCursor:      mailboxFolderProjectionStringPointer(projection.NextCursor),
		ProjectionState: projection.ProjectionState,
		ProjectionAsOf:  mailboxFolderProjectionTimePointer(projection.ProjectionAsOf),
	}
	for index, message := range projection.Messages {
		result.Messages[index] = mailboxThreadDetailProjectionMessageCopy(message)
	}
	return result
}

func mailboxThreadDetailProjectionMessageCopy(message MailboxThreadDetailProjectionMessage) MailboxThreadDetailProjectionMessage {
	result := MailboxThreadDetailProjectionMessage{
		ID:              message.ID,
		ThreadID:        message.ThreadID,
		From:            mailboxThreadDetailProjectionAddressCopy(message.From),
		To:              mailboxThreadDetailProjectionAddressesCopy(message.To),
		CC:              mailboxThreadDetailProjectionAddressesCopy(message.CC),
		Subject:         message.Subject,
		Snippet:         message.Snippet,
		SentAt:          mailboxFolderProjectionTimePointer(message.SentAt),
		ReceivedAt:      mailboxFolderProjectionTimePointer(message.ReceivedAt),
		Priority:        message.Priority,
		SizeBytes:       message.SizeBytes,
		HasAttachments:  message.HasAttachments,
		IsDraft:         message.IsDraft,
		FolderIDs:       append([]uuid.UUID{}, message.FolderIDs...),
		Seen:            message.Seen,
		Flagged:         message.Flagged,
		Answered:        message.Answered,
		Junk:            message.Junk,
		CalendarMethod:  copyMailCalendarMethod(message.CalendarMethod),
		CalendarEventID: mailboxFolderProjectionUUIDPointer(message.CalendarEventID),
		createdAt:       message.createdAt.UTC(),
	}
	return result
}

func mailboxThreadDetailProjectionAddressCopy(address MailboxThreadDetailProjectionAddress) MailboxThreadDetailProjectionAddress {
	return MailboxThreadDetailProjectionAddress{Email: address.Email, Name: mailboxFolderProjectionStringPointer(address.Name)}
}

func mailboxThreadDetailProjectionAddressesCopy(addresses []MailboxThreadDetailProjectionAddress) []MailboxThreadDetailProjectionAddress {
	result := make([]MailboxThreadDetailProjectionAddress, len(addresses))
	for index, address := range addresses {
		result[index] = mailboxThreadDetailProjectionAddressCopy(address)
	}
	return result
}

func mailboxThreadDetailProjectionQueryValid(query MailboxThreadDetailProjectionQuery) bool {
	return query.ThreadID != uuid.Nil && query.Limit >= 1 && query.Limit <= maxMailboxThreadProjectionLimit && (query.Cursor == nil || mailboxThreadDetailProjectionCursorValid(*query.Cursor))
}

func mailboxThreadDetailProjectionQueryForScope(query MailboxThreadDetailProjectionQuery, scope MailboxFolderProjectionScope) (MailboxThreadDetailProjectionQuery, error) {
	if scope.ActorID == uuid.Nil || scope.OrganizationID == uuid.Nil || scope.AccountID == uuid.Nil || !mailboxThreadDetailProjectionQueryValid(query) {
		return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	result := mailboxThreadDetailProjectionQueryCopy(query)
	if result.Cursor != nil && (result.Cursor.accountID != scope.AccountID || result.Cursor.threadID != result.ThreadID || result.Cursor.filterSignature != mailboxThreadDetailProjectionQuerySignature()) {
		return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	return result, nil
}

func mailboxThreadDetailProjectionCursorValid(cursor MailboxThreadDetailProjectionCursor) bool {
	return cursor.accountID != uuid.Nil && cursor.threadID != uuid.Nil && cursor.id != uuid.Nil && syncValidTime(cursor.createdAt) && mailboxThreadProjectionFilterSignatureValid(cursor.filterSignature)
}

func mailboxThreadDetailProjectionQuerySignature() string {
	encoded, err := json.Marshal(mailboxThreadDetailProjectionFilterWire{Version: mailboxThreadProjectionCursorV1, Sort: mailboxThreadDetailProjectionSort})
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(encoded)
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func mailboxThreadDetailProjectionDecodeCursor(value string) (*MailboxThreadDetailProjectionCursor, error) {
	if value == "" || len(value) > maxMailboxThreadProjectionCursor {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) == 0 || len(decoded) > maxMailboxThreadProjectionCursor {
		return nil, errMailboxThreadProjectionRequestInvalid
	}

	var wire mailboxThreadDetailProjectionCursorWire
	if err := json.Unmarshal(decoded, &wire); err != nil || wire.Version != mailboxThreadProjectionCursorV1 {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	accountID, err := uuid.Parse(wire.AccountID)
	if err != nil || accountID == uuid.Nil {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	threadID, err := uuid.Parse(wire.ThreadID)
	if err != nil || threadID == uuid.Nil {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	id, err := uuid.Parse(wire.ID)
	if err != nil || id == uuid.Nil {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	createdAt, err := time.Parse(time.RFC3339Nano, wire.CreatedAt)
	if err != nil || createdAt.Location() != time.UTC || wire.CreatedAt != createdAt.UTC().Format(time.RFC3339Nano) || !strings.HasSuffix(wire.CreatedAt, "Z") {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	cursor := &MailboxThreadDetailProjectionCursor{accountID: accountID, threadID: threadID, createdAt: createdAt.UTC(), id: id, filterSignature: wire.FilterSignature}
	if !mailboxThreadDetailProjectionCursorValid(*cursor) {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	canonical, err := mailboxThreadDetailProjectionEncodeCursor(*cursor)
	if err != nil || canonical != value {
		return nil, errMailboxThreadProjectionRequestInvalid
	}
	return cursor, nil
}

func mailboxThreadDetailProjectionEncodeCursor(cursor MailboxThreadDetailProjectionCursor) (string, error) {
	if !mailboxThreadDetailProjectionCursorValid(cursor) {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	wire := mailboxThreadDetailProjectionCursorWire{
		Version:         mailboxThreadProjectionCursorV1,
		AccountID:       cursor.accountID.String(),
		ThreadID:        cursor.threadID.String(),
		CreatedAt:       cursor.createdAt.UTC().Format(time.RFC3339Nano),
		ID:              cursor.id.String(),
		FilterSignature: cursor.filterSignature,
	}
	encoded, err := json.Marshal(wire)
	if err != nil || len(encoded) == 0 || len(encoded) > maxMailboxThreadProjectionCursor {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	result := base64.RawURLEncoding.EncodeToString(encoded)
	if len(result) == 0 || len(result) > maxMailboxThreadProjectionCursor {
		return "", errMailboxThreadProjectionRequestInvalid
	}
	return result, nil
}

func validateMailboxThreadDetailProjection(projection MailboxThreadDetailProjection, query MailboxThreadDetailProjectionQuery, accountID uuid.UUID) error {
	if accountID == uuid.Nil || !mailboxThreadDetailProjectionQueryValid(query) || projection.Thread.ID != query.ThreadID || len(projection.Messages) > query.Limit || !validMailboxProjectionState(projection.ProjectionState) || !validMailboxThreadProjectionItem(projection.Thread) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionState == MailboxProjectionReady && projection.ProjectionAsOf == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionAsOf != nil && !mailboxEntValidTime(*projection.ProjectionAsOf) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	seen := make(map[uuid.UUID]struct{}, len(projection.Messages))
	for index, message := range projection.Messages {
		if !validMailboxThreadDetailProjectionMessage(message, query.ThreadID) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := seen[message.ID]; duplicate {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[message.ID] = struct{}{}
		if index > 0 && !mailboxThreadDetailProjectionMessageBefore(projection.Messages[index-1], message) {
			return mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}
	if projection.NextCursor == nil {
		return nil
	}
	if len(projection.Messages) != query.Limit {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	cursor, err := mailboxThreadDetailProjectionDecodeCursor(*projection.NextCursor)
	if err != nil || cursor.accountID != accountID || cursor.threadID != query.ThreadID || cursor.filterSignature != mailboxThreadDetailProjectionQuerySignature() {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	last := projection.Messages[len(projection.Messages)-1]
	if !cursor.createdAt.Equal(last.createdAt) || cursor.id != last.ID {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return nil
}

func validMailboxThreadDetailProjectionMessage(message MailboxThreadDetailProjectionMessage, threadID uuid.UUID) bool {
	if message.ID == uuid.Nil || message.ThreadID != threadID || !validMailboxThreadDetailProjectionAddress(message.From) || len(message.To) > 200 || len(message.CC) > 200 || !syncValidSafeText(message.Subject, 998, true) || !syncValidSafeText(message.Snippet, MaxEngineSnippetBytes, true) || !syncValidTime(message.createdAt) || message.SizeBytes < 0 || message.SizeBytes > MaxEngineMessageBytes || len(message.FolderIDs) > maxMailboxThreadProjectionLimit {
		return false
	}
	if message.SentAt != nil && !syncValidTime(*message.SentAt) || message.ReceivedAt != nil && !syncValidTime(*message.ReceivedAt) || message.CalendarEventID != nil && *message.CalendarEventID == uuid.Nil || !validMailCalendarMethod(message.CalendarMethod) {
		return false
	}
	switch message.Priority {
	case string(EnginePriorityLow), string(EnginePriorityNormal), string(EnginePriorityHigh):
	default:
		return false
	}
	for _, addresses := range [][]MailboxThreadDetailProjectionAddress{message.To, message.CC} {
		for _, address := range addresses {
			if !validMailboxThreadDetailProjectionAddress(address) {
				return false
			}
		}
	}
	for index, folderID := range message.FolderIDs {
		if folderID == uuid.Nil || index > 0 && !syncLessUUID(message.FolderIDs[index-1], folderID) {
			return false
		}
	}
	return true
}

func validMailboxThreadDetailProjectionAddress(address MailboxThreadDetailProjectionAddress) bool {
	if !syncValidSafeText(address.Email, 320, false) || strings.Count(address.Email, "@") != 1 {
		return false
	}
	return address.Name == nil || syncValidSafeText(*address.Name, 640, true) && utf8.RuneCountInString(*address.Name) <= 160
}

func mailboxThreadDetailProjectionMessageBefore(left, right MailboxThreadDetailProjectionMessage) bool {
	if !left.createdAt.Equal(right.createdAt) {
		return left.createdAt.Before(right.createdAt)
	}
	return syncLessUUID(left.ID, right.ID)
}
