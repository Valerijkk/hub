package tracker

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ChannelMessenger публикует сообщения автоматизации в каналы проекта.
type ChannelMessenger interface {
	PostAutomationMessage(ctx context.Context, channelID, senderID uuid.UUID, text string) error
}

// AutomationNotifier доставляет in-app уведомления при срабатывании автоматизаций.
type AutomationNotifier interface {
	NotifyAutomation(ctx context.Context, userID uuid.UUID, title, body, entityType, entityID string, projectID, actorID *uuid.UUID) error
}

// AutomationService управляет CRUD правил workflow-автоматизации (PROD-01).
type AutomationService struct {
	repo Repository
}

// NewAutomationService создаёт сервис автоматизаций.
func NewAutomationService(repo Repository) *AutomationService {
	return &AutomationService{repo: repo}
}

// ListRules возвращает все правила автоматизации проекта.
func (s *AutomationService) ListRules(ctx context.Context, projectID uuid.UUID) ([]AutomationRuleResponse, error) {
	rules, err := s.repo.ListAutomationRules(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to list automation rules")
	}
	out := make([]AutomationRuleResponse, 0, len(rules))
	for _, r := range rules {
		out = append(out, toAutomationRuleResponse(r))
	}
	return out, nil
}

// CreateRule создаёт правило автоматизации (макс. 3 на проект — лимит PROD-01).
func (s *AutomationService) CreateRule(ctx context.Context, projectID, userID uuid.UUID, req CreateAutomationRuleRequest) (*AutomationRuleResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, apperrors.NewBadRequest("name is required")
	}
	if err := validateAutomationRuleRequest(req.TriggerType, req.ActionType, req.TriggerConfig, req.ActionConfig); err != nil {
		return nil, apperrors.NewBadRequest(err.Error())
	}
	// A3W1-2-003: channel_id только из этого проекта (иначе пост в чужой канал).
	if err := s.ensureActionChannelInProject(ctx, projectID, req.ActionType, req.ActionConfig); err != nil {
		return nil, err
	}
	count, err := s.repo.CountAutomationRules(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("failed to count rules")
	}
	if count >= maxAutomationRulesPerProject {
		return nil, apperrors.NewBadRequest(fmt.Sprintf("maximum %d automation rules per project", maxAutomationRulesPerProject))
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	rule, err := s.repo.CreateAutomationRule(ctx, projectID, userID, CreateAutomationRuleInput{
		Name:          strings.TrimSpace(req.Name),
		Enabled:       enabled,
		TriggerType:   req.TriggerType,
		TriggerConfig: req.TriggerConfig,
		ActionType:    req.ActionType,
		ActionConfig:  req.ActionConfig,
	})
	if err != nil {
		return nil, apperrors.NewInternal("failed to create automation rule")
	}
	resp := toAutomationRuleResponse(rule)
	return &resp, nil
}

// UpdateRule обновляет существующее правило автоматизации.
func (s *AutomationService) UpdateRule(ctx context.Context, projectID, ruleID uuid.UUID, req UpdateAutomationRuleRequest) (*AutomationRuleResponse, error) {
	existing, err := s.repo.FindAutomationRule(ctx, projectID, ruleID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("automation rule not found")
		}
		return nil, apperrors.NewInternal("failed to load rule")
	}
	triggerType := string(existing.TriggerType)
	actionType := string(existing.ActionType)
	triggerCfg := existing.TriggerConfig
	actionCfg := existing.ActionConfig
	if req.TriggerConfig != nil {
		triggerCfg = req.TriggerConfig
	}
	if req.ActionConfig != nil {
		actionCfg = req.ActionConfig
	}
	if err := validateAutomationRuleRequest(triggerType, actionType, triggerCfg, actionCfg); err != nil {
		return nil, apperrors.NewBadRequest(err.Error())
	}
	// A3W1-2-003: channel_id только из этого проекта.
	if err := s.ensureActionChannelInProject(ctx, projectID, actionType, actionCfg); err != nil {
		return nil, err
	}
	updated, err := s.repo.UpdateAutomationRule(ctx, projectID, ruleID, UpdateAutomationRuleInput{
		Name:          req.Name,
		Enabled:       req.Enabled,
		TriggerConfig: req.TriggerConfig,
		ActionConfig:  req.ActionConfig,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("automation rule not found")
		}
		return nil, apperrors.NewInternal("failed to update rule")
	}
	resp := toAutomationRuleResponse(updated)
	return &resp, nil
}

// DeleteRule удаляет правило автоматизации.
func (s *AutomationService) DeleteRule(ctx context.Context, projectID, ruleID uuid.UUID) error {
	if err := s.repo.DeleteAutomationRule(ctx, projectID, ruleID); err != nil {
		if ent.IsNotFound(err) || err.Error() == "automation rule not found" {
			return apperrors.NewNotFound("automation rule not found")
		}
		return apperrors.NewInternal("failed to delete rule")
	}
	return nil
}

// ensureActionChannelInProject проверяет, что channel_id ∈ project при post_to_channel (A3W1-2-003).
func (s *AutomationService) ensureActionChannelInProject(
	ctx context.Context,
	projectID uuid.UUID,
	actionType string,
	actionCfg map[string]interface{},
) error {
	if actionType != ActionPostToChannel {
		return nil
	}
	raw := configString(actionCfg, "channel_id")
	channelID, err := uuid.Parse(raw)
	if err != nil {
		return apperrors.NewBadRequest("action_config.channel_id invalid")
	}
	if _, err := s.repo.FindChannelInProject(ctx, projectID, channelID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("канал не найден")
		}
		return apperrors.NewInternal("ошибка проверки канала")
	}
	return nil
}

func validateAutomationRuleRequest(triggerType, actionType string, triggerCfg, actionCfg map[string]interface{}) error {
	switch triggerType {
	case TriggerStatusChanged, TriggerIssueCreated, TriggerIssueAssigned:
	default:
		return fmt.Errorf("invalid trigger_type")
	}
	switch actionType {
	case ActionPostToChannel, ActionNotifyUser:
	default:
		return fmt.Errorf("invalid action_type")
	}
	if actionType == ActionPostToChannel {
		if configString(actionCfg, "channel_id") == "" {
			return fmt.Errorf("action_config.channel_id required")
		}
	}
	if triggerType == TriggerStatusChanged {
		if configString(triggerCfg, "to_status_id") == "" {
			return fmt.Errorf("trigger_config.to_status_id required for status_changed")
		}
	}
	if strings.TrimSpace(triggerType) == "" || strings.TrimSpace(actionType) == "" {
		return fmt.Errorf("trigger_type and action_type required")
	}
	return nil
}

func configString(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
