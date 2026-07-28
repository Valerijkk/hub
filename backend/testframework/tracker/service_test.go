// Package tracker содержит юнит-тесты для tracker.TrackerService.
// Покрывает: создание задач (все 9 типов), оптимистичную блокировку,
// workflow-валидацию, курсорную пагинацию, связи, метки, доски, кастомные поля.
package tracker_test

// Тесты сервиса трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/customfielddefinition"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Мок tracker.Repository
// ---------------------------------------------------------------------------

// mockRepo реализует tracker.Repository для тестирования tracker.TrackerService.
type mockRepo struct {
	issues          map[uuid.UUID]*ent.Issue
	issuesByKey     map[string]*ent.Issue
	issueLinks      map[uuid.UUID]*ent.IssueLink
	statuses        map[uuid.UUID]*ent.WorkflowStatus
	transitions     map[string]*ent.WorkflowTransition // ключ: "fromID:toID"
	comments        map[uuid.UUID]*ent.Comment
	activities      []*ent.ActivityLog
	timeLogs        map[uuid.UUID]*ent.TimeLog
	labels          map[uuid.UUID]*ent.Label
	labelGroups     map[uuid.UUID]*ent.LabelGroup
	boards          map[uuid.UUID]*ent.Board
	customFields    map[uuid.UUID]*ent.CustomFieldDefinition
	projectCounters map[uuid.UUID]int
	issueMessageLinks []*ent.IssueMessageLink

	// Для проверки update с оптимистичной блокировкой
	updateVersionError bool
}

func newMockRepo() *mockRepo {
	defaultStatusID := uuid.New()
	defaultStatus := &ent.WorkflowStatus{
		ID:        defaultStatusID,
		Name:      "To Do",
		Category:  workflowstatus.CategoryTodo,
		Color:     "#6B7280",
		SortOrder: 0,
		IsDefault: true,
	}

	return &mockRepo{
		issues:      make(map[uuid.UUID]*ent.Issue),
		issuesByKey: make(map[string]*ent.Issue),
		issueLinks:  make(map[uuid.UUID]*ent.IssueLink),
		statuses: map[uuid.UUID]*ent.WorkflowStatus{
			defaultStatusID: defaultStatus,
		},
		transitions:     make(map[string]*ent.WorkflowTransition),
		comments:        make(map[uuid.UUID]*ent.Comment),
		timeLogs:        make(map[uuid.UUID]*ent.TimeLog),
		labels:          make(map[uuid.UUID]*ent.Label),
		labelGroups:     make(map[uuid.UUID]*ent.LabelGroup),
		boards:          make(map[uuid.UUID]*ent.Board),
		customFields:    make(map[uuid.UUID]*ent.CustomFieldDefinition),
		projectCounters: make(map[uuid.UUID]int),
	}
}

func (m *mockRepo) CreateIssue(_ context.Context, projectID uuid.UUID, input tracker.CreateIssueInput) (*ent.Issue, error) {
	m.projectCounters[projectID]++
	num := m.projectCounters[projectID]
	key := fmt.Sprintf("TEST-%d", num)

	// Находим дефолтный статус
	var status *ent.WorkflowStatus
	if input.StatusID != nil {
		status = m.statuses[*input.StatusID]
	} else {
		for _, st := range m.statuses {
			if st.IsDefault {
				status = st
				break
			}
		}
	}

	id := uuid.New()
	i := &ent.Issue{
		ID:              id,
		IssueNumber:     num,
		IssueKey:        key,
		Title:           input.Title,
		IssueType:       issue.IssueType(input.IssueType),
		Priority:        issue.Priority(input.Priority),
		DescriptionJSON: input.DescriptionJSON,
		DescriptionText: input.DescriptionText,
		CustomFields:    input.CustomFields,
		EstimateMinutes: input.EstimateMinutes,
		DueDate:         input.DueDate,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Edges: ent.IssueEdges{
			Project:  &ent.Project{ID: projectID},
			Status:   status,
			Reporter: &ent.User{ID: input.ReporterID, Name: "Test User", Email: "test@test.com"},
		},
	}
	if input.AssigneeID != nil {
		i.Edges.Assignee = &ent.User{ID: *input.AssigneeID, Name: "Assignee", Email: "assignee@test.com"}
	}
	if input.ParentID != nil {
		i.Edges.Parent = &ent.Issue{ID: *input.ParentID, IssueKey: "TEST-PARENT", Title: "Parent"}
	}

	m.issues[id] = i
	m.issuesByKey[key] = i
	return i, nil
}

func (m *mockRepo) FindIssueByKey(_ context.Context, projectID uuid.UUID, issueKey string) (*ent.Issue, error) {
	i, ok := m.issuesByKey[issueKey]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if i.Edges.Project == nil || i.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return i, nil
}

