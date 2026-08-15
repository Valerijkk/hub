package mail_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestEvaluateSendPolicy_NormalizesValidInternalManualSend(t *testing.T) {
	input := validSendPolicyInput()
	input.To = []mail.EngineAddress{{Email: "TO@EXAMPLE.TEST", Name: "To Recipient"}}
	input.CC = []mail.EngineAddress{{Email: "Cc@Example.Test", Name: "Carbon Copy"}}
	input.BCC = []mail.EngineAddress{{Email: "bcc@example.test"}}
	input.PolicyMaxRecipients = 3

	decision, err := mail.EvaluateSendPolicy(input)
	if err != nil {
		t.Fatalf("EvaluateSendPolicy() error = %v", err)
	}
	wantTo := []mail.EngineAddress{{Email: "to@example.test", Name: "To Recipient"}}
	wantCC := []mail.EngineAddress{{Email: "cc@example.test", Name: "Carbon Copy"}}
	wantBCC := []mail.EngineAddress{{Email: "bcc@example.test"}}
	if !reflect.DeepEqual(decision.To, wantTo) || !reflect.DeepEqual(decision.CC, wantCC) || !reflect.DeepEqual(decision.BCC, wantBCC) {
		t.Fatalf("normalized recipients = %#v/%#v/%#v, want %#v/%#v/%#v", decision.To, decision.CC, decision.BCC, wantTo, wantCC, wantBCC)
	}
	if decision.UserRateState.Tokens != 4 || decision.InstanceRateState.Tokens != 4 || decision.RetryAfterSeconds != 0 {
		t.Fatalf("successful rate decision = %#v, want both buckets consumed once and retry 0", decision)
	}
}

func TestEvaluateSendPolicy_ExternalToggle(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		enabled bool
		wantErr error
	}{
		{name: "disabled rejects external recipient", enabled: false, wantErr: mail.ErrSendExternalDisabled},
		{name: "enabled accepts external recipient", enabled: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			input.ExternalMailEnabled = testCase.enabled
			input.To = []mail.EngineAddress{{Email: "person@external.test"}}

			decision, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want errors.Is(%v)", err, testCase.wantErr)
			}
			if testCase.wantErr != nil && !reflect.DeepEqual(decision, mail.SendPolicyDecision{}) {
				t.Fatalf("external rejection decision = %#v, want zero value", decision)
			}
			if testCase.wantErr == nil && decision.To[0].Email != "person@external.test" {
				t.Fatalf("enabled external recipient = %#v, want normalized recipient", decision.To)
			}
		})
	}
}

func TestEvaluateSendPolicy_RequiresActiveSendAuthorizedMatchingIdentity(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.SendPolicyInput)
	}{
		{name: "inactive identity", mutate: func(input *mail.SendPolicyInput) { input.Identity.Active = false }},
		{name: "send disallowed", mutate: func(input *mail.SendPolicyInput) { input.Identity.SendAllowed = false }},
		{name: "from differs", mutate: func(input *mail.SendPolicyInput) { input.From.Email = "other@example.test" }},
		{name: "identity is not canonical", mutate: func(input *mail.SendPolicyInput) { input.Identity.CanonicalAddress = "SENDER@EXAMPLE.TEST" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			testCase.mutate(&input)

			decision, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, mail.ErrSendIdentityNotAllowed) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want identity-not-allowed", err)
			}
			if !reflect.DeepEqual(decision, mail.SendPolicyDecision{}) {
				t.Fatalf("identity rejection decision = %#v, want zero value", decision)
			}
		})
	}
}

func TestEvaluateSendPolicy_MessageBounds(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		messageBytes int64
		wantErr      error
	}{
		{name: "zero is invalid", messageBytes: 0, wantErr: mail.ErrSendInvalid},
		{name: "minimum accepted", messageBytes: 1},
		{name: "policy maximum accepted", messageBytes: 1_048_576},
		{name: "above policy maximum rejected", messageBytes: 1_048_577, wantErr: mail.ErrSendTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			input.PolicyMaxMessageBytes = 1_048_576
			input.MessageBytes = testCase.messageBytes

			_, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want errors.Is(%v)", err, testCase.wantErr)
			}
		})
	}
}

