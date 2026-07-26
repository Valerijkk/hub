package ai

import (
	"context"
	"strings"

	"github.com/valerii/hub/backend/ent"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

// IssueLookup — context issue для @hub-ai slash (M3-015..016).
type IssueLookup interface {
	IssueContext(ctx context.Context, issueKey string) (title, description string, err error)
}

// ProjectByKey — резолв projectID по ключу проекта (часть issue key до последнего '-').
type ProjectByKey interface {
	FindByKey(ctx context.Context, key string) (*ent.Project, error)
}

// TrackerIssueLookup адаптирует TrackerService для bot commands.
type TrackerIssueLookup struct {
	tracker  *trackermod.TrackerService
	projects ProjectByKey
}

// NewTrackerIssueLookup — issue lookup через tracker + project key (A3-002 scope).
func NewTrackerIssueLookup(tracker *trackermod.TrackerService, projects ProjectByKey) *TrackerIssueLookup {
	return &TrackerIssueLookup{tracker: tracker, projects: projects}
}

// IssueContext — title и plain-text description issue.
func (l *TrackerIssueLookup) IssueContext(ctx context.Context, issueKey string) (string, string, error) {
	if l == nil || l.tracker == nil || l.projects == nil {
		return "", "", nil
	}
	projectKey := projectKeyFromIssueKey(issueKey)
	if projectKey == "" {
		return "", "", &ent.NotFoundError{}
	}
	proj, err := l.projects.FindByKey(ctx, projectKey)
	if err != nil {
		return "", "", err
	}
	issue, err := l.tracker.GetIssue(ctx, proj.ID, issueKey)
	if err != nil {
		return "", "", err
	}
	desc := issue.DescriptionText
	if desc == "" && issue.DescriptionJSON != nil {
		desc = "(rich text description — см. задачу в трекере)"
	}
	return issue.Title, desc, nil
}

func projectKeyFromIssueKey(issueKey string) string {
	i := strings.LastIndex(issueKey, "-")
	if i <= 0 {
		return ""
	}
	return issueKey[:i]
}
