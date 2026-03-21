package calls

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// callChatCollection — имя коллекции MongoDB для эфемерного чата звонков.
const callChatCollection = "call_chat_messages"

// CallChatMessage представляет сообщение эфемерного чата звонка.
// Сообщения хранятся в MongoDB и удаляются при завершении звонка (room_finished).
type CallChatMessage struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	CallID    string        `bson:"call_id" json:"call_id"`
	UserID    string        `bson:"user_id" json:"user_id"`
	UserName  string        `bson:"user_name" json:"user_name"`
	Text      string        `bson:"text" json:"text"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

// ChatRepository реализует доступ к эфемерному чату звонка через MongoDB.
// Коллекция call_chat_messages с составным индексом {call_id: 1, created_at: 1}.
type ChatRepository struct {
	collection *mongo.Collection
}

// NewChatRepository создаёт новый репозиторий чата звонков.
// Использует коллекцию "call_chat_messages" в указанной базе данных.
func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{
		collection: db.Collection(callChatCollection),
	}
}

// EnsureIndexes создаёт составной индекс {call_id: 1, created_at: 1}
// для оптимизации выборки сообщений по звонку в хронологическом порядке.
func (r *ChatRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "call_id", Value: 1},
				{Key: "created_at", Value: 1},
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("не удалось создать индексы чата звонков", "error", err)
		return err
	}

	slog.Info("MongoDB индексы для чата звонков созданы")
	return nil
}

// SendMessage сохраняет новое сообщение в чат звонка.
// Текст ограничен 2000 символами. Возвращает созданное сообщение с заполненным ID и CreatedAt.
func (r *ChatRepository) SendMessage(ctx context.Context, callID, userID, userName, text string) (*CallChatMessage, error) {
	if len([]rune(text)) > 2000 {
		return nil, fmt.Errorf("текст сообщения превышает 2000 символов")
	}

	msg := &CallChatMessage{
		ID:        bson.NewObjectID(),
		CallID:    callID,
		UserID:    userID,
		UserName:  userName,
		Text:      text,
		CreatedAt: time.Now(),
	}

	_, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить сообщение чата: %w", err)
	}

	return msg, nil
}

// ListMessages возвращает последние N сообщений чата звонка в хронологическом порядке.
// Сообщения сортируются по created_at DESC для выборки последних, затем реверсируются.
func (r *ChatRepository) ListMessages(ctx context.Context, callID string, limit int) ([]CallChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	filter := bson.M{"call_id": callID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сообщения чата: %w", err)
	}
	defer cursor.Close(ctx)

	var results []CallChatMessage
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("не удалось декодировать сообщения чата: %w", err)
	}

	// Реверсируем для хронологического порядка
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results, nil
}

// DeleteCallChat удаляет все сообщения чата указанного звонка.
// Вызывается при завершении звонка (room_finished) для обеспечения эфемерности чата.
func (r *ChatRepository) DeleteCallChat(ctx context.Context, callID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"call_id": callID})
	if err != nil {
		return fmt.Errorf("не удалось удалить чат звонка: %w", err)
	}
	return nil
}
