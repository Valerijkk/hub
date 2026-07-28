package messenger

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// messagesCollection — имя коллекции MongoDB для хранения сообщений.
const messagesCollection = "messages"

// MongoRepository реализует MongoDB-часть интерфейса Repository.
// Отвечает за хранение и поиск сообщений.
type MongoRepository struct {
	db *mongo.Database
}

// NewMongoRepository создаёт новый MongoDB-репозиторий.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

// messages возвращает коллекцию сообщений.
func (r *MongoRepository) messages() *mongo.Collection {
	return r.db.Collection(messagesCollection)
}

// EnsureIndexes создаёт составные индексы для коллекции сообщений:
// 1. (channel_id, created_at DESC) — для постраничной загрузки
// 2. (channel_id, thread_id, created_at DESC) — для тредов
// 3. text index на content_text — для полнотекстового поиска
func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	coll := r.messages()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "channel_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "channel_id", Value: 1},
				{Key: "thread_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "content_text", Value: "text"},
			},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("не удалось создать MongoDB индексы", "error", err)
		return err
	}

	slog.Info("MongoDB индексы для сообщений созданы")
	return nil
}

// InsertMessage сохраняет сообщение в MongoDB и возвращает сгенерированный ID.
func (r *MongoRepository) InsertMessage(ctx context.Context, doc *MessageDocument) (string, error) {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	_, err := r.messages().InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	return doc.ID, nil
}

// FindMessageByID возвращает сообщение по ID.
func (r *MongoRepository) FindMessageByID(ctx context.Context, messageID string) (*MessageDocument, error) {
	var doc MessageDocument
	err := r.messages().FindOne(ctx, bson.M{"_id": messageID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListMessages возвращает сообщения канала с курсорной пагинацией по created_at.
func (r *MongoRepository) ListMessages(ctx context.Context, channelID uuid.UUID, before *time.Time, limit int) ([]*MessageDocument, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	filter := bson.M{
		"channel_id": channelID,
		"thread_id":  bson.M{"$exists": false},
		"deleted_at": nil,
	}
	if before != nil {
		filter["created_at"] = bson.M{"$lt": *before}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.messages().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*MessageDocument
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// UpdateMessage обновляет контент сообщения и устанавливает edited_at.
func (r *MongoRepository) UpdateMessage(ctx context.Context, messageID string, contentJSON map[string]interface{}, contentText string) (*MessageDocument, error) {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"content_json": contentJSON,
			"content_text": contentText,
			"edited_at":    now,
		},
	}

	_, err := r.messages().UpdateByID(ctx, messageID, update)
	if err != nil {
		return nil, err
	}

	return r.FindMessageByID(ctx, messageID)
}

// SoftDeleteMessage выполняет мягкое удаление, устанавливая deleted_at.
func (r *MongoRepository) SoftDeleteMessage(ctx context.Context, messageID string) error {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": now,
		},
	}
	_, err := r.messages().UpdateByID(ctx, messageID, update)
	return err
}

// IncrementReplyCount увеличивает счётчик ответов на 1.
func (r *MongoRepository) IncrementReplyCount(ctx context.Context, messageID string) error {
	update := bson.M{
		"$inc": bson.M{"reply_count": 1},
	}
	_, err := r.messages().UpdateByID(ctx, messageID, update)
	return err
}

// DecrementReplyCount уменьшает счётчик ответов на 1.
func (r *MongoRepository) DecrementReplyCount(ctx context.Context, messageID string) error {
	update := bson.M{
		"$inc": bson.M{"reply_count": -1},
	}
	_, err := r.messages().UpdateByID(ctx, messageID, update)
	return err
}

// ListThreadReplies возвращает ответы в треде с курсорной пагинацией.
func (r *MongoRepository) ListThreadReplies(ctx context.Context, threadID string, before *time.Time, limit int) ([]*MessageDocument, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	filter := bson.M{
		"thread_id":  threadID,
		"deleted_at": nil,
	}
	if before != nil {
		filter["created_at"] = bson.M{"$lt": *before}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.messages().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*MessageDocument
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// AddReaction добавляет реакцию пользователя к сообщению.
func (r *MongoRepository) AddReaction(ctx context.Context, messageID, userID, emoji string) error {
	// Проверяем, существует ли уже запись для данного emoji
	filter := bson.M{
		"_id":             messageID,
		"reactions.emoji": emoji,
	}
	update := bson.M{
		"$addToSet": bson.M{
			"reactions.$.user_ids": userID,
		},
	}

	result, err := r.messages().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// MatchedCount>0: emoji уже есть — $addToSet либо добавил userID, либо no-op
	// при повторном клике (ModifiedCount==0). Нельзя $push второй ReactionEntry.
	if result.MatchedCount > 0 {
		return nil
	}

	// Emoji ещё нет — создаём новую запись реакции.
	filter = bson.M{"_id": messageID}
	update = bson.M{
		"$push": bson.M{
			"reactions": ReactionEntry{
				Emoji:   emoji,
				UserIDs: []string{userID},
			},
		},
	}
	_, err = r.messages().UpdateOne(ctx, filter, update)
	return err
}

// RemoveReaction удаляет реакцию пользователя с сообщения.
func (r *MongoRepository) RemoveReaction(ctx context.Context, messageID, userID, emoji string) error {
	filter := bson.M{
		"_id":             messageID,
		"reactions.emoji": emoji,
	}
	update := bson.M{
		"$pull": bson.M{
			"reactions.$.user_ids": userID,
		},
	}
	_, err := r.messages().UpdateOne(ctx, filter, update)
	return err
}

// FindReaction проверяет, поставил ли пользователь указанную реакцию.
func (r *MongoRepository) FindReaction(ctx context.Context, messageID, userID, emoji string) (bool, error) {
	filter := bson.M{
		"_id":             messageID,
		"reactions.emoji": emoji,
		"reactions": bson.M{
			"$elemMatch": bson.M{
				"emoji":    emoji,
				"user_ids": userID,
			},
		},
	}

	count, err := r.messages().CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SearchMessages выполняет полнотекстовый поиск по content_text в указанных каналах.
func (r *MongoRepository) SearchMessages(ctx context.Context, query string, channelIDs []uuid.UUID, limit int) ([]*MessageDocument, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if len(channelIDs) == 0 {
		return []*MessageDocument{}, nil
	}

	filter := bson.M{
		"$text":       bson.M{"$search": query},
		"deleted_at":  nil,
		"channel_id":  bson.M{"$in": channelIDs},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.messages().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*MessageDocument
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
