package tracker_test

// Тесты сервиса трекера задач (волна 115): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

type conflictWorkflowStatusRepo struct {
	mockRepo
}

func (r *conflictWorkflowStatusRepo) CreateWorkflowStatus(_ context.Context, _ uuid.UUID, name, _, _ string, _ int) (*ent.WorkflowStatus, error) {
	for _, st := range r.statuses {
		if st.Name == name {
			return nil, &ent.ConstraintError{}
		}
	}
	id := uuid.New()
	st := &ent.WorkflowStatus{ID: id, Name: name}
	r.statuses[id] = st
	return st, nil
}

type conflictLabelGroupRepo struct {
	mockRepo
	groups map[string]*ent.LabelGroup
}

func (r *conflictLabelGroupRepo) CreateLabelGroup(_ context.Context, _ uuid.UUID, name string) (*ent.LabelGroup, error) {
	if r.groups == nil {
		r.groups = make(map[string]*ent.LabelGroup)
	}
	if _, ok := r.groups[name]; ok {
		return nil, &ent.ConstraintError{}
	}
	g := &ent.LabelGroup{ID: uuid.New(), Name: name}
	r.groups[name] = g
	return g, nil
}

// TestCreateWorkflowStatus_DuplicateName проверяет сценарий: создание workflow status duplicate name.
func TestCreateWorkflowStatus_DuplicateName(t *testing.T) {
	repo := &conflictWorkflowStatusRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateWorkflowStatus(ctx, projectID, tracker.CreateWorkflowStatusRequest{
		Name: "Done", Category: "done", Color: "#00FF00",
	})
	require.NoError(t, err)

	_, err = svc.CreateWorkflowStatus(ctx, projectID, tracker.CreateWorkflowStatusRequest{
		Name: "Done", Category: "done", Color: "#111111",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "уже существует")
}

// TestCreateLabelGroup_DuplicateName проверяет сценарий: создание label group duplicate name.
func TestCreateLabelGroup_DuplicateName(t *testing.T) {
	repo := &conflictLabelGroupRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.CreateLabelGroup(ctx, projectID, tracker.CreateLabelGroupRequest{Name: "Epics"})
	require.NoError(t, err)

	_, err = svc.CreateLabelGroup(ctx, projectID, tracker.CreateLabelGroupRequest{Name: "Epics"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "уже существует")
}
