package mail_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
)

type mailboxOwnerRepositoryFake struct {
	result mail.MailboxOwnerRecord
	err    error
	calls  int
	actors []uuid.UUID
}

func (fake *mailboxOwnerRepositoryFake) ResolveMailboxOwner(_ context.Context, actorID uuid.UUID) (mail.MailboxOwnerRecord, error) {
	fake.calls++
	fake.actors = append(fake.actors, actorID)
	return fake.result, fake.err
}

var _ mail.MailboxOwnerRepository = (*mailboxOwnerRepositoryFake)(nil)

func TestMailboxOwnerContract_ExactFrozenABI(t *testing.T) {
	var (
		_ func(mail.MailboxOwnerRepository) (*mail.MailboxService, error)                                          = mail.NewMailboxService
		_ func(*mail.MailboxService, context.Context, uuid.UUID) (mail.MailboxOwner, error)                        = (*mail.MailboxService).ResolveOwner
		_ func(mail.MailboxOwner) mail.MailboxOwnerState                                                          = mail.MailboxOwner.State
		_ func(mail.MailboxOwner) (uuid.UUID, bool)                                                               = mail.MailboxOwner.AccountID
		_ func(mail.MailboxOwner) uuid.UUID                                                                       = mail.MailboxOwner.OrganizationID
		_ func(mail.MailboxOwner) (string, bool)                                                                  = mail.MailboxOwner.PrimaryAddress
		_ func(mail.MailboxOwner) (mail.MailboxOwnerRecoveryCode, bool)                                           = mail.MailboxOwner.RecoveryCode
		_ func(mail.MailboxOwner) (mail.MailboxProjectionState, bool)                                             = mail.MailboxOwner.ProjectionState
		_ func(*mail.MailboxServiceError) string                                                                  = (*mail.MailboxServiceError).Error
		_ func(*mail.MailboxServiceError) error                                                                   = (*mail.MailboxServiceError).Unwrap
		_ error                                                                                                   = (*mail.MailboxServiceError)(nil)
	)

	if reflect.TypeOf(mail.MailboxOwnerState("")).Kind() != reflect.String ||
		reflect.TypeOf(mail.MailboxProjectionState("")).Kind() != reflect.String ||
		reflect.TypeOf(mail.MailboxOwnerRecoveryCode("")).Kind() != reflect.String ||
		reflect.TypeOf(mail.MailboxServiceErrorKind("")).Kind() != reflect.String {
		t.Fatal("mailbox owner contract enums must retain string underlying types")
	}

	for _, test := range []struct {
		name string
		got  string
		want string
	}{
		{name: "owner not configured", got: string(mail.MailboxOwnerNotConfigured), want: "not_configured"},
		{name: "owner provisioning", got: string(mail.MailboxOwnerProvisioning), want: "provisioning"},
		{name: "owner active", got: string(mail.MailboxOwnerActive), want: "active"},
		{name: "owner suspended", got: string(mail.MailboxOwnerSuspended), want: "suspended"},
		{name: "owner error", got: string(mail.MailboxOwnerError), want: "error"},
		{name: "owner deleting", got: string(mail.MailboxOwnerDeleting), want: "deleting"},
		{name: "projection bootstrapping", got: string(mail.MailboxProjectionBootstrapping), want: "bootstrapping"},
		{name: "projection ready", got: string(mail.MailboxProjectionReady), want: "ready"},
		{name: "projection degraded", got: string(mail.MailboxProjectionDegraded), want: "degraded"},
		{name: "projection rebuilding", got: string(mail.MailboxProjectionRebuilding), want: "rebuilding"},
		{name: "service invalid", got: string(mail.MailboxServiceErrorInvalid), want: "invalid"},
		{name: "service not found", got: string(mail.MailboxServiceErrorNotFound), want: "not_found"},
		{name: "service unavailable", got: string(mail.MailboxServiceErrorUnavailable), want: "unavailable"},
		{name: "service deleting", got: string(mail.MailboxServiceErrorDeleting), want: "deleting"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("contract value = %q, want %q", test.got, test.want)
			}
		})
	}

	recordType := reflect.TypeOf(mail.MailboxOwnerRecord{})
	wantFields := []struct {
		name string
		typeValue reflect.Type
	}{
		{name: "ActorID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "OrganizationID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "AccountID", typeValue: reflect.TypeOf(uuid.UUID{})},
		{name: "EngineAccountID", typeValue: reflect.TypeOf((*string)(nil))},
		{name: "State", typeValue: reflect.TypeOf(mail.MailboxOwnerState(""))},
		{name: "ProjectionState", typeValue: reflect.TypeOf(mail.MailboxProjectionState(""))},
		{name: "PrimaryAddress", typeValue: reflect.TypeOf((*string)(nil))},
		{name: "RecoveryCode", typeValue: reflect.TypeOf((*mail.MailboxOwnerRecoveryCode)(nil))},
	}
	if recordType.NumField() != len(wantFields) {
		t.Fatalf("MailboxOwnerRecord field count = %d, want %d", recordType.NumField(), len(wantFields))
	}
	for index, want := range wantFields {
		field := recordType.Field(index)
		if field.Name != want.name || field.Type != want.typeValue || field.PkgPath != "" || field.Anonymous || field.Tag != "" {
			t.Fatalf("MailboxOwnerRecord field[%d] = {%q %v pkg=%q anonymous=%t tag=%q}, want exported %q %v without tag", index, field.Name, field.Type, field.PkgPath, field.Anonymous, field.Tag, want.name, want.typeValue)
		}
	}

	repositoryType := reflect.TypeOf((*mail.MailboxOwnerRepository)(nil)).Elem()
	if repositoryType.NumMethod() != 1 {
		t.Fatalf("MailboxOwnerRepository method count = %d, want 1", repositoryType.NumMethod())
	}
	repositoryMethod, exists := repositoryType.MethodByName("ResolveMailboxOwner")
	wantRepositoryMethod := reflect.TypeOf((func(context.Context, uuid.UUID) (mail.MailboxOwnerRecord, error))(nil))
	if !exists || repositoryMethod.Type != wantRepositoryMethod {
		t.Fatalf("MailboxOwnerRepository.ResolveMailboxOwner type = %v (exists=%t), want %v", repositoryMethod.Type, exists, wantRepositoryMethod)
	}

	ownerType := reflect.TypeOf(mail.MailboxOwner{})
	wantOwnerMethods := map[string]reflect.Type{
		"State":           reflect.TypeOf((func(mail.MailboxOwner) mail.MailboxOwnerState)(nil)),
		"AccountID":       reflect.TypeOf((func(mail.MailboxOwner) (uuid.UUID, bool))(nil)),
		"OrganizationID":  reflect.TypeOf((func(mail.MailboxOwner) uuid.UUID)(nil)),
		"PrimaryAddress":  reflect.TypeOf((func(mail.MailboxOwner) (string, bool))(nil)),
		"RecoveryCode":    reflect.TypeOf((func(mail.MailboxOwner) (mail.MailboxOwnerRecoveryCode, bool))(nil)),
		"ProjectionState": reflect.TypeOf((func(mail.MailboxOwner) (mail.MailboxProjectionState, bool))(nil)),
	}
	requireExactMailboxMethods(t, ownerType, wantOwnerMethods)

	serviceType := reflect.TypeOf((*mail.MailboxService)(nil))
	requireExactMailboxMethods(t, serviceType, map[string]reflect.Type{
		"ResolveOwner": reflect.TypeOf((func(*mail.MailboxService, context.Context, uuid.UUID) (mail.MailboxOwner, error))(nil)),
	})

	serviceErrorType := reflect.TypeOf((*mail.MailboxServiceError)(nil))
	requireExactMailboxMethods(t, serviceErrorType, map[string]reflect.Type{
		"Error":  reflect.TypeOf((func(*mail.MailboxServiceError) string)(nil)),
		"Unwrap": reflect.TypeOf((func(*mail.MailboxServiceError) error)(nil)),
	})
}

