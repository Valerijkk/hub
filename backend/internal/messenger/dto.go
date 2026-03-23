package messenger

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Request DTOs
// ---------------------------------------------------------------------------

// CreateChannelRequest — запрос на создание канала.
type CreateChannelRequest struct {
	ChannelType string    `json:"channel_type" validate:"required,oneof=dm group public private"`
	Name        string    `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description string    `json:"description,omitempty" validate:"omitempty,max=500"`
	ProjectID   string    `json:"project_id,omitempty"`
	IsReadonly  bool      `json:"is_readonly,omitempty"`
	MemberIDs   []string  `json:"member_ids,omitempty"`
}

// SendMessageRequest — запрос на отправку сообщения.
type SendMessageRequest struct {
	ContentJSON     map[string]interface{} `json:"content_json" validate:"required"`
	ContentText     string                 `json:"content_text,omitempty"`
	ThreadID        string                 `json:"thread_id,omitempty"`
	Attachments     []AttachmentRef        `json:"attachments,omitempty"`
	Mentions        []string               `json:"mentions,omitempty"`
	MentionGroups   []string               `json:"mention_groups,omitempty"`
	VoiceURL        string                 `json:"voice_url,omitempty"`
	VoiceDurationMs int                    `json:"voice_duration_ms,omitempty"`
}

// EditMessageRequest — запрос на редактирование сообщения.
type EditMessageRequest struct {
	ContentJSON map[string]interface{} `json:"content_json" validate:"required"`
	ContentText string                 `json:"content_text,omitempty"`
}

// ToggleReactionRequest — запрос на добавление/удаление реакции.
type ToggleReactionRequest struct {
	Emoji string `json:"emoji" validate:"required,max=32"`
}

// ForwardMessageRequest — запрос на пересылку сообщения.
type ForwardMessageRequest struct {
	TargetChannelID string `json:"target_channel_id" validate:"required"`
}

// SearchMessagesRequest — запрос на поиск сообщений.
type SearchMessagesRequest struct {
	Query     string `json:"query" query:"query" validate:"required,min=1"`
	ChannelID string `json:"channel_id,omitempty" query:"channel_id"`
	Limit     int    `json:"limit,omitempty" query:"limit" validate:"omitempty,min=1,max=100"`
}

// AddMembersRequest — запрос на добавление участников в канал.
type AddMembersRequest struct {
	MemberIDs []string `json:"member_ids" validate:"required,min=1"`
}

// AddBookmarkRequest — запрос на добавление закладки.
type AddBookmarkRequest struct {
	MessageID string `json:"message_id" validate:"required"`
	ChannelID string `json:"channel_id" validate:"required"`
}

// CreateDMRequest — запрос на создание/получение DM-канала.
type CreateDMRequest struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}

// ---------------------------------------------------------------------------
// Response DTOs
// ---------------------------------------------------------------------------

// ChannelResponse — ответ с данными канала.
type ChannelResponse struct {
	ID          string     `json:"id"`
	ChannelType string     `json:"channel_type"`
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	IsReadonly  bool       `json:"is_readonly"`
	ProjectID   string     `json:"project_id,omitempty"`
	CreatedBy   string     `json:"created_by"`
	IsStarred   bool       `json:"is_starred"`
	MutedUntil  *time.Time `json:"muted_until,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ChannelMemberResponse — ответ с данными участника канала.
type ChannelMemberResponse struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	UserName string    `json:"user_name,omitempty"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// MessageResponse — ответ с данными сообщения.
type MessageResponse struct {
	ID              string                 `json:"id"`
	ChannelID       string                 `json:"channel_id"`
	SenderID        string                 `json:"sender_id"`
	ContentJSON     map[string]interface{} `json:"content_json"`
	ContentText     string                 `json:"content_text,omitempty"`
	ThreadID        string                 `json:"thread_id,omitempty"`
	ReplyCount      int                    `json:"reply_count"`
	ForwardedFromID string                 `json:"forwarded_from_id,omitempty"`
	VoiceURL        string                 `json:"voice_url,omitempty"`
	VoiceDurationMs int                    `json:"voice_duration_ms,omitempty"`
	Attachments     []AttachmentRef        `json:"attachments,omitempty"`
	Reactions       []ReactionEntry        `json:"reactions,omitempty"`
	Mentions        []string               `json:"mentions,omitempty"`
	MentionGroups   []string               `json:"mention_groups,omitempty"`
	EditedAt        *time.Time             `json:"edited_at,omitempty"`
	DeletedAt       *time.Time             `json:"deleted_at,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
}

