package mail_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxAccountProvisioningWorkerNilProviderDeadlineSealsUnavailableRetry(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	claim := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("a", 64))
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	repository := &provisioningWorkerRepositoryFake{
		scan:  oneProvisioningWorkerID,
		claim: claimedProvisioningWorkerResult(claim),
		finalize: func(ctx context.Context, request mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) {
			deadline, ok := ctx.Deadline()
			if ctx.Err() != nil || !ok || !deadline.Equal(now.Add(10*time.Second)) {
				t.Fatalf("detached finalize context = err:%v deadline:%v ok:%t", ctx.Err(), deadline, ok)
			}
			if request.Outcome != mail.MailboxAccountProvisioningFinalizeRetry || request.Code != mail.MailboxAccountProvisioningCodeUnavailable {
				t.Fatalf("sealed finalize request = %#v", request)
			}
			return mail.MailboxAccountProvisioningFinalizeResult{Outcome: request.Outcome, Code: request.Code}, nil
		},
	}
	provisioner := &provisioningWorkerProvisionerFake{ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) {
		t.Fatal("provider must not be called after nil deadline context")
		return mail.EngineAccount{}, nil
	}}
	timer := &provisioningWorkerTimerFake{now: now, withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
		switch call {
		case 1:
			cancelParent()
			return nil, nil
		case 2:
			return context.WithDeadline(ctx, deadline)
		default:
			t.Fatalf("WithDeadline calls = %d", call)
			return nil, nil
		}
	})}
	worker := provisioningWorkerNew(t, repository, provisioner, timer)

	result, err := worker.RunOnce(parent)
	adversarialWorkerResult(t, result, mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}, err, nil)
	if provisioner.calls() != 0 || repository.finalizeCalls() != 1 {
		t.Fatalf("external calls = provider:%d finalizer:%d", provisioner.calls(), repository.finalizeCalls())
	}
	requests := repository.finalizeRequests()
	if len(requests) != 1 || requests[0].Outcome != mail.MailboxAccountProvisioningFinalizeRetry || requests[0].Code != mail.MailboxAccountProvisioningCodeUnavailable {
		t.Fatalf("sealed finalize requests = %#v", requests)
	}
}

func TestMailboxAccountProvisioningWorkerNilFinalizeDeadlineLeavesSuccessUnfinalized(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	claim := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("a", 64))
	repository := &provisioningWorkerRepositoryFake{
		scan:  oneProvisioningWorkerID,
		claim: claimedProvisioningWorkerResult(claim),
		finalize: func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) {
			t.Fatal("finalizer must not be called after nil deadline context")
			return mail.MailboxAccountProvisioningFinalizeResult{}, nil
		},
	}
	provisioner := &provisioningWorkerProvisionerFake{ensure: func(ctx context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) {
		deadline, ok := ctx.Deadline()
		if !ok || !deadline.Equal(now.Add(30*time.Second)) {
			t.Fatalf("provider deadline = %v, %t", deadline, ok)
		}
		return provisioningWorkerSuccess(spec), nil
	}}
	timer := &provisioningWorkerTimerFake{now: now, withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
		switch call {
		case 1:
			return context.WithDeadline(ctx, deadline)
		case 2:
			return nil, nil
		default:
			t.Fatalf("WithDeadline calls = %d", call)
			return nil, nil
		}
	})}
	worker := provisioningWorkerNew(t, repository, provisioner, timer)

	result, err := worker.RunOnce(context.Background())
	adversarialWorkerResult(t, result, mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}, err, mail.ErrMailboxAccountProvisioningWorkerPartialFailure)
	if provisioner.calls() != 1 || repository.finalizeCalls() != 0 || len(repository.finalizeRequests()) != 0 {
		t.Fatalf("external calls = provider:%d finalizer:%d requests:%#v", provisioner.calls(), repository.finalizeCalls(), repository.finalizeRequests())
	}
}
