package tms_test

// Тесты модуля TMS (волна 99): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMatchTestCase_ByAutomationIDLabel проверяет сценарий: match test case by автоматизация id label.
func TestMatchTestCase_ByAutomationIDLabel(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-99")
	repo.automationIDMap["auto-login"] = tc

	matcher := tms.NewMatcher(repo)
	result := tms.AllureResult{
		UUID:   "auto-1",
		Name:   "loginTest",
		Status: "passed",
		Labels: []tms.AllureLabel{{Name: "automation_id", Value: "auto-login"}},
	}

	matched, strategy, err := matcher.MatchTestCase(context.Background(), uuid.New(), &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tms.MatchByAutomationID, strategy)
}

// TestMatchTestCase_ByCaseKeyInName проверяет сценарий: match test case by case key in name.
func TestMatchTestCase_ByCaseKeyInName(t *testing.T) {
	repo := newMockMatcherRepo()
	tcID := uuid.New()
	tc := makeMockTestCase(tcID, "HUB-TC-7")
	repo.caseKeyMap["HUB-TC-7"] = tc

	matcher := tms.NewMatcher(repo)
	result := tms.AllureResult{
		UUID:     "ck-1",
		Name:     "should pass HUB-TC-7 scenario",
		Status:   "passed",
		FullName: "com.example.Suite.should pass HUB-TC-7 scenario",
	}

	matched, strategy, err := matcher.MatchTestCase(context.Background(), uuid.New(), &result)
	require.NoError(t, err)
	require.NotNil(t, matched)
	assert.Equal(t, tms.MatchByCaseKey, strategy)
}

// TestMatchStrategy_StringUnknown проверяет сценарий: match strategy string unknown.
func TestMatchStrategy_StringUnknown(t *testing.T) {
	assert.Equal(t, "unmapped", tms.MatchStrategy(99).String())
}
