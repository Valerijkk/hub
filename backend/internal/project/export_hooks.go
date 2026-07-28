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
