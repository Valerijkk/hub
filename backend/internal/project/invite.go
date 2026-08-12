package project

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/projectmember"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// GenerateInviteCode генерирует криптостойкий 8-символьный код приглашения.
func GenerateInviteCode() (string, error) {
	b := make([]byte, 5) // 5 байт = 8 символов base32
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate invite code: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// CreateInvite создаёт ссылку-приглашение в проект.
func (s *ProjectService) CreateInvite(ctx context.Context, projectID, creatorID uuid.UUID, role string, maxUses int, expiresAt *time.Time) (*InviteResponse, error) {
	code, err := GenerateInviteCode()
	if err != nil {
		return nil, apperrors.NewInternal("failed to generate invite code")
	}

	invRole := invite.ProjectRole(role)
	inv, err := s.repo.CreateInvite(ctx, projectID, creatorID, code, invRole, maxUses, expiresAt)
	if err != nil {
		return nil, apperrors.NewInternal("failed to create invite")
	}

	return toInviteResponse(inv, s.resolveFrontendURL()), nil
}

// JoinByInvite позволяет пользователю вступить в проект по invite-коду.
// Проверяет срок, max uses и что пользователь ещё не member.
func (s *ProjectService) JoinByInvite(ctx context.Context, code string, userID uuid.UUID) (*ProjectResponse, error) {
	inv, err := s.repo.FindInviteByCode(ctx, code)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("invite not found or invalid code")
		}
		return nil, apperrors.NewInternal("failed to find invite")
	}

	// Проверяем срок действия
	if inv.ExpiresAt != nil && inv.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.NewBadRequest("invite has expired")
	}

	// Быстрая проверка до атомарного TryConsumeInvite — экономим round-trip
	// в типичном случае «лимит уже исчерпан».
	if inv.MaxUses > 0 && inv.UseCount >= inv.MaxUses {
		return nil, apperrors.NewBadRequest("invite has reached maximum uses")
	}

	// project_id из edge invite — без него нельзя добавить member.
	proj := inv.Edges.Project
	if proj == nil {
		return nil, apperrors.NewInternal("invite has no associated project")
	}

	// Проверяем, не является ли пользователь уже участником
	_, err = s.repo.FindMember(ctx, proj.ID, userID)
	if err == nil {
		return nil, apperrors.NewConflict("user is already a member of this project")
	}
	if !ent.IsNotFound(err) {
		return nil, apperrors.NewInternal("failed to check membership")
	}

	// HI-06: атомарно резервируем место в инвайте перед AddMember. Без этого
	// два параллельных JoinByInvite могли пройти check-then-act и превысить
	// MaxUses.
	consumed, err := s.repo.TryConsumeInvite(ctx, inv.ID, inv.MaxUses)
	if err != nil {
		return nil, apperrors.NewInternal("failed to reserve invite slot")
	}
	if !consumed {
		return nil, apperrors.NewBadRequest("invite has reached maximum uses")
	}

	// Добавляем member с ролью из invite.
	// NOTE: если AddMember упадёт, UseCount уже инкрементирован — мы получим
	// одну «потерянную» единицу использования. Это допустимая компенсация за
	// атомарность лимита (best-effort delivery vs. exactly-once).
	role := projectmember.ProjectRole(string(inv.ProjectRole))
	_, err = s.repo.AddMember(ctx, proj.ID, userID, role)
	if err != nil {
		slog.Error("failed to add member via invite after consuming slot",
			"invite_id", inv.ID, "user_id", userID, "error", err)
		return nil, apperrors.NewInternal("failed to add member via invite")
	}

	// Сбрасываем кэш RBAC
	if s.cache != nil {
		if err := s.cache.InvalidateProject(ctx, proj.ID.String()); err != nil {
			slog.Error("failed to invalidate RBAC cache", "error", err, "project_id", proj.ID)
		}
	}

	count, err := s.repo.MemberCount(ctx, proj.ID)
	if err != nil {
		slog.Error("failed to get member count after invite join",
			"project_id", proj.ID, "user_id", userID, "error", err)
		return nil, apperrors.NewInternal("failed to get member count")
	}
	return toProjectResponse(proj, count, string(inv.ProjectRole)), nil
}