func TestEvaluateSendPolicy_EngineMessageMaximumAndPlusOne(t *testing.T) {
	input := validSendPolicyInput()
	input.PolicyMaxMessageBytes = mail.MaxEngineMessageBytes
	input.MessageBytes = mail.MaxEngineMessageBytes
	if _, err := mail.EvaluateSendPolicy(input); err != nil {
		t.Fatalf("engine message maximum error = %v", err)
	}
	input.MessageBytes++
	if _, err := mail.EvaluateSendPolicy(input); !errors.Is(err, mail.ErrSendTooLarge) {
		t.Fatalf("engine message maximum plus one error = %v, want too-large", err)
	}
}

func TestEvaluateSendPolicy_RecipientBoundsAndCrossListUniqueness(t *testing.T) {
	input := validSendPolicyInput()
	input.PolicyMaxRecipients = 300
	input.To = recipients(100, "to")
	input.CC = recipients(100, "cc")
	input.BCC = recipients(100, "bcc")
	if _, err := mail.EvaluateSendPolicy(input); err != nil {
		t.Fatalf("300 normalized partitioned recipients error = %v", err)
	}

	for _, testCase := range []struct {
		name    string
		mutate  func(*mail.SendPolicyInput)
		wantErr error
	}{
		{
			name: "empty recipient partition is invalid",
			mutate: func(input *mail.SendPolicyInput) {
				input.To, input.CC, input.BCC = nil, nil, nil
			},
			wantErr: mail.ErrSendInvalid,
		},
		{
			name: "one list above 100 is too large",
			mutate: func(input *mail.SendPolicyInput) {
				input.To = recipients(101, "too-many")
			},
			wantErr: mail.ErrSendTooLarge,
		},
		{
			name: "combined policy maximum plus one is too large",
			mutate: func(input *mail.SendPolicyInput) {
				input.PolicyMaxRecipients = 2
				input.To = recipients(2, "to")
				input.CC = recipients(1, "cc")
			},
			wantErr: mail.ErrSendTooLarge,
		},
		{
			name: "oversized normalized address maps to too large",
			mutate: func(input *mail.SendPolicyInput) {
				input.To = []mail.EngineAddress{{Email: strings.Repeat("a", 65) + "@example.test"}}
			},
			wantErr: mail.ErrSendTooLarge,
		},
		{
			name: "cross list canonical duplicate is rejected",
			mutate: func(input *mail.SendPolicyInput) {
				input.To = []mail.EngineAddress{{Email: "DUPLICATE@EXAMPLE.TEST"}}
				input.CC = []mail.EngineAddress{{Email: "duplicate@example.test"}}
			},
			wantErr: mail.ErrSendDuplicateRecipient,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			testCase.mutate(&input)
			_, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want errors.Is(%v)", err, testCase.wantErr)
			}
		})
	}
}

