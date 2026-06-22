package ai

import (
	"context"

	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

// IssueLookup — context issue для @hub-ai slash (M3-015..016).
type IssueLookup interface {
	IssueContext(ctx context.Context, issueKey string) (title, description string, err error)
}

// TrackerIssueLookup адаптирует TrackerService для bot commands.
type TrackerIssueLookup struct {
	tracker *trackermod.TrackerService
}

// NewTrackerIssueLookup — issue lookup через tracker.
func NewTrackerIssueLookup(tracker *trackermod.TrackerService) *TrackerIssueLookup {
	return &TrackerIssueLookup{tracker: tracker}
}

// IssueContext — title и plain-text description issue.
func (l *TrackerIssueLookup) IssueContext(ctx context.Context, issueKey string) (string, string, error) {
	if l == nil || l.tracker == nil {
		return "", "", nil
	}
	issue, err := l.tracker.GetIssue(ctx, issueKey)
	if err != nil {
		return "", "", err
	}
	desc := issue.DescriptionText
	if desc == "" && issue.DescriptionJSON != nil {
		desc = "(rich text description — см. задачу в трекере)"
	}
	return issue.Title, desc, nil
}
