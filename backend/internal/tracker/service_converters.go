package tracker

import (
	"github.com/valerii/hub/backend/ent"
)

// ---------------------------------------------------------------------------
// Конвертеры (Ent -> DTO)
// ---------------------------------------------------------------------------

// toIssueResponse конвертирует Ent Issue в DTO.
func toIssueResponse(i *ent.Issue) *IssueResponse {
	resp := &IssueResponse{
		ID:               i.ID.String(),
		IssueNumber:      i.IssueNumber,
		IssueKey:         i.IssueKey,
		Title:            i.Title,
		IssueType:        string(i.IssueType),
		Priority:         string(i.Priority),
		Rank:             i.Rank,
		Version:          i.Version,
		EstimateMinutes:  i.EstimateMinutes,
		StoryPoints:      i.StoryPoints,
		TimeSpentMinutes: i.TimeSpentMinutes,
		CustomFields:     i.CustomFields,
		DueDate:          i.DueDate,
		CreatedAt:        i.CreatedAt,
		UpdatedAt:        i.UpdatedAt,
	}

	// Конвертируем чеклист из JSONB в DTO
	if i.Checklist != nil {
		cl := make([]ChecklistItem, 0, len(i.Checklist))
		for _, raw := range i.Checklist {
			item := ChecklistItem{}
			if id, ok := raw["id"].(string); ok {
				item.ID = id
			}
			if text, ok := raw["text"].(string); ok {
				item.Text = text
			}
			if checked, ok := raw["checked"].(bool); ok {
				item.Checked = checked
			}
			cl = append(cl, item)
		}
		resp.Checklist = cl
	}

	if i.DescriptionJSON != nil {
		resp.DescriptionJSON = &i.DescriptionJSON
	}
	if i.DescriptionText != "" {
		resp.DescriptionText = i.DescriptionText
	}

	if i.Edges.Status != nil {
		resp.Status = *toWorkflowStatusResponse(i.Edges.Status)
	}
	if i.Edges.Assignee != nil {
		resp.Assignee = toUserBriefResponse(i.Edges.Assignee)
	}
	if i.Edges.Reporter != nil {
		resp.Reporter = toUserBriefResponse(i.Edges.Reporter)
	}
	if i.Edges.Parent != nil {
		resp.Parent = toIssueBriefResponse(i.Edges.Parent)
	}
	if i.Edges.Children != nil {
		resp.ChildrenCount = len(i.Edges.Children)
	}
	if i.Edges.Labels != nil {
		labels := make([]LabelResponse, 0, len(i.Edges.Labels))
		for _, l := range i.Edges.Labels {
			labels = append(labels, *toLabelResponse(l))
		}
		resp.Labels = labels
	}

	return resp
}

// toIssueBriefResponse конвертирует Ent Issue в краткий DTO.
func toIssueBriefResponse(i *ent.Issue) *IssueBriefResponse {
	status := ""
	if i.Edges.Status != nil {
		status = i.Edges.Status.Name
	}
	resp := &IssueBriefResponse{
		ID:          i.ID.String(),
		IssueKey:    i.IssueKey,
		Title:       i.Title,
		IssueType:   string(i.IssueType),
		Priority:    string(i.Priority),
		Status:      status,
		StoryPoints: i.StoryPoints,
	}

	// Конвертируем чеклист из JSONB в DTO
	if i.Checklist != nil {
		cl := make([]ChecklistItem, 0, len(i.Checklist))
		for _, raw := range i.Checklist {
			item := ChecklistItem{}
			if id, ok := raw["id"].(string); ok {
				item.ID = id
			}
			if text, ok := raw["text"].(string); ok {
				item.Text = text
			}
			if checked, ok := raw["checked"].(bool); ok {
				item.Checked = checked
			}
			cl = append(cl, item)
		}
		resp.Checklist = cl
	}

	return resp
}

// toUserBriefResponse конвертирует Ent User в краткий DTO.
func toUserBriefResponse(u *ent.User) *UserBriefResponse {
	return &UserBriefResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
	}
}

// toWorkflowStatusResponse конвертирует Ent WorkflowStatus в DTO.
func toWorkflowStatusResponse(st *ent.WorkflowStatus) *WorkflowStatusResponse {
	return &WorkflowStatusResponse{
		ID:        st.ID.String(),
		Name:      st.Name,
		Category:  string(st.Category),
		Color:     st.Color,
		SortOrder: st.SortOrder,
		IsDefault: st.IsDefault,
	}
}

