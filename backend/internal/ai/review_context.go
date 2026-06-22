package ai

import (
	"fmt"
	"strings"
)

// ReviewContext — unified PR/MR для /ai review (M3-017).
type ReviewContext struct {
	URL          string
	Title        string
	Body         string
	Author       string
	BaseBranch   string
	HeadBranch   string
	ChangedFiles []string
	PatchSummary string
}

func buildReviewPrompt(rc *ReviewContext) (string, string) {
	system := `Ты senior code reviewer. Проанализируй изменения и дай структурированное ревью на русском:
**Summary**, **Риски**, **Замечания по файлам**, **Рекомендации**. Будь конкретным.`
	if rc == nil {
		return system, "Проанализируй pull/merge request."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "URL: %s\n", rc.URL)
	fmt.Fprintf(&b, "Title: %s\n", rc.Title)
	fmt.Fprintf(&b, "Author: %s\n", rc.Author)
	if rc.BaseBranch != "" || rc.HeadBranch != "" {
		fmt.Fprintf(&b, "Branch: %s → %s\n", rc.HeadBranch, rc.BaseBranch)
	}
	if rc.Body != "" {
		fmt.Fprintf(&b, "\nDescription:\n%s\n", rc.Body)
	}
	if len(rc.ChangedFiles) > 0 {
		fmt.Fprintf(&b, "\nChanged files: %s\n", strings.Join(rc.ChangedFiles, ", "))
	}
	if rc.PatchSummary != "" {
		fmt.Fprintf(&b, "\nDiff excerpts:\n%s\n", rc.PatchSummary)
	}
	return system, strings.TrimSpace(b.String())
}
