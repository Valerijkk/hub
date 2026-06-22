package intranet

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

var (
	ErrCoffeeNotEnough = errors.New("not enough participants for matching")
	ErrCoffeeNotEnrolled = errors.New("not enrolled in random coffee")
)

// CoffeeNotifier уведомляет участников о новой паре (V4-05).
type CoffeeNotifier interface {
	NotifyCoffeeMatch(ctx context.Context, userID uuid.UUID, partner CoffeePartnerResponse) error
}

// GetCoffeeEnrollment возвращает статус участия.
func (s *Service) GetCoffeeEnrollment(ctx context.Context, userID uuid.UUID) (*CoffeeEnrollmentResponse, error) {
	en, err := s.repo.GetCoffeeEnrollment(ctx, userID)
	if ent.IsNotFound(err) {
		return &CoffeeEnrollmentResponse{Enrolled: false, IsActive: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get coffee enrollment: %w", err)
	}
	t := en.EnrolledAt.UTC()
	return &CoffeeEnrollmentResponse{
		Enrolled:   true,
		IsActive:   en.IsActive,
		EnrolledAt: &t,
	}, nil
}

// SetCoffeeEnrollment включает/выключает участие.
func (s *Service) SetCoffeeEnrollment(ctx context.Context, userID uuid.UUID, active bool) (*CoffeeEnrollmentResponse, error) {
	if _, err := s.repo.GetActiveUser(ctx, userID); err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("validate user: %w", err)
	}
	if _, err := s.repo.UpsertCoffeeEnrollment(ctx, userID, active); err != nil {
		return nil, fmt.Errorf("upsert enrollment: %w", err)
	}
	return s.GetCoffeeEnrollment(ctx, userID)
}

// GetMyCoffeeMatch возвращает пару из последнего раунда.
func (s *Service) GetMyCoffeeMatch(ctx context.Context, userID uuid.UUID) (*CoffeeMyMatchResponse, error) {
	rnd, err := s.repo.GetLatestCoffeeRound(ctx)
	if ent.IsNotFound(err) {
		return &CoffeeMyMatchResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest round: %w", err)
	}
	en, err := s.repo.GetCoffeeEnrollment(ctx, userID)
	if ent.IsNotFound(err) || (err == nil && !en.IsActive) {
		return &CoffeeMyMatchResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get enrollment: %w", err)
	}

	pair, err := s.repo.FindUserPairInRound(ctx, rnd.ID, userID)
	if ent.IsNotFound(err) {
		t := rnd.CreatedAt.UTC()
		return &CoffeeMyMatchResponse{
			RoundID:  &rnd.ID,
			RoundAt:  &t,
			Unpaired: rnd.UnpairedCount > 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find pair: %w", err)
	}

	var partner *ent.User
	switch {
	case pair.Edges.ParticipantA != nil && pair.Edges.ParticipantA.ID == userID:
		partner = pair.Edges.ParticipantB
	case pair.Edges.ParticipantB != nil && pair.Edges.ParticipantB.ID == userID:
		partner = pair.Edges.ParticipantA
	}
	t := rnd.CreatedAt.UTC()
	p := toCoffeePartner(partner)
	return &CoffeeMyMatchResponse{
		RoundID: &rnd.ID,
		RoundAt: &t,
		Partner: &p,
	}, nil
}

// GetCoffeeStats — число активных участников (admin preview).
func (s *Service) GetCoffeeStats(ctx context.Context, userID uuid.UUID) (*CoffeeStatsResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	users, err := s.repo.ListActiveCoffeeEnrollments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	return &CoffeeStatsResponse{ActiveEnrollments: len(users)}, nil
}

// RunCoffeeMatching запускает раунд Random Coffee (admin).
func (s *Service) RunCoffeeMatching(ctx context.Context, adminID uuid.UUID) (*CoffeeRoundResponse, error) {
	if err := s.requireAdmin(ctx, adminID); err != nil {
		return nil, err
	}
	users, err := s.repo.ListActiveCoffeeEnrollments(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enrollments: %w", err)
	}
	if len(users) < 2 {
		return nil, ErrCoffeeNotEnough
	}

	prevPairs, err := s.repo.GetPreviousRoundPairKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("previous pairs: %w", err)
	}

	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	shuffleUUIDs(ids)

	pairs, unpaired := buildCoffeePairs(ids, prevPairs)
	rnd, err := s.repo.CreateCoffeeRound(ctx, adminID, len(pairs), len(unpaired))
	if err != nil {
		return nil, fmt.Errorf("create round: %w", err)
	}

	userByID := make(map[uuid.UUID]*ent.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}

	createdPairs := make([]CoffeePairResponse, 0, len(pairs))
	for _, p := range pairs {
		cp, err := s.repo.CreateCoffeePair(ctx, rnd.ID, p[0], p[1])
		if err != nil {
			return nil, fmt.Errorf("create pair: %w", err)
		}
		cp.Edges.ParticipantA = userByID[p[0]]
		cp.Edges.ParticipantB = userByID[p[1]]
		createdPairs = append(createdPairs, toCoffeePairResponse(cp))

		if s.coffeeNotifier != nil {
			pa := toCoffeePartner(userByID[p[0]])
			pb := toCoffeePartner(userByID[p[1]])
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, p[0], pb)
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, p[1], pa)
		}
	}

	unpairedResp := make([]CoffeePartnerResponse, 0, len(unpaired))
	for _, uid := range unpaired {
		unpairedResp = append(unpairedResp, toCoffeePartner(userByID[uid]))
	}

	resp := CoffeeRoundResponse{
		ID:            rnd.ID,
		PairCount:     len(pairs),
		UnpairedCount: len(unpaired),
		Pairs:         createdPairs,
		Unpaired:      unpairedResp,
		CreatedAt:     rnd.CreatedAt,
	}
	if u, ok := userByID[adminID]; ok {
		resp.TriggeredBy = displayUserName(u)
	}
	return &resp, nil
}

// ListCoffeeRounds — история раундов (admin).
func (s *Service) ListCoffeeRounds(ctx context.Context, userID uuid.UUID) ([]CoffeeRoundResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return nil, err
	}
	rounds, err := s.repo.ListCoffeeRounds(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("list rounds: %w", err)
	}
	out := make([]CoffeeRoundResponse, 0, len(rounds))
	for _, rnd := range rounds {
		out = append(out, toCoffeeRoundResponse(rnd))
	}
	return out, nil
}

func buildCoffeePairs(ids []uuid.UUID, avoid map[string]struct{}) ([][2]uuid.UUID, []uuid.UUID) {
	if len(ids) < 2 {
		return nil, ids
	}
	remaining := append([]uuid.UUID(nil), ids...)
	var pairs [][2]uuid.UUID

	for len(remaining) >= 2 {
		a := remaining[0]
		bIdx := 1
		if len(remaining) > 2 {
			for i := 1; i < len(remaining); i++ {
				if _, bad := avoid[pairKey(a, remaining[i])]; !bad {
					bIdx = i
					break
				}
			}
		}
		b := remaining[bIdx]
		pairs = append(pairs, [2]uuid.UUID{a, b})
		remaining = append(remaining[:bIdx], remaining[bIdx+1:]...)
		remaining = remaining[1:]
	}

	var unpaired []uuid.UUID
	if len(remaining) == 1 {
		unpaired = []uuid.UUID{remaining[0]}
	}
	return pairs, unpaired
}

func shuffleUUIDs(ids []uuid.UUID) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})
}
