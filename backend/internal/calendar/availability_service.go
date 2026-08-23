package calendar

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// AvailabilityQuery is the bounded, actor-scoped request passed from the
// transport into the private availability source repository.
type AvailabilityQuery struct {
	UserIDs     []uuid.UUID
	From        time.Time
	To          time.Time
	Timezone    string
	SlotMinutes int
}

// AvailabilityService composes the authorized source repository with the pure
// busy/free normalization policy. It never returns a partial result.
type AvailabilityService struct {
	repository AvailabilitySourceRepository
}

// NewAvailabilityService creates the availability application service.
func NewAvailabilityService(repository AvailabilitySourceRepository) *AvailabilityService {
	return &AvailabilityService{repository: repository}
}

// QueryAvailability resolves the complete authorized user set before exposing
// its privacy-preserving busy/free result.
func (service *AvailabilityService) QueryAvailability(ctx context.Context, actorID uuid.UUID, query AvailabilityQuery) (AvailabilityResult, error) {
	if !service.ready() || actorID == uuid.Nil {
		return AvailabilityResult{}, errAvailabilityUnavailable
	}

	sourceQuery, _, err := validateAvailabilityQuery(query)
	if err != nil {
		return AvailabilityResult{}, err
	}
	users, err := service.repository.LoadAvailability(ctx, actorID, sourceQuery)
	if err != nil {
		return AvailabilityResult{}, err
	}
	if !availabilityUsersMatchQuery(sourceQuery.UserIDs, users) {
		return AvailabilityResult{}, ErrAvailabilitySourceInvalid
	}
	return ResolveAvailabilityPolicy(AvailabilityRequest{
		From:        sourceQuery.From,
		To:          sourceQuery.To,
		SlotMinutes: query.SlotMinutes,
		Users:       users,
	})
}

func validateAvailabilityQuery(query AvailabilityQuery) (AvailabilitySourceQuery, *time.Location, error) {
	sourceQuery, err := normalizeAvailabilitySourceQuery(AvailabilitySourceQuery{
		From:    query.From,
		To:      query.To,
		UserIDs: append([]uuid.UUID(nil), query.UserIDs...),
	})
	if err != nil {
		return AvailabilitySourceQuery{}, nil, err
	}
	if query.SlotMinutes < AvailabilityMinSlotMinutes || query.SlotMinutes > AvailabilityMaxSlotMinutes {
		return AvailabilitySourceQuery{}, nil, ErrAvailabilitySourceInvalid
	}
	location, err := availabilityResponseLocation(query.Timezone)
	if err != nil {
		return AvailabilitySourceQuery{}, nil, err
	}
	return sourceQuery, location, nil
}

func availabilityResponseLocation(timezone string) (*time.Location, error) {
	if timezone == "" || timezone == "Local" {
		return nil, ErrAvailabilitySourceInvalid
	}
	location, err := time.LoadLocation(timezone)
	if err != nil || location == time.Local || location.String() != timezone {
		return nil, ErrAvailabilitySourceInvalid
	}
	return location, nil
}

func availabilityUsersMatchQuery(userIDs []uuid.UUID, users []AvailabilityUserInput) bool {
	if len(userIDs) != len(users) {
		return false
	}
	for index, userID := range userIDs {
		if users[index].UserID != userID {
			return false
		}
	}
	return true
}

func (service *AvailabilityService) ready() bool {
	return service != nil && !nilAvailabilitySourceRepository(service.repository)
}

func nilAvailabilitySourceRepository(repository AvailabilitySourceRepository) bool {
	if repository == nil {
		return true
	}
	value := reflect.ValueOf(repository)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
