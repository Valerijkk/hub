// Package mongotest подключается к MongoDB для интеграционных тестов.
// Пропускает тест, когда MONGODB_TEST_URL не задан и localhost:27017 недоступен.
package mongotest

// Тесты модуля mongotest: интеграционные и unit-сценарии testframework.

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DefaultURI используется, когда MONGODB_TEST_URL пуст.
const DefaultURI = "mongodb://hub:hub@localhost:27017/?authSource=admin"

// Connect открывает эфемерную тестовую БД; удаляет её при очистке.
func Connect(t *testing.T, namePrefix string) *mongo.Database {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uri := os.Getenv("MONGODB_TEST_URL")
	if uri == "" {
		uri = DefaultURI
	}

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri).
		SetConnectTimeout(3*time.Second).
		SetServerSelectionTimeout(3*time.Second))
	if err != nil {
		t.Skip("MongoDB недоступен, пропускаем тест:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		t.Skip("MongoDB не отвечает, пропускаем тест:", err)
	}

	dbName := namePrefix + bson.NewObjectID().Hex()[:8]
	db := client.Database(dbName)

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_ = db.Drop(cleanupCtx)
		_ = client.Disconnect(cleanupCtx)
	})

	return db
}
