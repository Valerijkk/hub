package devdemo

import (
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"sync"
)

const mailFakeIDPrefix = "dev-mail-fake-v1-"

// MailFakeIDSource preserves the legacy bootstrap identifier sequence until
// the seeded control plane is ready, then switches once to a boot-unique
// runtime namespace.
type MailFakeIDSource struct {
	mu            sync.Mutex
	runtimeReady  bool
	runtimePrefix string
	sequence      uint64
}

// NewMailFakeIDSource returns a process-local fake ID source in its legacy
// bootstrap phase. Bootstrap IDs intentionally remain compatible with the
// persisted development control plane.
func NewMailFakeIDSource() *MailFakeIDSource {
	return &MailFakeIDSource{}
}

// NextID returns a serially unique fake provider identifier for the active
// phase. The phase switch and sequence reset are linearizable with callers.
func (source *MailFakeIDSource) NextID() string {
	source.mu.Lock()
	defer source.mu.Unlock()

	source.sequence++
	if source.runtimeReady {
		return source.runtimePrefix + "-" + strconv.FormatUint(source.sequence, 10)
	}
	return mailFakeIDPrefix + strconv.FormatUint(source.sequence, 10)
}

// ActivateRuntime switches the source from legacy bootstrap IDs to an
// entropy-namespaced runtime sequence. A failed read leaves all state intact,
// while a successful later call is deliberately idempotent.
func (source *MailFakeIDSource) ActivateRuntime(entropy io.Reader) error {
	if source == nil {
		return errors.New("mail fake ID source is nil")
	}

	source.mu.Lock()
	defer source.mu.Unlock()
	if source.runtimeReady {
		return nil
	}
	if entropy == nil {
		return errors.New("mail fake ID entropy is nil")
	}

	var namespace [16]byte
	if _, err := io.ReadFull(entropy, namespace[:]); err != nil {
		return err
	}

	source.runtimePrefix = mailFakeIDPrefix + hex.EncodeToString(namespace[:])
	source.sequence = 0
	source.runtimeReady = true
	return nil
}

// RuntimeReady reports whether the source has completed its one-way runtime
// activation successfully.
func (source *MailFakeIDSource) RuntimeReady() bool {
	if source == nil {
		return false
	}

	source.mu.Lock()
	defer source.mu.Unlock()
	return source.runtimeReady
}