func TestEvaluateSendPolicy_RejectsInvalidConfigurationTimeDomainHeaderAndBucketShapes(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.SendPolicyInput)
	}{
		{name: "zero now", mutate: func(input *mail.SendPolicyInput) { input.Now = time.Time{} }},
		{name: "message policy below hard floor", mutate: func(input *mail.SendPolicyInput) { input.PolicyMaxMessageBytes = 1_048_575 }},
		{name: "message policy above engine ceiling", mutate: func(input *mail.SendPolicyInput) { input.PolicyMaxMessageBytes = mail.MaxEngineMessageBytes + 1 }},
		{name: "recipient policy zero", mutate: func(input *mail.SendPolicyInput) { input.PolicyMaxRecipients = 0 }},
		{name: "recipient policy above hard ceiling", mutate: func(input *mail.SendPolicyInput) { input.PolicyMaxRecipients = 301 }},
		{name: "no internal domain", mutate: func(input *mail.SendPolicyInput) { input.InternalDomains = nil }},
		{name: "too many internal domains", mutate: func(input *mail.SendPolicyInput) { input.InternalDomains = domains(101) }},
		{name: "noncanonical internal domain", mutate: func(input *mail.SendPolicyInput) { input.InternalDomains = []string{"Example.Test"} }},
		{name: "duplicate internal domain", mutate: func(input *mail.SendPolicyInput) { input.InternalDomains = []string{"example.test", "example.test"} }},
		{name: "nonascii internal domain", mutate: func(input *mail.SendPolicyInput) { input.InternalDomains = []string{"b\u00fccher.test"} }},
		{name: "zero user rate", mutate: func(input *mail.SendPolicyInput) { input.UserRatePerMinute = 0 }},
		{name: "instance rate above maximum", mutate: func(input *mail.SendPolicyInput) { input.InstanceRatePerMinute = 1001 }},
		{name: "user state future", mutate: func(input *mail.SendPolicyInput) { input.UserRateState.LastRefillAt = input.Now.Add(time.Nanosecond) }},
		{name: "instance state zero time", mutate: func(input *mail.SendPolicyInput) { input.InstanceRateState.LastRefillAt = time.Time{} }},
		{name: "negative tokens", mutate: func(input *mail.SendPolicyInput) { input.UserRateState.Tokens = -1 }},
		{name: "tokens over capacity", mutate: func(input *mail.SendPolicyInput) { input.InstanceRateState.Tokens = input.InstanceRatePerMinute + 1 }},
		{name: "negative fractional remainder", mutate: func(input *mail.SendPolicyInput) { input.UserRateState.RefillRemainder = -time.Nanosecond }},
		{name: "fractional remainder reaches a minute", mutate: func(input *mail.SendPolicyInput) { input.UserRateState.RefillRemainder = time.Minute }},
		{name: "full bucket carries a remainder", mutate: func(input *mail.SendPolicyInput) { input.UserRateState.RefillRemainder = time.Nanosecond }},
		{name: "automated header has CRLF", mutate: func(input *mail.SendPolicyInput) { input.Automated = true; input.AutoResponse.AutoSubmitted = "no\r\n" }},
		{name: "automated header is oversized", mutate: func(input *mail.SendPolicyInput) { input.Automated = true; input.AutoResponse.Precedence = strings.Repeat("x", mail.MaxEngineHeaderValueBytes+1) }},
		{name: "too many x loop values", mutate: func(input *mail.SendPolicyInput) { input.Automated = true; input.AutoResponse.XLoop = make([]string, 17) }},
		{name: "loop token has control", mutate: func(input *mail.SendPolicyInput) { input.Automated = true; input.LoopToken = "loop\u0000token" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			testCase.mutate(&input)
			decision, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, mail.ErrSendInvalid) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want invalid", err)
			}
			if !reflect.DeepEqual(decision, mail.SendPolicyDecision{}) {
				t.Fatalf("invalid input decision = %#v, want zero value", decision)
			}
		})
	}
}

