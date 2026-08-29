package calendar

import (
	"time"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v3"
)

type calendarHTTPAvailabilityCheckRequest struct {
	AttendeeUserIDs calendarHTTPField[[]uuid.UUID] `json:"attendee_user_ids"`
	From            calendarHTTPField[time.Time]   `json:"from"`
	To              calendarHTTPField[time.Time]   `json:"to"`
	Timezone        calendarHTTPField[string]      `json:"timezone"`
}

type calendarHTTPAvailabilityCheckResponse struct {
	AllAvailable bool `json:"all_available"`
}

func (handler *AvailabilityHandler) CheckAvailability(c fiber.Ctx) error {
	if !handler.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	var request calendarHTTPAvailabilityCheckRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return calendarHTTPAvailabilityCheckBadRequest()
	}
	query, err := request.queryForActor(actorID)
	if err != nil {
		return calendarHTTPAvailabilityCheckBadRequest()
	}
	result, err := handler.service.QueryAvailability(c.Context(), actorID, query)
	if err != nil {
		return calendarHTTPAvailabilityError(c, err)
	}
	allAvailable := true
	for _, user := range result.Users {
		if len(user.Busy) != 0 {
			allAvailable = false
			break
		}
	}
	return c.Status(fiber.StatusOK).JSON(calendarHTTPAvailabilityCheckResponse{AllAvailable: allAvailable})
}

func (request calendarHTTPAvailabilityCheckRequest) queryForActor(actorID uuid.UUID) (AvailabilityQuery, error) {
	if actorID == uuid.Nil ||
		!request.AttendeeUserIDs.present || request.AttendeeUserIDs.null ||
		!request.From.present || request.From.null ||
		!request.To.present || request.To.null ||
		!request.Timezone.present || request.Timezone.null {
		return AvailabilityQuery{}, ErrAvailabilitySourceInvalid
	}
	attendeeUserIDs := request.AttendeeUserIDs.value
	if len(attendeeUserIDs) == 0 || len(attendeeUserIDs) > 49 {
		return AvailabilityQuery{}, ErrAvailabilitySourceInvalid
	}
	userIDs := make([]uuid.UUID, 1, len(attendeeUserIDs)+1)
	userIDs[0] = actorID
	seen := make(map[uuid.UUID]struct{}, len(attendeeUserIDs)+1)
	seen[actorID] = struct{}{}
	for _, attendeeUserID := range attendeeUserIDs {
		if attendeeUserID == uuid.Nil {
			return AvailabilityQuery{}, ErrAvailabilitySourceInvalid
		}
		if _, exists := seen[attendeeUserID]; exists {
			return AvailabilityQuery{}, ErrAvailabilitySourceInvalid
		}
		seen[attendeeUserID] = struct{}{}
		userIDs = append(userIDs, attendeeUserID)
	}

	query := AvailabilityQuery{
		UserIDs:     userIDs,
		From:        request.From.value,
		To:          request.To.value,
		Timezone:    request.Timezone.value,
		SlotMinutes: AvailabilityMinSlotMinutes,
	}
	if _, _, err := validateAvailabilityQuery(query); err != nil {
		return AvailabilityQuery{}, err
	}
	return query, nil
}

func calendarHTTPAvailabilityCheckBadRequest() error {
	return fiber.NewError(fiber.StatusBadRequest, "calendar availability request is invalid")
}

func calendarHTTPAvailabilityCheckKeyAllowed(key string) bool {
	switch key {
	case "attendee_user_ids", "from", "to", "timezone":
		return true
	default:
		return false
	}
}