func TestMailboxService_RejectsNilDependenciesAndInvalidResolutionInput(t *testing.T) {
	var nilError *mail.MailboxServiceError
	if got := nilError.Error(); got != "mailbox service error" || nilError.Unwrap() != nil {
		t.Fatalf("nil MailboxServiceError = (%q, %v), want fixed content-free error and nil unwrap", got, nilError.Unwrap())
	}
	unknownError := &mail.MailboxServiceError{Kind: mail.MailboxServiceErrorKind("unknown")}
	if got := unknownError.Error(); got != "mailbox service error" || unknownError.Unwrap() != nil {
		t.Fatalf("unknown MailboxServiceError = (%q, %v), want fixed content-free error and nil unwrap", got, unknownError.Unwrap())
	}

	if service, err := mail.NewMailboxService(nil); err == nil || service != nil || !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("NewMailboxService(nil) = (%v, %v), want nil service and invalid error", service, err)
	}

	var typedNil *mailboxOwnerRepositoryFake
	if service, err := mail.NewMailboxService(typedNil); err == nil || service != nil || !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("NewMailboxService(typed nil) = (%v, %v), want nil service and invalid error", service, err)
	}

	actorID := uuid.New()
	repository := &mailboxOwnerRepositoryFake{result: validActiveMailboxOwnerRecord(actorID)}
	service, err := mail.NewMailboxService(repository)
	if err != nil {
		t.Fatalf("NewMailboxService() unexpected error = %v", err)
	}

	if owner, err := service.ResolveOwner(nil, actorID); owner != (mail.MailboxOwner{}) || !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("ResolveOwner(nil context) = (%#v, %v), want zero owner and invalid error", owner, err)
	}
	if owner, err := service.ResolveOwner(context.Background(), uuid.Nil); owner != (mail.MailboxOwner{}) || !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("ResolveOwner(zero actor) = (%#v, %v), want zero owner and invalid error", owner, err)
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0 for invalid inputs", repository.calls)
	}

	var nilService *mail.MailboxService
	if owner, err := nilService.ResolveOwner(context.Background(), actorID); owner != (mail.MailboxOwner{}) || !errors.Is(err, mail.ErrMailboxServiceInvalid) {
		t.Fatalf("nil service ResolveOwner() = (%#v, %v), want zero owner and invalid error", owner, err)
	}
}

