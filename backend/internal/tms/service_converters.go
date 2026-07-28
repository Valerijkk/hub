package tms

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
)

// ---------------------------------------------------------------------------
// Конвертеры Ent → DTO
// ---------------------------------------------------------------------------

// toTestCaseResponse конвертирует Ent TestCase в DTO.
func toTestCaseResponse(tc *ent.TestCase) *TestCaseResponse {
	resp := &TestCaseResponse{
		ID:            tc.ID.String(),
		CaseNumber:    tc.CaseNumber,
		CaseKey:       tc.CaseKey,
		Title:         tc.Title,
		Description:   tc.Description,
		Precondition:  tc.Precondition,
		Postcondition: tc.Postcondition,
		CaseType:      string(tc.CaseType),
		Priority:      string(tc.Priority),
		Status:        string(tc.Status),
		Steps:         tc.Steps,
		FullName:      tc.FullName,
		ClassName:     tc.ClassName,
		MethodName:    tc.MethodName,
		ExternalID:    tc.ExternalID,
		AutomationID:  tc.AutomationID,
		Version:       tc.Version,
		CreatedAt:     tc.CreatedAt,
		UpdatedAt:     tc.UpdatedAt,
	}

	if tc.Edges.Suite != nil {
		resp.Suite = &SuiteBriefResponse{
			ID:   tc.Edges.Suite.ID.String(),
			Name: tc.Edges.Suite.Name,
		}
	}
	if tc.Edges.CreatedBy != nil {
		resp.CreatedBy = &UserBriefResponse{
			ID:        tc.Edges.CreatedBy.ID.String(),
			Name:      tc.Edges.CreatedBy.Name,
			Email:     tc.Edges.CreatedBy.Email,
			AvatarURL: tc.Edges.CreatedBy.AvatarURL,
		}
	}
	if tc.Edges.Labels != nil {
		labels := make([]LabelResponse, 0, len(tc.Edges.Labels))
		for _, l := range tc.Edges.Labels {
			labels = append(labels, LabelResponse{
				ID:    l.ID.String(),
				Name:  l.Name,
				Color: l.Color,
			})
		}
		resp.Labels = labels
	}
	if tc.Edges.Components != nil {
		components := make([]ComponentResponse, 0, len(tc.Edges.Components))
		for _, c := range tc.Edges.Components {
			components = append(components, ComponentResponse{
				ID:          c.ID.String(),
				Name:        c.Name,
				Description: c.Description,
				CreatedAt:   c.CreatedAt,
			})
		}
		resp.Components = components
	}
	if tc.Edges.LinkedIssues != nil {
		issues := make([]IssueBriefResponse, 0, len(tc.Edges.LinkedIssues))
		for _, i := range tc.Edges.LinkedIssues {
			issues = append(issues, IssueBriefResponse{
				ID:       i.ID.String(),
				IssueKey: i.IssueKey,
				Title:    i.Title,
			})
		}
		resp.LinkedIssues = issues
	}

	return resp
}

// toTestCaseVersionResponse конвертирует Ent TestCaseVersion в DTO.
func toTestCaseVersionResponse(v *ent.TestCaseVersion) *TestCaseVersionResponse {
	resp := &TestCaseVersionResponse{
		ID:                v.ID.String(),
		VersionNumber:     v.VersionNumber,
		Snapshot:          v.Snapshot,
		ChangeDescription: v.ChangeDescription,
		CreatedAt:         v.CreatedAt,
	}
	if v.Edges.CreatedBy != nil {
		resp.CreatedBy = &UserBriefResponse{
			ID:        v.Edges.CreatedBy.ID.String(),
			Name:      v.Edges.CreatedBy.Name,
			Email:     v.Edges.CreatedBy.Email,
			AvatarURL: v.Edges.CreatedBy.AvatarURL,
		}
	}
	return resp
}

// toSuiteResponse конвертирует Ent TestSuite в DTO.
func toSuiteResponse(suite *ent.TestSuite) *SuiteResponse {
	resp := &SuiteResponse{
		ID:          suite.ID.String(),
		Name:        suite.Name,
		Description: suite.Description,
		SortOrder:   suite.SortOrder,
		CreatedAt:   suite.CreatedAt,
		UpdatedAt:   suite.UpdatedAt,
	}
	if suite.Edges.Parent != nil {
		resp.ParentID = suite.Edges.Parent.ID.String()
	}
	return resp
}

// toComponentResponse конвертирует Ent Component в DTO.
func toComponentResponse(c *ent.Component) *ComponentResponse {
	return &ComponentResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}

