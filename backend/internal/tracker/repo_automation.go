package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/automationrule"
)

func (r *EntRepository) ListAutomationRules(ctx context.Context, projectID uuid.UUID) ([]*ent.AutomationRule, error) {
	return r.client.AutomationRule.Query().
		Where(automationrule.ProjectIDEQ(projectID)).
		Order(ent.Asc(automationrule.FieldCreatedAt)).
		All(ctx)
}

func (r *EntRepository) CountAutomationRules(ctx context.Context, projectID uuid.UUID) (int, error) {
	return r.client.AutomationRule.Query().
		Where(automationrule.ProjectIDEQ(projectID)).
		Count(ctx)
}

func (r *EntRepository) CreateAutomationRule(ctx context.Context, projectID, createdBy uuid.UUID, input CreateAutomationRuleInput) (*ent.AutomationRule, error) {
	triggerCfg := input.TriggerConfig
	if triggerCfg == nil {
		triggerCfg = map[string]interface{}{}
	}
	actionCfg := input.ActionConfig
	if actionCfg == nil {
		actionCfg = map[string]interface{}{}
	}
	return r.client.AutomationRule.Create().
		SetProjectID(projectID).
		SetCreatedByID(createdBy).
		SetName(input.Name).
		SetEnabled(input.Enabled).
		SetTriggerType(automationrule.TriggerType(input.TriggerType)).
		SetTriggerConfig(triggerCfg).
		SetActionType(automationrule.ActionType(input.ActionType)).
		SetActionConfig(actionCfg).
		Save(ctx)
}

func (r *EntRepository) FindAutomationRule(ctx context.Context, projectID, ruleID uuid.UUID) (*ent.AutomationRule, error) {
	return r.client.AutomationRule.Query().
		Where(
			automationrule.IDEQ(ruleID),
			automationrule.ProjectIDEQ(projectID),
		).
		Only(ctx)
}

func (r *EntRepository) UpdateAutomationRule(ctx context.Context, ruleID uuid.UUID, input UpdateAutomationRuleInput) (*ent.AutomationRule, error) {
	upd := r.client.AutomationRule.UpdateOneID(ruleID)
	if input.Name != nil {
		upd.SetName(*input.Name)
	}
	if input.Enabled != nil {
		upd.SetEnabled(*input.Enabled)
	}
	if input.TriggerConfig != nil {
		upd.SetTriggerConfig(input.TriggerConfig)
	}
	if input.ActionConfig != nil {
		upd.SetActionConfig(input.ActionConfig)
	}
	return upd.Save(ctx)
}

func (r *EntRepository) DeleteAutomationRule(ctx context.Context, projectID, ruleID uuid.UUID) error {
	n, err := r.client.AutomationRule.Delete().
		Where(
			automationrule.IDEQ(ruleID),
			automationrule.ProjectIDEQ(projectID),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("automation rule not found")
	}
	return nil
}

func (r *EntRepository) ListEnabledAutomationRules(ctx context.Context, projectID uuid.UUID, triggerType string) ([]*ent.AutomationRule, error) {
	return r.client.AutomationRule.Query().
		Where(
			automationrule.ProjectIDEQ(projectID),
			automationrule.EnabledEQ(true),
			automationrule.TriggerTypeEQ(automationrule.TriggerType(triggerType)),
		).
		All(ctx)
}

// CreateAutomationRuleInput — входные данные репозитория для создания правила.
type CreateAutomationRuleInput struct {
	Name          string
	Enabled       bool
	TriggerType   string
	TriggerConfig map[string]interface{}
	ActionType    string
	ActionConfig  map[string]interface{}
}

// UpdateAutomationRuleInput — входные данные репозитория для обновления правила.
type UpdateAutomationRuleInput struct {
	Name          *string
	Enabled       *bool
	TriggerConfig map[string]interface{}
	ActionConfig  map[string]interface{}
}
