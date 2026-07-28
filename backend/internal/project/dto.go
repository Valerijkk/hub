package project

import "time"

// CreateProjectRequest — тело запроса на создание проекта.
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Key         string `json:"key" validate:"required,uppercase,min=2,max=10,alphanum"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

// UpdateProjectRequest — тело запроса на обновление проекта.
type UpdateProjectRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	AvatarURL   *string `json:"avatar_url"`
	Version     int     `json:"version" validate:"required,min=1"`
}

// AddMemberRequest — тело запроса на добавление участника.
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=developer tester viewer"`
}

// UpdateMemberRoleRequest — тело запроса на смену роли участника.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner developer tester viewer"`
}

// CreateInviteRequest — тело запроса на создание приглашения.
type CreateInviteRequest struct {
	Role      string     `json:"role" validate:"required,oneof=developer tester viewer"`
	MaxUses   int        `json:"max_uses"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateTeamRequest — тело запроса на создание команды.
type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
}

// AddTeamMemberRequest — тело запроса на добавление в команду.
type AddTeamMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// JoinByInviteRequest — тело запроса на вступление по invite-коду.
type JoinByInviteRequest struct {
	Code string `json:"code" validate:"required"`
}

// DTO ответов

// ProjectResponse — представление проекта в API-ответах.
type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description string    `json:"description,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Status      string    `json:"status"`
	MemberCount int       `json:"member_count"`
	MyRole      string    `json:"my_role,omitempty"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MemberResponse — представление участника проекта в API-ответах.
type MemberResponse struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// InviteResponse — представление приглашения в API-ответах.
type InviteResponse struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Role      string     `json:"role"`
	MaxUses   int        `json:"max_uses"`
	UseCount  int        `json:"use_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Link      string     `json:"link"`
	CreatedAt time.Time  `json:"created_at"`
}

// TeamResponse — представление команды в API-ответах.
type TeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count"`
}

// TeamMemberResponse — представление участника команды в API-ответах.
type TeamMemberResponse struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
}
