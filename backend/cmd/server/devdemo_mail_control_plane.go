//go:build !qa && !release

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	appent "github.com/valerii/hub/backend/ent"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const (
	devMailControlNamespace                    = "hub.dev.mail.control-plane.v1"
	devMailControlDomain                       = "hub.local"
	devMailControlGeneration             int64 = 1
	devMailControlHashVersion             int32 = 1
	devMailControlMaxMessageBytes         int64 = 5 * 1024 * 1024
	devMailControlMaxAttachmentBytes      int64 = 1 * 1024 * 1024
	devMailControlMaxRecipients           int32 = 25
	devMailControlSendRatePerMinute       int32 = 60
	devMailControlInstanceSendRateMinute int32 = 120
	devMailControlTrashRetentionDays      int32 = 30
	devMailControlJunkRetentionDays       int32 = 30
)

var (
	errDevMailControlInvalid     = errors.New("dev mail control-plane input is invalid")
	errDevMailControlConflict    = errors.New("dev mail control-plane state conflicts")
	errDevMailControlUnavailable = errors.New("dev mail control-plane is unavailable")
)

// devMailControlEventCutoff is before every deterministic demo-source and
// dynamic fake Mail message; it is not scheduler time.
func devMailControlEventCutoff() time.Time {
	return time.Date(2026, time.August, 30, 0, 0, 0, 0, time.UTC)
}

type devMailControlFailure string

const (
	devMailControlFailureInvalid     devMailControlFailure = "invalid"
	devMailControlFailureConflict    devMailControlFailure = "conflict"
	devMailControlFailureUnavailable devMailControlFailure = "unavailable"
)

type devMailControlStage string

const (
	devMailControlStageInput    devMailControlStage = "input"
	devMailControlStageProvider devMailControlStage = "provider"
	devMailControlStageUsers    devMailControlStage = "users"
	devMailControlStageDomain   devMailControlStage = "domain"
	devMailControlStagePolicy   devMailControlStage = "policy"
	devMailControlStageAccount  devMailControlStage = "account"
	devMailControlStageIdentity devMailControlStage = "identity"
	devMailControlStageSettings devMailControlStage = "settings"
	devMailControlStageCommit   devMailControlStage = "commit"
	devMailControlStageResult   devMailControlStage = "result"
)

type devMailControlError struct {
	stage   devMailControlStage
	failure devMailControlFailure
}

func (err *devMailControlError) Error() string {
	if err == nil {
		return "dev mail control-plane reconciliation failed"
	}
	return "dev mail control-plane " + string(err.stage) + " reconciliation failed"
}

func (err *devMailControlError) Unwrap() error {
	if err == nil {
		return nil
	}
	switch err.failure {
	case devMailControlFailureInvalid:
		return errDevMailControlInvalid
	case devMailControlFailureConflict:
		return errDevMailControlConflict
	default:
		return errDevMailControlUnavailable
	}
}

func newDevMailControlError(stage devMailControlStage, failure devMailControlFailure) error {
	return &devMailControlError{stage: stage, failure: failure}
}

// devMailControlPlaneAccount is server-private output for the later Fake Mail
// adapter. It deliberately contains only durable account and provider handles.
type devMailControlPlaneAccount struct {
	accountID               uuid.UUID
	engineAccountID         string
	enginePrimaryIdentityID string
}

// devMailControlPlaneResult reports fixed aggregate reconciliation outcomes
// without retaining caller-owned inputs, credentials, or message data.
type devMailControlPlaneResult struct {
	accounts map[string]devMailControlPlaneAccount

	domainsTotal    int
	domainsCreated  int
	domainsReplayed int

	accountsTotal    int
	accountsCreated  int
	accountsReplayed int

	identitiesTotal    int
	identitiesCreated  int
	identitiesReplayed int

	settingsTotal    int
	settingsCreated  int
	settingsReplayed int

	policiesTotal    int
	policiesCreated  int
	policiesReplayed int
}

type devMailControlDomainPlan struct {
	id              uuid.UUID
	organizationID  uuid.UUID
	domain          string
	provisioningKey string
	desiredHash     [sha256.Size]byte
}

