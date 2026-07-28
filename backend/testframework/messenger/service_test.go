package messenger_test

// Тесты сервиса мессенджера: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/messenger"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


// ---------------------------------------------------------------------------
// Мок NATS-клиента
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
// Вспомогательная функция: создаёт сервис с mock-зависимостями.
// ---------------------------------------------------------------------------

func newTestService() (*messenger.MessengerService, *mockRepository, *mockNATSClient) {
	repo := newMockRepository()
	nats := &mockNATSClient{}
	svc := messenger.ExportNewMessengerServiceTest(repo)
	return svc, repo, nats
}

// ---------------------------------------------------------------------------
// Тесты
// ---------------------------------------------------------------------------

// TestCreateDM_Dedup проверяет сценарий: создание dm дедупликация.
func TestCreateDM_Dedup(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userA := uuid.New()
	userB := uuid.New()

	// Первый вызов создаёт DM.
	ch1, err := svc.CreateOrGetDM(ctx, userA, userB)
	require.NoError(t, err)
	require.NotNil(t, ch1)

	// Повторный вызов с теми же участниками возвращает тот же канал.
	ch2, err := svc.CreateOrGetDM(ctx, userB, userA)
	require.NoError(t, err)
	assert.Equal(t, ch1.ID, ch2.ID, "DM должен быть дедуплицирован")

	// DM с самим собой — ошибка.
	_, err = svc.CreateOrGetDM(ctx, userA, userA)
	assert.Error(t, err, "DM с самим собой запрещён")
}

// TestSendMessage_MemberCheck проверяет сценарий: отправка message участник проверка.
func TestSendMessage_MemberCheck(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	outsiderID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "hello"},
		ContentText: "hello",
	}

	// Не-участник не может отправить сообщение.
	_, err := svc.SendMessage(ctx, channelID, outsiderID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "участником")

	// Участник канала может отправить сообщение.
	resp, err := svc.SendMessage(ctx, channelID, ownerID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestSendMessage_ReadOnlyBlocked проверяет сценарий: отправка message read only blocked.
func TestSendMessage_ReadOnlyBlocked(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", true, ownerID, memberID)

	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "test"},
		ContentText: "test",
	}

	// Обычный участник не может писать в readonly-канал.
	_, err := svc.SendMessage(ctx, channelID, memberID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "только для чтения")

	// Обычный участник не может писать в readonly-канал.
	resp, err := svc.SendMessage(ctx, channelID, ownerID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestEditMessage_AuthorOnly проверяет сценарий: edit message author only.
func TestEditMessage_AuthorOnly(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	authorID := uuid.New()
	otherID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, authorID, otherID)

	doc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    authorID,
		ContentJSON: map[string]interface{}{"text": "original"},
		ContentText: "original",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	editReq := messenger.EditMessageRequest{
		ContentJSON: map[string]interface{}{"text": "edited"},
		ContentText: "edited",
	}

	_, err := svc.EditMessage(ctx, channelID, doc.ID, otherID, editReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "свои сообщения")

	resp, err := svc.EditMessage(ctx, channelID, doc.ID, authorID, editReq)
	require.NoError(t, err)
	assert.Equal(t, "edited", resp.ContentText)
	assert.NotNil(t, resp.EditedAt)
}

// TestDeleteMessage_AdminCanDeleteAny проверяет сценарий: удаление message админ can удаление any.
func TestDeleteMessage_AdminCanDeleteAny(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	// Нельзя редактировать чужое сообщение.
	doc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    memberID,
		ContentJSON: map[string]interface{}{"text": "member message"},
		ContentText: "member message",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	// После edit version увеличивается.
	err := svc.DeleteMessage(ctx, doc.ID, ownerID, channelID.String(), false)
	require.NoError(t, err)

	// Проверяем soft delete канала.
	assert.NotNil(t, doc.DeletedAt)
}

// TestToggleReaction_AddRemove проверяет сценарий: toggle reaction add remove.
func TestToggleReaction_AddRemove(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	doc := &messenger.MessageDocument{
		ID:        uuid.New().String(),
		ChannelID: channelID,
		SenderID:  uuid.New(),
		CreatedAt: time.Now(),
	}
	repo.messages[doc.ID] = doc

	added, err := svc.ToggleReaction(ctx, channelID, doc.ID, userID, "thumbsup")
	require.NoError(t, err)
	assert.True(t, added, "первый toggle должен добавить реакцию")

	added, err = svc.ToggleReaction(ctx, channelID, doc.ID, userID, "thumbsup")
	require.NoError(t, err)
	assert.False(t, added, "второй toggle должен удалить реакцию")
}

