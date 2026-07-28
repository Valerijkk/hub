package intranet

import (
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// CreateNewsRequest — создание новости (admin).
type CreateNewsRequest struct {
	Title     string             `json:"title"`
	Body      string             `json:"body"`
	Pinned    bool               `json:"pinned"`
	Published bool               `json:"published"`
	Poll      *CreatePollPayload `json:"poll,omitempty"`
}

// CreatePollPayload — опциональный опрос при создании новости (FR-INTR-01).
type CreatePollPayload struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

// UpdateNewsRequest — обновление новости (admin).
type UpdateNewsRequest struct {
	Title     *string `json:"title,omitempty"`
	Body      *string `json:"body,omitempty"`
	Pinned    *bool   `json:"pinned,omitempty"`
	Published *bool   `json:"published,omitempty"`
	Version   int     `json:"version"`
}

// VoteNewsPollRequest — голос в news-опросе.
type VoteNewsPollRequest struct {
	OptionID uuid.UUID `json:"option_id"`
}

// UpdateNewsPollRequest — правка news-опроса (FR-INTR-03).
type UpdateNewsPollRequest struct {
	Question *string  `json:"question,omitempty"`
	Options  []string `json:"options,omitempty"`
	Closed   *bool    `json:"closed,omitempty"`
	Version  int      `json:"version"`
}

// NewsPollOptionResponse — вариант с агрегатами.
type NewsPollOptionResponse struct {
	ID         uuid.UUID `json:"id"`
	Text       string    `json:"text"`
	VotesCount int       `json:"votes_count"`
	Percent    *float64  `json:"percent,omitempty"`
}

// NewsPollResponse — GET/vote ответ по опросу.
type NewsPollResponse struct {
	ID            uuid.UUID                `json:"id"`
	Question      string                   `json:"question"`
	Closed        bool                     `json:"closed"`
	Version       int                      `json:"version"`
	VotedOptionID *uuid.UUID               `json:"voted_option_id"`
	Options       []NewsPollOptionResponse `json:"options"`
	TotalVotes    int                      `json:"total_votes"`
}

// NewsPostResponse — DTO новости для API.
type NewsPostResponse struct {
	ID         uuid.UUID         `json:"id"`
	Title      string            `json:"title"`
	Body       string            `json:"body"`
	Pinned     bool              `json:"pinned"`
	Published  bool              `json:"published"`
	Version    int               `json:"version"`
	AuthorID   uuid.UUID         `json:"author_id"`
	AuthorName string            `json:"author_name,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Poll       *NewsPollResponse `json:"poll,omitempty"`
}

// NewsListResponse — пагинированный список новостей.
type NewsListResponse struct {
	Items   []NewsPostResponse `json:"items"`
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	PerPage int                `json:"per_page"`
}

// DirectoryUserResponse — карточка сотрудника в справочнике.
type DirectoryUserResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name,omitempty"`
	DisplayName string    `json:"display_name"`
	Department  string    `json:"department,omitempty"`
	Position    string    `json:"position,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Timezone    string    `json:"timezone,omitempty"`
	Birthday    *string   `json:"birthday,omitempty"`
}

// DirectoryDepartmentGroup — сотрудники одного отдела.
type DirectoryDepartmentGroup struct {
	Department string                  `json:"department"`
	Users      []DirectoryUserResponse `json:"users"`
}

// DirectoryResponse — дерево отделов со списком сотрудников.
type DirectoryResponse struct {
	Departments []DirectoryDepartmentGroup `json:"departments"`
	Total       int                        `json:"total"`
}

// CreateKudoRequest — отправка благодарности коллеге.
type CreateKudoRequest struct {
	RecipientID uuid.UUID `json:"recipient_id"`
	Message     string    `json:"message"`
	IsPublic    bool      `json:"is_public"`
}

// KudoResponse — DTO kudos для API.
type KudoResponse struct {
	ID            uuid.UUID `json:"id"`
	Message       string    `json:"message"`
	IsPublic      bool      `json:"is_public"`
	SenderID      uuid.UUID `json:"sender_id"`
	SenderName    string    `json:"sender_name,omitempty"`
	RecipientID   uuid.UUID `json:"recipient_id"`
	RecipientName string    `json:"recipient_name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// KudoListResponse — пагинированный список kudos.
type KudoListResponse struct {
	Items   []KudoResponse `json:"items"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"per_page"`
}

func toDirectoryUserResponse(u *ent.User) DirectoryUserResponse {
	displayName := u.DisplayName
	if displayName == "" {
		displayName = u.Name
	}
	if displayName == "" {
		displayName = u.Email
	}
	var birthday *string
	if u.Birthday != nil {
		formatted := u.Birthday.UTC().Format("2006-01-02")
		birthday = &formatted
	}
	return DirectoryUserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		DisplayName: displayName,
		Department:  u.Department,
		Position:    u.Position,
		AvatarURL:   u.AvatarURL,
		Timezone:    u.Timezone,
		Birthday:    birthday,
	}
}

func toNewsPostResponse(p *ent.NewsPost) NewsPostResponse {
	resp := NewsPostResponse{
		ID:        p.ID,
		Title:     p.Title,
		Body:      p.Body,
		Pinned:    p.Pinned,
		Published: p.Published,
		Version:   p.Version,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if p.Edges.Author != nil {
		resp.AuthorID = p.Edges.Author.ID
		if p.Edges.Author.DisplayName != "" {
			resp.AuthorName = p.Edges.Author.DisplayName
		} else if p.Edges.Author.Name != "" {
			resp.AuthorName = p.Edges.Author.Name
		}
	}
	return resp
}

func toKudoResponse(k *ent.Kudo) KudoResponse {
	resp := KudoResponse{
		ID:        k.ID,
		Message:   k.Message,
		IsPublic:  k.IsPublic,
		CreatedAt: k.CreatedAt,
	}
	if k.Edges.Sender != nil {
		resp.SenderID = k.Edges.Sender.ID
		resp.SenderName = displayUserName(k.Edges.Sender)
	}
	if k.Edges.Recipient != nil {
		resp.RecipientID = k.Edges.Recipient.ID
		resp.RecipientName = displayUserName(k.Edges.Recipient)
	}
	return resp
}

func displayUserName(u *ent.User) string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Email
}