func TestMailboxService_ResolvesOwnerStatesThroughActorOnlyRepositoryCall(t *testing.T) {
	actorID := uuid.New()
	organizationID := uuid.New()
	accountID := uuid.New()
	engineAccountID := "engine-account-01"
	primaryAddress := "owner@example.test"
	recoveryCode := mail.MailboxOwnerRecoveryCode("MAILBOX_UNAVAILABLE")

	tests := []struct {
		name              string
		record            mail.MailboxOwnerRecord
		wantState         mail.MailboxOwnerState
		wantAccount       bool
		wantPrimary       bool
		wantRecovery      bool
		wantProjection    bool
		wantRecoveryCode  mail.MailboxOwnerRecoveryCode
		wantPrimaryValue  string
		wantProjectionVal mail.MailboxProjectionState
	}{
		{
			name: "not configured",
			record: mail.MailboxOwnerRecord{
				ActorID:        actorID,
				OrganizationID: organizationID,
				State:          mail.MailboxOwnerNotConfigured,
			},
			wantState: mail.MailboxOwnerNotConfigured,
		},
		{
			name: "provisioning",
			record: mail.MailboxOwnerRecord{
				ActorID:         actorID,
				OrganizationID:  organizationID,
				AccountID:       accountID,
				State:           mail.MailboxOwnerProvisioning,
				ProjectionState: mail.MailboxProjectionBootstrapping,
				RecoveryCode:    &recoveryCode,
			},
			wantState:         mail.MailboxOwnerProvisioning,
			wantAccount:       true,
			wantRecovery:      true,
			wantProjection:    true,
			wantRecoveryCode:  recoveryCode,
			wantProjectionVal: mail.MailboxProjectionBootstrapping,
		},
		{
			name: "active",
			record: mail.MailboxOwnerRecord{
				ActorID:         actorID,
				OrganizationID:  organizationID,
				AccountID:       accountID,
				EngineAccountID: &engineAccountID,
				State:           mail.MailboxOwnerActive,
				ProjectionState: mail.MailboxProjectionReady,
				PrimaryAddress:  &primaryAddress,
			},
			wantState:         mail.MailboxOwnerActive,
			wantAccount:       true,
			wantPrimary:       true,
			wantPrimaryValue:  primaryAddress,
			wantProjection:    true,
			wantProjectionVal: mail.MailboxProjectionReady,
		},
		{
			name: "suspended",
			record: mail.MailboxOwnerRecord{
				ActorID:         actorID,
				OrganizationID:  organizationID,
				AccountID:       accountID,
				EngineAccountID: &engineAccountID,
				State:           mail.MailboxOwnerSuspended,
				ProjectionState: mail.MailboxProjectionDegraded,
				PrimaryAddress:  &primaryAddress,
			},
			wantState:         mail.MailboxOwnerSuspended,
			wantAccount:       true,
			wantPrimary:       true,
			wantPrimaryValue:  primaryAddress,
			wantProjection:    true,
			wantProjectionVal: mail.MailboxProjectionDegraded,
		},
		{
			name: "error",
			record: mail.MailboxOwnerRecord{
				ActorID:         actorID,
				OrganizationID:  organizationID,
				AccountID:       accountID,
				State:           mail.MailboxOwnerError,
				ProjectionState: mail.MailboxProjectionDegraded,
				RecoveryCode:    &recoveryCode,
			},
			wantState:         mail.MailboxOwnerError,
			wantAccount:       true,
			wantRecovery:      true,
			wantProjection:    true,
			wantRecoveryCode:  recoveryCode,
			wantProjectionVal: mail.MailboxProjectionDegraded,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &mailboxOwnerRepositoryFake{result: test.record}
			service, err := mail.NewMailboxService(repository)
			if err != nil {
				t.Fatalf("NewMailboxService() unexpected error = %v", err)
			}

			owner, err := service.ResolveOwner(context.Background(), actorID)
			if err != nil {
				t.Fatalf("ResolveOwner() unexpected error = %v", err)
			}
			if repository.calls != 1 || len(repository.actors) != 1 || repository.actors[0] != actorID {
				t.Fatalf("repository actor-only calls = %#v (%d), want exactly [%s]", repository.actors, repository.calls, actorID)
			}
			if got := owner.State(); got != test.wantState {
				t.Fatalf("owner State() = %q, want %q", got, test.wantState)
			}
			if got := owner.OrganizationID(); got != organizationID {
				t.Fatalf("owner OrganizationID() = %s, want %s", got, organizationID)
			}
			wantAccountID := accountID
			if !test.wantAccount {
				wantAccountID = uuid.Nil
			}
			if got, ok := owner.AccountID(); got != wantAccountID || ok != test.wantAccount {
				t.Fatalf("owner AccountID() = (%s, %t), want (%s, %t)", got, ok, wantAccountID, test.wantAccount)
			}
			if got, ok := owner.PrimaryAddress(); got != test.wantPrimaryValue || ok != test.wantPrimary {
				t.Fatalf("owner PrimaryAddress() = (%q, %t), want (%q, %t)", got, ok, test.wantPrimaryValue, test.wantPrimary)
			}
			if got, ok := owner.RecoveryCode(); got != test.wantRecoveryCode || ok != test.wantRecovery {
				t.Fatalf("owner RecoveryCode() = (%q, %t), want (%q, %t)", got, ok, test.wantRecoveryCode, test.wantRecovery)
			}
			if got, ok := owner.ProjectionState(); got != test.wantProjectionVal || ok != test.wantProjection {
				t.Fatalf("owner ProjectionState() = (%q, %t), want (%q, %t)", got, ok, test.wantProjectionVal, test.wantProjection)
			}
		})
	}
}

