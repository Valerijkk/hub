package calls

import (
	"time"

	"github.com/google/uuid"
)

// CreateCallRequest содержит параметры для создания нового звонка.
type CreateCallRequest struct {
	ChannelID uuid.UUID `json:"channel_id"` // ID канала, в котором создаётся звонок
}

// JoinCallRequest содержит параметры для присоединения к звонку.
type JoinCallRequest struct {
	CallID uuid.UUID `json:"call_id"` // ID звонка для присоединения
}

// RaiseHandRequest содержит параметры для поднятия/опускания руки.
type RaiseHandRequest struct {
	Raised bool `json:"raised"` // true — поднять руку, false — опустить
}

// CallResponse содержит данные звонка для ответа клиенту.
type CallResponse struct {
	ID               uuid.UUID  `json:"id"`
	RoomName         string     `json:"room_name"`
	ChannelID        *uuid.UUID `json:"channel_id,omitempty"`
	Status           string     `json:"status"`
	CreatedBy        uuid.UUID  `json:"created_by"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	EndedAt          *time.Time `json:"ended_at,omitempty"`
	ParticipantCount int        `json:"participant_count"`
	Token            string     `json:"token,omitempty"`
	InviteCode       string     `json:"invite_code,omitempty"`
	Title            string     `json:"title,omitempty"`
}

// GuestJoinRequest содержит параметры для гостевого входа в звонок.
type GuestJoinRequest struct {
	InviteCode string `json:"invite_code"` // Invite-код звонка
	GuestName  string `json:"guest_name"`  // Имя гостя
}

// GuestJoinResponse содержит данные для подключения гостя к LiveKit.
type GuestJoinResponse struct {
	ID         uuid.UUID `json:"id"`
	RoomName   string    `json:"room_name"`
	Token      string    `json:"token"`
	LivekitURL string    `json:"livekit_url"`
}

// CallParticipantResponse содержит данные участника звонка.
type CallParticipantResponse struct {
	UserID   uuid.UUID  `json:"user_id"`
	UserName string     `json:"user_name"`
	JoinedAt time.Time  `json:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty"`
}
