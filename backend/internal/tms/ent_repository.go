package tms

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/component"
	"github.com/valerii/hub/backend/ent/environment"
	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/label"
	"github.com/valerii/hub/backend/ent/predicate"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testcaseversion"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/testsuite"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// testCaseProjectIDEQ — A8-002 query rewrite: FK column `project_test_cases`
// instead of EXISTS via HasProjectWith. Composite index (project_test_cases,
// deleted_at) is on TestCase schema (A8-002 FIXED).
func testCaseProjectIDEQ(projectID uuid.UUID) predicate.TestCase {
	return predicate.TestCase(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(testcase.ProjectColumn), projectID))
	})
}

// EntRepository реализует интерфейс Repository с помощью Ent ORM.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository создаёт новый Ent-репозиторий TMS.
func NewEntRepository(client *ent.Client) *EntRepository {
	return &EntRepository{client: client}
}

// ---------------------------------------------------------------------------
// Счётчики
// ---------------------------------------------------------------------------

// IncrementTestCaseCounter атомарно увеличивает счётчик тест-кейсов проекта.
func (r *EntRepository) IncrementTestCaseCounter(ctx context.Context, projectID uuid.UUID) (int, error) {
	p, err := r.client.Project.UpdateOneID(projectID).
		AddTestCaseCounter(1).
		Save(ctx)
	if err != nil {
		return 0, fmt.Errorf("инкремент счётчика тест-кейсов: %w", err)
	}
	return p.TestCaseCounter, nil
}

// ---------------------------------------------------------------------------
// Тест-кейсы
// ---------------------------------------------------------------------------

// CreateTestCase создаёт новый тест-кейс в проекте.
func (r *EntRepository) CreateTestCase(ctx context.Context, projectID uuid.UUID, input CreateTestCaseInput) (*ent.TestCase, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("начало транзакции: %w", err)
	}

	// Атомарный инкремент счётчика
	p, err := tx.Project.UpdateOneID(projectID).
		AddTestCaseCounter(1).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("инкремент счётчика тест-кейсов: %w", err)
	}

	caseNumber := p.TestCaseCounter
	caseKey := fmt.Sprintf("%s-TC-%d", p.Key, caseNumber)

	create := tx.TestCase.Create().
		SetCaseNumber(caseNumber).
		SetCaseKey(caseKey).
		SetTitle(input.Title).
		SetProjectID(projectID).
		SetCreatedByID(input.CreatedByID)

	if input.Description != "" {
		create = create.SetDescription(input.Description)
	}
	if input.Precondition != "" {
		create = create.SetPrecondition(input.Precondition)
	}
	if input.Postcondition != "" {
		create = create.SetPostcondition(input.Postcondition)
	}
	if input.CaseType != "" {
		create = create.SetCaseType(testcase.CaseType(input.CaseType))
	}
	if input.Priority != "" {
		create = create.SetPriority(testcase.Priority(input.Priority))
	}
	if input.Status != "" {
		create = create.SetStatus(testcase.Status(input.Status))
	}
	if len(input.Steps) > 0 {
		create = create.SetSteps(input.Steps)
	}
	if input.FullName != "" {
		create = create.SetFullName(input.FullName)
	}
	if input.ClassName != "" {
		create = create.SetClassName(input.ClassName)
	}
	if input.MethodName != "" {
		create = create.SetMethodName(input.MethodName)
	}
	if input.ExternalID != "" {
		create = create.SetExternalID(input.ExternalID)
	}
	if input.AutomationID != "" {
		create = create.SetAutomationID(input.AutomationID)
	}
	if input.SuiteID != nil {
		create = create.SetSuiteID(*input.SuiteID)
	}
	if len(input.LabelIDs) > 0 {
		create = create.AddLabelIDs(input.LabelIDs...)
	}
	if len(input.ComponentIDs) > 0 {
		create = create.AddComponentIDs(input.ComponentIDs...)
	}

	created, err := create.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("создание тест-кейса: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("коммит транзакции: %w", err)
	}

	// Подгружаем связи
	return r.GetTestCase(ctx, created.ID)
}

// GetTestCase возвращает тест-кейс по ID с подгруженными связями.
func (r *EntRepository) GetTestCase(ctx context.Context, id uuid.UUID) (*ent.TestCase, error) {
	return r.client.TestCase.Query().
		Where(testcase.ID(id), testcase.DeletedAtIsNil()).
		WithSuite().
		WithCreatedBy().
		WithLabels().
		WithComponents().
		WithLinkedIssues().
		Only(ctx)
}