func TestEvaluateSendPolicy_RejectsMalformedAutoResponseTokensBeforeRateMutation(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.SendPolicyInput)
	}{
		{name: "precedence comma", mutate: func(input *mail.SendPolicyInput) { input.AutoResponse.Precedence = "bulk,other" }},
		{name: "precedence whitespace only", mutate: func(input *mail.SendPolicyInput) { input.AutoResponse.Precedence = " \t" }},
		{name: "loop token comma", mutate: func(input *mail.SendPolicyInput) { input.LoopToken = "loop-token,other" }},
		{name: "loop token whitespace only", mutate: func(input *mail.SendPolicyInput) { input.LoopToken = "\t " }},
		{name: "x loop whitespace only", mutate: func(input *mail.SendPolicyInput) { input.AutoResponse.XLoop = []string{"\t "} }},
		{name: "x loop internal whitespace", mutate: func(input *mail.SendPolicyInput) { input.AutoResponse.XLoop = []string{"loop token"} }},
		{name: "x auto suppress empty segment", mutate: func(input *mail.SendPolicyInput) { input.AutoResponse.XAutoResponseSuppress = "all,,other" }},
		{name: "x auto suppress semicolon", mutate: func(input *mail.SendPolicyInput) { input.AutoResponse.XAutoResponseSuppress = "all;other" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			input.Automated = true
			testCase.mutate(&input)
			originalUserState := input.UserRateState
			originalInstanceState := input.InstanceRateState

			decision, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, mail.ErrSendInvalid) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want invalid", err)
			}
			if !reflect.DeepEqual(decision, mail.SendPolicyDecision{}) {
				t.Fatalf("malformed token decision = %#v, want zero value", decision)
			}
			if input.UserRateState != originalUserState || input.InstanceRateState != originalInstanceState {
				t.Fatalf("malformed metadata mutated input rate states: %#v/%#v", input.UserRateState, input.InstanceRateState)
			}
		})
	}
}

func TestEvaluateSendPolicy_AcceptsExactMetadataAndDomainBounds(t *testing.T) {
	t.Run("one hundred canonical internal domains", func(t *testing.T) {
		input := validSendPolicyInput()
		input.InternalDomains = domains(100)
		input.To = []mail.EngineAddress{{Email: "recipient@domain-0.example.test"}}
		input.ExternalMailEnabled = false
		if _, err := mail.EvaluateSendPolicy(input); err != nil {
			t.Fatalf("100 canonical internal domains error = %v", err)
		}
	})

	t.Run("sixteen valid nonmatching x loop values", func(t *testing.T) {
		input := validSendPolicyInput()
		input.Automated = true
		input.LoopToken = "loop-token"
		input.AutoResponse.XLoop = loopValues(16)
		if _, err := mail.EvaluateSendPolicy(input); err != nil {
			t.Fatalf("16 nonmatching x loop values error = %v", err)
		}
	})

	t.Run("maximum safe precedence header", func(t *testing.T) {
		input := validSendPolicyInput()
		input.Automated = true
		input.AutoResponse.Precedence = strings.Repeat("a", mail.MaxEngineHeaderValueBytes)
		if _, err := mail.EvaluateSendPolicy(input); err != nil {
			t.Fatalf("998 byte precedence header error = %v", err)
		}
	})
}

func TestEvaluateSendPolicy_ManualIgnoresAutoResponseMetadata(t *testing.T) {
	input := validSendPolicyInput()
	input.Automated = false
	input.AutoResponse = mail.SendAutoResponseMetadata{
		NullReversePath:      true,
		AutoSubmitted:        "auto-replied\r\n",
		Precedence:           strings.Repeat("x", mail.MaxEngineHeaderValueBytes+1),
		XLoop:                make([]string, 17),
		XAutoResponseSuppress: "all",
	}
	input.LoopToken = "loop\u0000token"

	decision, err := mail.EvaluateSendPolicy(input)
	if err != nil {
		t.Fatalf("manual send with ignored auto-response metadata error = %v", err)
	}
	if decision.UserRateState.Tokens != 4 || decision.InstanceRateState.Tokens != 4 {
		t.Fatalf("manual send did not consume both buckets: %#v", decision)
	}
}

