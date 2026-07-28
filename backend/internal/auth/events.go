package auth

// SessionCreatedEvent публикуется в NATS на "auth.session.created" при входе пользователя.
type SessionCreatedEvent struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
	IPAddress string `json:"ip_address"`
}

// SessionEndedEvent публикуется в NATS на "auth.session.{userID}.{logout|revoked}"
// после Logout/Revoke, чтобы WS Bridge вызвал UnregisterBySession (A2W3-4-002 / TASK-109).
type SessionEndedEvent struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Reason    string `json:"reason"` // logout | revoked
}

// UserRegisteredEvent публикуется в NATS на "auth.user.registered" при регистрации нового пользователя.
type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Method string `json:"method"` // magic_link only (ADR-AUTH-01)
}