// FindTestCaseByKey возвращает тест-кейс по ключу внутри projectID (A3-003 IDOR guard).
func (r *EntRepository) FindTestCaseByKey(ctx context.Context, projectID uuid.UUID, caseKey string) (*ent.TestCase, error) {
	return r.client.TestCase.Query().
		Where(
			testcase.CaseKey(caseKey),
			testcase.DeletedAtIsNil(),
			testcase.HasProjectWith(entproject.ID(projectID)),
		).
		WithSuite().
		WithCreatedBy().
		WithLabels().
		WithComponents().
		WithLinkedIssues().
		Only(ctx)
}

// FindTestCaseByExternalID возвращает тест-кейс по внешнему ID аннотации.
func (r *EntRepository) FindTestCaseByExternalID(ctx context.Context, projectID uuid.UUID, externalID string) (*ent.TestCase, error) {
	return r.client.TestCase.Query().
		Where(
			testcase.ExternalID(externalID),
			testcase.HasProjectWith(entproject.ID(projectID)),
		).
		Only(ctx)
}

// FindTestCaseByAutomationID возвращает тест-кейс по automation_id (явная привязка автотестов).
func (r *EntRepository) FindTestCaseByAutomationID(ctx context.Context, projectID uuid.UUID, automationID string) (*ent.TestCase, error) {
	return r.client.TestCase.Query().
		Where(
			testcase.AutomationID(automationID),
			testcase.HasProjectWith(entproject.ID(projectID)),
		).
		Only(ctx)
}

// FindTestCaseByFullName возвращает тест-кейс по полному имени (Allure matching).
func (r *EntRepository) FindTestCaseByFullName(ctx context.Context, projectID uuid.UUID, fullName string) (*ent.TestCase, error) {
	return r.client.TestCase.Query().
		Where(
			testcase.FullName(fullName),
			testcase.HasProjectWith(entproject.ID(projectID)),
		).
		Only(ctx)
}

// FindTestCaseByClassMethod возвращает тест-кейс по имени класса и метода.
func (r *EntRepository) FindTestCaseByClassMethod(ctx context.Context, projectID uuid.UUID, className, methodName string) (*ent.TestCase, error) {
	return r.client.TestCase.Query().
		Where(
			testcase.ClassName(className),
			testcase.MethodName(methodName),
			testcase.HasProjectWith(entproject.ID(projectID)),
		).
		Only(ctx)
}

// ListTestCases возвращает пагинированный список тест-кейсов с фильтрацией.
func (r *EntRepository) ListTestCases(ctx context.Context, projectID uuid.UUID, params ListTestCasesParams) ([]*ent.TestCase, int, error) {
	query := r.client.TestCase.Query().
		Where(testCaseProjectIDEQ(projectID), testcase.DeletedAtIsNil())

	if params.SuiteID != nil {
		query = query.Where(testcase.HasSuiteWith(testsuite.ID(*params.SuiteID)))
	}
	if params.Status != "" {
		query = query.Where(testcase.StatusEQ(testcase.Status(params.Status)))
	}
	if params.Priority != "" {
		query = query.Where(testcase.PriorityEQ(testcase.Priority(params.Priority)))
	}
	if params.Search != "" {
		query = query.Where(testcase.TitleContainsFold(params.Search))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт тест-кейсов: %w", err)
	}

	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	items, err := query.
		WithSuite().
		WithCreatedBy().
		WithLabels().
		WithComponents().
		Order(ent.Asc(testcase.FieldCreatedAt)).
		Offset(params.Offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("получение тест-кейсов: %w", err)
	}

	return items, total, nil
}

// UpdateTestCase обновляет тест-кейс с optimistic lock по Version (A3-004).
func (r *EntRepository) UpdateTestCase(ctx context.Context, id uuid.UUID, input UpdateTestCaseInput) (*ent.TestCase, error) {
	update := r.client.TestCase.Update().
		Where(
			testcase.IDEQ(id),
			testcase.VersionEQ(input.Version),
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
		return nil, fmt.Errorf("обновление тест-кейса: %w", err)
	}
	if affected == 0 {
		return nil, apperrors.NewConflict("version mismatch: тест-кейс был изменён другим пользователем")
	}

	return r.GetTestCase(ctx, id)
}

// DeleteTestCase выполняет мягкое удаление тест-кейса (устанавливает deleted_at).
func (r *EntRepository) DeleteTestCase(ctx context.Context, id uuid.UUID) error {
	return r.client.TestCase.UpdateOneID(id).
		SetDeletedAt(time.Now().UTC()).
		Exec(ctx)
}

// CountTestCasesInProject считает тест-кейсы из ids внутри projectID (P1-BE-12).
func (r *EntRepository) CountTestCasesInProject(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return r.client.TestCase.Query().
		Where(
			testcase.IDIn(ids...),
			testCaseProjectIDEQ(projectID),
			testcase.DeletedAtIsNil(),
		).
		Count(ctx)
}