// toTestRunResponse конвертирует Ent TestRun в DTO.
func toTestRunResponse(run *ent.TestRun) *TestRunResponse {
	resp := &TestRunResponse{
		ID:           run.ID.String(),
		Title:        run.Title,
		Description:  run.Description,
		RunType:      string(run.RunType),
		Status:       string(run.Status),
		TotalCount:   run.TotalCount,
		PassedCount:  run.PassedCount,
		FailedCount:  run.FailedCount,
		SkippedCount: run.SkippedCount,
		BrokenCount:  run.BrokenCount,
		UnknownCount: run.UnknownCount,
		AllureRunID:  run.AllureRunID,
		StartedAt:    run.StartedAt,
		CompletedAt:  run.CompletedAt,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
	}
	if run.Edges.CreatedBy != nil {
		resp.CreatedBy = &UserBriefResponse{
			ID:        run.Edges.CreatedBy.ID.String(),
			Name:      run.Edges.CreatedBy.Name,
			Email:     run.Edges.CreatedBy.Email,
			AvatarURL: run.Edges.CreatedBy.AvatarURL,
		}
	}
	return resp
}

// toTestResultResponse конвертирует Ent TestResult в DTO.
func toTestResultResponse(r *ent.TestResult) *TestResultResponse {
	resp := &TestResultResponse{
		ID:            r.ID.String(),
		Status:        string(r.Status),
		Comment:       r.Comment,
		StepResults:   r.StepResults,
		DurationMs:    r.DurationMs,
		AllureUUID:    r.AllureUUID,
		ErrorMessage:  r.ErrorMessage,
		ErrorTrace:    r.ErrorTrace,
		IsFlaky:       r.IsFlaky,
		MatchStrategy: string(r.MatchStrategy),
		StartedAt:     r.StartedAt,
		CompletedAt:   r.CompletedAt,
		CreatedAt:     r.CreatedAt,
	}
	if r.Edges.TestCase != nil {
		resp.TestCaseKey = r.Edges.TestCase.CaseKey
	}
	if r.Edges.ExecutedBy != nil {
		resp.ExecutedBy = &UserBriefResponse{
			ID:        r.Edges.ExecutedBy.ID.String(),
			Name:      r.Edges.ExecutedBy.Name,
			Email:     r.Edges.ExecutedBy.Email,
			AvatarURL: r.Edges.ExecutedBy.AvatarURL,
		}
	}
	return resp
}

// toTestRunDetailResponse конвертирует Ent TestRun (с results и assignees) в DTO.
func toTestRunDetailResponse(run *ent.TestRun) *TestRunDetailResponse {
	resp := &TestRunDetailResponse{
		TestRunResponse: *toTestRunResponse(run),
	}

	// Результаты
	resp.Results = make([]TestResultDetailResponse, 0, len(run.Edges.Results))
	for _, r := range run.Edges.Results {
		detail := TestResultDetailResponse{
			ID:          r.ID.String(),
			Status:      string(r.Status),
			Comment:     r.Comment,
			StepResults: r.StepResults,
			DurationMs:  r.DurationMs,
			IsFlaky:     r.IsFlaky,
			StartedAt:   r.StartedAt,
			CompletedAt: r.CompletedAt,
			CreatedAt:   r.CreatedAt,
		}
		if r.Edges.TestCase != nil {
			tc := r.Edges.TestCase
			detail.TestCase = &TestCaseBriefDTO{
				ID:       tc.ID.String(),
				CaseKey:  tc.CaseKey,
				Title:    tc.Title,
				CaseType: string(tc.CaseType),
				Priority: string(tc.Priority),
			}
		}
		if r.Edges.ExecutedBy != nil {
			detail.ExecutedBy = &UserBriefResponse{
				ID:        r.Edges.ExecutedBy.ID.String(),
				Name:      r.Edges.ExecutedBy.Name,
				Email:     r.Edges.ExecutedBy.Email,
				AvatarURL: r.Edges.ExecutedBy.AvatarURL,
			}
		}
		if r.Edges.Files != nil {
			detail.Files = make([]FileBriefResponse, 0, len(r.Edges.Files))
			for _, f := range r.Edges.Files {
				detail.Files = append(detail.Files, FileBriefResponse{
					ID:       f.ID.String(),
					FileName: f.Filename,
					FileSize: f.Size,
					MimeType: f.ContentType,
				})
			}
		}
		resp.Results = append(resp.Results, detail)
	}

	// Назначенные тестировщики
	resp.Assignees = make([]UserBriefResponse, 0, len(run.Edges.Assignees))
	for _, u := range run.Edges.Assignees {
		resp.Assignees = append(resp.Assignees, UserBriefResponse{
			ID:        u.ID.String(),
			Name:      u.Name,
			Email:     u.Email,
			AvatarURL: u.AvatarURL,
		})
	}

	return resp
}