func TestMailboxService_DeletingReturnsTypedOwnerWithoutEngineAccess(t *testing.T) {
	actorID := uuid.New()
	record := validActiveMailboxOwnerRecord(actorID)
	record.State = mail.MailboxOwnerDeleting
	record.EngineAccountID = nil
	repository := &mailboxOwnerRepositoryFake{result: record}
	service, err := mail.NewMailboxService(repository)
	if err != nil {
		t.Fatalf("NewMailboxService() unexpected error = %v", err)
	}

	owner, err := service.ResolveOwner(context.Background(), actorID)
	if err != nil {
		t.Fatalf("ResolveOwner() unexpected error = %v", err)
	}
	if got := owner.State(); got != mail.MailboxOwnerDeleting {
		t.Fatalf("owner State() = %q, want %q", got, mail.MailboxOwnerDeleting)
	}
	if got, ok := owner.AccountID(); got != record.AccountID || !ok {
		t.Fatalf("owner AccountID() = (%s, %t), want (%s, true)", got, ok, record.AccountID)
	}
	if got, ok := owner.PrimaryAddress(); got != *record.PrimaryAddress || !ok {
		t.Fatalf("owner PrimaryAddress() = (%q, %t), want (%q, true)", got, ok, *record.PrimaryAddress)
	}
	if got, ok := owner.ProjectionState(); got != record.ProjectionState || !ok {
		t.Fatalf("owner ProjectionState() = (%q, %t), want (%q, true)", got, ok, record.ProjectionState)
	}
	if got, ok := owner.RecoveryCode(); got != "" || ok {
		t.Fatalf("owner RecoveryCode() = (%q, %t), want (\"\", false)", got, ok)
	}
	if repository.calls != 1 {
		t.Fatalf("repository calls = %d, want 1", repository.calls)
	}
}