// CountLabelsInProject считает метки из ids внутри projectID (P1-BE-12).
func (r *EntRepository) CountLabelsInProject(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return r.client.Label.Query().
		Where(
			label.IDIn(ids...),
			label.HasProjectWith(entproject.ID(projectID)),
		).
		Count(ctx)
}

// ---------------------------------------------------------------------------
// Версионирование
// ---------------------------------------------------------------------------

// CreateTestCaseVersion создаёт снимок версии тест-кейса.
func (r *EntRepository) CreateTestCaseVersion(ctx context.Context, testCaseID, userID uuid.UUID, versionNumber int, snapshot map[string]interface{}, description string) (*ent.TestCaseVersion, error) {
	create := r.client.TestCaseVersion.Create().
		SetVersionNumber(versionNumber).
		SetSnapshot(snapshot).
		SetTestCaseID(testCaseID).
		SetCreatedByID(userID)

	if description != "" {
		create = create.SetChangeDescription(description)
	}

	return create.Save(ctx)
}

// ListTestCaseVersions возвращает все версии тест-кейса.
func (r *EntRepository) ListTestCaseVersions(ctx context.Context, testCaseID uuid.UUID) ([]*ent.TestCaseVersion, error) {
	return r.client.TestCaseVersion.Query().
		Where(testcaseversion.HasTestCaseWith(testcase.ID(testCaseID))).
		WithCreatedBy().
		Order(ent.Desc(testcaseversion.FieldVersionNumber)).
		All(ctx)
}

// GetTestCaseVersion возвращает версию по ID только если она принадлежит testCaseID (A3-007).
func (r *EntRepository) GetTestCaseVersion(ctx context.Context, testCaseID, versionID uuid.UUID) (*ent.TestCaseVersion, error) {
	return r.client.TestCaseVersion.Query().
		Where(
			testcaseversion.ID(versionID),
			testcaseversion.HasTestCaseWith(testcase.ID(testCaseID)),
		).
		WithCreatedBy().
		WithTestCase().
		Only(ctx)
}

// ---------------------------------------------------------------------------
// Папки/сьюты
// ---------------------------------------------------------------------------

// CreateSuite создаёт новую папку/секцию тест-кейсов.
func (r *EntRepository) CreateSuite(ctx context.Context, projectID uuid.UUID, input CreateSuiteInput) (*ent.TestSuite, error) {
	create := r.client.TestSuite.Create().
		SetName(input.Name).
		SetProjectID(projectID).
		SetSortOrder(input.SortOrder)

	if input.Description != "" {
		create = create.SetDescription(input.Description)
	}
	if input.ParentID != nil {
		create = create.SetParentID(*input.ParentID)
	}

	return create.Save(ctx)
}

// GetSuite возвращает папку по ID.
func (r *EntRepository) GetSuite(ctx context.Context, id uuid.UUID) (*ent.TestSuite, error) {
	return r.client.TestSuite.Query().
		Where(testsuite.ID(id)).
		WithParent().
		WithChildren().
		Only(ctx)
}

// GetSuiteInProject возвращает папку только внутри projectID (P1-BE-12 ParentID/SuiteID FK).
func (r *EntRepository) GetSuiteInProject(ctx context.Context, projectID, id uuid.UUID) (*ent.TestSuite, error) {
	return r.client.TestSuite.Query().
		Where(
			testsuite.IDEQ(id),
			testsuite.HasProjectWith(entproject.ID(projectID)),
		).
		WithParent().
		Only(ctx)
}

// ListSuites возвращает все папки проекта (для построения дерева).
func (r *EntRepository) ListSuites(ctx context.Context, projectID uuid.UUID) ([]*ent.TestSuite, error) {
	return r.client.TestSuite.Query().
		Where(testsuite.HasProjectWith(entproject.ID(projectID))).
		WithParent().
		WithChildren().
		Order(ent.Asc(testsuite.FieldSortOrder)).
		All(ctx)
}

// UpdateSuite обновляет папку только внутри projectID (P1-BE-10).
func (r *EntRepository) UpdateSuite(ctx context.Context, projectID, id uuid.UUID, input UpdateSuiteInput) (*ent.TestSuite, error) {
	update := r.client.TestSuite.UpdateOneID(id).
		Where(testsuite.HasProjectWith(entproject.ID(projectID)))

	if input.Name != nil {
		update = update.SetName(*input.Name)
	}
	if input.Description != nil {
		update = update.SetDescription(*input.Description)
	}
	if input.SortOrder != nil {
		update = update.SetSortOrder(*input.SortOrder)
	}

	_, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление папки: %w", err)
	}

	return r.GetSuite(ctx, id)
}

