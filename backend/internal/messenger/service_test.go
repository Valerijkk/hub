package messenger

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
)

// ---------------------------------------------------------------------------
// Mock Repository
// ---------------------------------------------------------------------------

// mockRepository реализует интерфейс Repository для unit-тестов.
type mockRepository struct {
	// Каналы
	channels       map[uuid.UUID]*ent.Channel
	members        map[uuid.UUID]map[uuid.UUID]*ent.ChannelMember // channelID -> userID -> member
	dmChannels     map[string]*ent.Channel                        // dmKey -> channel
	// Сообщения
	messages       map[string]*MessageDocument
	// Пины
	pins           map[string]bool // channelID:messageID -> exists
	// Закладки
	bookmarks      map[uuid.UUID][]*ent.Bookmark
	bookmarksByID  map[uuid.UUID]*ent.Bookmark
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		channels:      make(map[uuid.UUID]*ent.Channel),
		members:       make(map[uuid.UUID]map[uuid.UUID]*ent.ChannelMember),
		dmChannels:    make(map[string]*ent.Channel),
		messages:      make(map[string]*MessageDocument),
		pins:          make(map[string]bool),
		bookmarks:     make(map[uuid.UUID][]*ent.Bookmark),
		bookmarksByID: make(map[uuid.UUID]*ent.Bookmark),
	}
}

