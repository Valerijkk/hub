package intranet

import (
	"time"

	"github.com/google/uuid"
)

// CoffeeEnrollmentResponse — статус участия пользователя.
type CoffeeEnrollmentResponse struct {
	Enrolled  bool       `json:"enrolled"`
	IsActive  bool       `json:"is_active"`
	EnrolledAt *time.Time `json:"enrolled_at,omitempty"`
}

// CoffeePartnerResponse — напарник в текущем/прошлом раунде.
type CoffeePartnerResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name,omitempty"`
	Email       string    `json:"email,omitempty"`
	Department  string    `json:"department,omitempty"`
}

// CoffeePairResponse — пара в раунде.
type CoffeePairResponse struct {
	ID           uuid.UUID             `json:"id"`
	ParticipantA CoffeePartnerResponse `json:"participant_a"`
	ParticipantB CoffeePartnerResponse `json:"participant_b"`
}

// CoffeeTripleResponse — тройка при нечётном N (FR-INTR-10).
type CoffeeTripleResponse struct {
	ParticipantA CoffeePartnerResponse `json:"participant_a"`
	ParticipantB CoffeePartnerResponse `json:"participant_b"`
	ParticipantC CoffeePartnerResponse `json:"participant_c"`
}

// CoffeeRoundResponse — результат раунда matching.
type CoffeeRoundResponse struct {
	ID            uuid.UUID              `json:"id"`
	PairCount     int                    `json:"pair_count"`
	TripleCount   int                    `json:"triple_count"`
	UnpairedCount int                    `json:"unpaired_count"`
	Pairs         []CoffeePairResponse   `json:"pairs"`
	Triples       []CoffeeTripleResponse `json:"triples,omitempty"`
	Unpaired      []CoffeePartnerResponse `json:"unpaired,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	TriggeredBy   string                 `json:"triggered_by,omitempty"`
}

// CoffeeMyMatchResponse — текущий match пользователя (последний раунд).
type CoffeeMyMatchResponse struct {
	RoundID   *uuid.UUID            `json:"round_id,omitempty"`
	RoundAt   *time.Time            `json:"round_at,omitempty"`
	Partner   *CoffeePartnerResponse `json:"partner,omitempty"`
	Unpaired  bool                  `json:"unpaired"`
}

// CoffeeStatsResponse — для admin перед запуском.
type CoffeeStatsResponse struct {
	ActiveEnrollments int `json:"active_enrollments"`
}
