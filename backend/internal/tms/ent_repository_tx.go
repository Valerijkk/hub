package tms

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// snapshotUpdater — Ent-репозиторий с атомарным snapshot+update (A2W2-2-002).
type snapshotUpdater interface {
	UpdateTestCaseWithSnapshot(ctx context.Context, id, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string, input UpdateTestCaseInput) (*ent.TestCase, error)
}

// snapshotAndUpdateTestCase создаёт версию и обновляет кейс; на Ent — в одной tx.
func (s *TMSService) snapshotAndUpdateTestCase(
	ctx context.Context,
	id, userID uuid.UUID,
	versionNumber int,
	snapshot map[string]interface{},
	description string,
	input UpdateTestCaseInput,
) (*ent.TestCase, error) {
	if su, ok := s.repo.(snapshotUpdater); ok {
		return su.UpdateTestCaseWithSnapshot(ctx, id, userID, versionNumber, snapshot, description, input)
	}
	// Mock / non-Ent fallback: последовательный путь (unit-тесты на mock).
	if _, err := s.repo.CreateTestCaseVersion(ctx, id, userID, versionNumber, snapshot, description); err != nil {
		return nil, apperrors.NewInternal("ошибка создания версии")
	}
	return s.repo.UpdateTestCase(ctx, id, input)
}