func (m *mockRepository) CreateChannel(_ context.Context, input CreateChannelInput) (*ent.Channel, error) {
	ch := &ent.Channel{
		ID:          uuid.New(),
		ChannelType: channel.ChannelType(input.ChannelType),
		Name:        input.Name,
		Description: input.Description,
		IsReadonly:  input.IsReadonly,
		CreatedBy:   input.CreatedBy,
		ProjectID:   input.ProjectID,
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
	return ch, nil
}

func (m *mockRepository) ListUserChannels(_ context.Context, userID uuid.UUID) ([]*ent.Channel, error) {
	var result []*ent.Channel
	for chID, members := range m.members {
		if _, ok := members[userID]; ok {
			if ch, ok := m.channels[chID]; ok {
				result = append(result, ch)
			}
		}
	}
	return result, nil
}

func (m *mockRepository) UpdateChannel(_ context.Context, channelID uuid.UUID, input UpdateChannelInput) (*ent.Channel, error) {
	ch, ok := m.channels[channelID]
	if !ok {
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
	return ch, nil
}

func (m *mockRepository) DeleteChannel(_ context.Context, channelID uuid.UUID) error {
	delete(m.channels, channelID)
	delete(m.members, channelID)
	return nil
}

func (m *mockRepository) IsMember(_ context.Context, channelID, userID uuid.UUID) (bool, error) {
	members, ok := m.members[channelID]
	if !ok {
		return false, nil
	}
	_, exists := members[userID]
	return exists, nil
}

func (m *mockRepository) GetMemberRole(_ context.Context, channelID, userID uuid.UUID) (string, error) {
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
	if !ok {
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

// --- Сообщения ---

func (m *mockRepository) InsertMessage(_ context.Context, doc *MessageDocument) (string, error) {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	m.messages[doc.ID] = doc
	return doc.ID, nil
}

func (m *mockRepository) FindMessageByID(_ context.Context, messageID string) (*MessageDocument, error) {
	doc, ok := m.messages[messageID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return doc, nil
}

func (m *mockRepository) ListMessages(_ context.Context, channelID uuid.UUID, _ *time.Time, limit int) ([]*MessageDocument, error) {
	var result []*MessageDocument
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

func (m *mockRepository) UpdateMessage(_ context.Context, messageID string, contentJSON map[string]interface{}, contentText string) (*MessageDocument, error) {
	doc, ok := m.messages[messageID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	now := time.Now()
	doc.ContentJSON = contentJSON
	doc.ContentText = contentText
	doc.EditedAt = &now
	return doc, nil
}

func (m *mockRepository) SoftDeleteMessage(_ context.Context, messageID string) error {
	doc, ok := m.messages[messageID]
	if !ok {
		return &ent.NotFoundError{}
	}
	now := time.Now()
	doc.DeletedAt = &now
	return nil
}

func (m *mockRepository) IncrementReplyCount(_ context.Context, messageID string) error {
	doc, ok := m.messages[messageID]
	if !ok {
		return nil
	}
	doc.ReplyCount++
	return nil
}

func (m *mockRepository) DecrementReplyCount(_ context.Context, messageID string) error {
	doc, ok := m.messages[messageID]
	if !ok {
		return nil
	}
	doc.ReplyCount--
	return nil
}

func (m *mockRepository) ListThreadReplies(_ context.Context, threadID string, _ *time.Time, limit int) ([]*MessageDocument, error) {
	var result []*MessageDocument
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
	doc, ok := m.messages[messageID]
	if !ok {
		return &ent.NotFoundError{}
	}
	for i, r := range doc.Reactions {
		if r.Emoji == emoji {
			doc.Reactions[i].UserIDs = append(doc.Reactions[i].UserIDs, userID)
			return nil
		}
	}
	doc.Reactions = append(doc.Reactions, ReactionEntry{Emoji: emoji, UserIDs: []string{userID}})
	return nil
}

func (m *mockRepository) RemoveReaction(_ context.Context, messageID, userID, emoji string) error {
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

func (m *mockRepository) SearchMessages(_ context.Context, query string, channelID *uuid.UUID, limit int) ([]*MessageDocument, error) {
	var result []*MessageDocument
	for _, doc := range m.messages {
		if doc.DeletedAt != nil {
			continue
		}
		if channelID != nil && doc.ChannelID != *channelID {
			continue
		}
		// Простой поиск по подстроке для тестов
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

func (m *mockRepository) ListPins(_ context.Context, _ uuid.UUID) ([]*ent.PinnedMessage, error) {
	return nil, nil
}

// --- Закладки ---

func (m *mockRepository) AddBookmark(_ context.Context, userID uuid.UUID, messageID string, channelID uuid.UUID) (*ent.Bookmark, error) {
	b := &ent.Bookmark{ID: uuid.New(), UserID: userID, MessageID: messageID, ChannelID: channelID, CreatedAt: time.Now()}
	m.bookmarks[userID] = append(m.bookmarks[userID], b)
	m.bookmarksByID[b.ID] = b
	return b, nil
}

func (m *mockRepository) RemoveBookmark(_ context.Context, bookmarkID uuid.UUID) error {
	delete(m.bookmarksByID, bookmarkID)
	return nil
}

func (m *mockRepository) ListBookmarks(_ context.Context, userID uuid.UUID) ([]*ent.Bookmark, error) {
	return m.bookmarks[userID], nil
}

func (m *mockRepository) EnsureIndexes(_ context.Context) error {
	return nil
}

// ---------------------------------------------------------------------------
// Mock NATS client
// ---------------------------------------------------------------------------

type mockNATSClient struct {
	published []publishedMessage
}

type publishedMessage struct {
	subject string
	data    []byte
}

func (m *mockNATSClient) Publish(subject string, data []byte) error {
	m.published = append(m.published, publishedMessage{subject: subject, data: data})
	return nil
}

// ---------------------------------------------------------------------------
// Helper: создаёт сервис с mock-зависимостями
// ---------------------------------------------------------------------------

func newTestService() (*MessengerService, *mockRepository, *mockNATSClient) {
	repo := newMockRepository()
	nats := &mockNATSClient{}
	svc := &MessengerService{
		repo: repo,
		nats: nil, // NATS client не используется напрямую в тестах
	}
	// Подменяем publishEvent чтобы не паниковать при nil nats
	// Для этого создаём сервис с фейковым NATS
	svc.nats = nil
	return svc, repo, nats
}

// setupChannelWithMembers создаёт канал с участниками для тестов.
func setupChannelWithMembers(repo *mockRepository, channelType string, isReadonly bool, ownerID uuid.UUID, memberIDs ...uuid.UUID) uuid.UUID {
	ch := &ent.Channel{
		ID:          uuid.New(),
		ChannelType: channel.ChannelType(channelType),
		IsReadonly:  isReadonly,
		CreatedBy:   ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.channels[ch.ID] = ch
	repo.members[ch.ID] = make(map[uuid.UUID]*ent.ChannelMember)
	repo.members[ch.ID][ownerID] = &ent.ChannelMember{
		ID: uuid.New(), ChannelID: ch.ID, UserID: ownerID,
		Role: channelmember.RoleOwner, JoinedAt: time.Now(),
	}
	for _, uid := range memberIDs {
		repo.members[ch.ID][uid] = &ent.ChannelMember{
			ID: uuid.New(), ChannelID: ch.ID, UserID: uid,
			Role: channelmember.RoleMember, JoinedAt: time.Now(),
		}
	}
	return ch.ID
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestCreateDM_Dedup проверяет дедупликацию DM-каналов.
func TestCreateDM_Dedup(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userA := uuid.New()
	userB := uuid.New()

	// Первый вызов создаёт DM
	ch1, err := svc.CreateOrGetDM(ctx, userA, userB)
	require.NoError(t, err)
	require.NotNil(t, ch1)

	// Второй вызов с теми же пользователями возвращает тот же канал
	ch2, err := svc.CreateOrGetDM(ctx, userB, userA)
	require.NoError(t, err)
	assert.Equal(t, ch1.ID, ch2.ID, "DM должен быть дедуплицирован")

	// DM с самим собой — ошибка
	_, err = svc.CreateOrGetDM(ctx, userA, userA)
	assert.Error(t, err, "DM с самим собой запрещён")
}

// TestSendMessage_MemberCheck проверяет, что не-участник не может отправить сообщение.
func TestSendMessage_MemberCheck(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	req := SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "hello"},
		ContentText: "hello",
	}

	// Не-участник не может отправить
	_, err := svc.SendMessage(ctx, channelID, outsiderID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "участником")

	// Участник может отправить
	resp, err := svc.SendMessage(ctx, channelID, ownerID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestSendMessage_ReadOnlyBlocked проверяет, что обычный участник не может писать в readonly канал.
func TestSendMessage_ReadOnlyBlocked(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", true, ownerID, memberID)

	req := SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "test"},
		ContentText: "test",
	}

	// Обычный участник не может писать в readonly канал
	_, err := svc.SendMessage(ctx, channelID, memberID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только для чтения")

	// Owner может писать в readonly канал
	resp, err := svc.SendMessage(ctx, channelID, ownerID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestEditMessage_AuthorOnly проверяет, что редактировать может только автор.
func TestEditMessage_AuthorOnly(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	authorID := uuid.New()
	otherID := uuid.New()
	channelID := uuid.New()

	// Создаём сообщение напрямую в repo
	doc := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    authorID,
		ContentJSON: map[string]interface{}{"text": "original"},
		ContentText: "original",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	editReq := EditMessageRequest{
		ContentJSON: map[string]interface{}{"text": "edited"},
		ContentText: "edited",
	}

	// Другой пользователь не может редактировать
	_, err := svc.EditMessage(ctx, doc.ID, otherID, editReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "свои сообщения")

	// Автор может редактировать
	resp, err := svc.EditMessage(ctx, doc.ID, authorID, editReq)
	require.NoError(t, err)
	assert.Equal(t, "edited", resp.ContentText)
	assert.NotNil(t, resp.EditedAt)
}

// TestDeleteMessage_AdminCanDeleteAny проверяет, что admin/owner может удалять чужие сообщения.
func TestDeleteMessage_AdminCanDeleteAny(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	// Создаём сообщение от member
	doc := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    memberID,
		ContentJSON: map[string]interface{}{"text": "member message"},
		ContentText: "member message",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	// Owner может удалить чужое сообщение (через channelID проверку)
	err := svc.DeleteMessage(ctx, doc.ID, ownerID, channelID.String(), false)
	require.NoError(t, err)

	// Проверяем soft delete
	assert.NotNil(t, doc.DeletedAt)
}

// TestToggleReaction_AddRemove проверяет toggle-логику реакций.
func TestToggleReaction_AddRemove(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := uuid.New()

	doc := &MessageDocument{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		SenderID:  uuid.New(),
		CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	// Добавляем реакцию
	added, err := svc.ToggleReaction(ctx, doc.ID, userID, "thumbsup")
	require.NoError(t, err)
	assert.True(t, added, "первый toggle должен добавить реакцию")

	// Удаляем реакцию (toggle)
	added, err = svc.ToggleReaction(ctx, doc.ID, userID, "thumbsup")
	require.NoError(t, err)
	assert.False(t, added, "второй toggle должен удалить реакцию")
}

// TestThreadReply_IncrementsReplyCount проверяет инкремент reply_count при ответе в тред.
func TestThreadReply_IncrementsReplyCount(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	// Отправляем основное сообщение
	mainMsg, err := svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "main"},
		ContentText: "main",
	})
	require.NoError(t, err)

	// Отправляем ответ в тред
	_, err = svc.SendMessage(ctx, channelID, ownerID, SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "reply"},
		ContentText: "reply",
		ThreadID:    mainMsg.ID,
	})
	require.NoError(t, err)

	// Проверяем reply_count
	parentDoc := repo.messages[mainMsg.ID]
	assert.Equal(t, 1, parentDoc.ReplyCount, "reply_count должен быть инкрементирован")
}

// TestJoinChannel_PublicOnly проверяет, что присоединиться можно только к публичному каналу.
func TestJoinChannel_PublicOnly(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()

	// Публичный канал — можно вступить
	publicID := setupChannelWithMembers(repo, "public", false, ownerID)
	err := svc.JoinChannel(ctx, publicID, joinerID)
	assert.NoError(t, err)

	// Проверяем, что joiner стал участником
	isMember, _ := repo.IsMember(ctx, publicID, joinerID)
	assert.True(t, isMember)

	// Приватный канал — нельзя вступить
	privateID := setupChannelWithMembers(repo, "private", false, ownerID)
	err = svc.JoinChannel(ctx, privateID, joinerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "приватный")
}

// TestForwardMessage_CreatesNewMessage проверяет, что пересылка создаёт новое сообщение.
func TestForwardMessage_CreatesNewMessage(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	sourceChannelID := setupChannelWithMembers(repo, "public", false, userID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, userID)

	// Создаём исходное сообщение
	doc := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   sourceChannelID,
		SenderID:    userID,
		ContentJSON: map[string]interface{}{"text": "forward me"},
		ContentText: "forward me",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	// Пересылаем
	resp, err := svc.ForwardMessage(ctx, sourceChannelID, userID, doc.ID, targetChannelID.String())
	require.NoError(t, err)
	assert.NotEqual(t, doc.ID, resp.ID, "пересланное сообщение должно иметь новый ID")
	assert.Equal(t, doc.ID, resp.ForwardedFromID, "должен быть forwarded_from_id")
	assert.Equal(t, targetChannelID.String(), resp.ChannelID, "сообщение в целевом канале")
}

// TestSearchMessages_TextQuery проверяет полнотекстовый поиск по сообщениям.
func TestSearchMessages_TextQuery(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	channelID := uuid.New()

	// Создаём сообщения
	for _, text := range []string{"hello world", "goodbye world", "hello there"} {
		doc := &MessageDocument{
			ID:          uuid.New().String(),
			ChannelID:   channelID,
			SenderID:    uuid.New(),
			ContentText: text,
			CreatedAt:   time.Now(),
		}
		repo.messages[doc.ID] = doc
	}

	// Поиск "hello"
	results, err := svc.SearchMessages(ctx, "hello", channelID.String(), 10)
	require.NoError(t, err)
	assert.Len(t, results, 2, "должны найти 2 сообщения с 'hello'")

	// Поиск "goodbye"
	results, err = svc.SearchMessages(ctx, "goodbye", channelID.String(), 10)
	require.NoError(t, err)
	assert.Len(t, results, 1, "должны найти 1 сообщение с 'goodbye'")
}

// TestDeleteMessage_ThreadReplyDecrementsCount проверяет декремент reply_count при удалении ответа в треде.
func TestDeleteMessage_ThreadReplyDecrementsCount(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	// Основное сообщение
	mainDoc := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "main",
		ReplyCount:  1,
		CreatedAt:   time.Now(),
	}
	repo.messages[mainDoc.ID] = mainDoc

	// Ответ в тред
	replyDoc := &MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "reply",
		ThreadID:    mainDoc.ID,
		CreatedAt:   time.Now(),
	}
	repo.messages[replyDoc.ID] = replyDoc

	// Удаляем ответ
	err := svc.DeleteMessage(ctx, replyDoc.ID, ownerID, channelID.String(), false)
	require.NoError(t, err)

	// reply_count должен уменьшиться
	assert.Equal(t, 0, mainDoc.ReplyCount, "reply_count должен быть декрементирован")
}

// ---------------------------------------------------------------------------
// Phase 03-02: Тесты управления каналами
// ---------------------------------------------------------------------------

// TestJoinChannel_PublicAllowed проверяет, что любой пользователь может присоединиться к публичному каналу.
func TestJoinChannel_PublicAllowed(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()

	publicID := setupChannelWithMembers(repo, "public", false, ownerID)
	err := svc.JoinChannel(ctx, publicID, joinerID)
	require.NoError(t, err)

	isMember, _ := repo.IsMember(ctx, publicID, joinerID)
	assert.True(t, isMember, "пользователь должен стать участником публичного канала")
}

// TestJoinChannel_PrivateBlocked проверяет, что нельзя присоединиться к приватному каналу без приглашения.
func TestJoinChannel_PrivateBlocked(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()

	privateID := setupChannelWithMembers(repo, "private", false, ownerID)
	err := svc.JoinChannel(ctx, privateID, joinerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "приватный")
}

// TestJoinChannel_DMBlocked проверяет, что нельзя присоединиться к DM через JoinChannel.
func TestJoinChannel_DMBlocked(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()

	dmID := setupChannelWithMembers(repo, "dm", false, ownerID)
	err := svc.JoinChannel(ctx, dmID, joinerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DM")
}

// TestSendMessage_ReadOnlyBlocked_RegularUser проверяет, что обычный участник не может писать в readonly канал.
func TestSendMessage_ReadOnlyBlocked_RegularUser(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", true, ownerID, memberID)

	req := SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "test"},
		ContentText: "test",
	}

	_, err := svc.SendMessage(ctx, channelID, memberID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "только для чтения")
}

// TestSendMessage_ReadOnlyAllowed_Admin проверяет, что owner/admin может писать в readonly канал.
func TestSendMessage_ReadOnlyAllowed_Admin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", true, ownerID, memberID)

	req := SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "admin message"},
		ContentText: "admin message",
	}

	// Owner может писать в readonly канал
	resp, err := svc.SendMessage(ctx, channelID, ownerID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestAddMembers_RequiresAdmin проверяет, что обычный участник не может добавлять других.
func TestAddMembers_RequiresAdmin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	newUserID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	// Обычный member не может добавлять участников
	err := svc.AddMembers(ctx, channelID, memberID, []string{newUserID.String()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner или admin")
}

// TestAddMembers_AdminAllowed проверяет, что admin/owner может добавлять участников.
func TestAddMembers_AdminAllowed(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	newUserID := uuid.New()

	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	// Owner может добавлять участников
	err := svc.AddMembers(ctx, channelID, ownerID, []string{newUserID.String()})
	require.NoError(t, err)

	// Проверяем, что новый участник добавлен
	isMember, _ := repo.IsMember(ctx, channelID, newUserID)
	assert.True(t, isMember, "новый участник должен быть добавлен")
}

// TestCreateChannel_WithProjectID проверяет создание канала с привязкой к проекту.
func TestCreateChannel_WithProjectID(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()

	resp, err := svc.CreateChannel(ctx, userID, CreateChannelRequest{
		ChannelType: "public",
		Name:        "#project-hub",
		ProjectID:   projectID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, projectID.String(), resp.ProjectID, "канал должен быть привязан к проекту")
	assert.Equal(t, "#project-hub", resp.Name)
}

// TestCreateChannel_ReadOnlyFlag проверяет создание канала с флагом is_readonly.
func TestCreateChannel_ReadOnlyFlag(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	resp, err := svc.CreateChannel(ctx, userID, CreateChannelRequest{
		ChannelType: "public",
		Name:        "#hub-ci",
		IsReadonly:  true,
	})
	require.NoError(t, err)
	assert.True(t, resp.IsReadonly, "канал должен быть readonly")
	assert.Equal(t, "#hub-ci", resp.Name)
}