// ---------------------------------------------------------------------------
// Построение дерева папок
// ---------------------------------------------------------------------------

// buildSuiteTree строит дерево папок из плоского списка.
func buildSuiteTree(suites []*ent.TestSuite) []SuiteResponse {
	// id → ответ
	byID := make(map[uuid.UUID]*SuiteResponse, len(suites))
	for _, s := range suites {
		resp := toSuiteResponse(s)
		resp.Children = make([]SuiteResponse, 0)
		byID[s.ID] = resp
	}

	// Сборка дерева
	var roots []SuiteResponse
	for _, s := range suites {
		if s.Edges.Parent == nil {
			roots = append(roots, *byID[s.ID])
		} else {
			parentResp, ok := byID[s.Edges.Parent.ID]
			if ok {
				parentResp.Children = append(parentResp.Children, *byID[s.ID])
			}
		}
	}

	if roots == nil {
		roots = make([]SuiteResponse, 0)
	}
	return roots
}

// ---------------------------------------------------------------------------
// Вспомогательные функции снимков версий
// ---------------------------------------------------------------------------

// buildSnapshot создаёт снимок тест-кейса для версионирования.
// A2W2-2-010: ошибки Marshal/Unmarshal steps не глотаются.
func buildSnapshot(tc *ent.TestCase) (map[string]interface{}, error) {
	snapshot := map[string]interface{}{
		"title":         tc.Title,
		"description":   tc.Description,
		"precondition":  tc.Precondition,
		"postcondition": tc.Postcondition,
		"case_type":     string(tc.CaseType),
		"priority":      string(tc.Priority),
		"status":        string(tc.Status),
		"version":       tc.Version,
	}
	if len(tc.Steps) > 0 {
		stepsJSON, err := json.Marshal(tc.Steps)
		if err != nil {
			return nil, fmt.Errorf("сериализация steps для snapshot: %w", err)
		}
		var steps interface{}
		if err := json.Unmarshal(stepsJSON, &steps); err != nil {
			return nil, fmt.Errorf("десериализация steps для snapshot: %w", err)
		}
		snapshot["steps"] = steps
	}
	if tc.FullName != "" {
		snapshot["full_name"] = tc.FullName
	}
	if tc.ClassName != "" {
		snapshot["class_name"] = tc.ClassName
	}
	if tc.MethodName != "" {
		snapshot["method_name"] = tc.MethodName
	}
	if tc.ExternalID != "" {
		snapshot["external_id"] = tc.ExternalID
	}
	return snapshot, nil
}

// snapshotToUpdateInput конвертирует снимок версии в UpdateTestCaseInput для rollback.
func snapshotToUpdateInput(snapshot map[string]interface{}) UpdateTestCaseInput {
	input := UpdateTestCaseInput{}

	if v, ok := snapshot["title"].(string); ok {
		input.Title = &v
	}
	if v, ok := snapshot["description"].(string); ok {
		input.Description = &v
	}
	if v, ok := snapshot["precondition"].(string); ok {
		input.Precondition = &v
	}
	if v, ok := snapshot["postcondition"].(string); ok {
		input.Postcondition = &v
	}
	if v, ok := snapshot["case_type"].(string); ok {
		input.CaseType = &v
	}
	if v, ok := snapshot["priority"].(string); ok {
		input.Priority = &v
	}
	if v, ok := snapshot["status"].(string); ok {
		input.Status = &v
	}
	if v, ok := snapshot["steps"]; ok {
		stepsJSON, _ := json.Marshal(v)
		var steps []schema.TestStep
		if err := json.Unmarshal(stepsJSON, &steps); err == nil {
			input.Steps = &steps
		}
	}
	if v, ok := snapshot["full_name"].(string); ok {
		input.FullName = &v
	}
	if v, ok := snapshot["class_name"].(string); ok {
		input.ClassName = &v
	}
	if v, ok := snapshot["method_name"].(string); ok {
		input.MethodName = &v
	}
	if v, ok := snapshot["external_id"].(string); ok {
		input.ExternalID = &v
	}

	return input
}

// normalizeErrorPattern извлекает основной паттерн из сообщения об ошибке (первые 100 рун).
// P3-BE-04 / W2-A2-019 / TASK-060 residual: msg[:100] по байтам ломал UTF-8.
func normalizeErrorPattern(msg string) string {
	runes := []rune(msg)
	if len(runes) > 100 {
		return string(runes[:100])
	}
	return msg
}
