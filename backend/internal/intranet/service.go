package intranet

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

var (
	ErrNotFound        = errors.New("news post not found")
	ErrUserNotFound    = errors.New("directory user not found")
	ErrForbidden       = errors.New("admin access required")
	ErrInvalidInput    = errors.New("invalid news post input")
	ErrVersionConflict = errors.New("news post version conflict")
	ErrSelfKudo        = errors.New("cannot send kudos to yourself")
	ErrInvalidKudo     = errors.New("invalid kudos input")
)

const unassignedDepartment = "Без отдела"

// KudosNotifier доставляет уведомление получателю (MOD-03).
type KudosNotifier interface {
	NotifyKudosReceived(ctx context.Context, recipientID uuid.UUID, kudo KudoResponse, senderID uuid.UUID) error
}

// Service — бизнес-логика intranet.
type Service struct {
	repo            Repository
	notifier        KudosNotifier
	coffeeNotifier  CoffeeNotifier
}

// NewService создаёт сервис intranet.
func NewService(repo Repository, notifier KudosNotifier, coffeeNotifier CoffeeNotifier) *Service {
	return &Service{repo: repo, notifier: notifier, coffeeNotifier: coffeeNotifier}
}

// ListNews возвращает опубликованные новости для всех пользователей.
func (s *Service) ListNews(ctx context.Context, page, perPage int) (*NewsListResponse, error) {
	items, total, err := s.repo.ListNewsPosts(ctx, true, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list news: %w", err)
	}
	resp := make([]NewsPostResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toNewsPostResponse(item))
	}
	return &NewsListResponse{
		Items:   resp,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// ListNewsAdmin возвращает все новости (включая черновики) для админа.
func (s *Service) ListNewsAdmin(ctx context.Context, userID uuid.UUID, page, perPage int) (*NewsListResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	items, total, err := s.repo.ListNewsPosts(ctx, false, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list news admin: %w", err)
	}
	resp := make([]NewsPostResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toNewsPostResponse(item))
	}
	return &NewsListResponse{
		Items:   resp,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// GetNews возвращает одну новость (только published для member).
func (s *Service) GetNews(ctx context.Context, userID, id uuid.UUID, adminView bool) (*NewsPostResponse, error) {
	post, err := s.repo.GetNewsPost(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news: %w", err)
	}
	if !post.Published {
		if err := s.requireAdmin(ctx, userID); err != nil {
			return nil, ErrNotFound
		}
	} else if adminView {
		if err := s.requireAdmin(ctx, userID); err != nil {
			return nil, err
		}
	}
	resp := toNewsPostResponse(post)
	return &resp, nil
}

// CreateNews создаёт новость (admin).
func (s *Service) CreateNews(ctx context.Context, userID uuid.UUID, req CreateNewsRequest) (*NewsPostResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	if err := validateNewsInput(req.Title, req.Body); err != nil {
		return nil, err
	}
	post, err := s.repo.CreateNewsPost(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("create news: %w", err)
	}
	post, err = s.repo.GetNewsPost(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("reload news: %w", err)
	}
	resp := toNewsPostResponse(post)
	return &resp, nil
}

// UpdateNews обновляет новость (admin).
func (s *Service) UpdateNews(ctx context.Context, userID, id uuid.UUID, req UpdateNewsRequest) (*NewsPostResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	current, err := s.repo.GetNewsPost(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get news for update: %w", err)
	}
	if req.Version > 0 && current.Version != req.Version {
		return nil, ErrVersionConflict
	}
	if req.Title != nil || req.Body != nil {
		title := current.Title
		body := current.Body
		if req.Title != nil {
			title = *req.Title
		}
		if req.Body != nil {
			body = *req.Body
		}
		if err := validateNewsInput(title, body); err != nil {
			return nil, err
		}
	}
	post, err := s.repo.UpdateNewsPost(ctx, id, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update news: %w", err)
	}
	post, err = s.repo.GetNewsPost(ctx, post.ID)
	if err != nil {
		return nil, fmt.Errorf("reload news: %w", err)
	}
	resp := toNewsPostResponse(post)
	return &resp, nil
}

// DeleteNews — мягкое удаление (admin).
func (s *Service) DeleteNews(ctx context.Context, userID, id uuid.UUID) error {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return err
	}
	if err := s.repo.SoftDeleteNewsPost(ctx, id); err != nil {
		if ent.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete news: %w", err)
	}
	return nil
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

// ListDirectory возвращает активных сотрудников, сгруппированных по отделам.
func (s *Service) ListDirectory(ctx context.Context) (*DirectoryResponse, error) {
	users, err := s.repo.ListActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list directory users: %w", err)
	}

	groups := make(map[string][]DirectoryUserResponse)
	for _, u := range users {
		dept := strings.TrimSpace(u.Department)
		if dept == "" {
			dept = unassignedDepartment
		}
		groups[dept] = append(groups[dept], toDirectoryUserResponse(u))
	}

	deptNames := make([]string, 0, len(groups))
	for name := range groups {
		deptNames = append(deptNames, name)
	}
	sort.Slice(deptNames, func(i, j int) bool {
		if deptNames[i] == unassignedDepartment {
			return false
		}
		if deptNames[j] == unassignedDepartment {
			return true
		}
		return deptNames[i] < deptNames[j]
	})

	departments := make([]DirectoryDepartmentGroup, 0, len(deptNames))
	for _, name := range deptNames {
		departments = append(departments, DirectoryDepartmentGroup{
			Department: name,
			Users:      groups[name],
		})
	}

	return &DirectoryResponse{
		Departments: departments,
		Total:       len(users),
	}, nil
}

// GetDirectoryUser возвращает карточку активного сотрудника.
func (s *Service) GetDirectoryUser(ctx context.Context, id uuid.UUID) (*DirectoryUserResponse, error) {
	u, err := s.repo.GetActiveUser(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get directory user: %w", err)
	}
	resp := toDirectoryUserResponse(u)
	return &resp, nil
}

// ListKudosFeed возвращает публичную ленту благодарностей.
func (s *Service) ListKudosFeed(ctx context.Context, page, perPage int) (*KudoListResponse, error) {
	items, total, err := s.repo.ListPublicKudos(ctx, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list kudos feed: %w", err)
	}
	resp := make([]KudoResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toKudoResponse(item))
	}
	return &KudoListResponse{
		Items:   resp,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// ListMyReceivedKudos возвращает благодарности текущего пользователя.
func (s *Service) ListMyReceivedKudos(ctx context.Context, userID uuid.UUID, page, perPage int) (*KudoListResponse, error) {
	items, total, err := s.repo.ListReceivedKudos(ctx, userID, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("list received kudos: %w", err)
	}
	resp := make([]KudoResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toKudoResponse(item))
	}
	return &KudoListResponse{
		Items:   resp,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}

// CreateKudo отправляет благодарность коллеге.
func (s *Service) CreateKudo(ctx context.Context, senderID uuid.UUID, req CreateKudoRequest) (*KudoResponse, error) {
	if senderID == req.RecipientID {
		return nil, ErrSelfKudo
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" || len(msg) > 500 {
		return nil, ErrInvalidKudo
	}
	if _, err := s.repo.GetActiveUser(ctx, req.RecipientID); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("validate recipient: %w", err)
	}
	kudo, err := s.repo.CreateKudo(ctx, senderID, req.RecipientID, msg, req.IsPublic)
	if err != nil {
		return nil, fmt.Errorf("create kudo: %w", err)
	}
	kudo, err = s.repo.GetKudo(ctx, kudo.ID)
	if err != nil {
		return nil, fmt.Errorf("reload kudo: %w", err)
	}
	resp := toKudoResponse(kudo)
	if s.notifier != nil {
		if err := s.notifier.NotifyKudosReceived(ctx, req.RecipientID, resp, senderID); err != nil {
			return nil, fmt.Errorf("notify kudos: %w", err)
		}
	}
	return &resp, nil
}

func validateNewsInput(title, body string) error {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" {
		return ErrInvalidInput
	}
	if len(title) > 200 {
		return ErrInvalidInput
	}
	return nil
}