// DeleteSuite удаляет папку только внутри projectID (P1-BE-10).
func (r *EntRepository) DeleteSuite(ctx context.Context, projectID, id uuid.UUID) error {
	affected, err := r.client.TestSuite.Delete().
		Where(
			testsuite.IDEQ(id),
			testsuite.HasProjectWith(entproject.ID(projectID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// MoveSuite перемещает папку к другому родителю.
func (r *EntRepository) MoveSuite(ctx context.Context, id uuid.UUID, parentID *uuid.UUID) (*ent.TestSuite, error) {
	update := r.client.TestSuite.UpdateOneID(id)
	if parentID != nil {
		update = update.SetParentID(*parentID)
	} else {
		update = update.ClearParent()
	}

	_, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("перемещение папки: %w", err)
	}

	return r.GetSuite(ctx, id)
}

// ---------------------------------------------------------------------------
// Компоненты
// ---------------------------------------------------------------------------

// CreateComponent создаёт новый компонент в проекте.
func (r *EntRepository) CreateComponent(ctx context.Context, projectID uuid.UUID, name, description string) (*ent.Component, error) {
	create := r.client.Component.Create().
		SetName(name).
		SetProjectID(projectID)

	if description != "" {
		create = create.SetDescription(description)
	}

	return create.Save(ctx)
}

// ListComponents возвращает все компоненты проекта.
func (r *EntRepository) ListComponents(ctx context.Context, projectID uuid.UUID) ([]*ent.Component, error) {
	return r.client.Component.Query().
		Where(component.HasProjectWith(entproject.ID(projectID))).
		Order(ent.Asc(component.FieldName)).
		All(ctx)
}

// CountComponentsInProject считает компоненты из ids внутри projectID (P1-BE-12).
func (r *EntRepository) CountComponentsInProject(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	return r.client.Component.Query().
		Where(
			component.IDIn(ids...),
			component.HasProjectWith(entproject.ID(projectID)),
		).
		Count(ctx)
}

// DeleteComponent удаляет компонент только внутри projectID (P1-BE-10).
func (r *EntRepository) DeleteComponent(ctx context.Context, projectID, id uuid.UUID) error {
	affected, err := r.client.Component.Delete().
		Where(
			component.IDEQ(id),
			component.HasProjectWith(entproject.ID(projectID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Прогоны
// ---------------------------------------------------------------------------

// CreateTestRun создаёт новый прогон тестов.
func (r *EntRepository) CreateTestRun(ctx context.Context, projectID uuid.UUID, input CreateTestRunInput) (*ent.TestRun, error) {
	create := r.client.TestRun.Create().
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

	created, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание прогона: %w", err)
	}

	return r.GetTestRun(ctx, created.ID)
}

// GetTestRun возвращает прогон по ID с подгруженными связями.
// Без project-scope — для post-create / внутренних путей; HTTP-facing — GetTestRunInProject.
func (r *EntRepository) GetTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(testrun.ID(id)).
		WithCreatedBy().
		WithProject().
		Only(ctx)
}

// GetTestRunInProject возвращает прогон по ID только внутри projectID (A3-012 IDOR guard).
func (r *EntRepository) GetTestRunInProject(ctx context.Context, projectID, id uuid.UUID) (*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(
			testrun.ID(id),
			testrun.HasProjectWith(entproject.ID(projectID)),
		).
		WithCreatedBy().
		WithProject().
		Only(ctx)
}

// ListTestRuns возвращает пагинированный список прогонов проекта.
func (r *EntRepository) ListTestRuns(ctx context.Context, projectID uuid.UUID, params ListTestRunsParams) ([]*ent.TestRun, int, error) {
	query := r.client.TestRun.Query().
		Where(testrun.HasProjectWith(entproject.ID(projectID)))

	if params.Status != "" {
		query = query.Where(testrun.StatusEQ(testrun.Status(params.Status)))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт прогонов: %w", err)
	}

	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	items, err := query.
		WithCreatedBy().
		Order(ent.Desc(testrun.FieldCreatedAt)).
		Offset(params.Offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("получение прогонов: %w", err)
	}

	return items, total, nil
}

// UpdateTestRunStatus обновляет статус прогона.
func (r *EntRepository) UpdateTestRunStatus(ctx context.Context, id uuid.UUID, status string, timestamp *time.Time) (*ent.TestRun, error) {
	update := r.client.TestRun.UpdateOneID(id).
		SetStatus(testrun.Status(status))

	if status == "in_progress" && timestamp != nil {
		update = update.SetStartedAt(*timestamp)
	}
	if (status == "completed" || status == "aborted") && timestamp != nil {
		update = update.SetCompletedAt(*timestamp)
	}

	_, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление статуса прогона: %w", err)
	}

	return r.GetTestRun(ctx, id)
}

// ReopenTestRun переоткрывает завершённый прогон.
func (r *EntRepository) ReopenTestRun(ctx context.Context, id uuid.UUID) (*ent.TestRun, error) {
	_, err := r.client.TestRun.UpdateOneID(id).
		SetStatus(testrun.StatusInProgress).
		ClearCompletedAt().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("переоткрытие прогона: %w", err)
	}

	return r.GetTestRun(ctx, id)
}

// UpdateTestRunStats обновляет счётчики результатов прогона.
func (r *EntRepository) UpdateTestRunStats(ctx context.Context, id uuid.UUID, stats TestRunStats) (*ent.TestRun, error) {
	_, err := r.client.TestRun.UpdateOneID(id).
		SetTotalCount(stats.TotalCount).
		SetPassedCount(stats.PassedCount).
		SetFailedCount(stats.FailedCount).
		SetSkippedCount(stats.SkippedCount).
		SetBrokenCount(stats.BrokenCount).
		SetUnknownCount(stats.UnknownCount).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление статистики прогона: %w", err)
	}

	return r.GetTestRun(ctx, id)
}

// ---------------------------------------------------------------------------
// Результаты
// ---------------------------------------------------------------------------

// CreateTestResult создаёт результат выполнения теста.
func (r *EntRepository) CreateTestResult(ctx context.Context, runID uuid.UUID, input CreateTestResultInput) (*ent.TestResult, error) {
	create := r.client.TestResult.Create().
		SetRunID(runID).
		SetStatus(testresult.Status(input.Status))

	if input.Comment != "" {
		create = create.SetComment(input.Comment)
	}
	if len(input.StepResults) > 0 {
		create = create.SetStepResults(input.StepResults)
	}
	if input.DurationMs != nil {
		create = create.SetDurationMs(*input.DurationMs)
	}
	if input.AllureUUID != "" {
		create = create.SetAllureUUID(input.AllureUUID)
	}
	if input.ErrorMessage != "" {
		create = create.SetErrorMessage(input.ErrorMessage)
	}
	if input.ErrorTrace != "" {
		create = create.SetErrorTrace(input.ErrorTrace)
	}
	if input.IsFlaky {
		create = create.SetIsFlaky(true)
	}
	if input.MatchStrategy != "" {
		create = create.SetMatchStrategy(testresult.MatchStrategy(input.MatchStrategy))
	}
	if input.TestCaseID != nil {
		create = create.SetTestCaseID(*input.TestCaseID)
	}
	if input.ExecutedByID != nil {
		create = create.SetExecutedByID(*input.ExecutedByID)
	}
	if input.StartedAt != nil {
		create = create.SetStartedAt(*input.StartedAt)
	}
	if input.CompletedAt != nil {
		create = create.SetCompletedAt(*input.CompletedAt)
	}

	return create.Save(ctx)
}

// UpdateTestResult обновляет результат теста только если он принадлежит runID (A3-008).
func (r *EntRepository) UpdateTestResult(ctx context.Context, runID, id uuid.UUID, input UpdateTestResultInput) (*ent.TestResult, error) {
	update := r.client.TestResult.Update().
		Where(
			testresult.IDEQ(id),
			testresult.HasRunWith(testrun.ID(runID)),
		)

	if input.Status != nil {
		update = update.SetStatus(testresult.Status(*input.Status))
	}
	if input.Comment != nil {
		update = update.SetComment(*input.Comment)
	}
	if input.StepResults != nil {
		update = update.SetStepResults(*input.StepResults)
	}
	if input.IsFlaky != nil {
		update = update.SetIsFlaky(*input.IsFlaky)
	}
	if input.ExecutedByID != nil {
		update = update.SetExecutedByID(*input.ExecutedByID)
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление результата: %w", err)
	}
	if affected == 0 {
		return nil, &ent.NotFoundError{}
	}

	return r.client.TestResult.Get(ctx, id)
}

// ListTestResults возвращает все результаты прогона.
func (r *EntRepository) ListTestResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error) {
	return r.client.TestResult.Query().
		Where(testresult.HasRunWith(testrun.ID(runID))).
		WithTestCase().
		WithExecutedBy().
		Order(ent.Asc(testresult.FieldCreatedAt)).
		All(ctx)
}

// BatchCreateTestResults массово создаёт результаты (для импорта Allure).
func (r *EntRepository) BatchCreateTestResults(ctx context.Context, runID uuid.UUID, inputs []CreateTestResultInput) ([]*ent.TestResult, error) {
	bulk := make([]*ent.TestResultCreate, 0, len(inputs))
	for _, input := range inputs {
		create := r.client.TestResult.Create().
			SetRunID(runID).
			SetStatus(testresult.Status(input.Status))

		if input.Comment != "" {
			create = create.SetComment(input.Comment)
		}
		if len(input.StepResults) > 0 {
			create = create.SetStepResults(input.StepResults)
		}
		if input.DurationMs != nil {
			create = create.SetDurationMs(*input.DurationMs)
		}
		if input.AllureUUID != "" {
			create = create.SetAllureUUID(input.AllureUUID)
		}
		if input.ErrorMessage != "" {
			create = create.SetErrorMessage(input.ErrorMessage)
		}
		if input.ErrorTrace != "" {
			create = create.SetErrorTrace(input.ErrorTrace)
		}
		if input.IsFlaky {
			create = create.SetIsFlaky(true)
		}
		if input.MatchStrategy != "" {
			create = create.SetMatchStrategy(testresult.MatchStrategy(input.MatchStrategy))
		}
		if input.TestCaseID != nil {
			create = create.SetTestCaseID(*input.TestCaseID)
		}
		if input.ExecutedByID != nil {
			create = create.SetExecutedByID(*input.ExecutedByID)
		}
		if input.StartedAt != nil {
			create = create.SetStartedAt(*input.StartedAt)
		}
		if input.CompletedAt != nil {
			create = create.SetCompletedAt(*input.CompletedAt)
		}

		bulk = append(bulk, create)
	}

	return r.client.TestResult.CreateBulk(bulk...).Save(ctx)
}

// ---------------------------------------------------------------------------
// Привязка к задачам
// ---------------------------------------------------------------------------

// FindIssueInProject возвращает задачу по UUID только внутри projectID (A3-003-link IDOR guard).
// Исключает мягко удалённые задачи. Mutation: снять HasProjectWith → foreign issue находится.
func (r *EntRepository) FindIssueInProject(ctx context.Context, projectID, issueID uuid.UUID) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			entissue.IDEQ(issueID),
			entissue.DeletedAtIsNil(),
			entissue.HasProjectWith(entproject.IDEQ(projectID)),
		).
		Only(ctx)
}