func TestMailboxService_NondisclosesForeignAndMalformedRecords(t *testing.T) {
	actorID := uuid.New()
	secret := "internal-owner-secret"
	valid := validActiveMailboxOwnerRecord(actorID)
	foreign := valid
	foreign.ActorID = uuid.New()
	missingOrganization := valid
	missingOrganization.OrganizationID = uuid.Nil
	missingAccount := valid
	missingAccount.AccountID = uuid.Nil
	unknownState := valid
	unknownState.State = mail.MailboxOwnerState(secret)

	for _, test := range []struct {
		name   string
		record mail.MailboxOwnerRecord
	}{
		{name: "foreign actor", record: foreign},
		{name: "missing organization", record: missingOrganization},
		{name: "missing account", record: missingAccount},
		{name: "unknown state", record: unknownState},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &mailboxOwnerRepositoryFake{result: test.record}
			service, err := mail.NewMailboxService(repository)
			if err != nil {
				t.Fatalf("NewMailboxService() unexpected error = %v", err)
			}

			owner, err := service.ResolveOwner(context.Background(), actorID)
			if owner != (mail.MailboxOwner{}) {
				t.Fatalf("ResolveOwner() owner = %#v, want zero owner", owner)
			}
			requireMailboxServiceError(t, err, mail.MailboxServiceErrorNotFound, mail.ErrMailboxServiceNotFound)
			if strings.Contains(err.Error(), secret) {
				t.Fatalf("ResolveOwner() leaked record data in error %q", err)
			}
			if repository.calls != 1 {
				t.Fatalf("repository calls = %d, want 1", repository.calls)
			}
		})
	}
}

