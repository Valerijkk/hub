package mail

import "time"

const (
	mailboxSubmissionRateMaximumCapacity int32 = 1_000
	mailboxSubmissionRateMinuteNanos     int64 = int64(time.Minute)
	mailboxSubmissionRateMaximumVersion  int64 = 1<<63 - 1
)

type mailboxSubmissionRatePreparedBucket struct {
	original MailboxSubmissionRateBucketState
	state    MailboxSubmissionRateBucketState
	missing  bool
}

// MailboxSubmissionRateBucketState is the durable, provider-neutral state of
// one submission-rate bucket.
type MailboxSubmissionRateBucketState struct {
	Tokens               int32
	LastRefillAt         time.Time
	RefillRemainderNanos int64
	Version              int64
}

// MailboxSubmissionRateLimitInput contains the already-resolved capacities and
// durable bucket rows for one all-or-nothing submission authorization.
type MailboxSubmissionRateLimitInput struct {
	Now                  time.Time
	OrganizationCapacity int32
	UserCapacity         int32
	Organization         MailboxSubmissionRateBucketState
	User                 MailboxSubmissionRateBucketState
}

// MailboxSubmissionRateLimitDecision contains the updated durable states for
// one submission authorization decision.
type MailboxSubmissionRateLimitDecision struct {
	Allowed      bool
	RetryAt      *time.Time
	Organization MailboxSubmissionRateBucketState
	User         MailboxSubmissionRateBucketState
}

// MailboxSubmissionRateLimitErrorKind is the closed safe failure family for
// the persistence-state adapter.
type MailboxSubmissionRateLimitErrorKind string

const (
	// MailboxSubmissionRateLimitErrorInvalid rejects malformed state or a value
	// that cannot be represented safely by the durable adapter.
	MailboxSubmissionRateLimitErrorInvalid MailboxSubmissionRateLimitErrorKind = "invalid"
)

// MailboxSubmissionRateLimitError is a content-free typed rate-limit error.
type MailboxSubmissionRateLimitError struct {
	kind MailboxSubmissionRateLimitErrorKind
}

// Error returns the stable, content-free rate-limit adapter error text.
func (errorValue *MailboxSubmissionRateLimitError) Error() string {
	return "mailbox submission rate limit invalid"
}

// Kind returns the closed safe error classification.
func (errorValue *MailboxSubmissionRateLimitError) Kind() MailboxSubmissionRateLimitErrorKind {
	return MailboxSubmissionRateLimitErrorInvalid
}

// EvaluateMailboxSubmissionRateLimit adapts the existing provider-neutral send
// bucket refill semantics to durable dual-scope submission state. It performs
// no I/O, clock reads, provider calls, persistence, or synchronization.
func EvaluateMailboxSubmissionRateLimit(input MailboxSubmissionRateLimitInput) (MailboxSubmissionRateLimitDecision, error) {
	now := input.Now.UTC()
	if input.Now.IsZero() || !mailboxSubmissionRateValidCapacity(input.OrganizationCapacity) || !mailboxSubmissionRateValidCapacity(input.UserCapacity) {
		return MailboxSubmissionRateLimitDecision{}, mailboxSubmissionRateLimitError()
	}

	organization, err := mailboxSubmissionRatePrepareBucket(input.Organization, input.OrganizationCapacity, now)
	if err != nil {
		return MailboxSubmissionRateLimitDecision{}, err
	}
	user, err := mailboxSubmissionRatePrepareBucket(input.User, input.UserCapacity, now)
	if err != nil {
		return MailboxSubmissionRateLimitDecision{}, err
	}

	if organization.state.Tokens < 1 || user.state.Tokens < 1 {
		retryAt, err := mailboxSubmissionRateRetryAt(now, organization.state, input.OrganizationCapacity, user.state, input.UserCapacity)
		if err != nil {
			return MailboxSubmissionRateLimitDecision{}, err
		}
		organizationState, err := mailboxSubmissionRateFinalizeBucket(organization)
		if err != nil {
			return MailboxSubmissionRateLimitDecision{}, err
		}
		userState, err := mailboxSubmissionRateFinalizeBucket(user)
		if err != nil {
			return MailboxSubmissionRateLimitDecision{}, err
		}
		return MailboxSubmissionRateLimitDecision{
			RetryAt:      retryAt,
			Organization: organizationState,
			User:         userState,
		}, nil
	}

	organization.state.Tokens--
	user.state.Tokens--
	organizationState, err := mailboxSubmissionRateFinalizeBucket(organization)
	if err != nil {
		return MailboxSubmissionRateLimitDecision{}, err
	}
	userState, err := mailboxSubmissionRateFinalizeBucket(user)
	if err != nil {
		return MailboxSubmissionRateLimitDecision{}, err
	}
	return MailboxSubmissionRateLimitDecision{
		Allowed:      true,
		Organization: organizationState,
		User:         userState,
	}, nil
}

