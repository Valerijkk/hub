package backup

import "time"

// BackupResponse — ответ с данными записи бэкапа.
type BackupResponse struct {
	ID           string     `json:"id"`
	S3Key        string     `json:"s3_key"`
	FileSize     int64      `json:"file_size"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// RestoreRequest — запрос на восстановление (double confirm).
type RestoreRequest struct {
	Confirm string `json:"confirm" validate:"required"` // Должно быть "RESTORE"
}
