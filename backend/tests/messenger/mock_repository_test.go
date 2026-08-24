package messenger_test

// Тесты репозитория мессенджера: интеграционные и unit-сценарии testframework.
// W3-A3-011 / P3-BE-02: файл — mock-фикстура + контрактный Test* (не assert-free «мёртвый» *_test.go).

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/internal/messenger"
)

// ---------------------------------------------------------------------------
// Мок messenger.Repository
// ---------------------------------------------------------------------------

// mockRepository реализует интерфейс messenger.Repository для unit-тестов.
type mockRepository struct {
	// Каналы
	channels   map[uuid.UUID]*ent.Channel
	members    map[uuid.UUID]map[uuid.UUID]*ent.ChannelMember // channelID -> userID -> участник
	dmChannels map[string]*ent.Channel                        // dmKey -> канал
	// Сообщения
	messageMu                   sync.Mutex
	messages                    map[string]*messenger.MessageDocument
	messagesBySenderNonce       map[string]string
	createMessageCalls          int
	createMessageResultOverride *messenger.CreateMessageResult
	// Пины
	pins map[string]bool // channelID:messageID -> существует
	// Закладки
	bookmarks     map[uuid.UUID][]*ent.Bookmark
	bookmarksByID map[uuid.UUID]*ent.Bookmark
	// Приглашения
	invites        map[uuid.UUID]*ent.ChannelInvite
	invitesByToken map[string]*ent.ChannelInvite
	// Опросы
	polls       map[uuid.UUID]*ent.Poll
	pollOptions map[uuid.UUID][]*ent.PollOption // pollID -> варианты
	pollVotes   map[uuid.UUID][]*ent.PollVote   // pollID -> голоса
	// Упоминания
	displayNames         map[uuid.UUID]string
	projectMembersByRole map[uuid.UUID]map[string][]uuid.UUID
	projectMembers       map[uuid.UUID]map[uuid.UUID]bool // projectID -> userID
	// Состояние избранного/мьюта с ключом channelID:userID
	starred map[string]bool
	muted   map[string]*time.Time
	// Инъекция ошибок для тестов TASK-108 / TASK-144.
	addMembersErr             error
	addMembersErrExceptOwner  error // fail только role != "owner" (A3W3-2-002)
	isMemberErr               error
	findMessageIncludeDeleted bool
	deleteChannelErr          error
	addMembersCalls           int // счётчик вызовов AddMembers (для denial Join)
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		channels:              make(map[uuid.UUID]*ent.Channel),
		members:               make(map[uuid.UUID]map[uuid.UUID]*ent.ChannelMember),
		dmChannels:            make(map[string]*ent.Channel),
		messages:              make(map[string]*messenger.MessageDocument),
		messagesBySenderNonce: make(map[string]string),
		pins:                  make(map[string]bool),
		bookmarks:             make(map[uuid.UUID][]*ent.Bookmark),
		bookmarksByID:         make(map[uuid.UUID]*ent.Bookmark),
		invites:               make(map[uuid.UUID]*ent.ChannelInvite),
		invitesByToken:        make(map[string]*ent.ChannelInvite),
		polls:                 make(map[uuid.UUID]*ent.Poll),
		pollOptions:           make(map[uuid.UUID][]*ent.PollOption),
		pollVotes:             make(map[uuid.UUID][]*ent.PollVote),
		displayNames:          make(map[uuid.UUID]string),
		projectMembersByRole:  make(map[uuid.UUID]map[string][]uuid.UUID),
		projectMembers:        make(map[uuid.UUID]map[uuid.UUID]bool),
		starred:               make(map[string]bool),
		muted:                 make(map[string]*time.Time),
	}
}

