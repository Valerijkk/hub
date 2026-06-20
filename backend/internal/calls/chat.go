package calls

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

type chatMessageStore interface {
	insert(ctx context.Context, msg *CallChatMessage) error
	listByCallID(ctx context.Context, callID string, limit int) ([]CallChatMessage, error)
	deleteByCallID(ctx context.Context, callID string) error
	ensureIndexes(ctx context.Context) error
}

// ChatRepository реализует доступ к эфемерному чату звонка.
type ChatRepository struct {
	store chatMessageStore
}

// NewChatRepository создаёт новый репозиторий чата звонков.
// Использует коллекцию "call_chat_messages" в указанной базе данных.
func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{
		store: newMongoChatStore(db.Collection(callChatCollection)),
	}
}

// EnsureIndexes создаёт составной индекс {call_id: 1, created_at: 1}
// для оптимизации выборки сообщений по звонку в хронологическом порядке.
func (r *ChatRepository) EnsureIndexes(ctx context.Context) error {
	return r.store.ensureIndexes(ctx)
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

	if err := r.store.insert(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

// ListMessages возвращает последние N сообщений чата звонка в хронологическом порядке.
// Сообщения сортируются по created_at DESC для выборки последних, затем реверсируются.
func (r *ChatRepository) ListMessages(ctx context.Context, callID string, limit int) ([]CallChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	results, err := r.store.listByCallID(ctx, callID, limit)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results, nil
}

// DeleteCallChat удаляет все сообщения чата указанного звонка.
// Вызывается при завершении звонка (room_finished) для обеспечения эфемерности чата.
func (r *ChatRepository) DeleteCallChat(ctx context.Context, callID string) error {
	return r.store.deleteByCallID(ctx, callID)
}
