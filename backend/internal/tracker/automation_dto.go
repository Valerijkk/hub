package tracker

import (
	"time"

	"github.com/valerii/hub/backend/ent"
)

func toAutomationRuleResponse(r *ent.AutomationRule) AutomationRuleResponse {
	if r == nil {
		return AutomationRuleResponse{}
	}
	triggerCfg := r.TriggerConfig
	if triggerCfg == nil {
		triggerCfg = map[string]interface{}{}
	}
	actionCfg := r.ActionConfig
	if actionCfg == nil {
		actionCfg = map[string]interface{}{}
	}
	return AutomationRuleResponse{
		ID:            r.ID.String(),
		ProjectID:     r.ProjectID.String(),
		Name:          r.Name,
		Enabled:       r.Enabled,
		TriggerType:   string(r.TriggerType),
		TriggerConfig: triggerCfg,
		ActionType:    string(r.ActionType),
		ActionConfig:  actionCfg,
		CreatedByID:   r.CreatedByID.String(),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}

const maxAutomationRulesPerProject = 3

// Типы триггеров автоматизации (PROD-01).
const (
	TriggerStatusChanged = "status_changed"
	TriggerIssueCreated  = "issue_created"
	TriggerIssueAssigned = "issue_assigned"
)

// Типы действий автоматизации (PROD-01).
const (
	ActionPostToChannel = "post_to_channel"
	ActionNotifyUser    = "notify_user"
)

// AutomationRuleResponse — DTO правила автоматизации для API.
type AutomationRuleResponse struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"project_id"`
	Name          string                 `json:"name"`
	Enabled       bool                   `json:"enabled"`
	TriggerType   string                 `json:"trigger_type"`
	TriggerConfig map[string]interface{} `json:"trigger_config"`
	ActionType    string                 `json:"action_type"`
	ActionConfig  map[string]interface{} `json:"action_config"`
	CreatedByID   string                 `json:"created_by_id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateAutomationRuleRequest — тело POST-запроса на создание правила.
type CreateAutomationRuleRequest struct {
	Name          string                 `json:"name"`
	Enabled       *bool                  `json:"enabled"`
	TriggerType   string                 `json:"trigger_type"`
	TriggerConfig map[string]interface{} `json:"trigger_config"`
	ActionType    string                 `json:"action_type"`
	ActionConfig  map[string]interface{} `json:"action_config"`
}

// UpdateAutomationRuleRequest — тело PATCH-запроса на обновление правила.
type UpdateAutomationRuleRequest struct {
	Name          *string                `json:"name"`
	Enabled       *bool                  `json:"enabled"`
	TriggerConfig map[string]interface{} `json:"trigger_config"`
	ActionConfig  map[string]interface{} `json:"action_config"`
}
