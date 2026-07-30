//go:build hubtest

package project

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// ExportToProjectResponse конвертирует ent.Project в ProjectResponse для hubtest.
func ExportToProjectResponse(p *ent.Project, memberCount int, myRole string) *ProjectResponse {
	return toProjectResponse(p, memberCount, myRole)
}

// ExportToMemberResponse конвертирует ent.ProjectMember в MemberResponse для hubtest.
func ExportToMemberResponse(m *ent.ProjectMember) *MemberResponse {
	return toMemberResponse(m)
}

// ExportToInviteResponse конвертирует ent.Invite в InviteResponse для hubtest.
func ExportToInviteResponse(inv *ent.Invite, frontendURL string) *InviteResponse {
	return toInviteResponse(inv, frontendURL)
}

// ExportTemplateExportVersion возвращает версию формата экспорта шаблона.
func ExportTemplateExportVersion() int {
	return templateExportVersion
}

// ExportProjectServiceProvisionProject вызывает provisionProject для hubtest.
func ExportProjectServiceProvisionProject(s *ProjectService, ctx context.Context, projectID, ownerID uuid.UUID) {
	s.provisionProject(ctx, projectID, ownerID)
}

// ExportProjectServiceNotifyLinkDeleted вызывает notifyLinkDeleted для hubtest.
func ExportProjectServiceNotifyLinkDeleted(s *ProjectService, ctx context.Context, otherIssue *ent.Issue, deletedIssueKey, deletedProjectName string) {
	s.notifyLinkDeleted(ctx, otherIssue, deletedIssueKey, deletedProjectName)
}

// ExportToTemplateResponse конвертирует ent.ProjectTemplate в TemplateResponse для hubtest.
func ExportToTemplateResponse(t *ent.ProjectTemplate) *TemplateResponse {
	return toTemplateResponse(t)
}

// ExportSetAfterAddTeamLinkHook инъектирует mid-fail после AddTeamIDs внутри tx (TASK-088).
func ExportSetAfterAddTeamLinkHook(fn func(ctx context.Context) error) {
	afterAddTeamLinkHook = fn
}

// ExportSetAfterCreateTeamHook инъектирует mid-fail после CreateTeam внутри tx (A3W1-A5-003).
func ExportSetAfterCreateTeamHook(fn func(ctx context.Context) error) {
	afterCreateTeamHook = fn
}

// ExportSetAfterCrossLinkDeleteHook инъектирует mid-fail после DeleteIssueLink
// внутри SoftDeleteWithCrossLinkCleanup tx (A2W1-A5-001 / TASK-122).
func ExportSetAfterCrossLinkDeleteHook(fn func(ctx context.Context) error) {
	afterCrossLinkDeleteHook = fn
}

// ExportMaxTemplateImportBytes возвращает лимит импорта шаблона для hubtest (A3W1-A5-006).
func ExportMaxTemplateImportBytes() int {
	return maxTemplateImportBytes
}