func (m *mockRepo) FindIssueByID(_ context.Context, id uuid.UUID) (*ent.Issue, error) {
	i, ok := m.issues[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return i, nil
}

func (m *mockRepo) FindIssueInProject(_ context.Context, projectID uuid.UUID, id uuid.UUID) (*ent.Issue, error) {
	i, ok := m.issues[id]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if i.Edges.Project == nil || i.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return i, nil
}

func (m *mockRepo) ListIssues(_ context.Context, _ uuid.UUID, params tracker.ListIssuesParams) ([]*ent.Issue, int, error) {
	var result []*ent.Issue
	for _, i := range m.issues {
		// Простая фильтрация по типу
		if params.IssueType != "" && string(i.IssueType) != params.IssueType {
			continue
		}
		result = append(result, i)
	}

	total := len(result)

	// Курсорная пагинация
	if params.Cursor != "" {
		var filtered []*ent.Issue
		found := false
		for _, i := range result {
			if found {
				filtered = append(filtered, i)
			}
			if i.IssueKey == params.Cursor {
				found = true
			}
		}
		result = filtered
	}

	// Лимит
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if len(result) > limit {
		result = result[:limit]
	}

	return result, total, nil
}

func (m *mockRepo) UpdateIssue(_ context.Context, issueID uuid.UUID, input tracker.UpdateIssueInput) (*ent.Issue, error) {
	i, ok := m.issues[issueID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}

	if m.updateVersionError || i.Version != input.Version {
		return nil, apperrors.NewConflict("version mismatch")
	}

	if input.Title != nil {
		i.Title = *input.Title
	}
	if input.Priority != nil {
		i.Priority = issue.Priority(*input.Priority)
	}
	if input.StatusID != nil {
		if st, ok := m.statuses[*input.StatusID]; ok {
			i.Edges.Status = st
		}
	}
	i.Version++
	i.UpdatedAt = time.Now()
	return i, nil
}

func (m *mockRepo) DeleteIssue(_ context.Context, issueID uuid.UUID) error {
	i, ok := m.issues[issueID]
	if !ok {
		return &ent.NotFoundError{}
	}
	delete(m.issuesByKey, i.IssueKey)
	delete(m.issues, issueID)
	return nil
}

func (m *mockRepo) MoveIssue(_ context.Context, issueID uuid.UUID, statusID uuid.UUID, rank string, expectedVersion int) (*ent.Issue, error) {
	i, ok := m.issues[issueID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if expectedVersion > 0 && i.Version != expectedVersion {
		return nil, apperrors.NewConflict("version mismatch")
	}
	if st, ok := m.statuses[statusID]; ok {
		i.Edges.Status = st
	}
	i.Rank = rank
	i.Version++
	return i, nil
}

func (m *mockRepo) CreateIssueLink(_ context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error) {
	id := uuid.New()
	link := &ent.IssueLink{
		ID:         id,
		LinkType:   issuelink.LinkType(linkType),
		CustomType: customType,
		CreatedAt:  time.Now(),
		Edges: ent.IssueLinkEdges{
			Source: m.issues[sourceID],
			Target: m.issues[targetID],
		},
	}
	m.issueLinks[id] = link
	return link, nil
}

func (m *mockRepo) ListIssueLinks(_ context.Context, issueID uuid.UUID) ([]*ent.IssueLink, error) {
	var result []*ent.IssueLink
	for _, l := range m.issueLinks {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockRepo) ListProjectIssueLinks(_ context.Context, _ uuid.UUID) ([]*ent.IssueLink, error) {
	result := make([]*ent.IssueLink, 0, len(m.issueLinks))
	for _, l := range m.issueLinks {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockRepo) DeleteIssueLink(_ context.Context, sourceIssueID, linkID uuid.UUID) error {
	link, ok := m.issueLinks[linkID]
	if !ok {
		return &ent.NotFoundError{}
	}
	if link.Edges.Source == nil || link.Edges.Source.ID != sourceIssueID {
		return &ent.NotFoundError{}
	}
	delete(m.issueLinks, linkID)
	return nil
}

func (m *mockRepo) ListWorkflowStatuses(_ context.Context, _ uuid.UUID) ([]*ent.WorkflowStatus, error) {
	var result []*ent.WorkflowStatus
	for _, st := range m.statuses {
		result = append(result, st)
	}
	return result, nil
}

func (m *mockRepo) CreateWorkflowStatus(_ context.Context, _ uuid.UUID, name, category, color string, sortOrder int) (*ent.WorkflowStatus, error) {
	id := uuid.New()
	st := &ent.WorkflowStatus{
		ID:        id,
		Name:      name,
		Category:  workflowstatus.Category(category),
		Color:     color,
		SortOrder: sortOrder,
	}
	m.statuses[id] = st
	return st, nil
}

func (m *mockRepo) UpdateWorkflowStatus(_ context.Context, _, statusID uuid.UUID, name, color *string, sortOrder *int) (*ent.WorkflowStatus, error) {
	st, ok := m.statuses[statusID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if name != nil {
		st.Name = *name
	}
	if color != nil {
		st.Color = *color
	}
	if sortOrder != nil {
		st.SortOrder = *sortOrder
	}
	return st, nil
}

func (m *mockRepo) DeleteWorkflowStatus(_ context.Context, _, statusID uuid.UUID) error {
	delete(m.statuses, statusID)
	return nil
}

func (m *mockRepo) FindDefaultStatus(_ context.Context, _ uuid.UUID) (*ent.WorkflowStatus, error) {
	for _, st := range m.statuses {
		if st.IsDefault {
			return st, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (m *mockRepo) FindWorkflowStatusInProject(_ context.Context, _, statusID uuid.UUID) (*ent.WorkflowStatus, error) {
	st, ok := m.statuses[statusID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return st, nil
}

func (m *mockRepo) ListWorkflowTransitions(_ context.Context, _ uuid.UUID) ([]*ent.WorkflowTransition, error) {
	var result []*ent.WorkflowTransition
	for _, t := range m.transitions {
		result = append(result, t)
	}
	return result, nil
}

func (m *mockRepo) CreateWorkflowTransition(_ context.Context, _, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error) {
	id := uuid.New()
	t := &ent.WorkflowTransition{
		ID:        id,
		CreatedAt: time.Now(),
		Edges: ent.WorkflowTransitionEdges{
			FromStatus: m.statuses[fromStatusID],
			ToStatus:   m.statuses[toStatusID],
		},
	}
	key := fromStatusID.String() + ":" + toStatusID.String()
	m.transitions[key] = t
	return t, nil
}

func (m *mockRepo) DeleteWorkflowTransition(_ context.Context, _, transitionID uuid.UUID) error {
	for key, t := range m.transitions {
		if t.ID == transitionID {
			delete(m.transitions, key)
			return nil
		}
	}
	return nil
}

func (m *mockRepo) FindTransition(_ context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error) {
	key := fromStatusID.String() + ":" + toStatusID.String()
	t, ok := m.transitions[key]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	return t, nil
}

func (m *mockRepo) SeedDefaultWorkflow(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockRepo) CreateComment(_ context.Context, issueID, authorID uuid.UUID, input tracker.CreateCommentInput) (*ent.Comment, error) {
	id := uuid.New()
	mentions := make([]uuid.UUID, 0, len(input.Mentions))
	for _, raw := range input.Mentions {
		mid, err := uuid.Parse(raw)
		if err != nil {
			return nil, err
		}
		mentions = append(mentions, mid)
	}
	c := &ent.Comment{
		ID:         id,
		BodyJSON:   input.BodyJSON,
		BodyText:   input.BodyText,
		IsInternal: input.IsInternal,
		Mentions:   mentions,
		Version:    1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Edges: ent.CommentEdges{
			Author: &ent.User{ID: authorID, Name: "Test User", Email: "test@test.com"},
			Issue:  &ent.Issue{ID: issueID},
		},
	}
	m.comments[id] = c
	return c, nil
}

func (m *mockRepo) ListComments(_ context.Context, issueID uuid.UUID) ([]*ent.Comment, error) {
	var result []*ent.Comment
	for _, c := range m.comments {
		if c.Edges.Issue != nil && c.Edges.Issue.ID != issueID {
			continue
		}
		result = append(result, c)
	}
	return result, nil
}

func (m *mockRepo) UpdateComment(_ context.Context, issueID, authorID, commentID uuid.UUID, input tracker.UpdateCommentInput) (*ent.Comment, error) {
	c, ok := m.comments[commentID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}
	if c.Edges.Issue == nil || c.Edges.Issue.ID != issueID {
		return nil, &ent.NotFoundError{}
	}
	if c.Edges.Author == nil || c.Edges.Author.ID != authorID {
		return nil, &ent.NotFoundError{}
	}
	if c.Version == 0 {
		c.Version = 1
	}
	if c.Version != input.Version {
		return nil, &ent.NotFoundError{}
	}
	c.BodyJSON = input.BodyJSON
	c.BodyText = input.BodyText
	c.Version++
	return c, nil
}

func (m *mockRepo) DeleteComment(_ context.Context, issueID, authorID, commentID uuid.UUID) error {
	c, ok := m.comments[commentID]
	if !ok {
		return &ent.NotFoundError{}
	}
	if c.Edges.Issue == nil || c.Edges.Issue.ID != issueID {
		return &ent.NotFoundError{}
	}
	if c.Edges.Author == nil || c.Edges.Author.ID != authorID {
		return &ent.NotFoundError{}
	}
	delete(m.comments, commentID)
	return nil
}

func (m *mockRepo) ToggleReaction(_ context.Context, issueID, commentID, _ uuid.UUID, _ string) (bool, error) {
	c, ok := m.comments[commentID]
	if !ok {
		return false, &ent.NotFoundError{}
	}
	if c.Edges.Issue == nil || c.Edges.Issue.ID != issueID {
		return false, &ent.NotFoundError{}
	}
	return true, nil
}

func (m *mockRepo) LogActivity(_ context.Context, issueID, userID uuid.UUID, action string, changes []tracker.FieldChange) (*ent.ActivityLog, error) {
	schemaChanges := make([]schema.FieldChange, len(changes))
	for i, c := range changes {
		schemaChanges[i] = schema.FieldChange{Field: c.Field, OldValue: c.OldValue, NewValue: c.NewValue}
	}
	a := &ent.ActivityLog{
		ID:        uuid.New(),
		Action:    action,
		Changes:   schemaChanges,
		CreatedAt: time.Now(),
		Edges: ent.ActivityLogEdges{
			User: &ent.User{ID: userID, Name: "Test User", Email: "test@test.com"},
		},
	}
	m.activities = append(m.activities, a)
	return a, nil
}

func (m *mockRepo) ListActivities(_ context.Context, _ uuid.UUID) ([]*ent.ActivityLog, error) {
	return m.activities, nil
}

func (m *mockRepo) LogTime(_ context.Context, issueID, userID uuid.UUID, minutes int, description string) (*ent.TimeLog, error) {
	id := uuid.New()
	tl := &ent.TimeLog{
		ID:          id,
		Minutes:     minutes,
		Description: description,
		CreatedAt:   time.Now(),
		Edges: ent.TimeLogEdges{
			User: &ent.User{ID: userID, Name: "Test User", Email: "test@test.com"},
		},
	}
	m.timeLogs[id] = tl
	return tl, nil
}

func (m *mockRepo) StartTimer(_ context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error) {
	id := uuid.New()
	now := time.Now()
	tl := &ent.TimeLog{
		ID:        id,
		Minutes:   1,
		StartedAt: &now,
		CreatedAt: time.Now(),
		Edges: ent.TimeLogEdges{
			User: &ent.User{ID: userID, Name: "Test User", Email: "test@test.com"},
		},
	}
	m.timeLogs[id] = tl
	return tl, nil
}

func (m *mockRepo) StopTimer(_ context.Context, _, _ uuid.UUID) (*ent.TimeLog, error) {
	now := time.Now()
	for _, tl := range m.timeLogs {
		if tl.StartedAt != nil && tl.EndedAt == nil {
			tl.EndedAt = &now
			tl.Minutes = 5
			return tl, nil
		}
	}
	return nil, apperrors.NewNotFound("активный таймер не найден")
}

func (m *mockRepo) ListTimeLogs(_ context.Context, _ uuid.UUID) ([]*ent.TimeLog, error) {
	var result []*ent.TimeLog
	for _, tl := range m.timeLogs {
		result = append(result, tl)
	}
	return result, nil
}

func (m *mockRepo) CreateLabel(_ context.Context, _ uuid.UUID, name, color string, groupID *uuid.UUID) (*ent.Label, error) {
	id := uuid.New()
	l := &ent.Label{
		ID:        id,
		Name:      name,
		Color:     color,
		CreatedAt: time.Now(),
	}
	if groupID != nil {
		if g, ok := m.labelGroups[*groupID]; ok {
			l.Edges.Group = g
		}
	}
	m.labels[id] = l
	return l, nil
}

func (m *mockRepo) ListLabels(_ context.Context, _ uuid.UUID) ([]*ent.Label, error) {
	var result []*ent.Label
	for _, l := range m.labels {
		result = append(result, l)
	}
	return result, nil
}

func (m *mockRepo) DeleteLabel(_ context.Context, _, labelID uuid.UUID) error {
	delete(m.labels, labelID)
	return nil
}

func (m *mockRepo) CountLabelsInProject(_ context.Context, _ uuid.UUID, labelIDs []uuid.UUID) (int, error) {
	n := 0
	for _, id := range labelIDs {
		if _, ok := m.labels[id]; ok {
			n++
		}
	}
	return n, nil
}

func (m *mockRepo) CreateLabelGroup(_ context.Context, _ uuid.UUID, name string) (*ent.LabelGroup, error) {
	id := uuid.New()
	g := &ent.LabelGroup{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
	}
	m.labelGroups[id] = g
	return g, nil
}

func (m *mockRepo) ListLabelGroups(_ context.Context, _ uuid.UUID) ([]*ent.LabelGroup, error) {
	var result []*ent.LabelGroup
	for _, g := range m.labelGroups {
		result = append(result, g)
	}
	return result, nil
}

func (m *mockRepo) CreateBoard(_ context.Context, _ uuid.UUID, name string, filters map[string]interface{}) (*ent.Board, error) {
	id := uuid.New()
	b := &ent.Board{
		ID:        id,
		Name:      name,
		Filters:   filters,
		CreatedAt: time.Now(),
	}
	m.boards[id] = b
	return b, nil
}

func (m *mockRepo) ListBoards(_ context.Context, _ uuid.UUID) ([]*ent.Board, error) {
	var result []*ent.Board
	for _, b := range m.boards {
		result = append(result, b)
	}
	return result, nil
}

func (m *mockRepo) GetBoardData(_ context.Context, _, boardID uuid.UUID) (*ent.Board, error) {
	b, ok := m.boards[boardID]
	if !ok {
		return nil, &ent.NotFoundError{}
	}

	// Собираем статусы с задачами
	var statuses []*ent.WorkflowStatus
	for _, st := range m.statuses {
		stCopy := *st
		var statusIssues []*ent.Issue
		for _, i := range m.issues {
			if i.Edges.Status != nil && i.Edges.Status.ID == st.ID {
				statusIssues = append(statusIssues, i)
			}
		}
		stCopy.Edges.Issues = statusIssues
		statuses = append(statuses, &stCopy)
	}

	b.Edges.Project = &ent.Project{
		ID:  uuid.New(),
		Key: "TEST",
		Edges: ent.ProjectEdges{
			WorkflowStatuses: statuses,
		},
	}
	return b, nil
}

func (m *mockRepo) CreateCustomField(_ context.Context, _ uuid.UUID, input tracker.CreateCustomFieldInput) (*ent.CustomFieldDefinition, error) {
	id := uuid.New()
	cf := &ent.CustomFieldDefinition{
		ID:        id,
		Name:      input.Name,
		FieldType: customfielddefinition.FieldType(input.FieldType),
		Options:   input.Options,
		Required:  input.Required,
		CreatedAt: time.Now(),
	}
	m.customFields[id] = cf
	return cf, nil
}

func (m *mockRepo) ListCustomFields(_ context.Context, _ uuid.UUID) ([]*ent.CustomFieldDefinition, error) {
	var result []*ent.CustomFieldDefinition
	for _, cf := range m.customFields {
		result = append(result, cf)
	}
	return result, nil
}

func (m *mockRepo) DeleteCustomField(_ context.Context, _, fieldID uuid.UUID) error {
	delete(m.customFields, fieldID)
	return nil
}

// addStatusWithTransition — вспомогательный метод для добавления статуса и перехода.
func (m *mockRepo) addStatusWithTransition(name string, category workflowstatus.Category, fromStatusID uuid.UUID) uuid.UUID {
	id := uuid.New()
	m.statuses[id] = &ent.WorkflowStatus{
		ID:       id,
		Name:     name,
		Category: category,
		Color:    "#3B82F6",
	}
	key := fromStatusID.String() + ":" + id.String()
	m.transitions[key] = &ent.WorkflowTransition{
		ID: uuid.New(),
		Edges: ent.WorkflowTransitionEdges{
			FromStatus: m.statuses[fromStatusID],
			ToStatus:   m.statuses[id],
		},
	}
	return id
}

// getDefaultStatusID возвращает ID дефолтного статуса.
func (m *mockRepo) getDefaultStatusID() uuid.UUID {
	for id, st := range m.statuses {
		if st.IsDefault {
			return id
		}
	}
	return uuid.Nil
}

func (m *mockRepo) ListAutomationRules(context.Context, uuid.UUID) ([]*ent.AutomationRule, error) {
	return nil, nil
}
func (m *mockRepo) CountAutomationRules(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (m *mockRepo) CreateAutomationRule(context.Context, uuid.UUID, uuid.UUID, tracker.CreateAutomationRuleInput) (*ent.AutomationRule, error) {
	return nil, nil
}
func (m *mockRepo) FindAutomationRule(context.Context, uuid.UUID, uuid.UUID) (*ent.AutomationRule, error) {
	return nil, &ent.NotFoundError{}
}
func (m *mockRepo) UpdateAutomationRule(context.Context, uuid.UUID, uuid.UUID, tracker.UpdateAutomationRuleInput) (*ent.AutomationRule, error) {
	return nil, nil
}
func (m *mockRepo) DeleteAutomationRule(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (m *mockRepo) ListEnabledAutomationRules(context.Context, uuid.UUID, string) ([]*ent.AutomationRule, error) {
	return nil, nil
}

func (m *mockRepo) ListIssueMessageLinksByIssue(_ context.Context, issueID uuid.UUID) ([]*ent.IssueMessageLink, error) {
	var out []*ent.IssueMessageLink
	for _, l := range m.issueMessageLinks {
		if l.IssueID == issueID {
			out = append(out, l)
		}
	}
	return out, nil
}
func (m *mockRepo) CreateIssueMessageLink(_ context.Context, issueID, channelID, createdBy uuid.UUID, messageID string, mirror bool) (*ent.IssueMessageLink, error) {
	link := &ent.IssueMessageLink{
		ID:             uuid.New(),
		IssueID:        issueID,
		ChannelID:      channelID,
		MessageID:      messageID,
		MirrorToThread: mirror,
		CreatedByID:    createdBy,
		CreatedAt:      time.Now(),
	}
	m.issueMessageLinks = append(m.issueMessageLinks, link)
	return link, nil
}
func (m *mockRepo) ListIssueMessageLinksByThread(_ context.Context, channelID uuid.UUID, messageID string) ([]*ent.IssueMessageLink, error) {
	var out []*ent.IssueMessageLink
	for _, l := range m.issueMessageLinks {
		if l.ChannelID == channelID && l.MessageID == messageID {
			out = append(out, l)
		}
	}
	return out, nil
}
func (m *mockRepo) FindIssueMessageLink(_ context.Context, linkID uuid.UUID) (*ent.IssueMessageLink, error) {
	for _, l := range m.issueMessageLinks {
		if l.ID == linkID {
			return l, nil
		}
	}
	return nil, &ent.NotFoundError{}
}
func (m *mockRepo) DeleteIssueMessageLink(_ context.Context, linkID, issueID uuid.UUID) error {
	for i, l := range m.issueMessageLinks {
		if l.ID == linkID && l.IssueID == issueID {
			m.issueMessageLinks = append(m.issueMessageLinks[:i], m.issueMessageLinks[i+1:]...)
			return nil
		}
	}
	return &ent.NotFoundError{}
}

func (m *mockRepo) FindChannelInProject(_ context.Context, _, channelID uuid.UUID) (*ent.Channel, error) {
	return &ent.Channel{ID: channelID}, nil
}

func (m *mockRepo) IsChannelMember(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (m *mockRepo) GetDB() *ent.Client {
	return nil
}

// ---------------------------------------------------------------------------
// Helper для создания сервиса с мок-репозиторием
// ---------------------------------------------------------------------------

func newTestService() (*tracker.TrackerService, *mockRepo) {
	repo := newMockRepo()
	svc := tracker.NewTrackerService(repo, nil)
	return svc, repo
}

// ---------------------------------------------------------------------------
// Тесты
// ---------------------------------------------------------------------------

// TestCreateIssue проверяет сценарий: создание issue.
func TestCreateIssue(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	req := tracker.CreateIssueRequest{
		Title:     "Тестовая задача",
		IssueType: "task",
		Priority:  "medium",
	}

	resp, err := svc.CreateIssue(ctx, projectID, userID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "Тестовая задача", resp.Title)
	assert.Equal(t, "task", resp.IssueType)
	assert.Equal(t, "medium", resp.Priority)
	assert.Equal(t, "TEST-1", resp.IssueKey)
	assert.Equal(t, 1, resp.IssueNumber)
	assert.Equal(t, 1, resp.Version)
	assert.NotEmpty(t, resp.Status.ID)
}

// TestCreateIssueAllTypes проверяет сценарий: создание issue all types.
func TestCreateIssueAllTypes(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issueTypes := []string{
		"bug", "task", "story", "epic", "sub_task",
		"improvement", "tech_debt", "spike", "release",
	}

	for i, it := range issueTypes {
		t.Run(it, func(t *testing.T) {
			req := tracker.CreateIssueRequest{
				Title:     fmt.Sprintf("Issue type: %s", it),
				IssueType: it,
				Priority:  "medium",
			}

			resp, err := svc.CreateIssue(ctx, projectID, userID, req)
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, it, resp.IssueType)
			assert.Equal(t, i+1, resp.IssueNumber)
		})
	}
}

// TestUpdateIssueOptimisticLocking проверяет сценарий: обновление issue optimistic locking.
func TestUpdateIssueOptimisticLocking(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём задачу
	createReq := tracker.CreateIssueRequest{
		Title:     "Original",
		IssueType: "task",
		Priority:  "medium",
	}
	resp, err := svc.CreateIssue(ctx, projectID, userID, createReq)
	require.NoError(t, err)

	// Обновляем с правильной версией
	newTitle := "Updated"
	updateReq := tracker.UpdateIssueRequest{
		Title:   &newTitle,
		Version: 1,
	}
	updated, err := svc.UpdateIssue(ctx, projectID, resp.IssueKey, userID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Title)
	assert.Equal(t, 2, updated.Version)

	// Обновляем со старой версией — должен быть конфликт
	_ = repo
	updateReq2 := tracker.UpdateIssueRequest{
		Title:   &newTitle,
		Version: 1, // старая версия
	}
	_, err = svc.UpdateIssue(ctx, projectID, resp.IssueKey, userID, updateReq2)
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 409, appErr.Code)
}

// TestMoveIssueValidTransition проверяет сценарий: move issue корректный transition.
func TestMoveIssueValidTransition(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём второй статус и переход
	defaultStatusID := repo.getDefaultStatusID()
	inProgressID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultStatusID)

	// Создаём задачу
	createReq := tracker.CreateIssueRequest{
		Title:     "Move test",
		IssueType: "task",
		Priority:  "medium",
	}
	created, err := svc.CreateIssue(ctx, projectID, userID, createReq)
	require.NoError(t, err)

	// Перемещаем в In Progress (разрешённый переход)
	moveReq := tracker.MoveIssueRequest{
		StatusID: inProgressID.String(),
		Rank:     "aaa|bbb",
	}
	moved, err := svc.MoveIssue(ctx, projectID, created.IssueKey, userID, moveReq)
	require.NoError(t, err)
	assert.Equal(t, "In Progress", moved.Status.Name)
	assert.Equal(t, "aaa|bbb", moved.Rank)
}

// TestMoveIssueInvalidTransition проверяет отклонение при move issue некорректный transition.
func TestMoveIssueInvalidTransition(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём задачу
	createReq := tracker.CreateIssueRequest{
		Title:     "Invalid move",
		IssueType: "task",
		Priority:  "medium",
	}
	created, err := svc.CreateIssue(ctx, projectID, userID, createReq)
	require.NoError(t, err)

	// Пытаемся переместить в несуществующий статус (нет перехода)
	moveReq := tracker.MoveIssueRequest{
		StatusID: uuid.New().String(),
		Rank:     "aaa",
	}
	_, err = svc.MoveIssue(ctx, projectID, created.IssueKey, userID, moveReq)
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	// Несуществующий status → 404 (раньше мог быть 400 invalid transition).
	assert.Contains(t, []int{400, 404}, appErr.Code)
}

// TestCreateParentChild проверяет сценарий: создание parent child.
func TestCreateParentChild(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём родительскую задачу
	parentReq := tracker.CreateIssueRequest{
		Title:     "Parent Epic",
		IssueType: "epic",
		Priority:  "high",
	}
	parent, err := svc.CreateIssue(ctx, projectID, userID, parentReq)
	require.NoError(t, err)

	// Создаём дочернюю задачу
	childReq := tracker.CreateIssueRequest{
		Title:     "Child Task",
		IssueType: "task",
		Priority:  "medium",
		ParentID:  parent.ID,
	}
	child, err := svc.CreateIssue(ctx, projectID, userID, childReq)
	require.NoError(t, err)

	assert.NotNil(t, child.Parent)
	assert.Equal(t, parent.ID, child.Parent.ID)
}

// TestCreateIssueLink проверяет сценарий: создание issue link.
func TestCreateIssueLink(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём две задачи
	req1 := tracker.CreateIssueRequest{Title: "Blocker", IssueType: "bug", Priority: "critical"}
	issue1, err := svc.CreateIssue(ctx, projectID, userID, req1)
	require.NoError(t, err)

	req2 := tracker.CreateIssueRequest{Title: "Blocked", IssueType: "task", Priority: "medium"}
	issue2, err := svc.CreateIssue(ctx, projectID, userID, req2)
	require.NoError(t, err)

	// Создаём связь "blocks"
	linkReq := tracker.CreateIssueLinkRequest{
		TargetIssueID: issue2.ID,
		LinkType:      "blocks",
	}
	link, err := svc.CreateIssueLink(ctx, projectID, issue1.IssueKey, userID, linkReq)
	require.NoError(t, err)
	assert.Equal(t, "blocks", link.LinkType)
}

// TestWorkflowValidation проверяет сценарий: workflow валидация.
func TestWorkflowValidation(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()

	defaultID := repo.getDefaultStatusID()
	inProgressID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, defaultID)

	// Переход ToDo -> InProgress разрешён
	err := tracker.ExportTrackerServiceWorkflow(svc).ValidateTransition(ctx, defaultID, inProgressID)
	assert.NoError(t, err)

	// Переход InProgress -> ToDo НЕ разрешён (нет обратного перехода в моке)
	err = tracker.ExportTrackerServiceWorkflow(svc).ValidateTransition(ctx, inProgressID, defaultID)
	assert.Error(t, err)
}

// TestListIssuesPagination проверяет сценарий: список issues pagination.
func TestListIssuesPagination(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём 5 задач
	for i := 0; i < 5; i++ {
		req := tracker.CreateIssueRequest{
			Title:     fmt.Sprintf("Task %d", i+1),
			IssueType: "task",
			Priority:  "medium",
		}
		_, err := svc.CreateIssue(ctx, projectID, userID, req)
		require.NoError(t, err)
	}

	// Запрашиваем с лимитом 2
	listReq := tracker.ListIssuesRequest{Limit: 2}
	resp, err := svc.ListIssues(ctx, projectID, listReq)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(resp.Items), 2)
	assert.Equal(t, 5, resp.Total)
}

// TestCreateLabel проверяет сценарий: создание label.
func TestCreateLabel(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	// Создаём группу
	groupReq := tracker.CreateLabelGroupRequest{Name: "Priority Labels"}
	group, err := svc.CreateLabelGroup(ctx, projectID, groupReq)
	require.NoError(t, err)
	assert.Equal(t, "Priority Labels", group.Name)

	// Создаём метку в группе
	labelReq := tracker.CreateLabelRequest{
		Name:    "urgent",
		Color:   "#FF0000",
		GroupID: group.ID,
	}
	label, err := svc.CreateLabel(ctx, projectID, labelReq)
	require.NoError(t, err)
	assert.Equal(t, "urgent", label.Name)
	assert.Equal(t, "#FF0000", label.Color)

	// Список меток
	labels, err := svc.ListLabels(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, labels, 1)
}

// TestBoardData проверяет сценарий: board data.
func TestBoardData(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём доску
	boardReq := tracker.CreateBoardRequest{
		Name:    "Sprint Board",
		Filters: map[string]interface{}{"assignee": "me"},
	}
	board, err := svc.CreateBoard(ctx, projectID, boardReq)
	require.NoError(t, err)
	assert.Equal(t, "Sprint Board", board.Name)

	// Создаём задачу
	issueReq := tracker.CreateIssueRequest{
		Title:     "Board task",
		IssueType: "task",
		Priority:  "medium",
	}
	_, err = svc.CreateIssue(ctx, projectID, userID, issueReq)
	require.NoError(t, err)

	// Получаем данные доски
	boardID, _ := uuid.Parse(board.ID)
	_ = repo
	data, err := svc.GetBoardData(ctx, projectID, boardID, "")
	require.NoError(t, err)
	assert.Equal(t, "Sprint Board", data.Board.Name)
	assert.NotEmpty(t, data.Columns)
}

// TestCreateIssue_InvalidUUIDs проверяет отклонение при создание issue некорректный uui ds.
func TestCreateIssue_InvalidUUIDs(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		req  tracker.CreateIssueRequest
		err  string
	}{
		{
			name: "invalid assignee_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", AssigneeID: "not-uuid"},
			err:  "assignee_id",
		},
		{
			name: "invalid parent_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", ParentID: "bad"},
			err:  "parent_id",
		},
		{
			name: "invalid status_id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", StatusID: "bad"},
			err:  "status_id",
		},
		{
			name: "invalid label id",
			req:  tracker.CreateIssueRequest{Title: "T", IssueType: "task", Labels: []string{"bad-label"}},
			err:  "label ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateIssue(ctx, projectID, userID, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.err)
		})
	}
}

// TestGetIssue_NotFound проверяет сценарий: получение issue not found.
func TestGetIssue_NotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.GetIssue(ctx, projectID, "MISSING-1")
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 404, appErr.Code)
}