// TestThreadReply_IncrementsReplyCount проверяет сценарий: thread reply increments reply count.
func TestThreadReply_IncrementsReplyCount(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	// Проверяем soft delete канала.
	mainMsg, err := svc.SendMessage(ctx, channelID, ownerID, messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "main"},
		ContentText: "main",
	})
	require.NoError(t, err)

	// Добавляем реакцию на сообщение.
	_, err = svc.SendMessage(ctx, channelID, ownerID, messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "reply"},
		ContentText: "reply",
		ThreadID:    mainMsg.ID,
	})
	require.NoError(t, err)

	// Проверяем инкремент reply_count у родителя.
	parentDoc := repo.messages[mainMsg.ID]
	assert.Equal(t, 1, parentDoc.ReplyCount, "reply_count должен быть инкрементирован")
}

// TestJoinChannel_PublicOnly проверяет сценарий: join channel public only.
func TestJoinChannel_PublicOnly(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	joinerID := uuid.New()

	// Повторный toggle снимает реакцию.
	publicID := setupChannelWithMembers(repo, "public", false, ownerID)
	err := svc.JoinChannel(ctx, publicID, joinerID)
	assert.NoError(t, err)

	// Проверяем, что joiner стал участником публичного канала.
	isMember, _ := repo.IsMember(ctx, publicID, joinerID)
	assert.True(t, isMember)

	// Публикуем родительское сообщение.
	privateID := setupChannelWithMembers(repo, "private", false, ownerID)
	err = svc.JoinChannel(ctx, privateID, joinerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "приватный")
}

// TestForwardMessage_CreatesNewMessage проверяет сценарий: forward message creates new message.
func TestForwardMessage_CreatesNewMessage(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	sourceChannelID := setupChannelWithMembers(repo, "public", false, userID)
	targetChannelID := setupChannelWithMembers(repo, "public", false, userID)

	// Публикуем ответ в тред.
	doc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   sourceChannelID,
		SenderID:    userID,
		ContentJSON: map[string]interface{}{"text": "forward me"},
		ContentText: "forward me",
		CreatedAt:   time.Now(),
	}
	repo.messages[doc.ID] = doc

	// Проверяем инкремент reply_count.
	resp, err := svc.ForwardMessage(ctx, sourceChannelID, userID, doc.ID, targetChannelID.String())
	require.NoError(t, err)
	assert.NotEqual(t, doc.ID, resp.ID, "пересланное сообщение должно иметь новый ID")
	assert.Equal(t, doc.ID, resp.ForwardedFromID, "должен быть forwarded_from_id")
	assert.Equal(t, targetChannelID.String(), resp.ChannelID, "сообщение в целевом канале")
}

// TestSearchMessages_TextQuery проверяет сценарий: search сообщения text query.
func TestSearchMessages_TextQuery(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, userID)

	// Создаём публичный канал.
	for _, text := range []string{"hello world", "goodbye world", "hello there"} {
		doc := &messenger.MessageDocument{
			ID:          uuid.New().String(),
			ChannelID:   channelID,
			SenderID:    userID,
			ContentText: text,
			CreatedAt:   time.Now(),
		}
		repo.messages[doc.ID] = doc
	}

	// Поиск по тексту «hello».
	results, err := svc.SearchMessages(ctx, userID, "hello", channelID.String(), 10)
	require.NoError(t, err)
	assert.Len(t, results, 2, "должны найти 2 сообщения с 'hello'")

	// Поиск по тексту «goodbye».
	results, err = svc.SearchMessages(ctx, userID, "goodbye", channelID.String(), 10)
	require.NoError(t, err)
	assert.Len(t, results, 1, "должны найти 1 сообщение с 'goodbye'")
}

