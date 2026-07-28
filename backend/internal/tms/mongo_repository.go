package tms

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const allureResultsCollection = "allure_results"

// allureResultDocument — документ MongoDB для хранения сырого результата Allure.
type allureResultDocument struct {
	RunID      uuid.UUID    `bson:"run_id"`
	AllureUUID string       `bson:"allure_uuid"`
	Raw        AllureResult `bson:"raw"`
	CreatedAt  time.Time    `bson:"created_at"`
}

// MongoRepository реализует MongoDB-часть репозитория TMS.
// Отвечает за хранение сырых Allure JSON для аудита и отладки.
type MongoRepository struct {
	db *mongo.Database
}

// NewMongoRepository создаёт новый MongoDB-репозиторий TMS.
func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

// allureResults возвращает коллекцию allure_results.
func (r *MongoRepository) allureResults() *mongo.Collection {
	return r.db.Collection(allureResultsCollection)
}

// EnsureIndexes создаёт индексы для коллекции allure_results:
// 1. (run_id, created_at) — для выборки результатов прогона
// 2. (allure_uuid) — для поиска по UUID Allure
func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	coll := r.allureResults()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "run_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			Keys: bson.D{
				{Key: "allure_uuid", Value: 1},
			},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		slog.Error("не удалось создать MongoDB индексы TMS", "error", err)
		return fmt.Errorf("создание индексов allure_results: %w", err)
	}

	slog.Info("MongoDB индексы для allure_results созданы")
	return nil
}

// StoreAllureResults сохраняет массив результатов Allure в MongoDB.
// Использует InsertMany для пакетной вставки.
// A2W2-3-001: hard cap len(results).
func (r *MongoRepository) StoreAllureResults(ctx context.Context, runID uuid.UUID, results []AllureResult) error {
	if len(results) == 0 {
		return nil
	}
	if len(results) > maxMongoAllureResults {
		return fmt.Errorf("слишком много allure-результатов для Mongo (макс. %d)", maxMongoAllureResults)
	}

	now := time.Now()
	docs := make([]interface{}, 0, len(results))
	for _, result := range results {
		docs = append(docs, allureResultDocument{
			RunID:      runID,
			AllureUUID: result.UUID,
			Raw:        result,
			CreatedAt:  now,
		})
	}

	_, err := r.allureResults().InsertMany(ctx, docs)
	if err != nil {
		return fmt.Errorf("пакетная вставка allure результатов: %w", err)
	}
	return nil
}

// GetAllureResultsByRunID возвращает сырые результаты Allure для указанного прогона.
// A2W2-3-001: SetLimit против unbounded Mongo read.
func (r *MongoRepository) GetAllureResultsByRunID(ctx context.Context, runID uuid.UUID) ([]AllureResult, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(maxMongoAllureResults))

	cursor, err := r.allureResults().Find(ctx, bson.M{"run_id": runID}, opts)
	if err != nil {
		return nil, fmt.Errorf("поиск allure результатов по run_id: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []allureResultDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("десериализация allure результатов: %w", err)
	}

	results := make([]AllureResult, 0, len(docs))
	for _, doc := range docs {
		results = append(results, doc.Raw)
	}
	return results, nil
}

// GetAllureResultByUUID возвращает одиночный результат Allure по его UUID.
func (r *MongoRepository) GetAllureResultByUUID(ctx context.Context, allureUUID string) (*AllureResult, error) {
	var doc allureResultDocument
	err := r.allureResults().FindOne(ctx, bson.M{"allure_uuid": allureUUID}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("поиск allure результата по uuid: %w", err)
	}
	return &doc.Raw, nil
}