func mailboxSubmissionRateValidCapacity(value int32) bool {
	return value >= 1 && value <= mailboxSubmissionRateMaximumCapacity
}

func mailboxSubmissionRatePrepareBucket(input MailboxSubmissionRateBucketState, capacity int32, now time.Time) (mailboxSubmissionRatePreparedBucket, error) {
	if mailboxSubmissionRateMissingBucket(input) {
		return mailboxSubmissionRatePreparedBucket{
			original: input,
			state: MailboxSubmissionRateBucketState{
				Tokens:       capacity,
				LastRefillAt: now,
			},
			missing: true,
		}, nil
	}
	if !mailboxSubmissionRateValidExistingBucket(input, now) {
		return mailboxSubmissionRatePreparedBucket{}, mailboxSubmissionRateLimitError()
	}

	prepared := mailboxSubmissionRatePreparedBucket{original: input, state: input}
	prepared.state.LastRefillAt = prepared.state.LastRefillAt.UTC()
	rateState := SendRateState{
		Tokens:          int(prepared.state.Tokens),
		LastRefillAt:    prepared.state.LastRefillAt,
		RefillRemainder: time.Duration(prepared.state.RefillRemainderNanos),
	}
	if prepared.state.Tokens >= capacity {
		rateState = SendRateState{Tokens: int(capacity), LastRefillAt: now}
	} else {
		rateState = refillSendRateState(rateState, int(capacity), now)
	}
	prepared.state.Tokens = int32(rateState.Tokens)
	prepared.state.LastRefillAt = rateState.LastRefillAt.UTC()
	prepared.state.RefillRemainderNanos = int64(rateState.RefillRemainder)
	return prepared, nil
}

func mailboxSubmissionRateMissingBucket(value MailboxSubmissionRateBucketState) bool {
	return value.Version == 0 && value.Tokens == 0 && value.LastRefillAt.IsZero() && value.RefillRemainderNanos == 0
}

func mailboxSubmissionRateValidExistingBucket(value MailboxSubmissionRateBucketState, now time.Time) bool {
	return value.Version >= 1 &&
		value.Tokens >= 0 && value.Tokens <= mailboxSubmissionRateMaximumCapacity &&
		!value.LastRefillAt.IsZero() && value.LastRefillAt.Location() == time.UTC &&
		value.LastRefillAt == value.LastRefillAt.Round(0) && !value.LastRefillAt.After(now) &&
		value.RefillRemainderNanos >= 0 && value.RefillRemainderNanos < mailboxSubmissionRateMinuteNanos
}

func mailboxSubmissionRateEqualPersistedState(left, right MailboxSubmissionRateBucketState) bool {
	return left.Tokens == right.Tokens &&
		left.LastRefillAt == right.LastRefillAt &&
		left.RefillRemainderNanos == right.RefillRemainderNanos
}

func mailboxSubmissionRateFinalizeBucket(value mailboxSubmissionRatePreparedBucket) (MailboxSubmissionRateBucketState, error) {
	if value.missing {
		value.state.Version = 1
		return value.state, nil
	}
	if mailboxSubmissionRateEqualPersistedState(value.original, value.state) {
		return value.state, nil
	}
	if err := mailboxSubmissionRateIncrementVersion(&value.state); err != nil {
		return MailboxSubmissionRateBucketState{}, err
	}
	return value.state, nil
}

func mailboxSubmissionRateIncrementVersion(value *MailboxSubmissionRateBucketState) error {
	if value.Version >= mailboxSubmissionRateMaximumVersion {
		return mailboxSubmissionRateLimitError()
	}
	value.Version++
	return nil
}

func mailboxSubmissionRateRetryAt(now time.Time, organization MailboxSubmissionRateBucketState, organizationCapacity int32, user MailboxSubmissionRateBucketState, userCapacity int32) (*time.Time, error) {
	seconds := sendRetryAfterSeconds(
		SendRateState{Tokens: int(organization.Tokens), RefillRemainder: time.Duration(organization.RefillRemainderNanos)}, int(organizationCapacity),
		SendRateState{Tokens: int(user.Tokens), RefillRemainder: time.Duration(user.RefillRemainderNanos)}, int(userCapacity),
	)
	if seconds < 1 {
		return nil, mailboxSubmissionRateLimitError()
	}
	retryAt := now.Add(time.Duration(seconds) * time.Second).UTC()
	if !retryAt.After(now) {
		return nil, mailboxSubmissionRateLimitError()
	}
	return &retryAt, nil
}

func mailboxSubmissionRateLimitError() error {
	return &MailboxSubmissionRateLimitError{kind: MailboxSubmissionRateLimitErrorInvalid}
}