// LinkTestCaseToIssue привязывает тест-кейс к задаче.
func (r *EntRepository) LinkTestCaseToIssue(ctx context.Context, testCaseID, issueID uuid.UUID) error {
	return r.client.TestCase.UpdateOneID(testCaseID).
		AddLinkedIssueIDs(issueID).
		Exec(ctx)
}

// UnlinkTestCaseFromIssue отвязывает тест-кейс от задачи.
func (r *EntRepository) UnlinkTestCaseFromIssue(ctx context.Context, testCaseID, issueID uuid.UUID) error {
	return r.client.TestCase.UpdateOneID(testCaseID).
		RemoveLinkedIssueIDs(issueID).
		Exec(ctx)
}

// GetTestCaseLinkedIssues возвращает задачи, привязанные к тест-кейсу.
func (r *EntRepository) GetTestCaseLinkedIssues(ctx context.Context, testCaseID uuid.UUID) ([]*ent.Issue, error) {
	tc, err := r.client.TestCase.Query().
		Where(testcase.ID(testCaseID)).
		WithLinkedIssues().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение привязанных задач: %w", err)
	}
	return tc.Edges.LinkedIssues, nil
}

// GetIssueLinkedTestCases возвращает тест-кейсы, привязанные к задаче (обратная связь).
func (r *EntRepository) GetIssueLinkedTestCases(ctx context.Context, issueID uuid.UUID) ([]*ent.TestCase, error) {
	issue, err := r.client.Issue.Query().
		Where(entissue.ID(issueID)).
		WithLinkedTestCases().
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение привязанных тест-кейсов: %w", err)
	}
	return issue.Edges.LinkedTestCases, nil
}

