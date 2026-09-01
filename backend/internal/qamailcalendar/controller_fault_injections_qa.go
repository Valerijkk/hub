//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/aiprovider"
)

func (c *Controller) applyInboundRequestFixturePublishedProofFault(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	proof inboundRequestFixturePublishedFullProof,
) bool {
	persisted, persistedValid := newInboundRequestFixturePersistedFromPlan(plan)
	if c == nil ||
		ctx == nil ||
		tx == nil ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!persistedValid ||
		!proof.valid(plan, persisted) {
		return false
	}

	switch {
	case c.takeInboundRequestFixturePublishedFault(
		plan.nonce,
		inboundRequestFixturePublishedFaultOriginP3Duplicate,
	):
		_, err := tx.OutboxEvent.Create().
			SetID(inboundRequestFixtureUUID("published-fault-origin-p3-duplicate", plan.nonce)).
			SetEventType(inboundRequestFixturePublishedProjectOutboxMailReceivedSubject).
			SetOriginKind("mail_message").
			SetOriginID(proof.roots.mailMessageID).
			SetPayload([]byte("{}")).
			SetCreatedAt(plan.fixedAt).
			Save(ctx)
		return err == nil
	case c.takeInboundRequestFixturePublishedFault(
		plan.nonce,
		inboundRequestFixturePublishedFaultOriginP4MalformedEnvelope,
	):
		updated, err := tx.OutboxEvent.UpdateOneID(proof.outbox.p4OutboxID).
			SetPayload([]byte("{}")).
			Save(ctx)
		return err == nil &&
			updated != nil &&
			updated.ID == proof.outbox.p4OutboxID
	case c.takeInboundRequestFixturePublishedFault(
		plan.nonce,
		inboundRequestFixturePublishedFaultExternalUserAIProvider,
	):
		provider, err := tx.AIProvider.Create().
			SetID(inboundRequestFixtureUUID("published-fault-external-user-ai-provider", plan.nonce)).
			SetName("QA Published Cleanup Fault " + plan.nonce).
			SetProviderType(aiprovider.ProviderType("ollama")).
			SetBaseURL("https://qa-" + plan.nonce + ".invalid").
			SetModels([]string{"qa-published-cleanup-fault-" + plan.nonce}).
			SetEnabled(true).
			SetPriority(0).
			SetMaxRpm(1).
			SetRuResident(true).
			SetScope(aiprovider.Scope("user")).
			SetOwnerID(proof.roots.ownerUserID).
			SetCreatedAt(plan.fixedAt).
			SetUpdatedAt(plan.fixedAt).
			Save(ctx)
		return err == nil &&
			provider != nil &&
			provider.ID == inboundRequestFixtureUUID("published-fault-external-user-ai-provider", plan.nonce)
	default:
		return true
	}
}

func (c *Controller) applyInboundRequestFixturePublishedMaplessProofFault(
	ctx context.Context,
	tx *ent.Tx,
	plan inboundRequestFixturePlan,
	proof inboundRequestFixturePublishedFullProof,
) (bool, bool) {
	persisted, persistedValid := newInboundRequestFixturePersistedFromPlan(plan)
	if c == nil ||
		ctx == nil ||
		tx == nil ||
		!inboundRequestFixturePlanIsCanonical(plan) ||
		!persistedValid ||
		!proof.valid(plan, persisted) {
		return false, false
	}

	switch {
	case c.takeInboundRequestFixturePublishedFault(
		plan.nonce,
		inboundRequestFixturePublishedFaultCleanupMaplessPartial,
	):
		err := tx.OutboxEvent.DeleteOneID(proof.outbox.p5OutboxID).Exec(ctx)
		return true, err == nil
	case c.takeInboundRequestFixturePublishedFault(
		plan.nonce,
		inboundRequestFixturePublishedFaultCleanupMaplessAmbiguous,
	):
		_, err := tx.OutboxEvent.Create().
			SetID(inboundRequestFixtureUUID("published-fault-mapless-ambiguous", plan.nonce)).
			SetEventType(inboundRequestFixturePublishedProjectOutboxMailReceivedSubject).
			SetOriginKind("mail_message").
			SetOriginID(proof.roots.mailMessageID).
			SetPayload([]byte("{}")).
			SetCreatedAt(plan.fixedAt).
			Save(ctx)
		return true, err == nil
	default:
		return false, true
	}
}
