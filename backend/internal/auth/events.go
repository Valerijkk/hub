package auth

// SessionCreatedEvent публикуется в NATS на "auth.session.created" при входе пользователя.
type SessionCreatedEvent struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
	IPAddress string `json:"ip_address"`
}

// UserRegisteredEvent публикуется в NATS на "auth.user.registered" при регистрации нового пользователя.
type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Method string `json:"method"` // "magic_link" or "telegram"
}
