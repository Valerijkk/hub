//go:build qa && !release

package qamailcalendar

import (
	"context"

	"github.com/valerii/hub/backend/ent"
	organizationpredicate "github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

const threadSearchFixtureFreshProbeRowCount = 6 + threadSearchFixtureRowCount*3

// threadSearchFixtureProbeManifest reads one immutable thread-search fixture
// through a fresh controller client after a commit outcome is unknown. It
// performs no lifecycle/map work and deliberately never reuses a transaction.
func (c *Controller) threadSearchFixtureProbeManifest(
	ctx context.Context,
	plan threadSearchFixturePlan,
	persisted threadSearchFixturePersisted,
) ThreadSearchFixtureFreshProbeSummary {
	if c == nil ||
		ctx == nil ||
		c.client == nil ||
		!threadSearchFixturePlanMatchesCanonical(plan) ||
		!threadSearchFixturePersistedMatchesPlan(persisted, plan) {
		return ThreadSearchFixtureFreshProbeSummary{Failed: true}
	}

	summary := ThreadSearchFixtureFreshProbeSummary{Probed: true}
	observe := func(exact bool, err error) bool {
		switch {
		case err == nil && exact:
			summary.ExactRows++
			return true
		case err == nil:
			summary.MismatchedRows++
			return true
		case ent.IsNotFound(err):
			summary.AbsentRows++
			return true
		default:
			summary.Failed = true
			return false
		}
	}

	organization, err := c.client.Organization.Get(ctx, persisted.organizationID)
	if !observe(threadSearchFixtureOrganizationMatchesPlan(organization, plan), err) {
		return summary
	}

	owner, err := c.client.User.Get(ctx, persisted.ownerUserID)
	if err != nil {
		if !observe(false, err) {
			return summary
		}
	} else {
		ownerMatches := threadSearchFixtureOwnerMatchesPlan(owner, plan)
		if ownerMatches {
			ownerMatches, err = c.client.User.Query().
				Where(
					user.IDEQ(persisted.ownerUserID),
					user.HasOrganizationWith(organizationpredicate.IDEQ(plan.organizationID)),
				).
				Exist(ctx)
			if err != nil {
				summary.Failed = true
				return summary
			}
		}
		if !observe(ownerMatches, nil) {
			return summary
		}
	}

	domain, err := c.client.MailDomain.Get(ctx, persisted.mailDomainID)
	if !observe(threadSearchFixtureDomainMatchesPlan(domain, plan), err) {
		return summary
	}

	account, err := c.client.MailAccount.Get(ctx, persisted.mailAccountID)
	if !observe(threadSearchFixtureAccountMatchesPlan(account, plan), err) {
		return summary
	}

	identity, err := c.client.MailIdentity.Get(ctx, persisted.mailIdentityID)
	if !observe(threadSearchFixtureMailIdentityMatchesPlan(identity, plan), err) {
		return summary
	}

	folder, err := c.client.MailFolder.Get(ctx, persisted.mailFolderID)
	if !observe(threadSearchFixtureFolderMatchesPlan(folder, plan), err) {
		return summary
	}

	for index, row := range plan.rows {
		if persisted.threadIDs[index] != row.threadID ||
			persisted.messageIDs[index] != row.messageID ||
			persisted.messageFolderIDs[index] != row.messageFolderID {
			summary.Failed = true
			return summary
		}

		thread, err := c.client.MailThread.Get(ctx, persisted.threadIDs[index])
		if !observe(threadSearchFixtureThreadMatchesPlan(thread, plan, row), err) {
			return summary
		}

		message, err := c.client.MailMessage.Get(ctx, persisted.messageIDs[index])
		if !observe(threadSearchFixtureMessageMatchesPlan(message, plan, row), err) {
			return summary
		}

		messageFolder, err := c.client.MailMessageFolder.Get(ctx, persisted.messageFolderIDs[index])
		if !observe(threadSearchFixtureMessageFolderMatchesPlan(messageFolder, plan, row), err) {
			return summary
		}
	}

	if summary.ExactRows+summary.AbsentRows+summary.MismatchedRows != threadSearchFixtureFreshProbeRowCount {
		summary.Failed = true
		return summary
	}
	return summary
}
