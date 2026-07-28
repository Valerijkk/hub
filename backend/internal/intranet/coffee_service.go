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

	pairs, triples := buildCoffeeMatches(ids, prevPairs)
	// pair_count в БД = число бинарных пар + троек (групп матчинга); unpaired всегда 0 при N>=2.
	matchGroups := len(pairs) + len(triples)

	userByID := make(map[uuid.UUID]*ent.User, len(users))
	for _, u := range users {
		userByID[u.ID] = u
	}

	// W7-A2-004: все рёбра (пары + тройки как A-B/A-C) в одной tx с round.
	edges := make([][2]uuid.UUID, 0, len(pairs)+2*len(triples))
	for _, p := range pairs {
		edges = append(edges, [2]uuid.UUID{p[0], p[1]})
	}
	for _, t := range triples {
		edges = append(edges, [2]uuid.UUID{t[0], t[1]}, [2]uuid.UUID{t[0], t[2]})
	}
	rnd, created, err := s.repo.CreateCoffeeRoundWithPairs(ctx, adminID, matchGroups, 0, edges)
	if err != nil {
		return nil, fmt.Errorf("create round: %w", err)
	}

	createdPairs := make([]CoffeePairResponse, 0, len(created))
	edgeIdx := 0
	for _, p := range pairs {
		cp := created[edgeIdx]
		edgeIdx++
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

	createdTriples := make([]CoffeeTripleResponse, 0, len(triples))
	for _, t := range triples {
		// Тройка хранится как две рёберные пары (A-B, A-C) без schema migration.
		for _, edge := range [][2]uuid.UUID{{t[0], t[1]}, {t[0], t[2]}} {
			cp := created[edgeIdx]
			edgeIdx++
			cp.Edges.ParticipantA = userByID[edge[0]]
			cp.Edges.ParticipantB = userByID[edge[1]]
			createdPairs = append(createdPairs, toCoffeePairResponse(cp))
		}
		createdTriples = append(createdTriples, CoffeeTripleResponse{
			ParticipantA: toCoffeePartner(userByID[t[0]]),
			ParticipantB: toCoffeePartner(userByID[t[1]]),
			ParticipantC: toCoffeePartner(userByID[t[2]]),
		})
		if s.coffeeNotifier != nil {
			pa := toCoffeePartner(userByID[t[0]])
			pb := toCoffeePartner(userByID[t[1]])
			pc := toCoffeePartner(userByID[t[2]])
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, t[0], pb)
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, t[0], pc)
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, t[1], pa)
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, t[1], pc)
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, t[2], pa)
			_ = s.coffeeNotifier.NotifyCoffeeMatch(ctx, t[2], pb)
		}
	}

	resp := CoffeeRoundResponse{
		ID:            rnd.ID,
		PairCount:     len(pairs),
		TripleCount:   len(triples),
		UnpairedCount: 0,
		Pairs:         createdPairs,
		Triples:       createdTriples,
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

// buildCoffeeMatches — FR-INTR-10 / P1-BE-43: пары; при нечётном N — ровно одна тройка, без unpaired.
func buildCoffeeMatches(ids []uuid.UUID, avoid map[string]struct{}) (pairs [][2]uuid.UUID, triples [][3]uuid.UUID) {
	if len(ids) < 2 {
		return nil, nil
	}
	remaining := append([]uuid.UUID(nil), ids...)

	// Пока остаётся больше 3 — формируем обычные пары.
	for len(remaining) > 3 {
		a := remaining[0]
		bIdx := 1
		for i := 1; i < len(remaining); i++ {
			if _, bad := avoid[pairKey(a, remaining[i])]; !bad {
				bIdx = i
				break
			}
		}
		b := remaining[bIdx]
		pairs = append(pairs, [2]uuid.UUID{a, b})
		remaining = append(remaining[:bIdx], remaining[bIdx+1:]...)
		remaining = remaining[1:]
	}

	switch len(remaining) {
	case 3:
		triples = append(triples, [3]uuid.UUID{remaining[0], remaining[1], remaining[2]})
	case 2:
		pairs = append(pairs, [2]uuid.UUID{remaining[0], remaining[1]})
	}
	return pairs, triples
}

func shuffleUUIDs(ids []uuid.UUID) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})
}
