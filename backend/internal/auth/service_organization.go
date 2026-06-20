package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

const pendingUserTTL = 7 * 24 * time.Hour

// OrganizationService manages the single org per instance (ORG-01).
type OrganizationService struct {
	db *ent.Client
}

// NewOrganizationService creates OrganizationService.
func NewOrganizationService(db *ent.Client) *OrganizationService {
	return &OrganizationService{db: db}
}

// IsSystemAdmin returns true for founder and admin roles (ORG-04).
func IsSystemAdmin(role user.SystemRole) bool {
	return role == user.SystemRoleFounder || role == user.SystemRoleAdmin
}

// HasOrganization reports whether an org row exists.
func (s *OrganizationService) HasOrganization(ctx context.Context) (bool, error) {
	n, err := s.db.Organization.Query().Count(ctx)
	if err != nil {
		return false, fmt.Errorf("count organizations: %w", err)
	}
	return n > 0, nil
}

// GetOrganization returns the singleton organization.
func (s *OrganizationService) GetOrganization(ctx context.Context) (*ent.Organization, error) {
	org, err := s.db.Organization.Query().Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return org, nil
}

// ValidateInviteToken checks org invite token (ORG-05).
func (s *OrganizationService) ValidateInviteToken(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidInviteToken
	}
	ok, err := s.db.Organization.Query().
		Where(organization.InviteTokenEQ(token)).
		Exist(ctx)
	if err != nil {
		return fmt.Errorf("validate invite token: %w", err)
	}
	if !ok {
		return ErrInvalidInviteToken
	}
	return nil
}

// CreateOrganizationWithFounder creates org + founder user atomically (first registration).
func (s *OrganizationService) CreateOrganizationWithFounder(
	ctx context.Context,
	client *ent.Client,
	email string,
) (*ent.User, error) {
	token, err := generateInviteToken()
	if err != nil {
		return nil, err
	}
	org, err := client.Organization.Create().
		SetName("Hub").
		SetSlug("hub").
		SetInviteToken(token).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	u, err := client.User.Create().
		SetEmail(email).
		SetSystemRole(user.SystemRoleFounder).
		SetStatus(user.StatusActive).
		SetOrganization(org).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create founder: %w", err)
	}
	return u, nil
}

// CreatePendingMember creates a user awaiting admin approval (ORG-06).
func (s *OrganizationService) CreatePendingMember(
	ctx context.Context,
	client *ent.Client,
	email string,
	orgID uuid.UUID,
) (*ent.User, error) {
	now := time.Now().UTC()
	u, err := client.User.Create().
		SetEmail(email).
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusPending).
		SetPendingSince(now).
		SetOrganizationID(orgID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create pending user: %w", err)
	}
	return u, nil
}

// ListPendingUsers returns users in pending status.
func (s *OrganizationService) ListPendingUsers(ctx context.Context) ([]*ent.User, error) {
	users, err := s.db.User.Query().
		Where(user.StatusEQ(user.StatusPending)).
		Order(ent.Desc(user.FieldPendingSince)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pending users: %w", err)
	}
	return users, nil
}

// ApprovePendingUser activates a pending user (ORG-06).
func (s *OrganizationService) ApprovePendingUser(ctx context.Context, userID uuid.UUID) error {
	n, err := s.db.User.Update().
		Where(
			user.IDEQ(userID),
			user.StatusEQ(user.StatusPending),
		).
		SetStatus(user.StatusActive).
		ClearPendingSince().
		Save(ctx)
	if err != nil {
		return fmt.Errorf("approve pending user: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// RejectPendingUser removes a pending registration.
func (s *OrganizationService) RejectPendingUser(ctx context.Context, userID uuid.UUID) error {
	n, err := s.db.User.Delete().
		Where(
			user.IDEQ(userID),
			user.StatusEQ(user.StatusPending),
		).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("reject pending user: %w", err)
	}
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

// PurgeExpiredPending deletes pending users older than 7 days (ORG-06).
func (s *OrganizationService) PurgeExpiredPending(ctx context.Context) (int, error) {
	cutoff := time.Now().UTC().Add(-pendingUserTTL)
	n, err := s.db.User.Delete().
		Where(
			user.StatusEQ(user.StatusPending),
			user.PendingSinceLT(cutoff),
		).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("purge expired pending: %w", err)
	}
	return n, nil
}

func generateInviteToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ErrInvalidInviteToken indicates org invite token is missing or wrong.
var ErrInvalidInviteToken = fmt.Errorf("invalid invite token")