// toCommentResponse конвертирует Ent Comment в DTO.
func toCommentResponse(c *ent.Comment) *CommentResponse {
	mentions := make([]string, len(c.Mentions))
	for i, id := range c.Mentions {
		mentions[i] = id.String()
	}
	resp := &CommentResponse{
		ID:         c.ID.String(),
		BodyJSON:   c.BodyJSON,
		BodyText:   c.BodyText,
		IsInternal: c.IsInternal,
		Mentions:   mentions,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
	if c.Edges.Author != nil {
		resp.Author = *toUserBriefResponse(c.Edges.Author)
	}
	if c.Edges.Reactions != nil {
		resp.Reactions = aggregateReactions(c.Edges.Reactions)
	}
	return resp
}

// aggregateReactions группирует реакции по emoji.
func aggregateReactions(reactions []*ent.Reaction) []ReactionResponse {
	grouped := make(map[string]*ReactionResponse)
	for _, r := range reactions {
		if _, ok := grouped[r.Emoji]; !ok {
			grouped[r.Emoji] = &ReactionResponse{
				Emoji: r.Emoji,
				Users: make([]UserBriefResponse, 0),
			}
		}
		entry := grouped[r.Emoji]
		entry.Count++
		if r.Edges.User != nil {
			entry.Users = append(entry.Users, *toUserBriefResponse(r.Edges.User))
		}
	}

	result := make([]ReactionResponse, 0, len(grouped))
	for _, v := range grouped {
		result = append(result, *v)
	}
	return result
}

// toActivityLogResponse конвертирует Ent ActivityLog в DTO.
func toActivityLogResponse(a *ent.ActivityLog) *ActivityLogResponse {
	resp := &ActivityLogResponse{
		ID:        a.ID.String(),
		Action:    a.Action,
		CreatedAt: a.CreatedAt,
	}

	if a.Changes != nil {
		changes := make([]FieldChange, len(a.Changes))
		for i, c := range a.Changes {
			changes[i] = FieldChange{
				Field:    c.Field,
				OldValue: c.OldValue,
				NewValue: c.NewValue,
			}
		}
		resp.Changes = changes
	}

	if a.Edges.User != nil {
		resp.User = *toUserBriefResponse(a.Edges.User)
	}
	return resp
}

// toTimeLogResponse конвертирует Ent TimeLog в DTO.
func toTimeLogResponse(tl *ent.TimeLog) *TimeLogResponse {
	resp := &TimeLogResponse{
		ID:          tl.ID.String(),
		Minutes:     tl.Minutes,
		Description: tl.Description,
		StartedAt:   tl.StartedAt,
		EndedAt:     tl.EndedAt,
		CreatedAt:   tl.CreatedAt,
	}
	if tl.Edges.User != nil {
		resp.User = *toUserBriefResponse(tl.Edges.User)
	}
	return resp
}

// toLabelResponse конвертирует Ent Label в DTO.
func toLabelResponse(l *ent.Label) *LabelResponse {
	resp := &LabelResponse{
		ID:    l.ID.String(),
		Name:  l.Name,
		Color: l.Color,
	}
	if l.Edges.Group != nil {
		resp.GroupName = l.Edges.Group.Name
	}
	return resp
}

// toLabelGroupResponse конвертирует Ent LabelGroup в DTO.
func toLabelGroupResponse(g *ent.LabelGroup) *LabelGroupResponse {
	resp := &LabelGroupResponse{
		ID:   g.ID.String(),
		Name: g.Name,
	}
	if g.Edges.Labels != nil {
		labels := make([]LabelResponse, 0, len(g.Edges.Labels))
		for _, l := range g.Edges.Labels {
			labels = append(labels, *toLabelResponse(l))
		}
		resp.Labels = labels
	}
	return resp
}

// toBoardResponse конвертирует Ent Board в DTO.
func toBoardResponse(b *ent.Board) *BoardResponse {
	return &BoardResponse{
		ID:        b.ID.String(),
		Name:      b.Name,
		Filters:   b.Filters,
		IsDefault: b.IsDefault,
		CreatedAt: b.CreatedAt,
	}
}

// toBoardDataResponse конвертирует Ent Board с eager-loaded данными в DTO.
func toBoardDataResponse(b *ent.Board) *BoardDataResponse {
	resp := &BoardDataResponse{
		Board: *toBoardResponse(b),
	}

	if b.Edges.Project != nil && b.Edges.Project.Edges.WorkflowStatuses != nil {
		columns := make([]BoardColumnResponse, 0, len(b.Edges.Project.Edges.WorkflowStatuses))
		for _, st := range b.Edges.Project.Edges.WorkflowStatuses {
			col := BoardColumnResponse{
				Status: *toWorkflowStatusResponse(st),
			}
			if st.Edges.Issues != nil {
				issues := make([]IssueBriefResponse, 0, len(st.Edges.Issues))
				for _, i := range st.Edges.Issues {
					issues = append(issues, *toIssueBriefResponse(i))
				}
				col.Issues = issues
			}
			columns = append(columns, col)
		}
		resp.Columns = columns
	}

	return resp
}

// toCustomFieldResponse конвертирует Ent CustomFieldDefinition в DTO.
func toCustomFieldResponse(cf *ent.CustomFieldDefinition) *CustomFieldDefinitionResponse {
	return &CustomFieldDefinitionResponse{
		ID:        cf.ID.String(),
		Name:      cf.Name,
		FieldType: string(cf.FieldType),
		Options:   cf.Options,
		Required:  cf.Required,
		SortOrder: cf.SortOrder,
	}
}

// toIssueLinkResponse конвертирует Ent IssueLink в DTO с информацией о связанной задаче.
func toIssueLinkResponse(l *ent.IssueLink, related *ent.Issue) *IssueLinkResponse {
	resp := &IssueLinkResponse{
		ID:         l.ID.String(),
		LinkType:   string(l.LinkType),
		CustomType: l.CustomType,
	}
	if related != nil {
		resp.Issue = *toIssueBriefResponse(related)
	}
	return resp
}