func TestEvaluateSendPolicy_RefillsBucketsWithExactFractionalCarry(t *testing.T) {
	input := validSendPolicyInput()
	input.UserRatePerMinute = 3
	input.InstanceRatePerMinute = 3
	input.UserRateState = mail.SendRateState{Tokens: 1, LastRefillAt: input.Now.Add(-10 * time.Second), RefillRemainder: 30 * time.Second}
	input.InstanceRateState = mail.SendRateState{Tokens: 1, LastRefillAt: input.Now.Add(-10 * time.Second), RefillRemainder: 30 * time.Second}

	decision, err := mail.EvaluateSendPolicy(input)
	if err != nil {
		t.Fatalf("fractional carry evaluation error = %v", err)
	}
	if decision.UserRateState.Tokens != 1 || decision.InstanceRateState.Tokens != 1 || decision.UserRateState.RefillRemainder != 0 || decision.InstanceRateState.RefillRemainder != 0 {
		t.Fatalf("fractional carry decision = %#v, want one refilled then consumed token and no remainder", decision)
	}
	if !decision.UserRateState.LastRefillAt.Equal(input.Now) || !decision.InstanceRateState.LastRefillAt.Equal(input.Now) {
		t.Fatalf("refilled buckets must carry the injected now: %#v", decision)
	}
}

func TestEvaluateSendPolicy_ElapsedMinuteRefillsToCapacityThenConsumes(t *testing.T) {
	input := validSendPolicyInput()
	input.UserRatePerMinute = 2
	input.InstanceRatePerMinute = 2
	input.UserRateState = mail.SendRateState{Tokens: 0, LastRefillAt: input.Now.Add(-time.Minute)}
	input.InstanceRateState = mail.SendRateState{Tokens: 0, LastRefillAt: input.Now.Add(-time.Minute)}

	decision, err := mail.EvaluateSendPolicy(input)
	if err != nil {
		t.Fatalf("full refill evaluation error = %v", err)
	}
	if decision.UserRateState.Tokens != 1 || decision.InstanceRateState.Tokens != 1 || decision.UserRateState.RefillRemainder != 0 || decision.InstanceRateState.RefillRemainder != 0 {
		t.Fatalf("full refill decision = %#v, want capacity two then one consumption", decision)
	}
}

func TestEvaluateSendPolicy_ConsumesBothBucketsAtomically(t *testing.T) {
	input := validSendPolicyInput()
	input.UserRateState.Tokens = 1
	input.InstanceRateState.Tokens = 1
	decision, err := mail.EvaluateSendPolicy(input)
	if err != nil {
		t.Fatalf("dual available buckets error = %v", err)
	}
	if decision.UserRateState.Tokens != 0 || decision.InstanceRateState.Tokens != 0 {
		t.Fatalf("dual consumption = %#v, want both tokens consumed", decision)
	}

	input = validSendPolicyInput()
	input.UserRateState.Tokens = 0
	input.InstanceRateState.Tokens = 5
	decision, err = mail.EvaluateSendPolicy(input)
	if !errors.Is(err, mail.ErrSendRateLimited) {
		t.Fatalf("one empty bucket error = %v, want rate-limited", err)
	}
	if decision.UserRateState.Tokens != 0 || decision.InstanceRateState.Tokens != 5 {
		t.Fatalf("rate denial must not consume either bucket: %#v", decision)
	}
}

func TestEvaluateSendPolicy_RateRetryUsesCeilingAndHardBounds(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		userRate  int
		userState mail.SendRateState
		wantRetry int
	}{
		{name: "fractional wait rounds up", userRate: 3, userState: mail.SendRateState{Tokens: 0, RefillRemainder: time.Nanosecond}, wantRetry: 20},
		{name: "fastest legal bucket still returns one second", userRate: 1000, userState: mail.SendRateState{Tokens: 0}, wantRetry: 1},
		{name: "slowest legal bucket returns sixty seconds", userRate: 1, userState: mail.SendRateState{Tokens: 0}, wantRetry: 60},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			input.UserRatePerMinute = testCase.userRate
			input.UserRateState = testCase.userState
			input.UserRateState.LastRefillAt = input.Now
			input.InstanceRatePerMinute = testCase.userRate
			input.InstanceRateState = testCase.userState
			input.InstanceRateState.LastRefillAt = input.Now

			decision, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, mail.ErrSendRateLimited) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want rate-limited", err)
			}
			if decision.RetryAfterSeconds != testCase.wantRetry {
				t.Fatalf("RetryAfterSeconds = %d, want %d", decision.RetryAfterSeconds, testCase.wantRetry)
			}
		})
	}
}

