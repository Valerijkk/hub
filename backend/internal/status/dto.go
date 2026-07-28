package status

import "time"

// ComponentStatus — состояние одного компонента инфраструктуры.
// Отдаётся только авторизованным клиентам (см. GetStatusDetails).
type ComponentStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

// Response — статус Hub (V4-06).
// Публичный ответ содержит только агрегат (W7-A5-002: без id компонентов).
type Response struct {
	Status     string            `json:"status"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Components []ComponentStatus `json:"components,omitempty"`
}
