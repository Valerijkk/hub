package calendar

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/teambition/rrule-go"
	"github.com/valerii/hub/backend/ent"
)

const (
	defaultCallDuration = 60 * time.Minute
	maxRangeDays        = 366
)

var (
	ErrInvalidRange = errors.New("invalid calendar range")
	ErrRangeTooWide = errors.New("calendar range exceeds maximum")
)

// Service — агрегация событий календаря (MOD-04).
type Service struct {
	repo Repository
}

// NewService создаёт сервис календаря.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListEvents возвращает звонки, спринты и дедлайны пользователя за период.
func (s *Service) ListEvents(ctx context.Context, userID uuid.UUID, from, to time.Time) (*EventsResponse, error) {
	from = from.UTC()
	to = to.UTC()
	if err := validateRange(from, to); err != nil {
		return nil, err
	}

	projectIDs, err := s.repo.ListUserProjectIDs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user projects: %w", err)
	}

	events := make([]EventResponse, 0)

	calls, err := s.repo.ListScheduledCalls(ctx, projectIDs)
	if err != nil {
		return nil, fmt.Errorf("list scheduled calls: %w", err)
	}
	for _, sc := range calls {
		events = append(events, expandCallEvents(sc, from, to)...)
	}

	sprints, err := s.repo.ListSprintsInRange(ctx, projectIDs, from, to)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}
	for _, sp := range sprints {
		events = append(events, sprintEvent(sp))
	}

	issues, err := s.repo.ListIssuesDueInRange(ctx, projectIDs, from, to)
	if err != nil {
		return nil, fmt.Errorf("list issue deadlines: %w", err)
	}
	for _, iss := range issues {
		if ev, ok := deadlineEvent(iss); ok {
			events = append(events, ev)
		}
	}

	leaves, err := s.repo.ListApprovedLeaveInRange(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("list approved leave: %w", err)
	}
	for _, lv := range leaves {
		events = append(events, leaveEvent(lv))
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].StartsAt.Equal(events[j].StartsAt) {
			return events[i].Title < events[j].Title
		}
		return events[i].StartsAt.Before(events[j].StartsAt)
	})

	return &EventsResponse{Events: events, From: from, To: to}, nil
}

func validateRange(from, to time.Time) error {
	if to.Before(from) {
		return ErrInvalidRange
	}
	if to.Sub(from) > maxRangeDays*24*time.Hour {
		return ErrRangeTooWide
	}
	return nil
}

func expandCallEvents(sc *ent.ScheduledCall, from, to time.Time) []EventResponse {
	projectKey := ""
	if sc.Edges.Project != nil {
		projectKey = sc.Edges.Project.Key
	}
	url := fmt.Sprintf("/calls?channel=%s", sc.ChannelID)

	var occurrences []time.Time
	if sc.Rrule != "" {
		rule, err := parseRRule(sc.ScheduledAt, sc.Rrule)
		if err != nil {
			return nil
		}
		occurrences = rule.Between(from, to, true)
	} else if !sc.ScheduledAt.Before(from) && !sc.ScheduledAt.After(to) {
		occurrences = []time.Time{sc.ScheduledAt}
	}

	out := make([]EventResponse, 0, len(occurrences))
	for _, at := range occurrences {
		end := at.Add(defaultCallDuration)
		pid := sc.ProjectID
		out = append(out, EventResponse{
			ID:          fmt.Sprintf("call-%s-%s", sc.ID, at.UTC().Format(time.RFC3339)),
			Type:        EventTypeCall,
			Title:       sc.Title,
			Description: sc.Description,
			StartsAt:    at.UTC(),
			EndsAt:      &end,
			AllDay:      false,
			ProjectID:   &pid,
			ProjectKey:  projectKey,
			URL:         url,
		})
	}
	return out
}

func parseRRule(base time.Time, rruleStr string) (*rrule.RRule, error) {
	upper := strings.ToUpper(rruleStr)
	if !strings.Contains(upper, "DTSTART=") {
		dtstart := base.UTC().Format("20060102T150405Z")
		rruleStr = "DTSTART=" + dtstart + ";" + rruleStr
	}
	return rrule.StrToRRule(rruleStr)
}

func sprintEvent(sp *ent.Sprint) EventResponse {
	projectKey := ""
	var projectID *uuid.UUID
	if sp.Edges.Project != nil {
		projectKey = sp.Edges.Project.Key
		pid := sp.Edges.Project.ID
		projectID = &pid
	}
	start := sp.StartDate.UTC()
	end := sp.EndDate.UTC()
	url := ""
	if projectID != nil {
		url = fmt.Sprintf("/projects/%s/sprints/%s", projectID, sp.ID)
	}
	return EventResponse{
		ID:          "sprint-" + sp.ID.String(),
		Type:        EventTypeSprint,
		Title:       "Спринт: " + sp.Name,
		Description: sp.Goal,
		StartsAt:    start,
		EndsAt:      &end,
		AllDay:      true,
		ProjectID:   projectID,
		ProjectKey:  projectKey,
		URL:         url,
	}
}

func deadlineEvent(iss *ent.Issue) (EventResponse, bool) {
	if iss.DueDate == nil {
		return EventResponse{}, false
	}
	projectKey := ""
	var projectID *uuid.UUID
	if iss.Edges.Project != nil {
		projectKey = iss.Edges.Project.Key
		pid := iss.Edges.Project.ID
		projectID = &pid
	}
	due := iss.DueDate.UTC()
	url := ""
	if projectID != nil {
		url = fmt.Sprintf("/projects/%s/issues/%s", projectID, iss.ID)
	}
	return EventResponse{
		ID:         "deadline-" + iss.ID.String(),
		Type:       EventTypeDeadline,
		Title:      iss.IssueKey + ": " + iss.Title,
		StartsAt:   due,
		AllDay:     true,
		ProjectID:  projectID,
		ProjectKey: projectKey,
		URL:        url,
	}, true
}

func leaveEvent(lv *ent.LeaveRequest) EventResponse {
	name := "Сотрудник"
	if lv.Edges.Requester != nil {
		name = lv.Edges.Requester.DisplayName
		if name == "" {
			name = lv.Edges.Requester.Name
		}
		if name == "" {
			name = lv.Edges.Requester.Email
		}
	}
	start := lv.StartDate.UTC()
	end := lv.EndDate.UTC()
	typeLabel := map[string]string{
		"vacation": "Отпуск",
		"sick":     "Больничный",
		"personal": "Отгул",
		"other":    "Отсутствие",
	}[string(lv.LeaveType)]
	if typeLabel == "" {
		typeLabel = "Отсутствие"
	}
	title := fmt.Sprintf("%s: %s", name, typeLabel)
	return EventResponse{
		ID:        "leave-" + lv.ID.String(),
		Type:      EventTypeLeave,
		Title:     title,
		StartsAt:  start,
		EndsAt:    &end,
		AllDay:    true,
		URL:       "/hr",
	}
}
