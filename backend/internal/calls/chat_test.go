package calls

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// setupTestMongoDB создаёт подключение к тестовой MongoDB (integration).
// URL: MONGODB_TEST_URL или docker-compose localhost. Без Mongo — skip.
func setupTestMongoDB(t *testing.T) *mongo.Database {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uri := os.Getenv("MONGODB_TEST_URL")
	if uri == "" {
		uri = "mongodb://hub:hub@localhost:27017/?authSource=admin"
	}

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetConnectTimeout(2 * time.Second).
		SetServerSelectionTimeout(2 * time.Second))
	if err != nil {
		t.Skip("MongoDB недоступен, пропускаем тест:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("MongoDB не отвечает, пропускаем тест:", err)
	}

	dbName := "hub_test_calls_" + bson.NewObjectID().Hex()[:8]
	db := client.Database(dbName)

	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})

	return db
}

// TestSendMessage проверяет сохранение сообщения чата звонка в MongoDB.
// Сообщение должно получить ID и CreatedAt после вставки.
func TestSendMessage(t *testing.T) {
	db := setupTestMongoDB(t)
	repo := NewChatRepository(db)
	ctx := context.Background()

	msg, err := repo.SendMessage(ctx, "call-123", "user-1", "User One", "Привет!")
	require.NoError(t, err)
	require.NotNil(t, msg)

	assert.False(t, msg.ID.IsZero(), "ID должен быть заполнен")
	assert.Equal(t, "call-123", msg.CallID)
	assert.Equal(t, "user-1", msg.UserID)
	assert.Equal(t, "User One", msg.UserName)
	assert.Equal(t, "Привет!", msg.Text)
	assert.False(t, msg.CreatedAt.IsZero(), "CreatedAt должен быть заполнен")
}

// TestSendMessage_TextTooLong проверяет ограничение длины текста до 2000 символов.
func TestSendMessage_TextTooLong(t *testing.T) {
	db := setupTestMongoDB(t)
	repo := NewChatRepository(db)
	ctx := context.Background()

	longText := make([]byte, 2001)
	for i := range longText {
		longText[i] = 'a'
	}

	_, err := repo.SendMessage(ctx, "call-123", "user-1", "User", string(longText))
	assert.Error(t, err, "текст длиннее 2000 символов должен быть отклонён")
}

// TestListMessages проверяет получение сообщений в хронологическом порядке с учётом limit.
func TestListMessages(t *testing.T) {
	db := setupTestMongoDB(t)
	repo := NewChatRepository(db)
	ctx := context.Background()

	// Вставляем 3 сообщения с разным временем
	for i, text := range []string{"Первое", "Второе", "Третье"} {
		_, err := repo.SendMessage(ctx, "call-list", "user-1", "User", text)
		require.NoError(t, err)
		// Небольшая задержка чтобы CreatedAt различался
		if i < 2 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// Получаем все 3
	msgs, err := repo.ListMessages(ctx, "call-list", 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 3)
	// Проверяем хронологический порядок (первое сообщение первое)
	assert.Equal(t, "Первое", msgs[0].Text)
	assert.Equal(t, "Третье", msgs[2].Text)

	// Проверяем limit
	msgs, err = repo.ListMessages(ctx, "call-list", 2)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
}

// TestDeleteCallChat проверяет удаление всех сообщений чата звонка.
// После удаления ListMessages должен вернуть пустой список.
func TestDeleteCallChat(t *testing.T) {
	db := setupTestMongoDB(t)
	repo := NewChatRepository(db)
	ctx := context.Background()

	// Вставляем сообщения в два разных звонка
	_, err := repo.SendMessage(ctx, "call-del", "user-1", "User", "Msg 1")
	require.NoError(t, err)
	_, err = repo.SendMessage(ctx, "call-del", "user-2", "User2", "Msg 2")
	require.NoError(t, err)
	_, err = repo.SendMessage(ctx, "call-other", "user-1", "User", "Other call msg")
	require.NoError(t, err)

	// Удаляем сообщения call-del
	err = repo.DeleteCallChat(ctx, "call-del")
	require.NoError(t, err)

	// Проверяем: call-del пуст
	msgs, err := repo.ListMessages(ctx, "call-del", 10)
	require.NoError(t, err)
	assert.Empty(t, msgs, "после удаления чат должен быть пустым")

	// Проверяем: call-other не затронут
	msgs, err = repo.ListMessages(ctx, "call-other", 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 1, "другой звонок не должен быть затронут")
}

// TestEnsureIndexes проверяет создание индексов без ошибок.
func TestEnsureIndexes(t *testing.T) {
	db := setupTestMongoDB(t)
	repo := NewChatRepository(db)
	ctx := context.Background()

	err := repo.EnsureIndexes(ctx)
	require.NoError(t, err)

	// Повторный вызов тоже не должен падать
	err = repo.EnsureIndexes(ctx)
	require.NoError(t, err)
}
