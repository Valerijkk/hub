package messenger

import (
	"context"
	"errors"
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
		{
			Keys: bson.D{
				{Key: "sender_id", Value: 1},
				{Key: "client_nonce", Value: 1},
			},
			Options: options.Index().
				SetName("messages_sender_client_nonce_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.D{{
					Key: "$and",
					Value: bson.A{
						bson.D{{Key: "client_nonce", Value: bson.D{{Key: "$type", Value: "string"}}}},
						bson.D{{Key: "client_nonce", Value: bson.D{{Key: "$gt", Value: ""}}}},
					},
				}}),
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

// CreateMessageIdempotent atomically creates a message or returns its canonical nonce record.
func (r *MongoRepository) CreateMessageIdempotent(ctx context.Context, doc *MessageDocument) (*CreateMessageResult, error) {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	if doc.ClientNonce == "" {
		if _, err := r.messages().InsertOne(ctx, doc); err != nil {
			return nil, err
		}
		return &CreateMessageResult{Message: doc, Created: true}, nil
	}

	filter := bson.M{
		"sender_id":    doc.SenderID,
		"client_nonce": doc.ClientNonce,
	}
	update := bson.M{"$setOnInsert": doc}
	result, err := r.messages().UpdateOne(ctx, filter, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return nil, err
		}

		canonical, findErr := r.FindMessageByClientNonce(ctx, doc.SenderID, doc.ClientNonce)
		if findErr != nil {
			return nil, findErr
		}
		return &CreateMessageResult{Message: canonical, Created: false}, nil
	}

	canonical, err := r.FindMessageByClientNonce(ctx, doc.SenderID, doc.ClientNonce)
	if err != nil {
		return nil, err
	}
	return &CreateMessageResult{Message: canonical, Created: result.UpsertedCount == 1}, nil
}

// FindMessageByClientNonce intentionally includes soft-deleted messages because the nonce remains reserved.
func (r *MongoRepository) FindMessageByClientNonce(ctx context.Context, senderID uuid.UUID, clientNonce string) (*MessageDocument, error) {
	var doc MessageDocument
	err := r.messages().FindOne(ctx, bson.M{
		"sender_id":    senderID,
		"client_nonce": clientNonce,
	}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// FindMessageByID возвращает сообщение по ID.
// AUDIT-TM-003: только живые (deleted_at nil) — иначе Edit/React/Pin мутируют trash.
func (r *MongoRepository) FindMessageByID(ctx context.Context, messageID string) (*MessageDocument, error) {
	var doc MessageDocument
	err := r.messages().FindOne(ctx, bson.M{
		"_id":        messageID,
		"deleted_at": nil,
	}).Decode(&doc)
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
// AUDIT-TM-003: $set только при deleted_at nil — soft-deleted не редактируются.
func (r *MongoRepository) UpdateMessage(ctx context.Context, messageID string, contentJSON map[string]interface{}, contentText string) (*MessageDocument, error) {
	now := time.Now()
	filter := bson.M{
		"_id":        messageID,
		"deleted_at": nil,
	}
	update := bson.M{
		"$set": bson.M{
			"content_json": contentJSON,
			"content_text": contentText,
			"edited_at":    now,
		},
	}

	res, err := r.messages().UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, mongo.ErrNoDocuments
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
// A2W3-3-004: после MatchedCount-фикса concurrent first-react всё ещё TOCTOU —
// оба видят emoji absent и оба $push. Push только при "reactions.emoji" $ne;
// проигравший гонку добивает userID через $addToSet.
func (r *MongoRepository) AddReaction(ctx context.Context, messageID, userID, emoji string) error {
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

	// Emoji отсутствует — атомарный push только пока emoji ещё нет.
	filter = bson.M{
		"_id":             messageID,
		"reactions.emoji": bson.M{"$ne": emoji},
	}
	update = bson.M{
		"$push": bson.M{
			"reactions": ReactionEntry{
				Emoji:   emoji,
				UserIDs: []string{userID},
			},
		},
	}
	result, err = r.messages().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount > 0 {
		return nil
	}

	// Гонка: peer уже создал ReactionEntry — добавляем userID через $addToSet.
	filter = bson.M{
		"_id":             messageID,
		"reactions.emoji": emoji,
	}
	update = bson.M{
		"$addToSet": bson.M{
			"reactions.$.user_ids": userID,
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
		"$text":      bson.M{"$search": query},
		"deleted_at": nil,
		"channel_id": bson.M{"$in": channelIDs},
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
