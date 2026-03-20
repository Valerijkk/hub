package project

import "time"

// CreateProjectRequest is the request body for creating a project.
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Key         string `json:"key" validate:"required,uppercase,min=2,max=10,alphanum"`
	Description string `json:"description,omitempty" validate:"max=2000"`
}

// UpdateProjectRequest is the request body for updating a project.
type UpdateProjectRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	AvatarURL   *string `json:"avatar_url"`
}

// AddMemberRequest is the request body for adding a member to a project.
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required,oneof=developer tester viewer"`
}

// UpdateMemberRoleRequest is the request body for changing a member's role.
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner developer tester viewer"`
}

// CreateInviteRequest is the request body for creating a project invite.
type CreateInviteRequest struct {
	Role      string     `json:"role" validate:"required,oneof=developer tester viewer"`
	MaxUses   int        `json:"max_uses"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// CreateTeamRequest is the request body for creating a team.
type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description,omitempty"`
}

// AddTeamMemberRequest is the request body for adding a user to a team.
type AddTeamMemberRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// JoinByInviteRequest is the request body for joining a project via invite code.
type JoinByInviteRequest struct {
	Code string `json:"code" validate:"required"`
}

// Response DTOs

// ProjectResponse represents a project in API responses.
type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Key         string    `json:"key"`
	Description string    `json:"description,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Status      string    `json:"status"`
	MemberCount int       `json:"member_count"`
	MyRole      string    `json:"my_role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MemberResponse represents a project member in API responses.
type MemberResponse struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// InviteResponse represents a project invite in API responses.
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

// TeamResponse represents a team in API responses.
type TeamResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count"`
}

// TeamMemberResponse represents a team member in API responses.
type TeamMemberResponse struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	JoinedAt  time.Time `json:"joined_at"`
}
