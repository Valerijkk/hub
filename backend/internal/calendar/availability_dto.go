package calendar

import (
	"time"

	"github.com/google/uuid"
)

type calendarHTTPAvailabilityRequest struct {
	UserIDs     []uuid.UUID `json:"user_ids"`
	From        time.Time   `json:"from"`
	To          time.Time   `json:"to"`
	Timezone    string      `json:"timezone"`
	SlotMinutes int         `json:"slot_minutes"`
}

func (request calendarHTTPAvailabilityRequest) query() (AvailabilityQuery, error) {
	query := AvailabilityQuery{
		UserIDs:     append([]uuid.UUID(nil), request.UserIDs...),
		From:        request.From,
		To:          request.To,
		Timezone:    request.Timezone,
		SlotMinutes: request.SlotMinutes,
	}
	if _, _, err := validateAvailabilityQuery(query); err != nil {
		return AvailabilityQuery{}, err
	}
	return query, nil
}

type calendarHTTPAvailabilityBusyResponse struct {
	From   time.Time          `json:"from"`
	To     time.Time          `json:"to"`
	Source AvailabilitySource `json:"source"`
}

type calendarHTTPAvailabilityFreeResponse struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type calendarHTTPAvailabilityUserResponse struct {
	UserID uuid.UUID                              `json:"user_id"`
	Busy   []calendarHTTPAvailabilityBusyResponse `json:"busy"`
	Free   []calendarHTTPAvailabilityFreeResponse `json:"free"`
}

type calendarHTTPAvailabilityResponse struct {
	Users []calendarHTTPAvailabilityUserResponse `json:"users"`
}

func calendarHTTPAvailabilityResponseFor(result AvailabilityResult, timezone string) (calendarHTTPAvailabilityResponse, error) {
	location, err := availabilityResponseLocation(timezone)
	if err != nil {
		return calendarHTTPAvailabilityResponse{}, err
	}

	response := calendarHTTPAvailabilityResponse{
		Users: make([]calendarHTTPAvailabilityUserResponse, 0, len(result.Users)),
	}
	for _, user := range result.Users {
		if user.UserID == uuid.Nil {
			return calendarHTTPAvailabilityResponse{}, ErrAvailabilitySourceInvalid
		}
		entry := calendarHTTPAvailabilityUserResponse{
			UserID: user.UserID,
			Busy:   make([]calendarHTTPAvailabilityBusyResponse, 0, len(user.Busy)),
			Free:   make([]calendarHTTPAvailabilityFreeResponse, 0, len(user.Free)),
		}
		for _, busy := range user.Busy {
			from, fromOK := canonicalAvailabilityTime(busy.From)
			to, toOK := canonicalAvailabilityTime(busy.To)
			if !fromOK || !toOK || !to.After(from) || busy.Source.Validate() != nil {
				return calendarHTTPAvailabilityResponse{}, ErrAvailabilitySourceInvalid
			}
			entry.Busy = append(entry.Busy, calendarHTTPAvailabilityBusyResponse{
				From: from.In(location), To: to.In(location), Source: busy.Source,
			})
		}
		for _, free := range user.Free {
			from, fromOK := canonicalAvailabilityTime(free.From)
			to, toOK := canonicalAvailabilityTime(free.To)
			if !fromOK || !toOK || !to.After(from) {
				return calendarHTTPAvailabilityResponse{}, ErrAvailabilitySourceInvalid
			}
			entry.Free = append(entry.Free, calendarHTTPAvailabilityFreeResponse{
				From: from.In(location), To: to.In(location),
			})
		}
		response.Users = append(response.Users, entry)
	}
	return response, nil
}
