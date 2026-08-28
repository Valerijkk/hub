package devdemo

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
	devmod "github.com/valerii/hub/backend/internal/dev"
)

type organizationSeedResult = organizationResult

type organizationSeeder struct {
	transactions *transactionRunner
}

type organizationStageError struct {
	message string
	cause   error
}

func (err *organizationStageError) Error() string { return err.message }
func (err *organizationStageError) Unwrap() error { return err.cause }

func newOrganizationSeeder(transactions *transactionRunner) (*organizationSeeder, error) {
	if transactions == nil {
		return nil, errors.New("devdemo organization seed: transaction runner is required")
	}
	return &organizationSeeder{transactions: transactions}, nil
}

func (seeder *organizationSeeder) seed(ctx context.Context, input devmod.MailCalendarDemoSeedInput) (organizationSeedResult, error) {
	if seeder == nil || seeder.transactions == nil {
		return organizationSeedResult{}, errors.New("devdemo organization seed: seeder is required")
	}
	if ctx == nil {
		return organizationSeedResult{}, errors.New("devdemo organization seed: context is required")
	}
	users, err := validateMailCalendarDemoSeedInput(input)
	if err != nil {
		return organizationSeedResult{}, err
	}

	for attempt := 0; attempt < demoOrganizationMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return organizationSeedResult{}, errors.New("devdemo organization seed: context ended")
		}
		result := organizationSeedResult{}
		err := seeder.transactions.run(ctx, demoOrganizationAdvisoryKey, func(client *ent.Client) error {
			var reconcileErr error
			result, reconcileErr = reconcileOrganization(ctx, client, input.ProjectID, users)
			return reconcileErr
		})
		if err == nil {
			return result, nil
		}
		if ctx.Err() != nil {
			return organizationSeedResult{}, errors.New("devdemo organization seed: context ended")
		}
		if !isRetryableOrganizationError(err) {
			return organizationSeedResult{}, errors.New("devdemo organization seed: reconciliation failed")
		}
	}
	return organizationSeedResult{}, errors.New("devdemo organization seed: retries exhausted")
}

func reconcileOrganization(ctx context.Context, client *ent.Client, projectID uuid.UUID, users []demoUser) (organizationSeedResult, error) {
	organizations, err := client.Organization.Query().All(ctx)
	if err != nil {
		return organizationSeedResult{}, newOrganizationStageError("devdemo organization seed: organization query failed", err)
	}

	var result organizationSeedResult
	var organization *ent.Organization
	switch len(organizations) {
	case 0:
		token, err := newDemoToken()
		if err != nil {
			return organizationSeedResult{}, err
		}
		organization, err = client.Organization.Create().
			SetName(demoOrganizationName).
			SetSlug(demoOrganizationSlug).
			SetInviteToken(token).
			Save(ctx)
		if err != nil {
			return organizationSeedResult{}, newOrganizationStageError("devdemo organization seed: organization create failed", err)
		}
		result.created = true
	case 1:
		organization = organizations[0]
		if organization.Name != demoOrganizationName || organization.Slug != demoOrganizationSlug {
			return organizationSeedResult{}, errors.New("devdemo organization seed: organization identity mismatch")
		}
	default:
		return organizationSeedResult{}, errors.New("devdemo organization seed: organization cardinality mismatch")
	}
	result.organizationID = organization.ID

	for _, expected := range users {
		stored, err := client.User.Query().
			Where(user.IDEQ(expected.id)).
			WithOrganization().
			Only(ctx)
		if err != nil {
			return organizationSeedResult{}, newOrganizationStageError("devdemo organization seed: user query failed", err)
		}
		canonicalEmail, ok := canonicalStoredEmail(stored.Email)
		if !ok || canonicalEmail != expected.email {
			return organizationSeedResult{}, errors.New("devdemo organization seed: user identity mismatch")
		}
		if stored.Edges.Organization == nil {
			if _, err := client.User.UpdateOneID(stored.ID).SetOrganization(organization).Save(ctx); err != nil {
				return organizationSeedResult{}, newOrganizationStageError("devdemo organization seed: user attach failed", err)
			}
			result.usersAttached++
			continue
		}
		if stored.Edges.Organization.ID != organization.ID {
			return organizationSeedResult{}, errors.New("devdemo organization seed: user organization mismatch")
		}
	}

	storedProject, err := client.Project.Query().
		Where(project.IDEQ(projectID)).
		WithOrganization().
		Only(ctx)
	if err != nil {
		return organizationSeedResult{}, newOrganizationStageError("devdemo organization seed: project query failed", err)
	}
	if storedProject.Key != demoProjectKey {
		return organizationSeedResult{}, errors.New("devdemo organization seed: project identity mismatch")
	}
	if storedProject.Edges.Organization == nil {
		if _, err := client.Project.UpdateOneID(storedProject.ID).SetOrganization(organization).Save(ctx); err != nil {
			return organizationSeedResult{}, newOrganizationStageError("devdemo organization seed: project attach failed", err)
		}
		result.projectAttached = true
	} else if storedProject.Edges.Organization.ID != organization.ID {
		return organizationSeedResult{}, errors.New("devdemo organization seed: project organization mismatch")
	}

	return result, nil
}

func canonicalStoredEmail(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	address, err := mail.ParseAddress(trimmed)
	if err != nil || address.Name != "" || address.Address != trimmed {
		return "", false
	}
	return strings.ToLower(address.Address), true
}

func newOrganizationStageError(message string, cause error) error {
	return &organizationStageError{message: message, cause: cause}
}