// TestDeleteIssue_NotFound проверяет сценарий: удаление issue not found.
func TestDeleteIssue_NotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	err := svc.DeleteIssue(ctx, projectID, "MISSING-1", uuid.New())
	require.Error(t, err)
	appErr, ok := err.(*apperrors.AppError)
	require.True(t, ok)
	assert.Equal(t, 404, appErr.Code)
}

// TestCreateIssueLink_InvalidTargetID проверяет отклонение при создание issue link некорректный target id.
func TestCreateIssueLink_InvalidTargetID(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Source", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	_, err = svc.CreateIssueLink(ctx, projectID, created.IssueKey, userID, tracker.CreateIssueLinkRequest{
		TargetIssueID: "not-a-uuid",
		LinkType:      "blocks",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target_issue_id")
}

// TestListIssues_InvalidAssignee проверяет отклонение при список issues некорректный assignee.
func TestListIssues_InvalidAssignee(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.ListIssues(ctx, uuid.New(), tracker.ListIssuesRequest{Assignee: "bad-uuid"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "assignee")
}

// TestCommentsAndTimeTracking проверяет сценарий: comments and time tracking.
func TestCommentsAndTimeTracking(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issue, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Commentable", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	comment, err := svc.CreateComment(ctx, projectID, issue.IssueKey, userID, tracker.CreateCommentRequest{
		BodyJSON: map[string]interface{}{"text": "hi"},
		BodyText: "hi",
	})
	require.NoError(t, err)
	assert.Equal(t, "hi", comment.BodyText)

	comments, err := svc.ListComments(ctx, projectID, issue.IssueKey)
	require.NoError(t, err)
	assert.Len(t, comments, 1)

	tl, err := svc.LogTime(ctx, projectID, issue.IssueKey, userID, tracker.LogTimeRequest{Minutes: 30, Description: "work"})
	require.NoError(t, err)
	assert.Equal(t, 30, tl.Minutes)

	timer, err := svc.StartTimer(ctx, projectID, issue.IssueKey, userID)
	require.NoError(t, err)
	assert.NotNil(t, timer.StartedAt)

	stopped, err := svc.StopTimer(ctx, projectID, issue.IssueKey, userID)
	require.NoError(t, err)
	assert.NotNil(t, stopped.EndedAt)

	logs, err := svc.ListTimeLogs(ctx, projectID, issue.IssueKey)
	require.NoError(t, err)
	assert.NotEmpty(t, logs)
}

