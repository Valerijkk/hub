package project_test

// Тесты модуля проектов: интеграционные и unit-сценарии testframework.
import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/project"
)

type stubTrackerProv struct {
	workflowCalls int
	boardCalls    int
	workflowErr   error
	boardErr      error
}

func (s *stubTrackerProv) SeedDefaultWorkflow(_ context.Context, _ uuid.UUID) error {
	s.workflowCalls++
	return s.workflowErr
}

func (s *stubTrackerProv) CreateDefaultBoard(_ context.Context, _ uuid.UUID) error {
	s.boardCalls++
	return s.boardErr
}

type stubTMSProv struct {
	calls int
	err   error
}

func (s *stubTMSProv) CreateDefaultSuite(_ context.Context, _ uuid.UUID) error {
	s.calls++
	return s.err
}

type stubDashProv struct {
	calls int
	err   error
}

func (s *stubDashProv) CreateDefaultDashboard(_ context.Context, _, _ uuid.UUID) error {
	s.calls++
	return s.err
}

// TestProjectService_ProvisionerSettersAndProvision проверяет бизнес-логику: project сервис провизионинг setters and provision.
func TestProjectService_ProvisionerSettersAndProvision(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	ownerID := uuid.New()

	tracker := &stubTrackerProv{}
	tms := &stubTMSProv{}
	dash := &stubDashProv{}

	svc.SetTrackerProvisioner(tracker)
	svc.SetTMSProvisioner(tms)
	svc.SetDashboardProvisioner(dash)

	project.ExportProjectServiceProvisionProject(svc, ctx, projectID, ownerID)
	assert.Equal(t, 1, tracker.workflowCalls)
	assert.Equal(t, 1, tracker.boardCalls)
	assert.Equal(t, 1, tms.calls)
	assert.Equal(t, 1, dash.calls)
}

// TestProjectService_ProvisionProject_LogsErrorsWithoutPanic гарантирует отсутствие паники/гонок при project сервис provision project logs errors without паника.
func TestProjectService_ProvisionProject_LogsErrorsWithoutPanic(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	tracker := &stubTrackerProv{workflowErr: errors.New("wf fail"), boardErr: errors.New("board fail")}
	tms := &stubTMSProv{err: errors.New("tms fail")}
	dash := &stubDashProv{err: errors.New("dash fail")}

	svc.SetTrackerProvisioner(tracker)
	svc.SetTMSProvisioner(tms)
	svc.SetDashboardProvisioner(dash)

	require.NotPanics(t, func() {
		project.ExportProjectServiceProvisionProject(svc, ctx, uuid.New(), uuid.New())
	})
}

// TestProjectService_ProvisionProject_NilProvisioners проверяет бизнес-логику: project сервис provision project nil provisioners.
func TestProjectService_ProvisionProject_NilProvisioners(t *testing.T) {
	svc, _, _ := newTestService()
	require.NotPanics(t, func() {
		project.ExportProjectServiceProvisionProject(svc, context.Background(), uuid.New(), uuid.New())
	})
}
