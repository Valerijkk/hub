package mail_test

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestEvaluateMailboxSubmissionRateLimit_DualBucketDecisions(t *testing.T) {
	local := time.FixedZone("submission-rate-test", 3*60*60)
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, local)
	utcNow := now.UTC()
	largeNow := time.Unix(math.MaxInt64/2, 0).UTC()

	for _, testCase := range []struct {
		name          string
		input         mail.MailboxSubmissionRateLimitInput
		wantAllowed   bool
		wantRetryAt   *time.Time
		wantOrg       mail.MailboxSubmissionRateBucketState
		wantUser      mail.MailboxSubmissionRateBucketState
	}{
		{
			name: "both buckets debit exactly once and return UTC state",
			input: rateLimitInput(now, 4, 5,
				rateBucket(2, utcNow, 0, 1),
				rateBucket(3, utcNow, 0, 9)),
			wantAllowed: true,
			wantOrg:     rateBucket(1, utcNow, 0, 2),
			wantUser:    rateBucket(2, utcNow, 0, 10),
		},
		{
			name: "one empty bucket denies without debit or a version change to unchanged state",
			input: rateLimitInput(utcNow, 10, 5,
				rateBucket(0, utcNow, 0, 5),
				rateBucket(3, utcNow, 0, 8)),
			wantAllowed: false,
			wantRetryAt: rateTime(utcNow.Add(6 * time.Second)),
			wantOrg:     rateBucket(0, utcNow, 0, 5),
			wantUser:    rateBucket(3, utcNow, 0, 8),
		},
		{
			name: "both empty buckets use the maximum conservative retry wait",
			input: rateLimitInput(utcNow, 10, 2,
				rateBucket(0, utcNow, 0, 1),
				rateBucket(0, utcNow, 1, 1)),
			wantAllowed: false,
			wantRetryAt: rateTime(utcNow.Add(30 * time.Second)),
			wantOrg:     rateBucket(0, utcNow, 0, 1),
			wantUser:    rateBucket(0, utcNow, 1, 1),
		},
		{
			name: "fractional boundary earns exactly one token before the debit",
			input: rateLimitInput(utcNow, 10, 2,
				rateBucket(0, utcNow.Add(-time.Nanosecond), 59_999_999_990, 1),
				rateBucket(1, utcNow, 0, 1)),
			wantAllowed: true,
			wantOrg:     rateBucket(0, utcNow, 0, 2),
			wantUser:    rateBucket(0, utcNow, 0, 2),
		},
		{
			name: "capacity shrink clamps and clears fractional credit in one final version update",
			input: rateLimitInput(utcNow, 3, 4,
				rateBucket(8, utcNow.Add(-30*time.Second), 10, 3),
				rateBucket(4, utcNow, 9, 2)),
			wantAllowed: true,
			wantOrg:     rateBucket(2, utcNow, 0, 4),
			wantUser:    rateBucket(3, utcNow, 0, 3),
		},
		{
			name: "missing rows persist their final debited state with version one",
			input: rateLimitInput(utcNow, 3, 4,
				mail.MailboxSubmissionRateBucketState{},
				mail.MailboxSubmissionRateBucketState{}),
			wantAllowed: true,
			wantOrg:     rateBucket(2, utcNow, 0, 1),
			wantUser:    rateBucket(3, utcNow, 0, 1),
		},
		{
			name: "very large elapsed interval fills safely without arithmetic overflow",
			input: rateLimitInput(largeNow, 1_000, 2,
				rateBucket(0, time.Unix(0, 0).UTC(), 0, 1),
				rateBucket(1, largeNow, 0, 1)),
			wantAllowed: true,
			wantOrg:     rateBucket(999, largeNow, 0, 2),
			wantUser:    rateBucket(0, largeNow, 0, 2),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			before := testCase.input

			decision, err := mail.EvaluateMailboxSubmissionRateLimit(testCase.input)
			if err != nil {
				t.Fatalf("EvaluateMailboxSubmissionRateLimit() error = %v", err)
			}
			if !reflect.DeepEqual(testCase.input, before) {
				t.Fatal("EvaluateMailboxSubmissionRateLimit() mutated its input")
			}
			if decision.Allowed != testCase.wantAllowed {
				t.Fatalf("Allowed = %t, want %t", decision.Allowed, testCase.wantAllowed)
			}
			if decision.Organization != testCase.wantOrg {
				t.Fatalf("Organization = %#v, want %#v", decision.Organization, testCase.wantOrg)
			}
			if decision.User != testCase.wantUser {
				t.Fatalf("User = %#v, want %#v", decision.User, testCase.wantUser)
			}
			if testCase.wantRetryAt == nil {
				if decision.RetryAt != nil {
					t.Fatalf("RetryAt = %v, want nil", *decision.RetryAt)
				}
				return
			}
			if decision.RetryAt == nil {
				t.Fatal("RetryAt = nil, want non-nil")
			}
			if *decision.RetryAt != *testCase.wantRetryAt {
				t.Fatalf("RetryAt = %v, want %v", *decision.RetryAt, *testCase.wantRetryAt)
			}
			if !decision.RetryAt.After(testCase.input.Now) {
				t.Fatal("RetryAt must be after Now")
			}
			if decision.RetryAt.Location() != time.UTC {
				t.Fatal("RetryAt must be canonical UTC")
			}
		})
	}
}

