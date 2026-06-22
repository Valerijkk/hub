package tms

import (
	"github.com/valerii/hub/backend/ent/schema"
)

// CombinedRepository объединяет EntRepository (PostgreSQL) и MongoRepository (MongoDB)
// в единую реализацию интерфейса Repository.
// Go struct embedding автоматически промотирует все методы обоих репозиториев,
// устраняя необходимость в ручных делегирующих обёртках.
type CombinedRepository struct {
	*EntRepository   // Тест-кейсы, сьюты, прогоны, результаты, компоненты (PostgreSQL)
	*MongoRepository // Allure-результаты (MongoDB)
}

// NewCombinedRepository создаёт комбинированный репозиторий из Ent и MongoDB частей.
func NewCombinedRepository(entRepo *EntRepository, mongoRepo *MongoRepository) *CombinedRepository {
	return &CombinedRepository{EntRepository: entRepo, MongoRepository: mongoRepo}
}

// Подавляем предупреждение о неиспользуемом импорте schema.
var _ = schema.TestStep{}