func TestEvaluateSendPolicy_RateRetryUsesTheSlowestEmptyBucket(t *testing.T) {
	input := validSendPolicyInput()
	input.UserRatePerMinute = 3
	input.UserRateState = mail.SendRateState{Tokens: 0, LastRefillAt: input.Now, RefillRemainder: 30 * time.Second}
	input.InstanceRatePerMinute = 2
	input.InstanceRateState = mail.SendRateState{Tokens: 0, LastRefillAt: input.Now}

	decision, err := mail.EvaluateSendPolicy(input)
	if !errors.Is(err, mail.ErrSendRateLimited) {
		t.Fatalf("EvaluateSendPolicy() error = %v, want rate-limited", err)
	}
	if decision.RetryAfterSeconds != 30 {
		t.Fatalf("RetryAfterSeconds = %d, want slowest exact bucket retry 30", decision.RetryAfterSeconds)
	}
}

func TestEvaluateSendPolicy_AutomatedSuppressionGrammar(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		metadata       mail.SendAutoResponseMetadata
		loopToken      string
		wantSuppressed bool
	}{
		{name: "null reverse path", metadata: mail.SendAutoResponseMetadata{NullReversePath: true}, wantSuppressed: true},
		{name: "auto submitted automatic", metadata: mail.SendAutoResponseMetadata{AutoSubmitted: "auto-replied"}, wantSuppressed: true},
		{name: "auto submitted no accepts htab ows", metadata: mail.SendAutoResponseMetadata{AutoSubmitted: "\tNo\t"}},
		{name: "precedence bulk accepts htab ows", metadata: mail.SendAutoResponseMetadata{Precedence: "\tBuLk\t"}, wantSuppressed: true},
		{name: "precedence list", metadata: mail.SendAutoResponseMetadata{Precedence: "LIST"}, wantSuppressed: true},
		{name: "precedence junk", metadata: mail.SendAutoResponseMetadata{Precedence: "junk"}, wantSuppressed: true},
		{name: "other precedence is allowed", metadata: mail.SendAutoResponseMetadata{Precedence: "first-class"}},
		{name: "x loop matches token", metadata: mail.SendAutoResponseMetadata{XLoop: []string{" unrelated ", " LOOP-Token\t"}}, loopToken: "loop-token", wantSuppressed: true},
		{name: "x loop misses token", metadata: mail.SendAutoResponseMetadata{XLoop: []string{"different"}}, loopToken: "loop-token"},
		{name: "x auto suppress all accepts htab ows", metadata: mail.SendAutoResponseMetadata{XAutoResponseSuppress: "\tnone\t, \tALL\t"}, wantSuppressed: true},
		{name: "x auto suppress autoreply", metadata: mail.SendAutoResponseMetadata{XAutoResponseSuppress: "autoreply"}, wantSuppressed: true},
		{name: "x auto suppress autorespond", metadata: mail.SendAutoResponseMetadata{XAutoResponseSuppress: "x, autorespond, y"}, wantSuppressed: true},
		{name: "x auto suppress other token", metadata: mail.SendAutoResponseMetadata{XAutoResponseSuppress: "none, delivery"}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			input.Automated = true
			input.AutoResponse = testCase.metadata
			input.LoopToken = testCase.loopToken

			decision, err := mail.EvaluateSendPolicy(input)
			if testCase.wantSuppressed {
				if !errors.Is(err, mail.ErrSendAutoResponseSuppressed) {
					t.Fatalf("EvaluateSendPolicy() error = %v, want auto-response-suppressed", err)
				}
				if !reflect.DeepEqual(decision, mail.SendPolicyDecision{}) {
					t.Fatalf("suppressed decision = %#v, want zero value", decision)
				}
				return
			}
			if err != nil {
				t.Fatalf("allowed automated response error = %v", err)
			}
			if decision.UserRateState.Tokens != 4 || decision.InstanceRateState.Tokens != 4 {
				t.Fatalf("allowed automated response must consume both buckets: %#v", decision)
			}
		})
	}
}