func TestMailboxService_ValidatesRecoveryCodesAndOpaqueOwnerFields(t *testing.T) {
	actorID := uuid.New()
	valid := validActiveMailboxOwnerRecord(actorID)

	maxEngine := strings.Repeat("x", mail.MaxEngineOpaqueIDBytes)
	valid.EngineAccountID = &maxEngine
	if owner, err := resolveMailboxOwner(t, actorID, valid); err != nil || owner.State() != mail.MailboxOwnerActive {
		t.Fatalf("ResolveOwner(max opaque engine ID) = (%#v, %v), want active owner", owner, err)
	}

	badEngineLength := valid
	overlongEngine := strings.Repeat("x", mail.MaxEngineOpaqueIDBytes+1)
	badEngineLength.EngineAccountID = &overlongEngine
	requireMailboxOwnerNotFound(t, actorID, badEngineLength)

	badEngineControl := valid
	controlEngine := "engine\naccount"
	badEngineControl.EngineAccountID = &controlEngine
	requireMailboxOwnerNotFound(t, actorID, badEngineControl)

	badAddressControl := valid
	controlAddress := "owner\n@example.test"
	badAddressControl.PrimaryAddress = &controlAddress
	requireMailboxOwnerNotFound(t, actorID, badAddressControl)

	badAddressCanonical := valid
	nonCanonicalAddress := "OWNER@example.test"
	badAddressCanonical.PrimaryAddress = &nonCanonicalAddress
	requireMailboxOwnerNotFound(t, actorID, badAddressCanonical)

	badAddressLength := valid
	overlongAddress := strings.Repeat("a", mail.MaxEngineAddressEmailBytes+1)
	badAddressLength.PrimaryAddress = &overlongAddress
	requireMailboxOwnerNotFound(t, actorID, badAddressLength)

	suspendedWithoutEngine := valid
	suspendedWithoutEngine.State = mail.MailboxOwnerSuspended
	suspendedWithoutEngine.EngineAccountID = nil
	requireMailboxOwnerNotFound(t, actorID, suspendedWithoutEngine)

	badRecovery := valid
	badRecovery.State = mail.MailboxOwnerError
	badRecovery.EngineAccountID = nil
	badRecovery.PrimaryAddress = nil
	unknownRecovery := mail.MailboxOwnerRecoveryCode("MAILBOX_INTERNAL_SECRET")
	badRecovery.RecoveryCode = &unknownRecovery
	requireMailboxOwnerNotFound(t, actorID, badRecovery)

	for _, code := range []mail.MailboxOwnerRecoveryCode{
		"MAILBOX_ENGINE_DISABLED",
		"MAILBOX_DOMAIN_REQUIRED",
		"MAILBOX_PROVISIONING_RETRY",
		"MAILBOX_SUSPENDED",
		"MAILBOX_UNAVAILABLE",
	} {
		t.Run(string(code), func(t *testing.T) {
			record := valid
			record.State = mail.MailboxOwnerError
			record.EngineAccountID = nil
			record.PrimaryAddress = nil
			record.RecoveryCode = &code
			owner, err := resolveMailboxOwner(t, actorID, record)
			if err != nil {
				t.Fatalf("ResolveOwner() unexpected error = %v", err)
			}
			if got, ok := owner.RecoveryCode(); !ok || got != code {
				t.Fatalf("owner RecoveryCode() = (%q, %t), want (%q, true)", got, ok, code)
			}
		})
	}
}

