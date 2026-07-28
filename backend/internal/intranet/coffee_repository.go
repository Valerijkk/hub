package intranet

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/coffeeenrollment"
	"github.com/valerii/hub/backend/ent/coffeepair"
	"github.com/valerii/hub/backend/ent/coffeeround"
	"github.com/valerii/hub/backend/ent/user"
)

func (r *EntRepository) GetCoffeeEnrollment(ctx context.Context, userID uuid.UUID) (*ent.CoffeeEnrollment, error) {
	return r.client.CoffeeEnrollment.Query().
		Where(coffeeenrollment.HasUserWith(user.ID(userID))).
		Only(ctx)
}

func (r *EntRepository) UpsertCoffeeEnrollment(ctx context.Context, userID uuid.UUID, active bool) (*ent.CoffeeEnrollment, error) {
	existing, err := r.GetCoffeeEnrollment(ctx, userID)
	if ent.IsNotFound(err) {
		return r.client.CoffeeEnrollment.Create().
			SetUserID(userID).
			SetIsActive(active).
			Save(ctx)
	}
	if err != nil {
		return nil, err
	}
	return r.client.CoffeeEnrollment.UpdateOneID(existing.ID).
		SetIsActive(active).
		Save(ctx)
}

func (r *EntRepository) ListActiveCoffeeEnrollments(ctx context.Context) ([]*ent.User, error) {
	return r.client.User.Query().
		Where(
			user.StatusEQ(user.StatusActive),
			user.HasCoffeeEnrollmentWith(coffeeenrollment.IsActive(true)),
		).
		All(ctx)
}

func (r *EntRepository) CreateCoffeeRound(ctx context.Context, adminID uuid.UUID, pairCount, unpairedCount int) (*ent.CoffeeRound, error) {
	return r.client.CoffeeRound.Create().
		SetTriggeredByID(adminID).
		SetPairCount(pairCount).
		SetUnpairedCount(unpairedCount).
		Save(ctx)
}

func (r *EntRepository) CreateCoffeePair(ctx context.Context, roundID, userA, userB uuid.UUID) (*ent.CoffeePair, error) {
	if userA.String() > userB.String() {
		userA, userB = userB, userA
	}
	return r.client.CoffeePair.Create().
		SetRoundID(roundID).
		SetParticipantAID(userA).
		SetParticipantBID(userB).
		Save(ctx)
}

func (r *EntRepository) GetLatestCoffeeRound(ctx context.Context) (*ent.CoffeeRound, error) {
	return r.client.CoffeeRound.Query().
		Order(ent.Desc(coffeeround.FieldCreatedAt)).
		WithTriggeredBy().
		WithPairs(func(q *ent.CoffeePairQuery) {
			q.WithParticipantA().WithParticipantB()
		}).
		First(ctx)
}

func (r *EntRepository) GetPreviousRoundPairKeys(ctx context.Context) (map[string]struct{}, error) {
	// A7-006: вызывается ДО CreateCoffeeRound — «предыдущий» = latest (First, без Offset).
	prev, err := r.client.CoffeeRound.Query().
		Order(ent.Desc(coffeeround.FieldCreatedAt)).
		WithPairs(func(q *ent.CoffeePairQuery) {
			q.WithParticipantA().WithParticipantB()
		}).
		First(ctx)
	if ent.IsNotFound(err) {
		return map[string]struct{}{}, nil
	}
	if err != nil {
		return nil, err
	}
	keys := make(map[string]struct{}, len(prev.Edges.Pairs))
	for _, p := range prev.Edges.Pairs {
		if p.Edges.ParticipantA != nil && p.Edges.ParticipantB != nil {
			keys[pairKey(p.Edges.ParticipantA.ID, p.Edges.ParticipantB.ID)] = struct{}{}
		}
	}
	return keys, nil
}

func (r *EntRepository) ListCoffeeRounds(ctx context.Context, limit int) ([]*ent.CoffeeRound, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return r.client.CoffeeRound.Query().
		Order(ent.Desc(coffeeround.FieldCreatedAt)).
		Limit(limit).
		WithTriggeredBy().
		WithPairs(func(q *ent.CoffeePairQuery) {
			q.WithParticipantA().WithParticipantB()
		}).
		All(ctx)
}

func (r *EntRepository) FindUserPairInRound(ctx context.Context, roundID, userID uuid.UUID) (*ent.CoffeePair, error) {
	// First: участник тройки имеет 2 рёберные пары — Only() ломал бы GetMyCoffeeMatch.
	return r.client.CoffeePair.Query().
		Where(
			coffeepair.HasRoundWith(coffeeround.ID(roundID)),
			coffeepair.Or(
				coffeepair.HasParticipantAWith(user.ID(userID)),
				coffeepair.HasParticipantBWith(user.ID(userID)),
			),
		).
		WithParticipantA().
		WithParticipantB().
		First(ctx)
}

func pairKey(a, b uuid.UUID) string {
	if a.String() > b.String() {
		a, b = b, a
	}
	return a.String() + "|" + b.String()
}

func toCoffeePartner(u *ent.User) CoffeePartnerResponse {
	if u == nil {
		return CoffeePartnerResponse{}
	}
	name := u.DisplayName
	if name == "" {
		name = u.Name
	}
	if name == "" {
		name = u.Email
	}
	return CoffeePartnerResponse{
		UserID:      u.ID,
		DisplayName: name,
		Email:       u.Email,
		Department:  u.Department,
	}
}

func toCoffeePairResponse(p *ent.CoffeePair) CoffeePairResponse {
	return CoffeePairResponse{
		ID:           p.ID,
		ParticipantA: toCoffeePartner(p.Edges.ParticipantA),
		ParticipantB: toCoffeePartner(p.Edges.ParticipantB),
	}
}

func toCoffeeRoundResponse(rnd *ent.CoffeeRound) CoffeeRoundResponse {
	resp := CoffeeRoundResponse{
		ID:            rnd.ID,
		PairCount:     rnd.PairCount,
		UnpairedCount: rnd.UnpairedCount,
		CreatedAt:     rnd.CreatedAt,
		Pairs:         make([]CoffeePairResponse, 0),
	}
	if rnd.Edges.TriggeredBy != nil {
		resp.TriggeredBy = displayUserName(rnd.Edges.TriggeredBy)
	}
	for _, p := range rnd.Edges.Pairs {
		resp.Pairs = append(resp.Pairs, toCoffeePairResponse(p))
	}
	return resp
}