func TestEvaluateSendPolicy_DecisionDoesNotAliasOrMutateInputs(t *testing.T) {
	input := validSendPolicyInput()
	input.To = []mail.EngineAddress{{Email: "TO@EXAMPLE.TEST", Name: "Original"}}
	input.CC = []mail.EngineAddress{{Email: "cc@example.test"}}
	originalTo := append([]mail.EngineAddress(nil), input.To...)
	originalCC := append([]mail.EngineAddress(nil), input.CC...)

	decision, err := mail.EvaluateSendPolicy(input)
	if err != nil {
		t.Fatalf("EvaluateSendPolicy() error = %v", err)
	}
	if !reflect.DeepEqual(input.To, originalTo) || !reflect.DeepEqual(input.CC, originalCC) {
		t.Fatalf("input recipients mutated: %#v/%#v", input.To, input.CC)
	}
	input.To[0].Email = "changed@example.test"
	input.CC[0].Email = "changed-cc@example.test"
	if decision.To[0].Email != "to@example.test" || decision.CC[0].Email != "cc@example.test" {
		t.Fatalf("decision aliases input recipients: %#v/%#v", decision.To, decision.CC)
	}
	decision.To[0].Email = "decision-change@example.test"
	if input.To[0].Email != "changed@example.test" {
		t.Fatalf("input aliases decision recipient: %#v", input.To)
	}
}

func TestSendPolicyError_KindsAreStableAndContentFree(t *testing.T) {
	for _, testCase := range []struct {
		kind     mail.SendErrorKind
		sentinel error
		literal  string
		text     string
	}{
		{kind: mail.SendErrorInvalid, sentinel: mail.ErrSendInvalid, literal: "invalid", text: "mail send invalid"},
		{kind: mail.SendErrorTooLarge, sentinel: mail.ErrSendTooLarge, literal: "too_large", text: "mail send exceeds bounds"},
		{kind: mail.SendErrorDuplicateRecipient, sentinel: mail.ErrSendDuplicateRecipient, literal: "duplicate_recipient", text: "mail send duplicate recipient"},
		{kind: mail.SendErrorIdentityNotAllowed, sentinel: mail.ErrSendIdentityNotAllowed, literal: "identity_not_allowed", text: "mail send identity not allowed"},
		{kind: mail.SendErrorExternalDisabled, sentinel: mail.ErrSendExternalDisabled, literal: "external_disabled", text: "mail send external delivery disabled"},
		{kind: mail.SendErrorAutoResponseSuppressed, sentinel: mail.ErrSendAutoResponseSuppressed, literal: "auto_response_suppressed", text: "mail send auto response suppressed"},
		{kind: mail.SendErrorRateLimited, sentinel: mail.ErrSendRateLimited, literal: "rate_limited", text: "mail send rate limited"},
	} {
		t.Run(string(testCase.kind), func(t *testing.T) {
			err := &mail.SendPolicyError{Kind: testCase.kind}
			if string(testCase.kind) != testCase.literal {
				t.Fatalf("SendErrorKind = %q, want %q", testCase.kind, testCase.literal)
			}
			if !errors.Is(err, testCase.sentinel) {
				t.Fatalf("errors.Is(%v, %v) = false", err, testCase.sentinel)
			}
			if testCase.sentinel.Error() != testCase.text || err.Error() != testCase.text {
				t.Fatalf("sentinel/Error text = %q/%q, want %q", testCase.sentinel.Error(), err.Error(), testCase.text)
			}
		})
	}

	var nilError *mail.SendPolicyError
	if nilError.Error() != "mail send policy error" || errors.Unwrap(nilError) != nil {
		t.Fatalf("nil SendPolicyError must be generic and not unwrap: %q / %v", nilError.Error(), errors.Unwrap(nilError))
	}
	unknown := &mail.SendPolicyError{Kind: mail.SendErrorKind("unknown-HOSTILE-SEND-MARKER@example.test")}
	if unknown.Error() != "mail send policy error" || errors.Unwrap(unknown) != nil {
		t.Fatalf("unknown SendPolicyError must be generic and not unwrap: %q / %v", unknown.Error(), errors.Unwrap(unknown))
	}
}

