package tracker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// AutomationEngine выполняет включённые правила по событиям трекера (PROD-01).
type AutomationEngine struct {
	repo      Repository
	messenger ChannelMessenger
	notifier  AutomationNotifier
}

// NewAutomationEngine создаёт движок выполнения правил.
func NewAutomationEngine(repo Repository, messenger ChannelMessenger, notifier AutomationNotifier) *AutomationEngine {
	return &AutomationEngine{repo: repo, messenger: messenger, notifier: notifier}
}

// HandleStatusChanged запускает правила при смене статуса задачи.
// A2W1-A3-001 / TASK-088: ошибка ListEnabledAutomationRules пробрасывается → consumer Nak.
func (e *AutomationEngine) HandleStatusChanged(ctx context.Context, evt IssueMovedEvent) error {
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return nil
	}
	rules, err := e.repo.ListEnabledAutomationRules(ctx, projectID, TriggerStatusChanged)
	if err != nil {
		slog.Error("automation: list rules", "error", err)
		return fmt.Errorf("list automation rules: %w", err)
	}
	for _, rule := range rules {
		if configString(rule.TriggerConfig, "to_status_id") != evt.NewStatusID {
			continue
		}
		if err := e.executeRule(ctx, rule, evt.IssueKey, evt.IssueID, evt.ProjectID, evt.MovedByID, ""); err != nil {
			return err
		}
	}
	return nil
}

// HandleIssueCreated запускает правила при создании задачи.
func (e *AutomationEngine) HandleIssueCreated(ctx context.Context, evt IssueCreatedEvent) error {
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return nil
	}
	rules, err := e.repo.ListEnabledAutomationRules(ctx, projectID, TriggerIssueCreated)
	if err != nil {
		slog.Error("automation: list rules", "error", err)
		return fmt.Errorf("list automation rules: %w", err)
	}
	for _, rule := range rules {
		filterType := configString(rule.TriggerConfig, "issue_type")
		if filterType != "" && filterType != evt.IssueType {
			continue
		}
		if err := e.executeRule(ctx, rule, evt.IssueKey, evt.IssueID, evt.ProjectID, evt.CreatedByID, evt.Title); err != nil {
			return err
		}
	}
	return nil
}

// HandleIssueAssigned запускает правила при смене исполнителя.
func (e *AutomationEngine) HandleIssueAssigned(ctx context.Context, evt IssueAssignedEvent) error {
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return nil
	}
	rules, err := e.repo.ListEnabledAutomationRules(ctx, projectID, TriggerIssueAssigned)
	if err != nil {
		slog.Error("automation: list rules", "error", err)
		return fmt.Errorf("list automation rules: %w", err)
	}
	for _, rule := range rules {
		if err := e.executeRule(ctx, rule, evt.IssueKey, evt.IssueID, evt.ProjectID, evt.AssignedByID, evt.AssigneeID); err != nil {
			return err
		}
	}
	return nil
}

// executeRule выполняет действие правила. A3W1-A3-002: ошибка Post/Notify
// пробрасывается → consumer Nak (не silent Ack).
func (e *AutomationEngine) executeRule(ctx context.Context, rule *ent.AutomationRule, issueKey, issueID, projectID, actorID, extra string) error {
	switch string(rule.ActionType) {
	case ActionPostToChannel:
		if e.messenger == nil {
			return nil
		}
		channelIDStr := configString(rule.ActionConfig, "channel_id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			slog.Warn("automation: invalid channel_id", "rule_id", rule.ID)
			return nil
		}
		senderID := rule.CreatedByID
		if actorUUID, err := uuid.Parse(actorID); err == nil {
			senderID = actorUUID
		}
		template := configString(rule.ActionConfig, "message_template")
		if template == "" {
			template = "Задача {{issue_key}} обновлена автоматизацией «{{rule_name}}»"
		}
		text := renderAutomationTemplate(template, issueKey, rule.Name, extra)
		if err := e.messenger.PostAutomationMessage(ctx, channelID, senderID, text); err != nil {
			slog.Warn("automation: post message failed", "rule_id", rule.ID, "error", err)
			return fmt.Errorf("automation post message: %w", err)
		}
	case ActionNotifyUser:
		if e.notifier == nil {
			return nil
		}
		targetIDStr := extra
		if targetIDStr == "" {
			return nil
		}
		userID, err := uuid.Parse(targetIDStr)
		if err != nil {
			return nil
		}
		projUUID, _ := uuid.Parse(projectID)
		var projPtr *uuid.UUID
		if projUUID != uuid.Nil {
			projPtr = &projUUID
		}
		var actorPtr *uuid.UUID
		if actorUUID, err := uuid.Parse(actorID); err == nil {
			actorPtr = &actorUUID
		}
		title := fmt.Sprintf("Автоматизация: %s", rule.Name)
		body := fmt.Sprintf("Задача %s — сработало правило «%s»", issueKey, rule.Name)
		if err := e.notifier.NotifyAutomation(ctx, userID, title, body, "issue", issueID, projPtr, actorPtr); err != nil {
			slog.Warn("automation: notify failed", "rule_id", rule.ID, "error", err)
			return fmt.Errorf("automation notify: %w", err)
		}
	}
	return nil
}

func renderAutomationTemplate(tmpl, issueKey, ruleName, extra string) string {
	out := strings.ReplaceAll(tmpl, "{{issue_key}}", issueKey)
	out = strings.ReplaceAll(out, "{{rule_name}}", ruleName)
	out = strings.ReplaceAll(out, "{{title}}", extra)
	return out
}