func (m *mockRepository) CreateChannel(_ context.Context, input messenger.CreateChannelInput) (*ent.Channel, error) {
	ch := &ent.Channel{
		ID:          uuid.New(),
		ChannelType: channel.ChannelType(input.ChannelType),
		Name:        input.Name,
		Description: input.Description,
		IsReadonly:  input.IsReadonly,
		CreatedBy:   input.CreatedBy,
		ProjectID:   input.ProjectID,
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if input.DMKey != nil {
		ch.DmKey = input.DMKey
		m.dmChannels[*input.DMKey] = ch
	}
	m.channels[ch.ID] = ch
	m.members[ch.ID] = make(map[uuid.UUID]*ent.ChannelMember)
	return ch, nil
}

func (m *mockRepository) GetChannel(_ context.Context, channelID uuid.UUID) (*ent.Channel, error) {
	ch, ok := m.channels[channelID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if ch.DeletedAt != nil {
		return nil, &ent.NotFoundError{}
	}
	return ch, nil
}

func (m *mockRepository) ListUserChannels(_ context.Context, userID uuid.UUID) ([]*ent.Channel, error) {
	var result []*ent.Channel
	for chID, members := range m.members {
		if _, ok := members[userID]; ok {
			if ch, ok := m.channels[chID]; ok && ch.DeletedAt == nil {
				result = append(result, ch)
			}
		}
	}
	return result, nil
}

func (m *mockRepository) UpdateChannel(_ context.Context, channelID uuid.UUID, input messenger.UpdateChannelInput) (*ent.Channel, error) {
	ch, ok := m.channels[channelID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if ch.Version == 0 {
		ch.Version = 1
	}
	if ch.DeletedAt != nil || ch.Version != input.Version {
		return nil, &ent.NotFoundError{}
	}
	if input.Name != nil {
		ch.Name = *input.Name
	}
	if input.Description != nil {
		ch.Description = *input.Description
	}
	if input.IsReadonly != nil {
		ch.IsReadonly = *input.IsReadonly
	}
	ch.Version++
	return ch, nil
}

func (m *mockRepository) DeleteChannel(_ context.Context, channelID uuid.UUID) error {
	if m.deleteChannelErr != nil {
		return m.deleteChannelErr
	}
	ch, ok := m.channels[channelID]
	if !ok {
		return &ent.NotFoundError{}
	}
	now := time.Now().UTC()
	ch.DeletedAt = &now
	return nil
}

func (m *mockRepository) IsMember(_ context.Context, channelID, userID uuid.UUID) (bool, error) {
	if m.isMemberErr != nil {
		return false, m.isMemberErr
	}
	ch, ok := m.channels[channelID]
	if !ok || ch.DeletedAt != nil {
		return false, nil
	}
	members, ok := m.members[channelID]
	if !ok {
		return false, nil
	}
	_, exists := members[userID]
	return exists, nil
}

func (m *mockRepository) GetMemberRole(_ context.Context, channelID, userID uuid.UUID) (string, error) {
	ch, ok := m.channels[channelID]
	if !ok || ch.DeletedAt != nil {
		return "", &ent.NotFoundError{}
	}
	members, ok := m.members[channelID]
	if !ok {
		return "", &ent.NotFoundError{}
	}
	member, ok := members[userID]
	if !ok {
		return "", &ent.NotFoundError{}
	}
	return string(member.Role), nil
}

func (m *mockRepository) GetChannelMemberIDs(_ context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	if ch, ok := m.channels[channelID]; !ok || ch.DeletedAt != nil {
		return nil, nil
	}
	members, ok := m.members[channelID]
	if !ok {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(members))
	for uid := range members {
		ids = append(ids, uid)
	}
	return ids, nil
}

func (m *mockRepository) AddMembers(_ context.Context, channelID uuid.UUID, userIDs []uuid.UUID, role string) error {
	m.addMembersCalls++
	if m.addMembersErr != nil {
		return m.addMembersErr
	}
	if m.addMembersErrExceptOwner != nil && role != "owner" {
		return m.addMembersErrExceptOwner
	}
	if _, ok := m.members[channelID]; !ok {
		m.members[channelID] = make(map[uuid.UUID]*ent.ChannelMember)
	}
	for _, uid := range userIDs {
		m.members[channelID][uid] = &ent.ChannelMember{
			ID:        uuid.New(),
			ChannelID: channelID,
			UserID:    uid,
			Role:      channelmember.Role(role),
			JoinedAt:  time.Now(),
		}
	}
	return nil
}

func (m *mockRepository) RemoveMember(_ context.Context, channelID, userID uuid.UUID) error {
	if members, ok := m.members[channelID]; ok {
		delete(members, userID)
	}
	return nil
}

func (m *mockRepository) FindDMChannel(_ context.Context, dmKey string) (*ent.Channel, error) {
	ch, ok := m.dmChannels[dmKey]
	if !ok || ch.DeletedAt != nil {
		return nil, &ent.NotFoundError{}
	}
	return ch, nil
}

func (m *mockRepository) ListMembers(_ context.Context, channelID uuid.UUID) ([]*ent.ChannelMember, error) {
	members, ok := m.members[channelID]
	if !ok {
		return nil, nil
	}
	result := make([]*ent.ChannelMember, 0, len(members))
	for _, mem := range members {
		result = append(result, mem)
	}
	return result, nil
}

func (m *mockRepository) FilterUsersSharingChannel(_ context.Context, callerID uuid.UUID, candidateIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(candidateIDs) == 0 {
		return nil, nil
	}
	want := make(map[uuid.UUID]struct{}, len(candidateIDs))
	for _, id := range candidateIDs {
		want[id] = struct{}{}
	}
	shared := make(map[uuid.UUID]struct{})
	for chID, members := range m.members {
		ch, ok := m.channels[chID]
		if !ok || ch.DeletedAt != nil {
			continue
		}
		if _, ok := members[callerID]; !ok {
			continue
		}
		for uid := range members {
			if _, ok := want[uid]; ok {
				shared[uid] = struct{}{}
			}
		}
	}
	out := make([]uuid.UUID, 0, len(shared))
	for id := range shared {
		out = append(out, id)
	}
	return out, nil
}

// --- Сообщения ---

func (m *mockRepository) InsertMessage(_ context.Context, doc *messenger.MessageDocument) (string, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	m.messages[doc.ID] = doc
	return doc.ID, nil
}

func messageSenderNonceKey(senderID uuid.UUID, clientNonce string) string {
	return senderID.String() + "\x00" + clientNonce
}

func (m *mockRepository) CreateMessageIdempotent(_ context.Context, doc *messenger.MessageDocument) (*messenger.CreateMessageResult, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	m.createMessageCalls++
	if m.createMessageResultOverride != nil {
		return m.createMessageResultOverride, nil
	}
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	created := true
	canonical := doc
	if doc.ClientNonce != "" {
		key := messageSenderNonceKey(doc.SenderID, doc.ClientNonce)
		if messageID, ok := m.messagesBySenderNonce[key]; ok {
			canonical = m.messages[messageID]
			created = false
		} else {
			m.messages[doc.ID] = doc
			m.messagesBySenderNonce[key] = doc.ID
		}
	} else {
		m.messages[doc.ID] = doc
	}

	return &messenger.CreateMessageResult{Message: canonical, Created: created}, nil
}

func (m *mockRepository) FindMessageByClientNonce(_ context.Context, senderID uuid.UUID, clientNonce string) (*messenger.MessageDocument, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	messageID, ok := m.messagesBySenderNonce[messageSenderNonceKey(senderID, clientNonce)]
	if !ok {
		return nil, nil
	}
	return m.messages[messageID], nil
}

func (m *mockRepository) FindMessageByID(_ context.Context, messageID string) (*messenger.MessageDocument, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok || (doc.DeletedAt != nil && !m.findMessageIncludeDeleted) {
		return nil, &ent.NotFoundError{}
	}
	return doc, nil
}

func (m *mockRepository) ListMessages(_ context.Context, channelID uuid.UUID, _ *time.Time, limit int) ([]*messenger.MessageDocument, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	var result []*messenger.MessageDocument
	for _, doc := range m.messages {
		if doc.ChannelID == channelID && doc.ThreadID == "" && doc.DeletedAt == nil {
			result = append(result, doc)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockRepository) UpdateMessage(_ context.Context, messageID string, contentJSON map[string]interface{}, contentText string) (*messenger.MessageDocument, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok || doc.DeletedAt != nil {
		return nil, &ent.NotFoundError{}
	}
	now := time.Now()
	doc.ContentJSON = contentJSON
	doc.ContentText = contentText
	doc.EditedAt = &now
	return doc, nil
}

func (m *mockRepository) SoftDeleteMessage(_ context.Context, messageID string) error {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok {
		return &ent.NotFoundError{}
	}
	now := time.Now()
	doc.DeletedAt = &now
	return nil
}

func (m *mockRepository) IncrementReplyCount(_ context.Context, messageID string) error {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok {
		return nil
	}
	doc.ReplyCount++
	return nil
}

func (m *mockRepository) DecrementReplyCount(_ context.Context, messageID string) error {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok {
		return nil
	}
	doc.ReplyCount--
	return nil
}

func (m *mockRepository) ListThreadReplies(_ context.Context, threadID string, _ *time.Time, limit int) ([]*messenger.MessageDocument, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	var result []*messenger.MessageDocument
	for _, doc := range m.messages {
		if doc.ThreadID == threadID && doc.DeletedAt == nil {
			result = append(result, doc)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// --- Реакции ---

func (m *mockRepository) AddReaction(_ context.Context, messageID, userID, emoji string) error {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok {
		return &ent.NotFoundError{}
	}
	for i, r := range doc.Reactions {
		if r.Emoji == emoji {
			for _, uid := range r.UserIDs {
				if uid == userID {
					return nil
				}
			}
			doc.Reactions[i].UserIDs = append(doc.Reactions[i].UserIDs, userID)
			return nil
		}
	}
	doc.Reactions = append(doc.Reactions, messenger.ReactionEntry{Emoji: emoji, UserIDs: []string{userID}})
	return nil
}

func (m *mockRepository) RemoveReaction(_ context.Context, messageID, userID, emoji string) error {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok {
		return &ent.NotFoundError{}
	}
	for i, r := range doc.Reactions {
		if r.Emoji == emoji {
			newIDs := make([]string, 0)
			for _, uid := range r.UserIDs {
				if uid != userID {
					newIDs = append(newIDs, uid)
				}
			}
			doc.Reactions[i].UserIDs = newIDs
			return nil
		}
	}
	return nil
}

func (m *mockRepository) FindReaction(_ context.Context, messageID, userID, emoji string) (bool, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	doc, ok := m.messages[messageID]
	if !ok {
		return false, nil
	}
	for _, r := range doc.Reactions {
		if r.Emoji == emoji {
			for _, uid := range r.UserIDs {
				if uid == userID {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// --- Поиск ---

func (m *mockRepository) SearchMessages(_ context.Context, query string, channelIDs []uuid.UUID, limit int) ([]*messenger.MessageDocument, error) {
	m.messageMu.Lock()
	defer m.messageMu.Unlock()

	allowed := make(map[uuid.UUID]struct{}, len(channelIDs))
	for _, id := range channelIDs {
		allowed[id] = struct{}{}
	}
	var result []*messenger.MessageDocument
	for _, doc := range m.messages {
		if doc.DeletedAt != nil {
			continue
		}
		if len(allowed) == 0 {
			continue
		}
		if _, ok := allowed[doc.ChannelID]; !ok {
			continue
		}
		if contains(doc.ContentText, query) {
			result = append(result, doc)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// --- Пины ---

func (m *mockRepository) AddPin(_ context.Context, channelID uuid.UUID, messageID string, _ uuid.UUID) error {
	m.pins[channelID.String()+":"+messageID] = true
	return nil
}

func (m *mockRepository) RemovePin(_ context.Context, channelID uuid.UUID, messageID string) error {
	delete(m.pins, channelID.String()+":"+messageID)
	return nil
}

func (m *mockRepository) FindPin(_ context.Context, channelID uuid.UUID, messageID string) (*ent.PinnedMessage, error) {
	key := channelID.String() + ":" + messageID
	if _, ok := m.pins[key]; ok {
		return &ent.PinnedMessage{ID: uuid.New()}, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockRepository) ListPins(_ context.Context, channelID uuid.UUID) ([]*ent.PinnedMessage, error) {
	var result []*ent.PinnedMessage
	for key := range m.pins {
		parts := strings.Split(key, ":")
		if len(parts) == 2 && parts[0] == channelID.String() {
			result = append(result, &ent.PinnedMessage{
				ID:        uuid.New(),
				ChannelID: channelID,
				MessageID: parts[1],
				PinnedBy:  uuid.New(),
				PinnedAt:  time.Now(),
			})
		}
	}
	return result, nil
}

// --- Закладки ---

func (m *mockRepository) AddBookmark(_ context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) (*ent.Bookmark, error) {
	b := &ent.Bookmark{ID: uuid.New(), UserID: userID, MessageID: messageID, ChannelID: channelID, CreatedAt: time.Now()}
	m.bookmarks[userID] = append(m.bookmarks[userID], b)
	m.bookmarksByID[b.ID] = b
	return b, nil
}

func (m *mockRepository) RemoveBookmark(_ context.Context, userID, bookmarkID uuid.UUID) error {
	b, ok := m.bookmarksByID[bookmarkID]
	if !ok || b.UserID != userID {
		return &ent.NotFoundError{}
	}
	delete(m.bookmarksByID, bookmarkID)
	if list, ok := m.bookmarks[b.UserID]; ok {
		filtered := make([]*ent.Bookmark, 0, len(list))
		for _, item := range list {
			if item.ID != bookmarkID {
				filtered = append(filtered, item)
			}
		}
		m.bookmarks[b.UserID] = filtered
	}
	return nil
}

func (m *mockRepository) ListBookmarks(_ context.Context, userID uuid.UUID) ([]*ent.Bookmark, error) {
	return m.bookmarks[userID], nil
}

func memberKey(channelID, userID uuid.UUID) string {
	return channelID.String() + ":" + userID.String()
}

func (m *mockRepository) requireMember(channelID, userID uuid.UUID) error {
	members, ok := m.members[channelID]
	if !ok {
		return &ent.NotFoundError{}
	}
	if _, ok := members[userID]; !ok {
		return &ent.NotFoundError{}
	}
	return nil
}

func (m *mockRepository) ToggleStar(_ context.Context, channelID, userID uuid.UUID) (bool, error) {
	if err := m.requireMember(channelID, userID); err != nil {
		return false, err
	}
	key := memberKey(channelID, userID)
	m.starred[key] = !m.starred[key]
	return m.starred[key], nil
}

func (m *mockRepository) MuteChannel(_ context.Context, channelID, userID uuid.UUID, duration string) (time.Time, error) {
	if err := m.requireMember(channelID, userID); err != nil {
		return time.Time{}, err
	}
	until := time.Now().Add(time.Hour)
	if duration == "forever" {
		until = time.Now().Add(365 * 24 * time.Hour)
	}
	m.muted[memberKey(channelID, userID)] = &until
	return until, nil
}

func (m *mockRepository) UnmuteChannel(_ context.Context, channelID, userID uuid.UUID) error {
	if err := m.requireMember(channelID, userID); err != nil {
		return err
	}
	delete(m.muted, memberKey(channelID, userID))
	return nil
}

func (m *mockRepository) IsChannelMuted(_ context.Context, channelID, userID uuid.UUID) (bool, error) {
	until, ok := m.muted[memberKey(channelID, userID)]
	if !ok || until == nil {
		return false, nil
	}
	return until.After(time.Now()), nil
}

func (m *mockRepository) GetUserChannelMemberships(_ context.Context, userID uuid.UUID) ([]*ent.ChannelMember, error) {
	var result []*ent.ChannelMember
	for _, members := range m.members {
		if mem, ok := members[userID]; ok {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *mockRepository) EnsureIndexes(_ context.Context) error {
	return nil
}

// --- Упоминания ---

func (m *mockRepository) ListProjectMembersByRole(_ context.Context, projectID uuid.UUID, role string) ([]uuid.UUID, error) {
	byRole, ok := m.projectMembersByRole[projectID]
	if !ok {
		return nil, nil
	}
	return byRole[role], nil
}

func (m *mockRepository) IsProjectMember(_ context.Context, projectID, userID uuid.UUID) (bool, error) {
	members, ok := m.projectMembers[projectID]
	if !ok {
		return false, nil
	}
	return members[userID], nil
}

func (m *mockRepository) GetUserDisplayName(_ context.Context, userID uuid.UUID) (string, error) {
	return m.displayNames[userID], nil
}

// --- Приглашения в каналы ---

func (m *mockRepository) CreateChannelInvite(_ context.Context, input messenger.CreateChannelInviteInput) (*ent.ChannelInvite, error) {
	inv := &ent.ChannelInvite{
		ID:        uuid.New(),
		ChannelID: input.ChannelID,
		CreatorID: input.CreatorID,
		Token:     input.Token,
		MaxUses:   input.MaxUses,
		UseCount:  0,
		IsRevoked: false,
		ExpiresAt: input.ExpiresAt,
		CreatedAt: time.Now(),
	}
	m.invites[inv.ID] = inv
	m.invitesByToken[inv.Token] = inv
	return inv, nil
}

func (m *mockRepository) FindChannelInviteByToken(_ context.Context, token string) (*ent.ChannelInvite, error) {
	inv, ok := m.invitesByToken[token]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return inv, nil
}

func (m *mockRepository) IncrementInviteUses(_ context.Context, inviteID uuid.UUID, maxUses int) error {
	inv, ok := m.invites[inviteID]
	if !ok {
		return &ent.NotFoundError{}
	}
	if maxUses > 0 && inv.UseCount >= maxUses {
		return messenger.ErrInviteUsesExhausted
	}
	inv.UseCount++
	return nil
}

func (m *mockRepository) DecrementInviteUses(_ context.Context, inviteID uuid.UUID) error {
	inv, ok := m.invites[inviteID]
	if !ok {
		return &ent.NotFoundError{}
	}
	if inv.UseCount <= 0 {
		return &ent.NotFoundError{}
	}
	inv.UseCount--
	return nil
}

func (m *mockRepository) RevokeChannelInvite(_ context.Context, channelID, inviteID uuid.UUID) error {
	inv, ok := m.invites[inviteID]
	if !ok || inv.ChannelID != channelID {
		return &ent.NotFoundError{}
	}
	inv.IsRevoked = true
	return nil
}

func (m *mockRepository) ListChannelInvites(_ context.Context, channelID uuid.UUID) ([]*ent.ChannelInvite, error) {
	var result []*ent.ChannelInvite
	for _, inv := range m.invites {
		if inv.ChannelID == channelID && !inv.IsRevoked {
			result = append(result, inv)
		}
	}
	return result, nil
}

// --- Опросы ---

func (m *mockRepository) CreatePollWithOptions(_ context.Context, input messenger.CreatePollInput) (*ent.Poll, []*ent.PollOption, error) {
	p := &ent.Poll{
		ID:            uuid.New(),
		ChannelID:     input.ChannelID,
		CreatorID:     input.CreatorID,
		Question:      input.Question,
		IsAnonymous:   input.IsAnonymous,
		IsMultiChoice: input.IsMultiChoice,
		IsClosed:      false,
		CreatedAt:     time.Now(),
	}
	m.polls[p.ID] = p
	options := make([]*ent.PollOption, 0, len(input.Options))
	for i, text := range input.Options {
		opt := &ent.PollOption{
			ID:       uuid.New(),
			PollID:   p.ID,
			Text:     text,
			Position: i,
		}
		options = append(options, opt)
	}
	m.pollOptions[p.ID] = options
	return p, options, nil
}

func (m *mockRepository) GetPollWithOptions(_ context.Context, pollID uuid.UUID) (*ent.Poll, []*ent.PollOption, error) {
	p, ok := m.polls[pollID]
	if !ok {
		return nil, nil, &ent.NotFoundError{}
	}
	return p, m.pollOptions[pollID], nil
}

func (m *mockRepository) GetPoll(_ context.Context, pollID uuid.UUID) (*ent.Poll, error) {
	p, ok := m.polls[pollID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return p, nil
}

func (m *mockRepository) ClosePoll(_ context.Context, pollID uuid.UUID) error {
	p, ok := m.polls[pollID]
	if !ok {
		return &ent.NotFoundError{}
	}
	p.IsClosed = true
	return nil
}

func (m *mockRepository) VotePoll(_ context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error {
	// Удаляем предыдущие голоса пользователя в этом опросе.
	existing := m.pollVotes[pollID]
	filtered := existing[:0]
	for _, v := range existing {
		if v.UserID != userID {
			filtered = append(filtered, v)
		}
	}
	for _, oid := range optionIDs {
		filtered = append(filtered, &ent.PollVote{
			ID:       uuid.New(),
			PollID:   pollID,
			OptionID: oid,
			UserID:   userID,
		})
	}
	m.pollVotes[pollID] = filtered
	return nil
}

func (m *mockRepository) ListVotes(_ context.Context, pollID uuid.UUID) ([]*ent.PollVote, error) {
	return m.pollVotes[pollID], nil
}

// TestMockRepository_AddReactionContract — W3-A3-011: missing message → NotFound;
// existing message → reaction appended (mutation: no-op AddReaction → FAIL).
func TestMockRepository_AddReactionContract(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()
	msgID := uuid.New().String()
	userID := uuid.New()
	channelID := uuid.New()

	err := repo.AddReaction(ctx, msgID, userID.String(), "👍")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	_, err = repo.InsertMessage(ctx, &messenger.MessageDocument{
		ID: msgID, ChannelID: channelID, SenderID: userID, ContentText: "hi",
	})
	require.NoError(t, err)

	require.NoError(t, repo.AddReaction(ctx, msgID, userID.String(), "👍"))
	doc, err := repo.FindMessageByID(ctx, msgID)
	require.NoError(t, err)
	require.Len(t, doc.Reactions, 1)
	assert.Equal(t, "👍", doc.Reactions[0].Emoji)
	assert.Equal(t, []string{userID.String()}, doc.Reactions[0].UserIDs)

	list, err := repo.ListMessages(ctx, channelID, nil, 10)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}