// TestDeleteMessage_ThreadReplyDecrementsCount проверяет сценарий: удаление message thread reply decrements count.
func TestDeleteMessage_ThreadReplyDecrementsCount(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	channelID := setupChannelWithMembers(repo, "public", false, ownerID)

	// Присоединяемся к публичному каналу.
	mainDoc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "main",
		ReplyCount:  1,
		CreatedAt:   time.Now(),
	}
	repo.messages[mainDoc.ID] = mainDoc

	// Проверяем, что joiner стал участником.
	replyDoc := &messenger.MessageDocument{
		ID:          uuid.New().String(),
		ChannelID:   channelID,
		SenderID:    ownerID,
		ContentText: "reply",
		ThreadID:    mainDoc.ID,
		CreatedAt:   time.Now(),
	}
	repo.messages[replyDoc.ID] = replyDoc

	// Нельзя присоединиться к приватному каналу.
	err := svc.DeleteMessage(ctx, replyDoc.ID, ownerID, channelID.String(), false)
	require.NoError(t, err)

	// Проверяем инкремент reply_count у родителя.
	assert.Equal(t, 0, mainDoc.ReplyCount, "reply_count должен быть декрементирован")
}

// ---------------------------------------------------------------------------
// Вспомогательная функция: создаёт сервис с mock-зависимостями.
// ---------------------------------------------------------------------------

// TestJoinChannel_PublicAllowed проверяет сценарий: join channel public allowed.
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

// TestJoinChannel_PrivateBlocked проверяет сценарий: join channel private blocked.
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

// TestJoinChannel_DMBlocked проверяет сценарий: join channel dm blocked.
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

// TestSendMessage_ReadOnlyBlocked_RegularUser проверяет сценарий: отправка message read only blocked regular user.
func TestSendMessage_ReadOnlyBlocked_RegularUser(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", true, ownerID, memberID)

	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "test"},
		ContentText: "test",
	}

	_, err := svc.SendMessage(ctx, channelID, memberID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "только для чтения")
}

// TestSendMessage_ReadOnlyAllowed_Admin проверяет сценарий: отправка message read only allowed админ.
func TestSendMessage_ReadOnlyAllowed_Admin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", true, ownerID, memberID)

	req := messenger.SendMessageRequest{
		ContentJSON: map[string]interface{}{"text": "admin message"},
		ContentText: "admin message",
	}

	// Обычный участник не может писать в readonly-канал.
	resp, err := svc.SendMessage(ctx, channelID, ownerID, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
}

// TestAddMembers_RequiresAdmin проверяет сценарий: add members requires админ.
func TestAddMembers_RequiresAdmin(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()
	newUserID := uuid.New()

	channelID := setupChannelWithMembers(repo, "public", false, ownerID, memberID)

	// Создаём сообщение для пересылки.
	err := svc.AddMembers(ctx, channelID, memberID, []string{newUserID.String()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "owner или admin")
}

// TestAddMembers_AdminAllowed проверяет сценарий: add members админ allowed.
func TestAddMembers_AdminAllowed(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	newUserID := uuid.New()

	channelID := setupChannelWithMembers(repo, "private", false, ownerID)

	// Пересылаем сообщение в целевой канал.
	err := svc.AddMembers(ctx, channelID, ownerID, []string{newUserID.String()})
	require.NoError(t, err)

	// Создаём сообщения для поиска.
	isMember, _ := repo.IsMember(ctx, channelID, newUserID)
	assert.True(t, isMember, "новый участник должен быть добавлен")
}

// TestCreateChannel_WithProjectID проверяет сценарий: создание channel with project id.
func TestCreateChannel_WithProjectID(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()

	resp, err := svc.CreateChannel(ctx, userID, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "#project-hub",
		ProjectID:   projectID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, projectID.String(), resp.ProjectID, "канал должен быть привязан к проекту")
	assert.Equal(t, "#project-hub", resp.Name)
}

// TestCreateChannel_ReadOnlyFlag проверяет сценарий: создание channel read only flag.
func TestCreateChannel_ReadOnlyFlag(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	userID := uuid.New()

	resp, err := svc.CreateChannel(ctx, userID, messenger.CreateChannelRequest{
		ChannelType: "public",
		Name:        "#hub-ci",
		IsReadonly:  true,
	})
	require.NoError(t, err)
	assert.True(t, resp.IsReadonly, "канал должен быть readonly")
	assert.Equal(t, "#hub-ci", resp.Name)
}
