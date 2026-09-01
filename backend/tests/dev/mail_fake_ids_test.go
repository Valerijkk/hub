package dev_test

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
	"testing"

	devdemo "github.com/valerii/hub/backend/internal/devdemo"
)

func TestMailFakeIDSourceBootstrapIDsRemainLegacyCompatible(t *testing.T) {
	source := devdemo.NewMailFakeIDSource()

	for _, want := range []string{
		"dev-mail-fake-v1-1",
		"dev-mail-fake-v1-2",
	} {
		if got := source.NextID(); got != want {
			t.Errorf("NextID() = %q, want %q", got, want)
		}
	}
}

func TestMailFakeIDSourceActivationUsesEntropyScopedRuntimeIDs(t *testing.T) {
	source := devdemo.NewMailFakeIDSource()
	for _, want := range []string{
		"dev-mail-fake-v1-1",
		"dev-mail-fake-v1-2",
	} {
		if got := source.NextID(); got != want {
			t.Fatalf("bootstrap NextID() = %q, want %q", got, want)
		}
	}
	entropy := bytes.NewReader([]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	})

	if err := source.ActivateRuntime(entropy); err != nil {
		t.Fatalf("ActivateRuntime() error = %v", err)
	}

	for _, want := range []string{
		"dev-mail-fake-v1-000102030405060708090a0b0c0d0e0f-1",
		"dev-mail-fake-v1-000102030405060708090a0b0c0d0e0f-2",
	} {
		if got := source.NextID(); got != want {
			t.Errorf("NextID() = %q, want %q", got, want)
		}
	}
}

func TestMailFakeIDSourceRuntimeReadyTracksSuccessfulActivation(t *testing.T) {
	source := devdemo.NewMailFakeIDSource()
	if source.RuntimeReady() {
		t.Fatal("RuntimeReady() = true during bootstrap, want false")
	}

	if err := source.ActivateRuntime(bytes.NewReader(make([]byte, 15))); err == nil {
		t.Fatal("ActivateRuntime(short reader) error = nil, want non-nil")
	}
	if source.RuntimeReady() {
		t.Fatal("RuntimeReady() = true after failed activation, want false")
	}

	if err := source.ActivateRuntime(bytes.NewReader([]byte{
		0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
		0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
	})); err != nil {
		t.Fatalf("ActivateRuntime() error = %v", err)
	}
	if !source.RuntimeReady() {
		t.Fatal("RuntimeReady() = false after successful activation, want true")
	}

	if err := source.ActivateRuntime(mailFakeIDErrorReader{}); err != nil {
		t.Fatalf("idempotent ActivateRuntime() error = %v", err)
	}
	if !source.RuntimeReady() {
		t.Fatal("RuntimeReady() = false after idempotent activation, want true")
	}
}

func TestMailFakeIDSourceRepeatedActivationDoesNotRereadOrReset(t *testing.T) {
	source := devdemo.NewMailFakeIDSource()
	if err := source.ActivateRuntime(bytes.NewReader([]byte{
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	})); err != nil {
		t.Fatalf("first ActivateRuntime() error = %v", err)
	}

	if got, want := source.NextID(), "dev-mail-fake-v1-101112131415161718191a1b1c1d1e1f-1"; got != want {
		t.Fatalf("first runtime NextID() = %q, want %q", got, want)
	}

	if err := source.ActivateRuntime(mailFakeIDErrorReader{}); err != nil {
		t.Fatalf("second ActivateRuntime() error = %v, want idempotent success", err)
	}
	if got, want := source.NextID(), "dev-mail-fake-v1-101112131415161718191a1b1c1d1e1f-2"; got != want {
		t.Errorf("runtime NextID() after repeated activation = %q, want %q", got, want)
	}
}

func TestMailFakeIDSourceActivationFailurePreservesBootstrapAndAllowsRetry(t *testing.T) {
	tests := []struct {
		name    string
		entropy func() *bytes.Reader
	}{
		{name: "nil"},
		{
			name: "short",
			entropy: func() *bytes.Reader {
				return bytes.NewReader(make([]byte, 15))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := devdemo.NewMailFakeIDSource()

			var err error
			if tt.entropy == nil {
				err = source.ActivateRuntime(nil)
			} else {
				err = source.ActivateRuntime(tt.entropy())
			}
			if err == nil {
				t.Fatal("ActivateRuntime() error = nil, want non-nil")
			}
			if got, want := source.NextID(), "dev-mail-fake-v1-1"; got != want {
				t.Fatalf("bootstrap NextID() after failed activation = %q, want %q", got, want)
			}

			if err := source.ActivateRuntime(bytes.NewReader([]byte{
				0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
				0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
			})); err != nil {
				t.Fatalf("retry ActivateRuntime() error = %v", err)
			}
			if got, want := source.NextID(), "dev-mail-fake-v1-202122232425262728292a2b2c2d2e2f-1"; got != want {
				t.Errorf("runtime NextID() after retry = %q, want %q", got, want)
			}
		})
	}

	t.Run("reader error", func(t *testing.T) {
		source := devdemo.NewMailFakeIDSource()
		if err := source.ActivateRuntime(mailFakeIDErrorReader{}); err == nil {
			t.Fatal("ActivateRuntime(error reader) error = nil, want non-nil")
		}
		if got, want := source.NextID(), "dev-mail-fake-v1-1"; got != want {
			t.Fatalf("bootstrap NextID() after reader error = %q, want %q", got, want)
		}
		if err := source.ActivateRuntime(bytes.NewReader([]byte{
			0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
			0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		})); err != nil {
			t.Fatalf("retry ActivateRuntime() after reader error = %v", err)
		}
		if got, want := source.NextID(), "dev-mail-fake-v1-202122232425262728292a2b2c2d2e2f-1"; got != want {
			t.Errorf("runtime NextID() after reader-error retry = %q, want %q", got, want)
		}
	})
}

func TestMailFakeIDSourceConcurrentRuntimeIDsAreUnique(t *testing.T) {
	source := devdemo.NewMailFakeIDSource()
	if err := source.ActivateRuntime(bytes.NewReader([]byte{
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
		0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	})); err != nil {
		t.Fatalf("ActivateRuntime() error = %v", err)
	}

	const callers = 128
	ids := make(chan string, callers)
	var wg sync.WaitGroup
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			ids <- source.NextID()
		}()
	}
	wg.Wait()
	close(ids)

	seen := make(map[string]struct{}, callers)
	const runtimePrefix = "dev-mail-fake-v1-303132333435363738393a3b3c3d3e3f-"
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Errorf("duplicate runtime ID %q", id)
		}
		seen[id] = struct{}{}
		if len(id) <= len(runtimePrefix) || id[:len(runtimePrefix)] != runtimePrefix {
			t.Errorf("runtime ID %q does not carry the activated lowercase-hex namespace", id)
			continue
		}
		sequence, err := strconv.ParseUint(id[len(runtimePrefix):], 10, 64)
		if err != nil || sequence == 0 {
			t.Errorf("runtime ID %q has invalid positive decimal sequence", id)
		}
	}
	if got := len(seen); got != callers {
		t.Errorf("unique runtime ID count = %d, want %d", got, callers)
	}
}

type mailFakeIDErrorReader struct{}

func (mailFakeIDErrorReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy reader failed")
}