func TestEvaluateMailboxSubmissionRateLimit_RejectsInvalidState(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	valid := rateLimitInput(now, 2, 2,
		rateBucket(1, now, 0, 1),
		rateBucket(1, now, 0, 1))

	for _, testCase := range []struct {
		name  string
		input mail.MailboxSubmissionRateLimitInput
	}{
		{name: "zero now", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Now = time.Time{}; return value }()},
		{name: "organization capacity below minimum", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.OrganizationCapacity = 0; return value }()},
		{name: "user capacity above maximum", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.UserCapacity = 1_001; return value }()},
		{name: "zero version is not missing with tokens", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization = rateBucket(1, time.Time{}, 0, 0); return value }()},
		{name: "existing zero version", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.Version = 0; return value }()},
		{name: "negative tokens", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.Tokens = -1; return value }()},
		{name: "tokens above durable bound", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.Tokens = 1_001; return value }()},
		{name: "negative fractional remainder", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.RefillRemainderNanos = -1; return value }()},
		{name: "fractional remainder at one minute", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.RefillRemainderNanos = int64(time.Minute); return value }()},
		{name: "existing state with zero refill time", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.LastRefillAt = time.Time{}; return value }()},
		{name: "negative version", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.Version = -1; return value }()},
		{name: "backwards time", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.LastRefillAt = now.Add(time.Nanosecond); return value }()},
		{name: "same instant non UTC refill time", input: func() mail.MailboxSubmissionRateLimitInput { value := valid; value.Organization.LastRefillAt = now.In(time.FixedZone("non-utc", 3*60*60)); return value }()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			decision, err := mail.EvaluateMailboxSubmissionRateLimit(testCase.input)
			if err == nil {
				t.Fatal("EvaluateMailboxSubmissionRateLimit() error = nil, want typed invalid error")
			}
			var typed *mail.MailboxSubmissionRateLimitError
			if !errors.As(err, &typed) {
				t.Fatalf("error type = %T, want *MailboxSubmissionRateLimitError", err)
			}
			if typed.Kind() != mail.MailboxSubmissionRateLimitErrorInvalid {
				t.Fatalf("error kind = %q, want %q", typed.Kind(), mail.MailboxSubmissionRateLimitErrorInvalid)
			}
			if err.Error() != "mailbox submission rate limit invalid" {
				t.Fatalf("error text = %q", err.Error())
			}
			if errors.Unwrap(err) != nil {
				t.Fatal("typed invalid error must not expose a cause")
			}
			if decision != (mail.MailboxSubmissionRateLimitDecision{}) {
				t.Fatalf("decision = %#v, want zero decision", decision)
			}
		})
	}
}

func TestEvaluateMailboxSubmissionRateLimit_RejectsRetryTimeOverflow(t *testing.T) {
	now := time.Unix(math.MaxInt64-30, 0).UTC()
	input := rateLimitInput(now, 1, 1,
		rateBucket(0, now, 0, 1),
		rateBucket(1, now, 0, 1))

	decision, err := mail.EvaluateMailboxSubmissionRateLimit(input)
	if err == nil {
		t.Fatal("EvaluateMailboxSubmissionRateLimit() error = nil, want checked retry-time overflow error")
	}
	if decision != (mail.MailboxSubmissionRateLimitDecision{}) {
		t.Fatalf("decision = %#v, want zero decision", decision)
	}
}

func TestEvaluateMailboxSubmissionRateLimit_RetryAtIsDefensive(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	input := rateLimitInput(now, 10, 5,
		rateBucket(0, now, 0, 1),
		rateBucket(3, now, 0, 1))

	first, err := mail.EvaluateMailboxSubmissionRateLimit(input)
	if err != nil || first.RetryAt == nil {
		t.Fatalf("first decision = %#v, error = %v", first, err)
	}
	wantRetryAt := *first.RetryAt
	*first.RetryAt = first.RetryAt.Add(time.Hour)

	second, err := mail.EvaluateMailboxSubmissionRateLimit(input)
	if err != nil || second.RetryAt == nil {
		t.Fatalf("second decision = %#v, error = %v", second, err)
	}
	if *second.RetryAt != wantRetryAt {
		t.Fatalf("second RetryAt = %v, want %v", *second.RetryAt, wantRetryAt)
	}
}

func TestEvaluateMailboxSubmissionRateLimit_ReevaluatesDecisionAsANewAuthorization(t *testing.T) {
	now := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	input := rateLimitInput(now, 2, 2,
		rateBucket(2, now, 0, 1),
		rateBucket(2, now, 0, 1))

	first, err := mail.EvaluateMailboxSubmissionRateLimit(input)
	if err != nil || !first.Allowed {
		t.Fatalf("first decision = %#v, error = %v", first, err)
	}
	input.Organization = first.Organization
	input.User = first.User

	second, err := mail.EvaluateMailboxSubmissionRateLimit(input)
	if err != nil || !second.Allowed {
		t.Fatalf("second decision = %#v, error = %v", second, err)
	}
	if second.Organization != rateBucket(0, now, 0, 3) || second.User != rateBucket(0, now, 0, 3) {
		t.Fatalf("second decision states = %#v / %#v, want both empty at version 3", second.Organization, second.User)
	}
}

func rateLimitInput(now time.Time, organizationCapacity, userCapacity int32, organization, user mail.MailboxSubmissionRateBucketState) mail.MailboxSubmissionRateLimitInput {
	return mail.MailboxSubmissionRateLimitInput{
		Now:                  now,
		OrganizationCapacity: organizationCapacity,
		UserCapacity:         userCapacity,
		Organization:         organization,
		User:                 user,
	}
}

func rateBucket(tokens int32, lastRefillAt time.Time, remainderNanos, version int64) mail.MailboxSubmissionRateBucketState {
	return mail.MailboxSubmissionRateBucketState{
		Tokens:               tokens,
		LastRefillAt:         lastRefillAt,
		RefillRemainderNanos: remainderNanos,
		Version:              version,
	}
}

func rateTime(value time.Time) *time.Time {
	copyValue := value
	return &copyValue
}