func TestMailboxService_RepositoryErrorsAreUnavailableAndInputPointersCannotMutateOwner(t *testing.T) {
	actorID := uuid.New()
	secret := "database address owner@example.test"
	repository := &mailboxOwnerRepositoryFake{err: errors.New(secret)}
	service, err := mail.NewMailboxService(repository)
	if err != nil {
		t.Fatalf("NewMailboxService() unexpected error = %v", err)
	}
	owner, err := service.ResolveOwner(context.Background(), actorID)
	if owner != (mail.MailboxOwner{}) {
		t.Fatalf("ResolveOwner(repository error) owner = %#v, want zero owner", owner)
	}
	requireMailboxServiceError(t, err, mail.MailboxServiceErrorUnavailable, mail.ErrMailboxServiceUnavailable)
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("ResolveOwner(repository error) leaked %q in %q", secret, err)
	}

	primaryAddress := "owner@example.test"
	recoveryCode := mail.MailboxOwnerRecoveryCode("MAILBOX_PROVISIONING_RETRY")
	record := mail.MailboxOwnerRecord{
		ActorID:         actorID,
		OrganizationID:  uuid.New(),
		AccountID:       uuid.New(),
		State:           mail.MailboxOwnerError,
		ProjectionState: mail.MailboxProjectionDegraded,
		PrimaryAddress:  &primaryAddress,
		RecoveryCode:    &recoveryCode,
	}
	owner, err = resolveMailboxOwner(t, actorID, record)
	if err != nil {
		t.Fatalf("ResolveOwner() unexpected error = %v", err)
	}
	primaryAddress = "changed@example.test"
	recoveryCode = mail.MailboxOwnerRecoveryCode("MAILBOX_ENGINE_DISABLED")
	if got, ok := owner.PrimaryAddress(); !ok || got != "owner@example.test" {
		t.Fatalf("owner PrimaryAddress() after source mutation = (%q, %t), want original value", got, ok)
	}
	if got, ok := owner.RecoveryCode(); !ok || got != mail.MailboxOwnerRecoveryCode("MAILBOX_PROVISIONING_RETRY") {
		t.Fatalf("owner RecoveryCode() after source mutation = (%q, %t), want original value", got, ok)
	}
	if _, exists := reflect.TypeOf(mail.MailboxOwner{}).MethodByName("EngineAccountID"); exists {
		t.Fatal("MailboxOwner must not expose EngineAccountID")
	}
}

func validActiveMailboxOwnerRecord(actorID uuid.UUID) mail.MailboxOwnerRecord {
	engineAccountID := "engine-account-01"
	primaryAddress := "owner@example.test"
	return mail.MailboxOwnerRecord{
		ActorID:         actorID,
		OrganizationID:  uuid.New(),
		AccountID:       uuid.New(),
		EngineAccountID: &engineAccountID,
		State:           mail.MailboxOwnerActive,
		ProjectionState: mail.MailboxProjectionReady,
		PrimaryAddress:  &primaryAddress,
	}
}

func resolveMailboxOwner(t *testing.T, actorID uuid.UUID, record mail.MailboxOwnerRecord) (mail.MailboxOwner, error) {
	t.Helper()
	service, err := mail.NewMailboxService(&mailboxOwnerRepositoryFake{result: record})
	if err != nil {
		t.Fatalf("NewMailboxService() unexpected error = %v", err)
	}
	return service.ResolveOwner(context.Background(), actorID)
}

func requireMailboxOwnerNotFound(t *testing.T, actorID uuid.UUID, record mail.MailboxOwnerRecord) {
	t.Helper()
	owner, err := resolveMailboxOwner(t, actorID, record)
	if owner != (mail.MailboxOwner{}) {
		t.Fatalf("ResolveOwner() owner = %#v, want zero owner", owner)
	}
	requireMailboxServiceError(t, err, mail.MailboxServiceErrorNotFound, mail.ErrMailboxServiceNotFound)
}

func requireExactMailboxMethods(t *testing.T, typeValue reflect.Type, want map[string]reflect.Type) {
	t.Helper()
	if typeValue.NumMethod() != len(want) {
		t.Fatalf("%s method count = %d, want %d", typeValue, typeValue.NumMethod(), len(want))
	}
	for name, wantType := range want {
		method, exists := typeValue.MethodByName(name)
		if !exists || method.Type != wantType {
			t.Fatalf("%s.%s type = %v (exists=%t), want %v", typeValue, name, method.Type, exists, wantType)
		}
	}
}

func requireMailboxServiceError(t *testing.T, err error, wantKind mail.MailboxServiceErrorKind, wantSentinel error) {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil, want mailbox service error")
	}
	var serviceError *mail.MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError.Kind != wantKind {
		t.Fatalf("error = %#v, want MailboxServiceError kind %q", err, wantKind)
	}
	if !errors.Is(err, wantSentinel) {
		t.Fatalf("errors.Is(%v, %v) = false, want true", err, wantSentinel)
	}
}
