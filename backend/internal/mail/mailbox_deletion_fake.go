package mail

import (
	"context"
	"sync"
)

// FakeEngineAccountDeletionSeed establishes one deterministic fake account pair.
type FakeEngineAccountDeletionSeed struct {
	ProvisioningKey string
	EngineAccountID string
}

// FakeEngineAccountDeletionPortConfig supplies the fake account pairs.
type FakeEngineAccountDeletionPortConfig struct {
	Accounts []FakeEngineAccountDeletionSeed
}

// FakeMailObjectDeletionSeed establishes one deterministic fake object tuple.
type FakeMailObjectDeletionSeed struct {
	Bucket    string
	ObjectKey string
}

// FakeMailObjectDeletionPortConfig supplies the fake object tuples.
type FakeMailObjectDeletionPortConfig struct {
	Objects []FakeMailObjectDeletionSeed
}

// FakeEngineAccountDeletionPort is a deterministic in-memory account deletion port.
type FakeEngineAccountDeletionPort struct {
	mu                   sync.Mutex
	provisioningAccounts map[string]string
	engineAccounts       map[string]string
}

// FakeMailObjectDeletionPort is a deterministic in-memory object deletion port.
type FakeMailObjectDeletionPort struct {
	mu         sync.Mutex
	objects    map[fakeMailObjectDeletionKey]struct{}
	objectKeys map[string]string
}

type fakeMailObjectDeletionKey struct {
	bucket    string
	objectKey string
}

var _ EngineAccountDeletionPort = (*FakeEngineAccountDeletionPort)(nil)
var _ MailObjectDeletionPort = (*FakeMailObjectDeletionPort)(nil)

// NewFakeEngineAccountDeletionPort constructs a deterministic account deletion fake.
func NewFakeEngineAccountDeletionPort(config FakeEngineAccountDeletionPortConfig) (*FakeEngineAccountDeletionPort, error) {
	fake := &FakeEngineAccountDeletionPort{
		provisioningAccounts: make(map[string]string, len(config.Accounts)),
		engineAccounts:       make(map[string]string, len(config.Accounts)),
	}
	for _, seed := range config.Accounts {
		if !validEngineAccountDeletionIdentity(seed.ProvisioningKey, seed.EngineAccountID) {
			return nil, &EngineError{Kind: EngineErrorRejectedPolicy}
		}
		if _, exists := fake.provisioningAccounts[seed.ProvisioningKey]; exists {
			return nil, &EngineError{Kind: EngineErrorRejectedPolicy}
		}
		if _, exists := fake.engineAccounts[seed.EngineAccountID]; exists {
			return nil, &EngineError{Kind: EngineErrorRejectedPolicy}
		}
		fake.provisioningAccounts[seed.ProvisioningKey] = seed.EngineAccountID
		fake.engineAccounts[seed.EngineAccountID] = seed.ProvisioningKey
	}
	return fake, nil
}

// NewFakeMailObjectDeletionPort constructs a deterministic object deletion fake.
func NewFakeMailObjectDeletionPort(config FakeMailObjectDeletionPortConfig) (*FakeMailObjectDeletionPort, error) {
	fake := &FakeMailObjectDeletionPort{
		objects:    make(map[fakeMailObjectDeletionKey]struct{}, len(config.Objects)),
		objectKeys: make(map[string]string, len(config.Objects)),
	}
	for _, seed := range config.Objects {
		if !validMailObjectDeletionIdentity(seed.Bucket, seed.ObjectKey) {
			return nil, mailObjectDeletionError(MailObjectDeletionErrorRejectedPolicy)
		}
		key := fakeMailObjectDeletionKey{bucket: seed.Bucket, objectKey: seed.ObjectKey}
		if _, exists := fake.objects[key]; exists {
			return nil, mailObjectDeletionError(MailObjectDeletionErrorRejectedPolicy)
		}
		if _, exists := fake.objectKeys[seed.ObjectKey]; exists {
			return nil, mailObjectDeletionError(MailObjectDeletionErrorRejectedPolicy)
		}
		fake.objects[key] = struct{}{}
		fake.objectKeys[seed.ObjectKey] = seed.Bucket
	}
	return fake, nil
}

