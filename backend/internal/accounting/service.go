package accounting

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/journalentry"
)

var (
	ErrNotFound        = errors.New("accounting entity not found")
	ErrForbidden       = errors.New("admin access required")
	ErrInvalidInput    = errors.New("invalid accounting input")
	ErrUnbalanced      = errors.New("journal entry not balanced")
	ErrAlreadyPosted   = errors.New("journal entry already posted")
)

// Service — двойная бухгалтерия (V4-03).
type Service struct {
	repo Repository
}

// NewService создаёт accounting service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAccounts(ctx context.Context, userID uuid.UUID) ([]AccountResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureDefaultChart(ctx); err != nil {
		return nil, fmt.Errorf("ensure chart: %w", err)
	}
	items, err := s.repo.ListAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := make([]AccountResponse, 0, len(items))
	for _, a := range items {
		out = append(out, toAccountResponse(a))
	}
	return out, nil
}

func (s *Service) CreateAccount(ctx context.Context, userID uuid.UUID, req CreateAccountRequest) (*AccountResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	if req.Code == "" || req.Name == "" || req.AccountType == "" {
		return nil, ErrInvalidInput
	}
	a, err := s.repo.CreateAccount(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	resp := toAccountResponse(a)
	return &resp, nil
}

func (s *Service) ListJournal(ctx context.Context, userID uuid.UUID) ([]JournalEntryResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListJournalEntries(ctx, 50)
	if err != nil {
		return nil, fmt.Errorf("list journal: %w", err)
	}
	out := make([]JournalEntryResponse, 0, len(items))
	for _, e := range items {
		out = append(out, toJournalResponse(e))
	}
	return out, nil
}

func (s *Service) CreateJournal(ctx context.Context, userID uuid.UUID, req CreateJournalRequest) (*JournalEntryResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	req.Description = strings.TrimSpace(req.Description)
	if req.Description == "" || len(req.Lines) < 2 {
		return nil, ErrInvalidInput
	}
	entryDate, err := parseEntryDate(req.EntryDate)
	if err != nil {
		return nil, ErrInvalidInput
	}
	if err := validateLines(ctx, s.repo, req.Lines); err != nil {
		return nil, err
	}
	e, err := s.repo.CreateJournalEntry(ctx, userID, entryDate, req)
	if err != nil {
		return nil, fmt.Errorf("create journal: %w", err)
	}
	resp := toJournalResponse(e)
	return &resp, nil
}

func (s *Service) PostJournal(ctx context.Context, userID, entryID uuid.UUID) (*JournalEntryResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	current, err := s.repo.GetJournalEntry(ctx, entryID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get journal: %w", err)
	}
	if current.Status == journalentry.StatusPosted {
		return nil, ErrAlreadyPosted
	}
	lines := make([]JournalLineInput, 0, len(current.Edges.Lines))
	for _, l := range current.Edges.Lines {
		if l.Edges.Account == nil {
			return nil, ErrInvalidInput
		}
		lines = append(lines, JournalLineInput{
			AccountID:   l.Edges.Account.ID,
			DebitCents:  l.DebitCents,
			CreditCents: l.CreditCents,
		})
	}
	if err := validateLines(ctx, s.repo, lines); err != nil {
		return nil, err
	}
	e, err := s.repo.PostJournalEntry(ctx, entryID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("post journal: %w", err)
	}
	resp := toJournalResponse(e)
	return &resp, nil
}

func (s *Service) TrialBalance(ctx context.Context, userID uuid.UUID) ([]TrialBalanceRow, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	if err := s.repo.EnsureDefaultChart(ctx); err != nil {
		return nil, fmt.Errorf("ensure chart: %w", err)
	}
	return s.repo.TrialBalance(ctx)
}

func (s *Service) requireAdmin(ctx context.Context, userID uuid.UUID) error {
	ok, err := s.repo.IsSystemAdmin(ctx, userID)
	if err != nil {
		return fmt.Errorf("check admin: %w", err)
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func validateLines(ctx context.Context, repo Repository, lines []JournalLineInput) error {
	var debit, credit int64
	for _, l := range lines {
		if l.DebitCents < 0 || l.CreditCents < 0 {
			return ErrInvalidInput
		}
		if l.DebitCents > 0 && l.CreditCents > 0 {
			return ErrInvalidInput
		}
		if l.DebitCents == 0 && l.CreditCents == 0 {
			return ErrInvalidInput
		}
		ok, err := repo.AccountExists(ctx, l.AccountID)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvalidInput
		}
		debit += l.DebitCents
		credit += l.CreditCents
	}
	if debit != credit || debit == 0 {
		return ErrUnbalanced
	}
	return nil
}

func parseEntryDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Now().UTC().Truncate(24 * time.Hour), nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func toAccountResponse(a *ent.LedgerAccount) AccountResponse {
	return AccountResponse{
		ID:          a.ID,
		Code:        a.Code,
		Name:        a.Name,
		AccountType: AccountType(a.AccountType),
		IsActive:    a.IsActive,
		Version:     a.Version,
	}
}

func toJournalResponse(e *ent.JournalEntry) JournalEntryResponse {
	resp := JournalEntryResponse{
		ID:          e.ID,
		EntryDate:   e.EntryDate.UTC().Format("2006-01-02"),
		Description: e.Description,
		Reference:   e.Reference,
		Status:      EntryStatus(e.Status),
		Version:     e.Version,
		PostedAt:    e.PostedAt,
		CreatedAt:   e.CreatedAt,
	}
	if e.Edges.CreatedBy != nil {
		resp.CreatedBy = e.Edges.CreatedBy.ID
	}
	for _, l := range e.Edges.Lines {
		line := JournalLineResponse{
			ID:          l.ID,
			DebitCents:  l.DebitCents,
			CreditCents: l.CreditCents,
			Memo:        l.Memo,
		}
		if l.Edges.Account != nil {
			line.AccountID = l.Edges.Account.ID
			line.AccountCode = l.Edges.Account.Code
			line.AccountName = l.Edges.Account.Name
		}
		resp.Lines = append(resp.Lines, line)
	}
	return resp
}