// GetProjectIssuesWithTestCases возвращает задачи проекта с привязанными тест-кейсами.
func (r *EntRepository) GetProjectIssuesWithTestCases(ctx context.Context, projectID uuid.UUID) ([]*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(entissue.HasProjectWith(entproject.ID(projectID))).
		WithLinkedTestCases().
		All(ctx)
}

// CreateBugIssue создаёт задачу типа bug и привязывает к тест-кейсу.
func (r *EntRepository) CreateBugIssue(ctx context.Context, projectID, userID, testCaseID uuid.UUID, title, description, priority string) (*ent.Issue, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("начало транзакции: %w", err)
	}

	// Инкрементируем счётчик задач проекта
	p, err := tx.Project.UpdateOneID(projectID).
		AddTaskCounter(1).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("инкремент счётчика задач: %w", err)
	}

	issueKey := fmt.Sprintf("%s-%d", p.Key, p.TaskCounter)

	// Получаем дефолтный статус проекта
	statuses, err := tx.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(entproject.ID(projectID))).
		All(ctx)
	if err != nil || len(statuses) == 0 {
		_ = tx.Rollback()
		return nil, fmt.Errorf("получение дефолтного статуса: %w", err)
	}
	defaultStatusID := statuses[0].ID
	for _, st := range statuses {
		if st.IsDefault {
			defaultStatusID = st.ID
			break
		}
	}

	if priority == "" {
		priority = "high"
	}

	create := tx.Issue.Create().
		SetIssueNumber(p.TaskCounter).
		SetIssueKey(issueKey).
		SetTitle(title).
		SetIssueType(entissue.IssueType("bug")).
		SetPriority(entissue.Priority(priority)).
		SetProjectID(projectID).
		SetStatusID(defaultStatusID).
		SetReporterID(userID)

	if description != "" {
		create = create.SetDescriptionText(description)
	}

	issue, err := create.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("создание бага: %w", err)
	}

	// Привязываем к тест-кейсу только если он известен (A3-009).
	if testCaseID != uuid.Nil {
		err = tx.TestCase.UpdateOneID(testCaseID).
			AddLinkedIssueIDs(issue.ID).
			Exec(ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("привязка бага к тест-кейсу: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("коммит транзакции: %w", err)
	}

	return issue, nil
}

// GetLastResultForTestCase возвращает последний результат для тест-кейса.
func (r *EntRepository) GetLastResultForTestCase(ctx context.Context, testCaseID uuid.UUID) (*ent.TestResult, error) {
	return r.client.TestResult.Query().
		Where(testresult.HasTestCaseWith(testcase.ID(testCaseID))).
		Order(ent.Desc(testresult.FieldCreatedAt)).
		First(ctx)
}

// ---------------------------------------------------------------------------
// Расширенные операции прогонов
// ---------------------------------------------------------------------------

// AddTestCasesToRun массово создаёт TestResult записи (со статусом "unknown") для каждого тест-кейса.
func (r *EntRepository) AddTestCasesToRun(ctx context.Context, runID uuid.UUID, caseIDs []uuid.UUID) error {
	bulk := make([]*ent.TestResultCreate, 0, len(caseIDs))
	for _, caseID := range caseIDs {
		create := r.client.TestResult.Create().
			SetRunID(runID).
			SetTestCaseID(caseID).
			SetStatus(testresult.StatusUnknown)
		bulk = append(bulk, create)
	}
	_, err := r.client.TestResult.CreateBulk(bulk...).Save(ctx)
	if err != nil {
		return fmt.Errorf("массовое создание результатов: %w", err)
	}
	return nil
}

// GetTestResult возвращает результат по ID с подгруженными связями.
func (r *EntRepository) GetTestResult(ctx context.Context, resultID uuid.UUID) (*ent.TestResult, error) {
	return r.client.TestResult.Query().
		Where(testresult.ID(resultID)).
		WithTestCase().
		WithRun().
		WithExecutedBy().
		WithFiles().
		Only(ctx)
}

// ListTestRunResults возвращает результаты прогона с подгруженными тест-кейсами и файлами.
func (r *EntRepository) ListTestRunResults(ctx context.Context, runID uuid.UUID) ([]*ent.TestResult, error) {
	return r.client.TestResult.Query().
		Where(testresult.HasRunWith(testrun.ID(runID))).
		WithTestCase().
		WithExecutedBy().
		WithFiles().
		Order(ent.Asc(testresult.FieldCreatedAt)).
		All(ctx)
}

// AssignTestersToRun устанавливает M2M связь прогона с назначенными тестировщиками.
func (r *EntRepository) AssignTestersToRun(ctx context.Context, runID uuid.UUID, userIDs []uuid.UUID) error {
	return r.client.TestRun.UpdateOneID(runID).
		ClearAssignees().
		AddAssigneeIDs(userIDs...).
		Exec(ctx)
}

// GetTestRunWithResults возвращает прогон с результатами, тест-кейсами, файлами и назначенными.
// Без project-scope — для внутренних путей; HTTP-facing — GetTestRunWithResultsInProject.
func (r *EntRepository) GetTestRunWithResults(ctx context.Context, runID uuid.UUID) (*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(testrun.ID(runID)).
		WithCreatedBy().
		WithAssignees().
		WithResults(func(q *ent.TestResultQuery) {
			q.WithTestCase().
				WithExecutedBy().
				WithFiles().
				Order(ent.Asc(testresult.FieldCreatedAt))
		}).
		Only(ctx)
}

// GetTestRunWithResultsInProject — detail/export только внутри projectID (A3-012).
func (r *EntRepository) GetTestRunWithResultsInProject(ctx context.Context, projectID, runID uuid.UUID) (*ent.TestRun, error) {
	return r.client.TestRun.Query().
		Where(
			testrun.ID(runID),
			testrun.HasProjectWith(entproject.ID(projectID)),
		).
		WithCreatedBy().
		WithAssignees().
		WithResults(func(q *ent.TestResultQuery) {
			q.WithTestCase().
				WithExecutedBy().
				WithFiles().
				Order(ent.Asc(testresult.FieldCreatedAt))
		}).
		Only(ctx)
}

// AddFilesToResult привязывает файлы к результату теста.
func (r *EntRepository) AddFilesToResult(ctx context.Context, resultID uuid.UUID, fileIDs []uuid.UUID) error {
	return r.client.TestResult.UpdateOneID(resultID).
		AddFileIDs(fileIDs...).
		Exec(ctx)
}

// ---------------------------------------------------------------------------
// Тестовые окружения
// ---------------------------------------------------------------------------

// CreateEnvironment создаёт тестовое окружение в проекте.
func (r *EntRepository) CreateEnvironment(ctx context.Context, projectID uuid.UUID, input CreateEnvironmentInput) (*ent.Environment, error) {
	builder := r.client.Environment.Create().
		SetName(input.Name).
		SetEnvLabel(environment.EnvLabel(input.Label)).
		SetProjectID(projectID)
	if input.Config != nil {
		builder.SetConfig(input.Config)
	}
	return builder.Save(ctx)
}

// GetEnvironment возвращает окружение по ID.
func (r *EntRepository) GetEnvironment(ctx context.Context, envID uuid.UUID) (*ent.Environment, error) {
	return r.client.Environment.Get(ctx, envID)
}

// GetEnvironmentInProject возвращает окружение только внутри projectID (P1-BE-11).
func (r *EntRepository) GetEnvironmentInProject(ctx context.Context, projectID, envID uuid.UUID) (*ent.Environment, error) {
	return r.client.Environment.Query().
		Where(
			environment.IDEQ(envID),
			environment.HasProjectWith(entproject.ID(projectID)),
		).
		WithProject().
		Only(ctx)
}

// ListEnvironmentsByProject возвращает все окружения проекта.
func (r *EntRepository) ListEnvironmentsByProject(ctx context.Context, projectID uuid.UUID) ([]*ent.Environment, error) {
	return r.client.Project.Query().
		Where(entproject.ID(projectID)).
		QueryEnvironments().
		All(ctx)
}

// UpdateEnvironment обновляет окружение только внутри projectID (P1-BE-11).
func (r *EntRepository) UpdateEnvironment(ctx context.Context, projectID, envID uuid.UUID, input UpdateEnvironmentInput) (*ent.Environment, error) {
	builder := r.client.Environment.UpdateOneID(envID).
		Where(environment.HasProjectWith(entproject.ID(projectID)))
	if input.Name != nil {
		builder.SetName(*input.Name)
	}
	if input.Label != nil {
		builder.SetEnvLabel(environment.EnvLabel(*input.Label))
	}
	if input.Config != nil {
		builder.SetConfig(*input.Config)
	}
	return builder.Save(ctx)
}

// DeleteEnvironment удаляет окружение только внутри projectID (P1-BE-11).
func (r *EntRepository) DeleteEnvironment(ctx context.Context, projectID, envID uuid.UUID) error {
	affected, err := r.client.Environment.Delete().
		Where(
			environment.IDEQ(envID),
			environment.HasProjectWith(entproject.ID(projectID)),
		).
		Exec(ctx)
	if err != nil {
		return err
	}
	if affected == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

// GetLatestTestRunForEnvironment возвращает последний прогон проекта в указанном окружении.
// Если прогонов нет — возвращает (nil, nil) без ошибки.
func (r *EntRepository) GetLatestTestRunForEnvironment(ctx context.Context, projectID, envID uuid.UUID) (*ent.TestRun, error) {
	runs, err := r.client.TestRun.Query().
		Where(
			testrun.HasProjectWith(entproject.ID(projectID)),
			testrun.HasEnvironmentWith(environment.ID(envID)),
		).
		Order(ent.Desc(testrun.FieldCreatedAt)).
		Limit(1).
		All(ctx)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		return nil, nil
	}
	return runs[0], nil
}
