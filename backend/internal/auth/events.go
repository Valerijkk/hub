package auth

// SessionCreatedEvent is published to NATS on "auth.session.created" when a user logs in.
type SessionCreatedEvent struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	Device    string `json:"device"`
	IPAddress string `json:"ip_address"`
}

// UserRegisteredEvent is published to NATS on "auth.user.registered" when a new user signs up.
type UserRegisteredEvent struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Method string `json:"method"` // "magic_link" or "telegram"
}
