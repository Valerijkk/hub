//go:build hubtest

package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/valerii/hub/backend/ent"
)

// ExportParseIssueEventKind экспортирует parseIssueEventKind для hubtest-сборки.
func ExportParseIssueEventKind(subject string) string {
	return parseIssueEventKind(subject)
}

// ExportAutomationConsumerProcessMessage экспортирует processMessage consumer для hubtest.
func ExportAutomationConsumerProcessMessage(c *AutomationConsumer, ctx context.Context, msg jetstream.Msg) {
	c.processMessage(ctx, msg)
}

// ExportMaxAutomationRulesPerProject возвращает лимит правил автоматизации на проект.
func ExportMaxAutomationRulesPerProject() int {
	return maxAutomationRulesPerProject
}

// ExportRenderAutomationTemplate подставляет переменные в шаблон сообщения автоматизации.
func ExportRenderAutomationTemplate(tmpl, issueKey, ruleName, extra string) string {
	return renderAutomationTemplate(tmpl, issueKey, ruleName, extra)
}

// ExportValidateAutomationRuleRequest проверяет конфигурацию правила автоматизации.
func ExportValidateAutomationRuleRequest(triggerType, actionType string, triggerCfg, actionCfg map[string]interface{}) error {
	return validateAutomationRuleRequest(triggerType, actionType, triggerCfg, actionCfg)
}

// ExportAutomationEngineExecuteRule вызывает executeRule для white-box тестов.
func ExportAutomationEngineExecuteRule(e *AutomationEngine, ctx context.Context, rule *ent.AutomationRule, issueKey, issueID, projectID, actorID, extra string) error {
	return e.executeRule(ctx, rule, issueKey, issueID, projectID, actorID, extra)
}

// ExportToAutomationRuleResponse конвертирует ent.AutomationRule в AutomationRuleResponse для hubtest.
func ExportToAutomationRuleResponse(r *ent.AutomationRule) AutomationRuleResponse {
	return toAutomationRuleResponse(r)
}

// ExportToCommentResponse конвертирует ent.Comment в CommentResponse для hubtest.
func ExportToCommentResponse(c *ent.Comment) *CommentResponse {
	return toCommentResponse(c)
}

// ExportToActivityLogResponse конвертирует ent.ActivityLog в ActivityLogResponse для hubtest.
func ExportToActivityLogResponse(a *ent.ActivityLog) *ActivityLogResponse {
	return toActivityLogResponse(a)
}

// ExportToTimeLogResponse конвертирует ent.TimeLog в TimeLogResponse для hubtest.
func ExportToTimeLogResponse(tl *ent.TimeLog) *TimeLogResponse {
	return toTimeLogResponse(tl)
}

// ExportToLabelGroupResponse конвертирует ent.LabelGroup в LabelGroupResponse для hubtest.
func ExportToLabelGroupResponse(g *ent.LabelGroup) *LabelGroupResponse {
	return toLabelGroupResponse(g)
}

// ExportToBoardResponse конвертирует ent.Board в BoardResponse для hubtest.
func ExportToBoardResponse(b *ent.Board) *BoardResponse {
	return toBoardResponse(b)
}

// ExportToIssueResponse конвертирует ent.Issue в IssueResponse для hubtest.
func ExportToIssueResponse(i *ent.Issue) *IssueResponse {
	return toIssueResponse(i)
}

// ExportToIssueBriefResponse конвертирует ent.Issue в IssueBriefResponse для hubtest.
func ExportToIssueBriefResponse(i *ent.Issue) *IssueBriefResponse {
	return toIssueBriefResponse(i)
}

// ExportToBoardDataResponse конвертирует ent.Board в BoardDataResponse для hubtest.
func ExportToBoardDataResponse(b *ent.Board) *BoardDataResponse {
	return toBoardDataResponse(b)
}

// ExportToIssueLinkResponse конвертирует ent.IssueLink в IssueLinkResponse для hubtest.
func ExportToIssueLinkResponse(l *ent.IssueLink, related *ent.Issue) *IssueLinkResponse {
	return toIssueLinkResponse(l, related)
}

// ExportToCustomFieldResponse конвертирует ent.CustomFieldDefinition в CustomFieldDefinitionResponse для hubtest.
func ExportToCustomFieldResponse(cf *ent.CustomFieldDefinition) *CustomFieldDefinitionResponse {
	return toCustomFieldResponse(cf)
}

// ExportToWorkflowStatusResponse конвертирует ent.WorkflowStatus в WorkflowStatusResponse для hubtest.
func ExportToWorkflowStatusResponse(st *ent.WorkflowStatus) *WorkflowStatusResponse {
	return toWorkflowStatusResponse(st)
}

// ExportBuildFieldChanges строит список изменений полей задачи для hubtest.
func ExportBuildFieldChanges(existing *ent.Issue, req UpdateIssueRequest) []FieldChange {
	return buildFieldChanges(existing, req)
}

// ExportAggregateReactions агрегирует реакции комментария для hubtest.
func ExportAggregateReactions(reactions []*ent.Reaction) []ReactionResponse {
	return aggregateReactions(reactions)
}

// ExportExtractPlainText извлекает plain text из JSON описания для hubtest.
func ExportExtractPlainText(descJSON map[string]interface{}) string {
	return extractPlainText(descJSON)
}

// ExportBuildSwimlanes группирует задачи по swimlane для hubtest.
func ExportBuildSwimlanes(statuses []*ent.WorkflowStatus, swimlane string) []SwimlaneGroup {
	return buildSwimlanes(statuses, swimlane)
}

// ExportTrackerServiceWorkflow возвращает workflow-движок для hubtest.
func ExportTrackerServiceWorkflow(s *TrackerService) *WorkflowEngine {
	return s.workflow
}

// ExportTrackerServiceMirrorIssueCommentsToThreads вызывает mirrorIssueCommentsToThreads для hubtest.
func ExportTrackerServiceMirrorIssueCommentsToThreads(s *TrackerService, ctx context.Context, issueID, userID uuid.UUID, bodyText string) {
	s.mirrorIssueCommentsToThreads(ctx, issueID, userID, bodyText)
}

// ExportConfigString читает строковое значение из map конфигурации для hubtest.
func ExportConfigString(m map[string]interface{}, key string) string {
	return configString(m, key)
}

// ExportSwimlaneKeyLabel возвращает ключ и подпись swimlane для hubtest.
func ExportSwimlaneKeyLabel(iss *ent.Issue, swimlane string) (string, string) {
	return swimlaneKeyLabel(iss, swimlane)
}

// ExportLogOutboxPublishErr — P3-BE-01 / TASK-057: slog при ошибке post-commit Publish.
func ExportLogOutboxPublishErr(ctx context.Context, eventType string, err error) {
	logOutboxPublishErr(ctx, eventType, err)
}
