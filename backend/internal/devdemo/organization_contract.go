package devdemo

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	devmod "github.com/valerii/hub/backend/internal/dev"
)

const (
	demoOrganizationName                 = "Hub"
	demoOrganizationSlug                 = "hub"
	demoProjectKey                       = "DEMO"
	demoOrganizationMaxAttempts          = 3
	demoOrganizationAdvisoryKey    int64 = 0x48554244454D4F31
	demoOrganizationSlugConstraint       = "organizations_slug_key"
)

type demoUser struct {
	email string
	id    uuid.UUID
}

type organizationResult struct {
	organizationID  uuid.UUID
	created         bool
	usersAttached   int
	projectAttached bool
}

func validateMailCalendarDemoSeedInput(input devmod.MailCalendarDemoSeedInput) ([]demoUser, error) {
	if input.ProjectID == uuid.Nil {
		return nil, errors.New("devdemo organization seed: project is required")
	}
	if len(input.UserIDs) == 0 {
		return nil, errors.New("devdemo organization seed: users are required")
	}

	rawEmails := make([]string, 0, len(input.UserIDs))
	for rawEmail := range input.UserIDs {
		rawEmails = append(rawEmails, rawEmail)
	}
	sort.Strings(rawEmails)

	users := make([]demoUser, 0, len(input.UserIDs))
	seenIDs := make(map[uuid.UUID]struct{}, len(input.UserIDs))
	for _, rawEmail := range rawEmails {
		userID := input.UserIDs[rawEmail]
		if userID == uuid.Nil {
			return nil, errors.New("devdemo organization seed: user ID is required")
		}
		trimmedEmail := strings.TrimSpace(rawEmail)
		address, err := mail.ParseAddress(trimmedEmail)
		if err != nil || trimmedEmail != rawEmail || address.Name != "" || address.Address != rawEmail {
			return nil, errors.New("devdemo organization seed: user email is invalid")
		}
		canonicalEmail := strings.ToLower(address.Address)
		if rawEmail != canonicalEmail {
			return nil, errors.New("devdemo organization seed: user email is not canonical")
		}
		if _, exists := seenIDs[userID]; exists {
			return nil, errors.New("devdemo organization seed: user IDs must be unique")
		}
		seenIDs[userID] = struct{}{}
		users = append(users, demoUser{email: canonicalEmail, id: userID})
	}
	return users, nil
}

func newDemoToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate devdemo token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func isRetryableOrganizationError(err error) bool {
	var postgresError *pq.Error
	if !errors.As(err, &postgresError) || postgresError == nil {
		return false
	}
	switch string(postgresError.Code) {
	case "23505":
		return postgresError.Constraint == demoOrganizationSlugConstraint
	case "40001", "40P01":
		return true
	default:
		return false
	}
}