func TestEvaluateSendPolicy_ErrorsDoNotLeakHostileInputMarkers(t *testing.T) {
	marker := "HOSTILE-SEND-MARKER@external.test"
	for _, testCase := range []struct {
		name    string
		mutate  func(*mail.SendPolicyInput)
		wantErr error
	}{
		{
			name: "invalid address",
			mutate: func(input *mail.SendPolicyInput) {
				input.To[0].Email = marker + "\r\n"
			},
			wantErr: mail.ErrSendInvalid,
		},
		{
			name: "external domain",
			mutate: func(input *mail.SendPolicyInput) {
				input.ExternalMailEnabled = false
				input.To[0].Email = marker
			},
			wantErr: mail.ErrSendExternalDisabled,
		},
		{
			name: "automated loop header",
			mutate: func(input *mail.SendPolicyInput) {
				input.Automated = true
				input.LoopToken = marker
				input.AutoResponse.XLoop = []string{marker}
			},
			wantErr: mail.ErrSendAutoResponseSuppressed,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := validSendPolicyInput()
			testCase.mutate(&input)
			_, err := mail.EvaluateSendPolicy(input)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("EvaluateSendPolicy() error = %v, want errors.Is(%v)", err, testCase.wantErr)
			}
			if strings.Contains(err.Error(), marker) {
				t.Fatalf("error leaked hostile marker: %q", err.Error())
			}
		})
	}
}

func validSendPolicyInput() mail.SendPolicyInput {
	now := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	return mail.SendPolicyInput{
		Now:                   now,
		MessageBytes:          1,
		PolicyMaxMessageBytes: 1_048_576,
		PolicyMaxRecipients:   5,
		To:                    []mail.EngineAddress{{Email: "recipient@example.test"}},
		From:                  mail.EngineAddress{Email: "sender@example.test", Name: "Sender"},
		Identity:              mail.SendAuthorizedIdentity{CanonicalAddress: "sender@example.test", Active: true, SendAllowed: true},
		InternalDomains:       []string{"example.test"},
		ExternalMailEnabled:   true,
		UserRatePerMinute:     5,
		UserRateState:         mail.SendRateState{Tokens: 5, LastRefillAt: now},
		InstanceRatePerMinute: 5,
		InstanceRateState:     mail.SendRateState{Tokens: 5, LastRefillAt: now},
	}
}

func recipients(count int, prefix string) []mail.EngineAddress {
	values := make([]mail.EngineAddress, count)
	for index := range values {
		values[index] = mail.EngineAddress{Email: fmt.Sprintf("%s-%d@example.test", prefix, index)}
	}
	return values
}

func domains(count int) []string {
	values := make([]string, count)
	for index := range values {
		values[index] = fmt.Sprintf("domain-%d.example.test", index)
	}
	return values
}

func loopValues(count int) []string {
	values := make([]string, count)
	for index := range values {
		values[index] = fmt.Sprintf("loop-%02d", index)
	}
	return values
}
