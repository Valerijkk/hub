package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Доски
// ---------------------------------------------------------------------------

// CreateBoard создаёт доску в проекте.
func (s *TrackerService) CreateBoard(ctx context.Context, projectID uuid.UUID, req CreateBoardRequest) (*BoardResponse, error) {
	b, err := s.repo.CreateBoard(ctx, projectID, req.Name, req.Filters)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания доски")
	}
	return toBoardResponse(b), nil
}

// ListBoards возвращает все доски проекта.
func (s *TrackerService) ListBoards(ctx context.Context, projectID uuid.UUID) ([]BoardResponse, error) {
	boards, err := s.repo.ListBoards(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения досок")
	}
	result := make([]BoardResponse, 0, len(boards))
	for _, b := range boards {
		result = append(result, *toBoardResponse(b))
	}
	return result, nil
}

// GetBoardData возвращает данные доски с задачами, сгруппированными по статусам.
// Если передан параметр swimlane (assignee|priority|type), группирует задачи по этому полю.
func (s *TrackerService) GetBoardData(ctx context.Context, boardID uuid.UUID, swimlane string) (*BoardDataResponse, error) {
	b, err := s.repo.GetBoardData(ctx, boardID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("доска не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения данных доски")
	}
	resp := toBoardDataResponse(b)

	// Если swimlane задан, группируем задачи дополнительно
	if swimlane != "" && b.Edges.Project != nil && b.Edges.Project.Edges.WorkflowStatuses != nil {
		resp.Swimlanes = buildSwimlanes(b.Edges.Project.Edges.WorkflowStatuses, swimlane)
	}

	return resp, nil
}

// buildSwimlanes группирует задачи всех колонок по указанному полю (assignee/priority/type).
func buildSwimlanes(statuses []*ent.WorkflowStatus, swimlane string) []SwimlaneGroup {
	// Собираем уникальные ключи swimlane
	groupedIssues := make(map[string]map[string][]*ent.Issue) // groupKey -> statusID -> issues
	groupLabels := make(map[string]string)

	for _, st := range statuses {
		if st.Edges.Issues == nil {
			continue
		}
		for _, iss := range st.Edges.Issues {
			key, lbl := swimlaneKeyLabel(iss, swimlane)
			if _, ok := groupedIssues[key]; !ok {
				groupedIssues[key] = make(map[string][]*ent.Issue)
			}
			groupedIssues[key][st.ID.String()] = append(groupedIssues[key][st.ID.String()], iss)
			groupLabels[key] = lbl
		}
	}

	var swimlanes []SwimlaneGroup
	for key, statusMap := range groupedIssues {
		sg := SwimlaneGroup{
			Key:   key,
			Label: groupLabels[key],
		}
		for _, st := range statuses {
			col := BoardColumnResponse{
				Status: *toWorkflowStatusResponse(st),
			}
			if issues, ok := statusMap[st.ID.String()]; ok {
				briefs := make([]IssueBriefResponse, 0, len(issues))
				for _, i := range issues {
					briefs = append(briefs, *toIssueBriefResponse(i))
				}
				col.Issues = briefs
			}
			sg.Columns = append(sg.Columns, col)
		}
		swimlanes = append(swimlanes, sg)
	}
	return swimlanes
}

// swimlaneKeyLabel возвращает ключ и метку для группировки задачи по swimlane-полю.
func swimlaneKeyLabel(iss *ent.Issue, swimlane string) (string, string) {
	switch swimlane {
	case "assignee":
		if iss.Edges.Assignee != nil {
			a := iss.Edges.Assignee
			name := a.Name
			if name == "" {
				name = a.Email
			}
			return a.ID.String(), name
		}
		return "unassigned", "Не назначен"
	case "priority":
		p := string(iss.Priority)
		if p == "" {
			p = "medium"
		}
		return p, p
	case "type":
		t := string(iss.IssueType)
		return t, t
	default:
		return "all", "Все"
	}
}
