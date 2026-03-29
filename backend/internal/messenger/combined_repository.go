package messenger

// CombinedRepository объединяет EntRepository и MongoRepository
// в единую реализацию интерфейса Repository.
// Go struct embedding автоматически промотирует все методы обоих репозиториев,
// устраняя необходимость в ручных делегирующих обёртках.
type CombinedRepository struct {
	*EntRepository   // Каналы, участники, pins, bookmarks, star/mute (PostgreSQL)
	*MongoRepository // Сообщения, реакции, поиск, индексы (MongoDB)
}

// NewCombinedRepository создаёт комбинированный репозиторий из Ent и MongoDB частей.
func NewCombinedRepository(entRepo *EntRepository, mongoRepo *MongoRepository) *CombinedRepository {
	return &CombinedRepository{EntRepository: entRepo, MongoRepository: mongoRepo}
}
