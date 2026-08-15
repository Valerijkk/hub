package calendar

import (
	"time"

	"github.com/google/uuid"
)

// EventClock задаёт инъецируемое время для канонических календарных команд.
type EventClock interface {
	Now() time.Time
}

// EventIDGenerator задаёт инъецируемую генерацию application event id.
// Идентификатор envelope не совпадает с transport id строки outbox.
type EventIDGenerator interface {
	NewEventID() uuid.UUID
}

// EventSerializer задаёт чистую сериализацию iTIP без I/O и скрытых часов.
type EventSerializer interface {
	Serialize(message ITIPMessage) ([]byte, error)
}

// EventService объединяет чистые зависимости канонического календарного домена.
type EventService struct {
	repository EventRepository
	clock      EventClock
	ids        EventIDGenerator
	serializer EventSerializer
}

// NewEventService создаёт сервис без обращения к repository, часам, генератору или сериализатору.
func NewEventService(
	repository EventRepository,
	clock EventClock,
	ids EventIDGenerator,
	serializer EventSerializer,
) *EventService {
	return &EventService{
		repository: repository,
		clock:      clock,
		ids:        ids,
		serializer: serializer,
	}
}