// TestWorkflowTransitions_InvalidUUID проверяет отклонение при workflow transitions некорректный uuid.
func TestWorkflowTransitions_InvalidUUID(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	_, err := svc.CreateWorkflowTransition(ctx, uuid.New(), tracker.CreateWorkflowTransitionRequest{
		FromStatusID: "bad",
		ToStatusID:   uuid.New().String(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "from_status_id")

	_, err = svc.CreateWorkflowTransition(ctx, uuid.New(), tracker.CreateWorkflowTransitionRequest{
		FromStatusID: uuid.New().String(),
		ToStatusID:   "bad",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_status_id")
}

// TestCreateIssueFromPromotion проверяет сценарий: создание issue from promotion.
func TestCreateIssueFromPromotion(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	desc := map[string]interface{}{"type": "doc", "children": []interface{}{}}

	key, id, err := svc.CreateIssueFromPromotion(ctx, projectID, userID, "Promoted", "story", desc)
	require.NoError(t, err)
	assert.Equal(t, "TEST-1", key)
	assert.NotEmpty(t, id)
}

// TestCustomFieldValues проверяет сценарий: custom field values.
func TestCustomFieldValues(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	// Создаём определение кастомного поля
	cfReq := tracker.CreateCustomFieldRequest{
		Name:      "Sprint",
		FieldType: "select",
		Options:   []string{"Sprint 1", "Sprint 2", "Sprint 3"},
		Required:  false,
	}
	cf, err := svc.CreateCustomField(ctx, projectID, cfReq)
	require.NoError(t, err)
	assert.Equal(t, "Sprint", cf.Name)
	assert.Equal(t, "select", cf.FieldType)
	assert.Len(t, cf.Options, 3)

	// Создаём задачу с кастомным полем
	issueReq := tracker.CreateIssueRequest{
		Title:     "Custom field task",
		IssueType: "task",
		Priority:  "medium",
		CustomFields: map[string]interface{}{
			cf.ID: "Sprint 1",
		},
	}
	resp, err := svc.CreateIssue(ctx, projectID, userID, issueReq)
	require.NoError(t, err)
	assert.NotNil(t, resp.CustomFields)

	// Список кастомных полей
	fields, err := svc.ListCustomFields(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, fields, 1)
}