type devMailControlAccountPlan struct {
	id              uuid.UUID
	identityID      uuid.UUID
	settingsID      uuid.UUID
	organizationID  uuid.UUID
	userID          uuid.UUID
	email           string
	localPart       string
	provisioningKey string
	controlHash     [sha256.Size]byte
}

type devMailControlPolicyPlan struct {
	id             uuid.UUID
	organizationID uuid.UUID
}

type devMailControlPlan struct {
	now      time.Time
	domain   devMailControlDomainPlan
	policy   devMailControlPolicyPlan
	accounts []devMailControlAccountPlan
}

type devMailProviderAccount struct {
	accountID         string
	primaryIdentityID string
}

// reconcileDevMailControlPlane provisions the Fake provider before one Ent
// transaction. Existing rows are validated, never repaired or overwritten.
func reconcileDevMailControlPlane(
	ctx context.Context,
	client *appent.Client,
	engine *mailmod.FakeMailEngine,
	organizationID uuid.UUID,
	userIDs map[string]uuid.UUID,
	now time.Time,
) (devMailControlPlaneResult, error) {
	if !validDevMailControlContext(ctx) || client == nil || engine == nil || organizationID == uuid.Nil || !validDevMailControlNow(now) {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
	}
	if err := ctx.Err(); err != nil {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureUnavailable)
	}
	plan, err := newDevMailControlPlan(organizationID, userIDs, now)
	if err != nil {
		return devMailControlPlaneResult{}, err
	}

	domainState, err := engine.EnsureDomain(ctx, mailmod.DomainSpec{
		ProvisioningKey:   plan.domain.provisioningKey,
		Domain:            plan.domain.domain,
		DesiredGeneration: devMailControlGeneration,
		DesiredHash:       hex.EncodeToString(plan.domain.desiredHash[:]),
		Present:           true,
		Enabled:           true,
	})
	if err != nil || !validDevMailProviderDomain(domainState) {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageProvider, devMailControlFailureUnavailable)
	}

	providerAccounts := make(map[string]devMailProviderAccount, len(plan.accounts))
	for _, accountPlan := range plan.accounts {
		if err := ctx.Err(); err != nil {
			return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageProvider, devMailControlFailureUnavailable)
		}
		providerAccount, err := engine.EnsureAccount(ctx, mailmod.AccountSpec{
			ProvisioningKey:   accountPlan.provisioningKey,
			LocalPart:         accountPlan.localPart,
			Domain:            plan.domain.domain,
			DesiredGeneration: devMailControlGeneration,
			DesiredHash:       hex.EncodeToString(accountPlan.controlHash[:]),
			Enabled:           true,
			Primary: mailmod.AccountIdentitySpec{
				LocalPart: accountPlan.localPart,
				Domain:    plan.domain.domain,
				Enabled:   true,
			},
			Aliases: []mailmod.AccountIdentitySpec{},
		})
		if err != nil || !validDevMailProviderAccount(providerAccount, accountPlan) {
			return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageProvider, devMailControlFailureUnavailable)
		}
		providerAccounts[accountPlan.email] = devMailProviderAccount{
			accountID:         providerAccount.AccountID,
			primaryIdentityID: providerAccount.PrimaryIdentityID,
		}
	}

	tx, err := client.Tx(ctx)
	if err != nil || tx == nil {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageCommit, devMailControlFailureUnavailable)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	result, err := reconcileDevMailControlPlaneTransaction(ctx, tx, plan, domainState.DomainID, providerAccounts)
	if err != nil {
		return devMailControlPlaneResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return devMailControlPlaneResult{}, devMailControlTransactionFailure(devMailControlStageCommit, err)
	}
	committed = true
	if !validDevMailControlResult(result) {
		return devMailControlPlaneResult{}, newDevMailControlError(devMailControlStageResult, devMailControlFailureUnavailable)
	}
	return cloneDevMailControlResult(result), nil
}