// BookmarkResponse — ответ с данными закладки.
type BookmarkResponse struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	ChannelID string    `json:"channel_id"`
	CreatedAt time.Time `json:"created_at"`
}

// PinnedMessageResponse — ответ с данными закреплённого сообщения.
type PinnedMessageResponse struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	ChannelID string    `json:"channel_id"`
	PinnedBy  string    `json:"pinned_by"`
	PinnedAt  time.Time `json:"pinned_at"`
}

// ---------------------------------------------------------------------------
// MongoDB document и вложенные структуры
// ---------------------------------------------------------------------------

// MessageDocument — документ сообщения в MongoDB.
type MessageDocument struct {
	ID              string                 `bson:"_id,omitempty"`
	ChannelID       uuid.UUID              `bson:"channel_id"`
	SenderID        uuid.UUID              `bson:"sender_id"`
	ContentJSON     map[string]interface{} `bson:"content_json"`
	ContentText     string                 `bson:"content_text,omitempty"`
	ThreadID        string                 `bson:"thread_id,omitempty"`
	ReplyCount      int                    `bson:"reply_count"`
	ForwardedFromID string                 `bson:"forwarded_from_id,omitempty"`
	VoiceURL        string                 `bson:"voice_url,omitempty"`
	VoiceDurationMs int                    `bson:"voice_duration_ms,omitempty"`
	Attachments     []AttachmentRef        `bson:"attachments,omitempty"`
	Reactions       []ReactionEntry        `bson:"reactions,omitempty"`
	Mentions        []string               `bson:"mentions,omitempty"`
	MentionGroups   []string               `bson:"mention_groups,omitempty"`
	EditedAt        *time.Time             `bson:"edited_at,omitempty"`
	DeletedAt       *time.Time             `bson:"deleted_at,omitempty"`
	CreatedAt       time.Time              `bson:"created_at"`
}

// ReactionEntry — запись реакции на сообщение.
type ReactionEntry struct {
	Emoji   string   `json:"emoji" bson:"emoji"`
	UserIDs []string `json:"user_ids" bson:"user_ids"`
}

// AttachmentRef — ссылка на вложение к сообщению.
type AttachmentRef struct {
	FileID   string `json:"file_id" bson:"file_id"`
	FileName string `json:"file_name" bson:"file_name"`
	FileSize int64  `json:"file_size" bson:"file_size"`
	MimeType string `json:"mime_type" bson:"mime_type"`
	URL      string `json:"url" bson:"url"`
}

// toMessageResponse конвертирует MongoDB-документ в DTO ответа.
func toMessageResponse(doc *MessageDocument) *MessageResponse {
	return &MessageResponse{
		ID:              doc.ID,
		ChannelID:       doc.ChannelID.String(),
		SenderID:        doc.SenderID.String(),
		ContentJSON:     doc.ContentJSON,
		ContentText:     doc.ContentText,
		ThreadID:        doc.ThreadID,
		ReplyCount:      doc.ReplyCount,
		ForwardedFromID: doc.ForwardedFromID,
		VoiceURL:        doc.VoiceURL,
		VoiceDurationMs: doc.VoiceDurationMs,
		Attachments:     doc.Attachments,
		Reactions:       doc.Reactions,
		Mentions:        doc.Mentions,
		MentionGroups:   doc.MentionGroups,
		EditedAt:        doc.EditedAt,
		DeletedAt:       doc.DeletedAt,
		CreatedAt:       doc.CreatedAt,
	}
}
