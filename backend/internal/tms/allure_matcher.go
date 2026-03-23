package tms

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// caseKeyInNameRegex ищет ключ тест-кейса вида PROJ-TC-123 в имени автотеста.
// Алгоритм: паттерн UpperCase-TC-Number с границами слов.
// Это позволяет автоматически привязывать результат, если разработчик
// включил ключ тест-кейса в название теста или аннотацию.
var caseKeyInNameRegex = regexp.MustCompile(`\b([A-Z]+-TC-\d+)\b`)

// ---------------------------------------------------------------------------
// Матчер Allure → Тест-кейс
// ---------------------------------------------------------------------------
// Многостратегийный матчинг: аннотация > fullName > class+method.
// Если ни одна стратегия не сработала — результат считается «unmapped».
// ---------------------------------------------------------------------------

// MatchStrategy определяет стратегию, по которой был найден тест-кейс.
type MatchStrategy int

const (
	// MatchByAnnotation — найден по аннотации (link type=tms или label AS_ID/tms).
	MatchByAnnotation MatchStrategy = iota
	// MatchByAutomationID — найден по automation_id (явная привязка QA-инженером).
	MatchByAutomationID
	// MatchByCaseKey — найден по ключу тест-кейса (PROJ-TC-123) в имени автотеста.
	MatchByCaseKey
	// MatchByFullName — найден по полному имени (fullName).
	MatchByFullName
	// MatchByClassName — найден по имени класса (suite) + имени метода (name).
	MatchByClassName
)

// String возвращает строковое представление стратегии для хранения в БД.
func (s MatchStrategy) String() string {
	switch s {
	case MatchByAnnotation:
		return "annotation"
	case MatchByAutomationID:
		return "automation_id"
	case MatchByCaseKey:
		return "case_key"
	case MatchByFullName:
		return "fullname"
	case MatchByClassName:
		return "classmethod"
	default:
		return "unmapped"
	}
}

// Matcher выполняет сопоставление результатов Allure с тест-кейсами в базе.
type Matcher struct {
	repo Repository
}

// NewMatcher создаёт новый матчер с указанным репозиторием.
func NewMatcher(repo Repository) *Matcher {
	return &Matcher{repo: repo}
}

// MatchTestCase пытается найти тест-кейс для результата Allure.
// Стратегии проверяются в порядке приоритета:
//  1. Links с type="tms" — по name ищем FindTestCaseByExternalID
//  2. Labels с name="AS_ID" или name="tms" — по value ищем FindTestCaseByExternalID
//  3. automation_id — по tms-ссылкам и label "automation_id" ищем FindTestCaseByAutomationID
//  4. case_key в имени — паттерн PROJ-TC-123 в Name/FullName ищем FindTestCaseByKey
//  5. FullName — ищем FindTestCaseByFullName
//  6. Suite label + Name — ищем FindTestCaseByClassMethod
//
// Возвращает nil, -1, nil если совпадение не найдено (unmapped).
func (m *Matcher) MatchTestCase(ctx context.Context, projectID uuid.UUID, result *AllureResult) (*ent.TestCase, MatchStrategy, error) {
	// Стратегия 1: Links с type="tms"
	for _, link := range result.Links {
		if link.Type == "tms" && link.Name != "" {
			tc, err := m.repo.FindTestCaseByExternalID(ctx, projectID, link.Name)
			if err == nil && tc != nil {
				slog.Debug("allure матчинг по tms link", "external_id", link.Name, "case_key", tc.CaseKey)
				return tc, MatchByAnnotation, nil
			}
			if !ent.IsNotFound(err) && err != nil {
				return nil, -1, err
			}
		}
	}

	// Стратегия 2: Labels с name="AS_ID" или name="tms"
	for _, label := range result.Labels {
		if (label.Name == "AS_ID" || label.Name == "tms") && label.Value != "" {
			tc, err := m.repo.FindTestCaseByExternalID(ctx, projectID, label.Value)
			if err == nil && tc != nil {
				slog.Debug("allure матчинг по label", "label", label.Name, "value", label.Value, "case_key", tc.CaseKey)
				return tc, MatchByAnnotation, nil
			}
			if !ent.IsNotFound(err) && err != nil {
				return nil, -1, err
			}
		}
	}

	// Стратегия 2.5: automation_id — ищем в аннотациях и лейблах значение,
	// совпадающее с полем automation_id тест-кейса.
	// Это позволяет QA-инженеру задать явный ID, а разработчику —
	// указать его в @TmsLink или label "automation_id".
	for _, link := range result.Links {
		if link.Type == "tms" && link.Name != "" {
			tc, err := m.repo.FindTestCaseByAutomationID(ctx, projectID, link.Name)
			if err == nil && tc != nil {
				slog.Debug("allure матчинг по automation_id (tms link)", "automation_id", link.Name, "case_key", tc.CaseKey)
				return tc, MatchByAutomationID, nil
			}
		}
	}
	for _, label := range result.Labels {
		if label.Name == "automation_id" && label.Value != "" {
			tc, err := m.repo.FindTestCaseByAutomationID(ctx, projectID, label.Value)
			if err == nil && tc != nil {
				slog.Debug("allure матчинг по automation_id (label)", "automation_id", label.Value, "case_key", tc.CaseKey)
				return tc, MatchByAutomationID, nil
			}
		}
	}

	// Стратегия 2.7: case_key в имени теста — ищем паттерн PROJ-TC-123
	// в полях Name и FullName результата Allure. Позволяет разработчику
	// просто включить ключ тест-кейса в название теста.
	searchTexts := []string{result.Name, result.FullName}
	for _, text := range searchTexts {
		keys := caseKeyInNameRegex.FindAllString(text, -1)
		for _, key := range keys {
			tc, err := m.repo.FindTestCaseByKey(ctx, key)
			if err == nil && tc != nil {
				slog.Debug("allure матчинг по case_key в имени", "case_key", key, "name", result.Name)
				return tc, MatchByCaseKey, nil
			}
		}
	}

	// Стратегия 3: FullName
	if result.FullName != "" {
		tc, err := m.repo.FindTestCaseByFullName(ctx, projectID, result.FullName)
		if err == nil && tc != nil {
			slog.Debug("allure матчинг по fullName", "full_name", result.FullName, "case_key", tc.CaseKey)
			return tc, MatchByFullName, nil
		}
		if !ent.IsNotFound(err) && err != nil {
			return nil, -1, err
		}
	}

	// Стратегия 4: Suite (label) + Name
	suiteName := extractLabelValue(result.Labels, "suite")
	if suiteName != "" && result.Name != "" {
		tc, err := m.repo.FindTestCaseByClassMethod(ctx, projectID, suiteName, result.Name)
		if err == nil && tc != nil {
			slog.Debug("allure матчинг по class+method", "suite", suiteName, "name", result.Name, "case_key", tc.CaseKey)
			return tc, MatchByClassName, nil
		}
		if !ent.IsNotFound(err) && err != nil {
			return nil, -1, err
		}
	}

	// Ничего не найдено — unmapped
	slog.Debug("allure результат не сопоставлен с тест-кейсом", "uuid", result.UUID, "name", result.Name)
	return nil, -1, nil
}