func newDevMailControlPlan(organizationID uuid.UUID, userIDs map[string]uuid.UUID, now time.Time) (devMailControlPlan, error) {
	if organizationID == uuid.Nil || !validDevMailControlNow(now) || userIDs == nil {
		return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
	}
	domainID := devMailControlUUID("domain", organizationID.String(), devMailControlDomain)
	if domainID == uuid.Nil {
		return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
	}
	plan := devMailControlPlan{
		now: now,
		domain: devMailControlDomainPlan{
			id:              domainID,
			organizationID:  organizationID,
			domain:          devMailControlDomain,
			provisioningKey: "hub-devdemo-mail-domain-v1:" + domainID.String(),
			desiredHash:     devMailControlHash("domain", organizationID.String(), devMailControlDomain),
		},
		policy: devMailControlPolicyPlan{
			id:             devMailControlUUID("policy", organizationID.String()),
			organizationID: organizationID,
		},
	}
	if plan.policy.id == uuid.Nil {
		return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
	}
	for _, email := range []string{"pm@hub.local", "dev1@hub.local"} {
		userID, ok := userIDs[email]
		if !ok || userID == uuid.Nil {
			return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
		}
		localPart, ok := strings.CutSuffix(email, "@"+devMailControlDomain)
		if !ok || !validDevMailControlLocalPart(localPart) {
			return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
		}
		accountID := devMailControlUUID("account", organizationID.String(), userID.String(), email)
		identityID := devMailControlUUID("identity", organizationID.String(), userID.String(), email)
		settingsID := devMailControlUUID("settings", organizationID.String(), userID.String(), email)
		if accountID == uuid.Nil || identityID == uuid.Nil || settingsID == uuid.Nil || accountID == identityID || accountID == settingsID || identityID == settingsID {
			return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
		}
		plan.accounts = append(plan.accounts, devMailControlAccountPlan{
			id:              accountID,
			identityID:      identityID,
			settingsID:      settingsID,
			organizationID:  organizationID,
			userID:          userID,
			email:           email,
			localPart:       localPart,
			provisioningKey: "hub-devdemo-mail-account-v1:" + accountID.String(),
			controlHash:     devMailControlHash("account", organizationID.String(), userID.String(), email),
		})
	}
	if len(plan.accounts) != 2 || plan.accounts[0].userID == plan.accounts[1].userID {
		return devMailControlPlan{}, newDevMailControlError(devMailControlStageInput, devMailControlFailureInvalid)
	}
	sort.Slice(plan.accounts, func(left, right int) bool { return plan.accounts[left].email < plan.accounts[right].email })
	return plan, nil
}

func validDevMailProviderDomain(state mailmod.EngineDomainState) bool {
	return validDevMailControlOpaqueID(state.DomainID) &&
		state.State == mailmod.EngineDomainVerified && state.Applied &&
		len(state.DNSRecords) == 0 &&
		state.Verification == mailmod.EngineVerificationVerified &&
		state.DKIM == mailmod.EngineVerificationVerified
}

func validDevMailProviderAccount(account mailmod.EngineAccount, plan devMailControlAccountPlan) bool {
	return validDevMailControlOpaqueID(account.AccountID) &&
		validDevMailControlOpaqueID(account.PrimaryIdentityID) &&
		account.State == mailmod.EngineAccountActive &&
		account.AppliedGeneration == devMailControlGeneration &&
		account.AppliedHash == hex.EncodeToString(plan.controlHash[:]) &&
		len(account.AliasIdentityIDs) == 0
}

func validDevMailProviderAccountHandles(account devMailProviderAccount) bool {
	return validDevMailControlOpaqueID(account.accountID) && validDevMailControlOpaqueID(account.primaryIdentityID)
}

