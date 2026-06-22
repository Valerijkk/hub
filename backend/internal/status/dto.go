package status

import "time"

// ComponentStatus — состояние одного компонента инфраструктуры.
type ComponentStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

// Response — публичный статус Hub (V4-06).
type Response struct {
	Status     string            `json:"status"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Components []ComponentStatus `json:"components"`
}
