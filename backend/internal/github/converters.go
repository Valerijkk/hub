package github

import "github.com/valerii/hub/backend/ent"

// toIntegrationResponse конвертирует Ent GitHubIntegration в DTO ответа.
func toIntegrationResponse(ig *ent.GitHubIntegration) *IntegrationResponse {
	resp := &IntegrationResponse{
		ID:        ig.ID.String(),
		RepoOwner: ig.RepoOwner,
		RepoName:  ig.RepoName,
		CreatedAt: ig.CreatedAt,
	}
	if ig.MergeStatusID != nil {
		resp.MergeStatusID = ig.MergeStatusID.String()
	}
	return resp
}

// toGitHubLinkResponse конвертирует Ent GitHubLink в DTO ответа.
func toGitHubLinkResponse(l *ent.GitHubLink) *GitHubLinkResponse {
	return &GitHubLinkResponse{
		ID:         l.ID.String(),
		LinkType:   string(l.LinkType),
		ExternalID: l.ExternalID,
		Title:      l.Title,
		URL:        l.URL,
		Status:     string(l.Status),
		Author:     l.Author,
		Branch:     l.Branch,
		CreatedAt:  l.CreatedAt,
		UpdatedAt:  l.UpdatedAt,
	}
}