// UpdateTestCaseWithSnapshot — A2W2-2-002: CreateTestCaseVersion + UpdateTestCase в одной Ent tx.
func (r *EntRepository) UpdateTestCaseWithSnapshot(
	ctx context.Context,
	id, userID uuid.UUID,
	versionNumber int,
	snapshot map[string]interface{},
	description string,
	input UpdateTestCaseInput,
) (*ent.TestCase, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("начало транзакции snapshot+update: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	create := tx.TestCaseVersion.Create().
		SetVersionNumber(versionNumber).
		SetSnapshot(snapshot).
		SetTestCaseID(id).
		SetCreatedByID(userID)
	if description != "" {
		create = create.SetChangeDescription(description)
	}
	if _, err := create.Save(ctx); err != nil {
		return nil, fmt.Errorf("создание версии в tx: %w", err)
	}

	update := tx.TestCase.Update().
		Where(
			testcase.IDEQ(id),
			testcase.VersionEQ(input.Version),
			testcase.DeletedAtIsNil(),
		).
		AddVersion(1)

	if input.Title != nil {
		update = update.SetTitle(*input.Title)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.Precondition != nil {
		update = update.SetPrecondition(*input.Precondition)
	}
	if input.Postcondition != nil {
		update = update.SetPostcondition(*input.Postcondition)
	}
	if input.CaseType != nil {
		update = update.SetCaseType(testcase.CaseType(*input.CaseType))
	}
	if input.Priority != nil {
		update = update.SetPriority(testcase.Priority(*input.Priority))
	}
	if input.Status != nil {
		update = update.SetStatus(testcase.Status(*input.Status))
	}
	if input.Steps != nil {
		update = update.SetSteps(*input.Steps)
	}
	if input.FullName != nil {
		update = update.SetFullName(*input.FullName)
	}
	if input.ClassName != nil {
		update = update.SetClassName(*input.ClassName)
	}
	if input.MethodName != nil {
		update = update.SetMethodName(*input.MethodName)
	}
	if input.ExternalID != nil {
		update = update.SetExternalID(*input.ExternalID)
	}
	if input.AutomationID != nil {
		update = update.SetAutomationID(*input.AutomationID)
	}
	if input.SuiteID != nil {
		update = update.SetSuiteID(*input.SuiteID)
	}
	if input.ClearSuite {
		update = update.ClearSuite()
	}
	if input.LabelIDs != nil {
		update = update.ClearLabels().AddLabelIDs(*input.LabelIDs...)
	}
	if input.ComponentIDs != nil {
		update = update.ClearComponents().AddComponentIDs(*input.ComponentIDs...)
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление тест-кейса в tx: %w", err)
	}
	if affected == 0 {
		alive, existErr := tx.TestCase.Query().
			Where(testcase.IDEQ(id), testcase.DeletedAtIsNil()).
			Exist(ctx)
		if existErr != nil {
			return nil, fmt.Errorf("проверка тест-кейса после conflict: %w", existErr)
		}
		if !alive {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewConflict("version mismatch: тест-кейс был изменён другим пользователем")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit snapshot+update: %w", err)
	}
	committed = true
	return r.GetTestCase(ctx, id)
}

// CreateAutomaticImportRun — A2W2-2-004: create run + batch results + status в одной Ent tx.
func (r *EntRepository) CreateAutomaticImportRun(
	ctx context.Context,
	projectID uuid.UUID,
	input CreateTestRunInput,
	results []CreateTestResultInput,
	stats TestRunStats,
) (*ent.TestRun, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("начало транзакции automatic import: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	create := tx.TestRun.Create().
		SetTitle(input.Title).
		SetProjectID(projectID).
		SetCreatedByID(input.CreatedByID)
	if input.Description != "" {
		create = create.SetDescription(input.Description)
	}
	if input.RunType != "" {
		create = create.SetRunType(testrun.RunType(input.RunType))
	}
	if input.AllureRunID != "" {
		create = create.SetAllureRunID(input.AllureRunID)
	}
	run, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание прогона в tx: %w", err)
	}

	if len(results) > 0 {
		bulk := make([]*ent.TestResultCreate, 0, len(results))
		for _, ri := range results {
			rc := tx.TestResult.Create().
				SetRunID(run.ID).
				SetStatus(testresult.Status(ri.Status))
			if ri.Comment != "" {
				rc = rc.SetComment(ri.Comment)
			}
			if len(ri.StepResults) > 0 {
				rc = rc.SetStepResults(ri.StepResults)
			}
			if ri.DurationMs != nil {
				rc = rc.SetDurationMs(*ri.DurationMs)
			}
			if ri.AllureUUID != "" {
				rc = rc.SetAllureUUID(ri.AllureUUID)
			}
			if ri.ErrorMessage != "" {
				rc = rc.SetErrorMessage(ri.ErrorMessage)
			}
			if ri.ErrorTrace != "" {
				rc = rc.SetErrorTrace(ri.ErrorTrace)
			}
			if ri.IsFlaky {
				rc = rc.SetIsFlaky(true)
			}
			if ri.MatchStrategy != "" {
				rc = rc.SetMatchStrategy(testresult.MatchStrategy(ri.MatchStrategy))
			}
			if ri.TestCaseID != nil {
				rc = rc.SetTestCaseID(*ri.TestCaseID)
			}
			if ri.ExecutedByID != nil {
				rc = rc.SetExecutedByID(*ri.ExecutedByID)
			}
			if ri.StartedAt != nil {
				rc = rc.SetStartedAt(*ri.StartedAt)
			}
			if ri.CompletedAt != nil {
				rc = rc.SetCompletedAt(*ri.CompletedAt)
			}
			bulk = append(bulk, rc)
		}
		if _, err := tx.TestResult.CreateBulk(bulk...).Save(ctx); err != nil {
			return nil, fmt.Errorf("batch results в tx: %w", err)
		}
	}

	now := time.Now().UTC()
	if _, err := tx.TestRun.UpdateOneID(run.ID).
		SetStatus(testrun.StatusInProgress).
		SetStartedAt(now).
		SetTotalCount(stats.TotalCount).
		SetPassedCount(stats.PassedCount).
		SetFailedCount(stats.FailedCount).
		SetSkippedCount(stats.SkippedCount).
		SetBrokenCount(stats.BrokenCount).
		SetUnknownCount(stats.UnknownCount).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("старт прогона в tx: %w", err)
	}
	if _, err := tx.TestRun.UpdateOneID(run.ID).
		SetStatus(testrun.StatusCompleted).
		SetCompletedAt(now).
		Save(ctx); err != nil {
		return nil, fmt.Errorf("завершение прогона в tx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit automatic import: %w", err)
	}
	committed = true
	return r.GetTestRun(ctx, run.ID)
}

// AbortOrphanTestRun — cleanup при частичном импорте (fallback без tx).
func (r *EntRepository) AbortOrphanTestRun(ctx context.Context, runID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.client.TestRun.UpdateOneID(runID).
		SetStatus(testrun.StatusAborted).
		SetCompletedAt(now).
		Save(ctx)
	return err
}
