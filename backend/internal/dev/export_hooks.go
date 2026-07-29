//go:build hubtest

package dev

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ExportDemoUser — строка демо-пользователя для hubtest.
type ExportDemoUser struct {
	Email       string
	Name        string
	DisplayName string
	SystemRole  string
}

// ExportDemoUsers возвращает заданных демо-пользователей для hubtest.
func ExportDemoUsers() []ExportDemoUser {
	out := make([]ExportDemoUser, len(demoUsers))
	for i, u := range demoUsers {
		out[i] = ExportDemoUser{
			Email:       u.Email,
			Name:        u.Name,
			DisplayName: u.DisplayName,
			SystemRole:  string(u.SystemRole),
		}
	}
	return out
}

// ExportDemoStatus — демо-статус workflow для hubtest.
type ExportDemoStatus struct {
	Name     string
	Color    string
	Category string
	Order    int
}

// ExportDemoStatuses возвращает заданные демо-статусы для hubtest.
func ExportDemoStatuses() []ExportDemoStatus {
	out := make([]ExportDemoStatus, len(demoStatuses))
	for i, s := range demoStatuses {
		out[i] = ExportDemoStatus{
			Name:     s.Name,
			Color:    s.Color,
			Category: string(s.Category),
			Order:    s.Order,
		}
	}
	return out
}

// ExportDemoIssue — демо-задача для hubtest.
type ExportDemoIssue struct {
	Key           string
	IssueType     string
	Title         string
	StatusName    string
	AssigneeEmail string
	Priority      string
}

// ExportDemoIssues возвращает заданные демо-задачи для hubtest.
func ExportDemoIssues() []ExportDemoIssue {
	out := make([]ExportDemoIssue, len(demoIssues))
	for i, iss := range demoIssues {
		out[i] = ExportDemoIssue{
			Key:           iss.Key,
			IssueType:     string(iss.IssueType),
			Title:         iss.Title,
			StatusName:    iss.StatusName,
			AssigneeEmail: iss.AssigneeEmail,
			Priority:      string(iss.Priority),
		}
	}
	return out
}

// ExportDemoTestStep — демо-шаг тест-кейса для hubtest.
type ExportDemoTestStep struct {
	Action         string
	ExpectedResult string
}

// ExportDemoTestCase — демо тест-кейс для hubtest.
type ExportDemoTestCase struct {
	CaseKey string
	Title   string
	Steps   []ExportDemoTestStep
}

// ExportDemoTestCases возвращает заданные демо тест-кейсы для hubtest.
func ExportDemoTestCases() []ExportDemoTestCase {
	out := make([]ExportDemoTestCase, len(demoTestCases))
	for i, tc := range demoTestCases {
		steps := make([]ExportDemoTestStep, len(tc.Steps))
		for j, step := range tc.Steps {
			steps[j] = ExportDemoTestStep{
				Action:         step.Action,
				ExpectedResult: step.ExpectedResult,
			}
		}
		out[i] = ExportDemoTestCase{
			CaseKey: tc.CaseKey,
			Title:   tc.Title,
			Steps:   steps,
		}
	}
	return out
}

// ExportDemoChannel — демо-канал для hubtest.
type ExportDemoChannel struct {
	Name string
	Type string
}

// ExportDemoChannels возвращает заданные демо-каналы для hubtest.
func ExportDemoChannels() []ExportDemoChannel {
	out := make([]ExportDemoChannel, len(demoChannels))
	for i, ch := range demoChannels {
		out[i] = ExportDemoChannel{
			Name: ch.Name,
			Type: string(ch.Type),
		}
	}
	return out
}

// ExportMethodOrder оборачивает methodOrder для hubtest.
func ExportMethodOrder(m string) int {
	return methodOrder(m)
}

// ExportMethodColor оборачивает methodColor для hubtest.
func ExportMethodColor(m string) string {
	return methodColor(m)
}

// ExportGuessModule оборачивает guessModule для hubtest.
func ExportGuessModule(path string) string {
	return guessModule(path)
}

// ExportRouteEntry — строка API-маршрута для hubtest.
type ExportRouteEntry struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// ExportCollectRoutes оборачивает collectRoutes для hubtest.
func ExportCollectRoutes(app *fiber.App) []ExportRouteEntry {
	routes := collectRoutes(app)
	out := make([]ExportRouteEntry, len(routes))
	for i, r := range routes {
		out[i] = ExportRouteEntry{Method: r.Method, Path: r.Path}
	}
	return out
}

// ExportSeedWorkflowTransitions — A3W7-5-002 mutation: Create-fail → err.
func ExportSeedWorkflowTransitions(h *Handler, ctx context.Context, projectID uuid.UUID, statusMap map[string]uuid.UUID) (int, error) {
	return h.seedWorkflowTransitions(ctx, projectID, statusMap)
}

// ExportSeedIssueLinks — A3W7-5-002.
func ExportSeedIssueLinks(h *Handler, ctx context.Context, issueMap map[string]uuid.UUID) (int, error) {
	return h.seedIssueLinks(ctx, issueMap)
}

// ExportSeedChannels — A3W7-5-002.
func ExportSeedChannels(h *Handler, ctx context.Context, projectID uuid.UUID, userIDs map[string]uuid.UUID) (int, error) {
	return h.seedChannels(ctx, projectID, userIDs)
}
