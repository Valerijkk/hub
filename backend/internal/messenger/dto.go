package messenger

import (
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// DTO запросов
// ---------------------------------------------------------------------------

// CreateChannelRequest — запрос на создание канала.
// Version используется только в UpdateChannel (TASK-095 optimistic lock).
type CreateChannelRequest struct {
	ChannelType string   `json:"channel_type" validate:"required,oneof=dm group public private"`
	Name        string   `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description string   `json:"description,omitempty" validate:"omitempty,max=500"`
	ProjectID   string   `json:"project_id,omitempty"`
	IsReadonly  bool     `json:"is_readonly,omitempty"`
	MemberIDs   []string `json:"member_ids,omitempty"`
	Version     int      `json:"version,omitempty"`
}

// SendMessageRequest — запрос на отправку сообщения.
type SendMessageRequest struct {
	ContentJSON     map[string]interface{} `json:"content_json" validate:"required"`
	ContentText     string                 `json:"content_text,omitempty"`
	ThreadID        string                 `json:"thread_id,omitempty"`
	ClientNonce     string                 `json:"client_nonce,omitempty" validate:"omitempty,min=8,max=128"`
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
// DTO ответов
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
	Version     int        `json:"version"`
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
	MessageType     string                 `json:"message_type,omitempty"`
	PollID          string                 `json:"poll_id,omitempty"`
	ContentJSON     map[string]interface{} `json:"content_json"`
	ContentText     string                 `json:"content_text,omitempty"`
	ThreadID        string                 `json:"thread_id,omitempty"`
	ClientNonce     string                 `json:"client_nonce,omitempty"`
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
// ChannelName / ContentPreview — enrich для UI (A3W9-2-001); пустые, если канал/сообщение недоступны.
type BookmarkResponse struct {
	ID             string    `json:"id"`
	MessageID      string    `json:"message_id"`
	ChannelID      string    `json:"channel_id"`
	ChannelName    string    `json:"channel_name,omitempty"`
	ContentPreview string    `json:"content_preview,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
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
	MessageType     string                 `bson:"message_type,omitempty"`
	PollID          string                 `bson:"poll_id,omitempty"`
	ContentJSON     map[string]interface{} `bson:"content_json"`
	ContentText     string                 `bson:"content_text,omitempty"`
	ThreadID        string                 `bson:"thread_id,omitempty"`
	ClientNonce     string                 `bson:"client_nonce,omitempty"`
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

// ---------------------------------------------------------------------------
// DTO опросов
// ---------------------------------------------------------------------------

// CreatePollRequest — запрос на создание опроса.
type CreatePollRequest struct {
	Question      string   `json:"question" validate:"required,min=1,max=500"`
	Options       []string `json:"options" validate:"required,min=2,max=10"`
	IsAnonymous   bool     `json:"is_anonymous"`
	IsMultiChoice bool     `json:"is_multi_choice"`
}

// VotePollRequest — запрос на голосование в опросе.
type VotePollRequest struct {
	OptionIDs []string `json:"option_ids" validate:"required,min=1"`
}

// PollOptionResponse — вариант ответа в опросе.
type PollOptionResponse struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Position int    `json:"position"`
	Count    int    `json:"count"`
	// VoterIDs содержит ID проголосовавших (пусто при анонимном голосовании).
	VoterIDs []string `json:"voter_ids,omitempty"`
	// Voted — проголосовал ли текущий пользователь за этот вариант.
	Voted bool `json:"voted"`
}

// PollResponse — ответ с данными опроса.
type PollResponse struct {
	ID            string               `json:"id"`
	ChannelID     string               `json:"channel_id"`
	CreatorID     string               `json:"creator_id"`
	Question      string               `json:"question"`
	IsAnonymous   bool                 `json:"is_anonymous"`
	IsMultiChoice bool                 `json:"is_multi_choice"`
	IsClosed      bool                 `json:"is_closed"`
	TotalVotes    int                  `json:"total_votes"`
	Options       []PollOptionResponse `json:"options"`
	CreatedAt     time.Time            `json:"created_at"`
}

// toMessageResponse конвертирует MongoDB-документ в DTO ответа.
func toMessageResponse(doc *MessageDocument) *MessageResponse {
	return &MessageResponse{
		ID:              doc.ID,
		ChannelID:       doc.ChannelID.String(),
		SenderID:        doc.SenderID.String(),
		MessageType:     doc.MessageType,
		PollID:          doc.PollID,
		ContentJSON:     doc.ContentJSON,
		ContentText:     doc.ContentText,
		ThreadID:        doc.ThreadID,
		ClientNonce:     doc.ClientNonce,
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

// ---------------------------------------------------------------------------
// DTO приглашений в канал
// ---------------------------------------------------------------------------

// CreateInviteLinkRequest — тело запроса для создания пригласительной ссылки.
type CreateInviteLinkRequest struct {
	MaxUses   int    `json:"max_uses"`
	ExpiresIn string `json:"expires_in"`
}

// ChannelInviteResponse — ответ с данными пригласительной ссылки.
type ChannelInviteResponse struct {
	ID        uuid.UUID  `json:"id"`
	ChannelID uuid.UUID  `json:"channel_id"`
	CreatorID uuid.UUID  `json:"creator_id"`
	Token     string     `json:"token"`
	InviteURL string     `json:"invite_url"`
	MaxUses   int        `json:"max_uses"`
	UseCount  int        `json:"use_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsRevoked bool       `json:"is_revoked"`
	CreatedAt time.Time  `json:"created_at"`
}

// InviteValidationResponse — результат валидации пригласительной ссылки.
type InviteValidationResponse struct {
	Valid       bool   `json:"valid"`
	ChannelName string `json:"channel_name,omitempty"`
	ChannelType string `json:"channel_type,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
	Message     string `json:"message,omitempty"`
}