func (f *FakeEngineAccountDeletionPort) DeleteAccount(_ context.Context, request EngineAccountDeleteRequest) (EngineAccountDeleteResult, error) {
	if err := validateEngineAccountDeleteRequest(request); err != nil {
		return EngineAccountDeleteResult{}, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	engineAccountID, provisioningFound := f.provisioningAccounts[request.ProvisioningKey]
	provisioningKey, engineFound := f.engineAccounts[request.EngineAccountID]
	if provisioningFound && engineFound && engineAccountID == request.EngineAccountID && provisioningKey == request.ProvisioningKey {
		delete(f.provisioningAccounts, request.ProvisioningKey)
		delete(f.engineAccounts, request.EngineAccountID)
		return EngineAccountDeleteResult{State: EngineAccountDeletionStateDeleted}, nil
	}
	if !provisioningFound && !engineFound {
		return EngineAccountDeleteResult{}, &EngineError{Kind: EngineErrorNotFound}
	}
	return EngineAccountDeleteResult{}, &EngineError{Kind: EngineErrorConflict}
}

func (f *FakeEngineAccountDeletionPort) ReconcileAccountAbsence(_ context.Context, request EngineAccountAbsenceRequest) (EngineAccountAbsenceResult, error) {
	if err := validateEngineAccountDeleteRequest(request); err != nil {
		return EngineAccountAbsenceResult{}, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	engineAccountID, provisioningFound := f.provisioningAccounts[request.ProvisioningKey]
	provisioningKey, engineFound := f.engineAccounts[request.EngineAccountID]
	if provisioningFound && engineFound && engineAccountID == request.EngineAccountID && provisioningKey == request.ProvisioningKey {
		return EngineAccountAbsenceResult{State: EngineAccountAbsenceStatePresent}, nil
	}
	if !provisioningFound && !engineFound {
		return EngineAccountAbsenceResult{State: EngineAccountAbsenceStateAuthoritativelyAbsent}, nil
	}
	return EngineAccountAbsenceResult{}, &EngineError{Kind: EngineErrorConflict}
}

func (f *FakeMailObjectDeletionPort) DeleteObject(_ context.Context, request MailObjectDeleteRequest) (MailObjectDeleteResult, error) {
	if err := validateMailObjectDeleteRequest(request); err != nil {
		return MailObjectDeleteResult{}, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	key := fakeMailObjectDeletionKey{bucket: request.Bucket, objectKey: request.ObjectKey}
	if _, exists := f.objects[key]; exists {
		delete(f.objects, key)
		delete(f.objectKeys, request.ObjectKey)
		return MailObjectDeleteResult{State: MailObjectDeletionStateDeleted}, nil
	}
	if _, exists := f.objectKeys[request.ObjectKey]; exists {
		return MailObjectDeleteResult{}, mailObjectDeletionError(MailObjectDeletionErrorConflict)
	}
	return MailObjectDeleteResult{}, mailObjectDeletionError(MailObjectDeletionErrorIndeterminate)
}

func (f *FakeMailObjectDeletionPort) ReconcileObjectAbsence(_ context.Context, request MailObjectAbsenceRequest) (MailObjectAbsenceResult, error) {
	if err := validateMailObjectDeleteRequest(request); err != nil {
		return MailObjectAbsenceResult{}, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	key := fakeMailObjectDeletionKey{bucket: request.Bucket, objectKey: request.ObjectKey}
	if _, exists := f.objects[key]; exists {
		return MailObjectAbsenceResult{State: MailObjectAbsenceStatePresent}, nil
	}
	if _, exists := f.objectKeys[request.ObjectKey]; exists {
		return MailObjectAbsenceResult{}, mailObjectDeletionError(MailObjectDeletionErrorConflict)
	}
	return MailObjectAbsenceResult{State: MailObjectAbsenceStateAuthoritativelyAbsent}, nil
}