func validDevMailControlPlan(plan devMailControlPlan) bool {
	if !validDevMailControlNow(plan.now) || plan.domain.id == uuid.Nil || plan.domain.organizationID == uuid.Nil || plan.domain.domain != devMailControlDomain || !validDevMailControlOpaqueID(plan.domain.provisioningKey) || plan.policy.id == uuid.Nil || plan.policy.organizationID != plan.domain.organizationID || len(plan.accounts) != 2 {
		return false
	}
	seenEmails := make(map[string]struct{}, len(plan.accounts))
	seenUsers := make(map[uuid.UUID]struct{}, len(plan.accounts))
	for _, account := range plan.accounts {
		expectedLocalPart, exactEmail := strings.CutSuffix(account.email, "@"+devMailControlDomain)
		if account.id == uuid.Nil || account.identityID == uuid.Nil || account.settingsID == uuid.Nil || account.organizationID != plan.domain.organizationID || account.userID == uuid.Nil || !validDevMailControlEmail(account.email) || !exactEmail || account.localPart != expectedLocalPart || !validDevMailControlLocalPart(account.localPart) || !validDevMailControlOpaqueID(account.provisioningKey) {
			return false
		}
		if _, exists := seenEmails[account.email]; exists {
			return false
		}
		if _, exists := seenUsers[account.userID]; exists {
			return false
		}
		seenEmails[account.email] = struct{}{}
		seenUsers[account.userID] = struct{}{}
	}
	_, hasPM := seenEmails["pm@hub.local"]
	_, hasDev1 := seenEmails["dev1@hub.local"]
	return hasPM && hasDev1
}

func validDevMailControlResult(result devMailControlPlaneResult) bool {
	if len(result.accounts) != 2 || result.domainsTotal != 1 || result.domainsCreated < 0 || result.domainsReplayed < 0 || result.domainsCreated+result.domainsReplayed != result.domainsTotal || result.accountsTotal != 2 || result.accountsCreated < 0 || result.accountsReplayed < 0 || result.accountsCreated+result.accountsReplayed != result.accountsTotal || result.identitiesTotal != 2 || result.identitiesCreated < 0 || result.identitiesReplayed < 0 || result.identitiesCreated+result.identitiesReplayed != result.identitiesTotal || result.settingsTotal != 2 || result.settingsCreated < 0 || result.settingsReplayed < 0 || result.settingsCreated+result.settingsReplayed != result.settingsTotal || result.policiesTotal != 1 || result.policiesCreated < 0 || result.policiesReplayed < 0 || result.policiesCreated+result.policiesReplayed != result.policiesTotal {
		return false
	}
	for _, email := range []string{"pm@hub.local", "dev1@hub.local"} {
		account, ok := result.accounts[email]
		if !ok || account.accountID == uuid.Nil || !validDevMailControlOpaqueID(account.engineAccountID) || !validDevMailControlOpaqueID(account.enginePrimaryIdentityID) {
			return false
		}
	}
	return true
}

func cloneDevMailControlResult(result devMailControlPlaneResult) devMailControlPlaneResult {
	cloned := result
	cloned.accounts = make(map[string]devMailControlPlaneAccount, len(result.accounts))
	for email, account := range result.accounts {
		cloned.accounts[email] = account
	}
	return cloned
}

func devMailControlUUID(kind string, fields ...string) uuid.UUID {
	if kind == "" || len(fields) == 0 {
		return uuid.Nil
	}
	parts := append([]string{devMailControlNamespace, kind}, fields...)
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(strings.Join(parts, "\x00")))
}

func devMailControlHash(kind string, fields ...string) [sha256.Size]byte {
	parts := append([]string{devMailControlNamespace, kind}, fields...)
	return sha256.Sum256([]byte(strings.Join(parts, "\x00")))
}

func validDevMailControlContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	value := reflect.ValueOf(ctx)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}

func validDevMailControlNow(now time.Time) bool {
	return !now.IsZero() && now.Location() == time.UTC && now.Nanosecond() == 0
}

func validDevMailControlEmail(email string) bool {
	return email == "pm@hub.local" || email == "dev1@hub.local"
}

func validDevMailControlLocalPart(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func validDevMailControlOpaqueID(value string) bool {
	if value == "" || len(value) > mailmod.MaxEngineOpaqueIDBytes || !utf8.ValidString(value) {
		return false
	}
	if _, err := uuid.Parse(value); err == nil {
		return false
	}
	for _, character := range value {
		if character < 0x21 || character == 0x7f {
			return false
		}
	}
	return true
}
