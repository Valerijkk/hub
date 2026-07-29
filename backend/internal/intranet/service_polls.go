package intranet

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// GetNewsPoll — агрегаты опроса (FR-INTR-02/03).
// Percent виден после собственного голоса или автору/admin поста.
func (s *Service) GetNewsPoll(ctx context.Context, userID, newsID uuid.UUID) (*NewsPollResponse, error) {
	post, err := s.repo.GetNewsPost(ctx, newsID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news for poll: %w", err)
	}
	poll, err := s.repo.GetNewsPollByPostID(ctx, newsID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news poll: %w", err)
	}
	return s.buildPollResponse(ctx, userID, post, poll)
}

// VoteNewsPoll — один голос на пользователя (FR-INTR-02, NFR-31).
func (s *Service) VoteNewsPoll(ctx context.Context, userID, newsID uuid.UUID, req VoteNewsPollRequest) (*NewsPollResponse, error) {
	if req.OptionID == uuid.Nil {
		return nil, ErrInvalidInput
	}
	post, err := s.repo.GetNewsPost(ctx, newsID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news for vote: %w", err)
	}
	poll, err := s.repo.GetNewsPollByPostID(ctx, newsID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news poll for vote: %w", err)
	}
	if poll.Closed {
		return nil, ErrPollClosed
	}
	if err := s.repo.VoteNewsPoll(ctx, poll.ID, req.OptionID, userID); err != nil {
		return nil, err
	}
	poll, err = s.repo.GetNewsPollByPostID(ctx, newsID)
	if err != nil {
		return nil, fmt.Errorf("reload news poll: %w", err)
	}
	return s.buildPollResponse(ctx, userID, post, poll)
}

// UpdateNewsPoll — автор или admin; options только до первого голоса (FR-INTR-03).
func (s *Service) UpdateNewsPoll(ctx context.Context, userID, newsID uuid.UUID, req UpdateNewsPollRequest) (*NewsPollResponse, error) {
	post, err := s.repo.GetNewsPost(ctx, newsID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news for poll update: %w", err)
	}
	if err := s.requireAuthorOrAdmin(ctx, userID, post); err != nil {
		return nil, err
	}
	// A3W7-2-002: version обязателен (как news TASK-095); VersionEQ в repo.
	if req.Version < 1 {
		return nil, ErrInvalidInput
	}
	poll, err := s.repo.GetNewsPollByPostID(ctx, newsID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news poll for update: %w", err)
	}
	if req.Question != nil {
		q := strings.TrimSpace(*req.Question)
		if q == "" || len(q) > 500 {
			return nil, ErrInvalidInput
		}
		req.Question = &q
	}
	var options []string
	if req.Options != nil {
		if err := validatePollOptions(req.Options); err != nil {
			return nil, err
		}
		n, err := s.repo.CountNewsPollVotes(ctx, poll.ID)
		if err != nil {
			return nil, fmt.Errorf("count poll votes: %w", err)
		}
		if n > 0 {
			return nil, ErrPollHasVotes
		}
		options = req.Options
	}
	updated, err := s.repo.UpdateNewsPoll(ctx, poll.ID, UpdateNewsPollInput{
		Question: req.Question,
		Options:  options,
		Closed:   req.Closed,
		Version:  req.Version,
	})
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVersionConflict
		}
		return nil, fmt.Errorf("update news poll: %w", err)
	}
	return s.buildPollResponse(ctx, userID, post, updated)
}

func (s *Service) buildPollResponse(ctx context.Context, userID uuid.UUID, post *ent.NewsPost, poll *ent.NewsPoll) (*NewsPollResponse, error) {
	var votedID *uuid.UUID
	vote, err := s.repo.GetUserNewsPollVote(ctx, poll.ID, userID)
	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("get user vote: %w", err)
	}
	if vote != nil {
		id := vote.OptionID
		votedID = &id
	}
	isAuthor := post.Edges.Author != nil && post.Edges.Author.ID == userID
	showPercent := votedID != nil || isAuthor
	if !showPercent {
		if ok, err := s.repo.IsSystemAdmin(ctx, userID); err == nil && ok {
			showPercent = true
		}
	}
	total := 0
	opts := make([]NewsPollOptionResponse, 0, len(poll.Edges.Options))
	for _, o := range poll.Edges.Options {
		total += o.VotesCount
		item := NewsPollOptionResponse{
			ID:         o.ID,
			Text:       o.Text,
			VotesCount: o.VotesCount,
		}
		opts = append(opts, item)
	}
	if showPercent && total > 0 {
		for i := range opts {
			p := float64(opts[i].VotesCount) * 100 / float64(total)
			opts[i].Percent = &p
		}
	} else if showPercent {
		for i := range opts {
			z := 0.0
			opts[i].Percent = &z
		}
	}
	return &NewsPollResponse{
		ID:            poll.ID,
		Question:      poll.Question,
		Closed:        poll.Closed,
		Version:       poll.Version,
		VotedOptionID: votedID,
		Options:       opts,
		TotalVotes:    total,
	}, nil
}

func validatePollPayload(p *CreatePollPayload) error {
	if p == nil {
		return nil
	}
	q := strings.TrimSpace(p.Question)
	if q == "" || len(q) > 500 {
		return ErrInvalidInput
	}
	p.Question = q
	return validatePollOptions(p.Options)
}

func validatePollOptions(options []string) error {
	if len(options) < 2 || len(options) > 10 {
		return ErrInvalidInput
	}
	for i, o := range options {
		t := strings.TrimSpace(o)
		if t == "" || len(t) > 200 {
			return ErrInvalidInput
		}
		options[i] = t
	}
	return nil
}

func (s *Service) requireAuthorOrAdmin(ctx context.Context, userID uuid.UUID, post *ent.NewsPost) error {
	if post.Edges.Author != nil && post.Edges.Author.ID == userID {
		return nil
	}
	return s.requireAdmin(ctx, userID)
}
