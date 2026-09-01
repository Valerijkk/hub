//go:build hubtest

package mail_test

import (
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
)

func TestExportMailboxEntStableThreadIDParity(t *testing.T) {
	t.Parallel()

	accountID := uuid.MustParse("4ec88cba-fd28-4c8a-ad24-0a530a0bbb0d")
	key := stableThreadIDNonzeroKey()

	got, err := mail.ExportMailboxEntStableThreadID(accountID, mail.ThreadKeyVersionV1, key)
	if err != nil {
		t.Fatalf("export stable thread ID: %v", err)
	}

	want := referenceVersionedThreadID(accountID, mail.ThreadKeyVersionV1, key)
	if got != want {
		t.Fatalf("versioned stable thread ID = %s, want %s", got, want)
	}
	legacy := referenceLegacyKeyOnlyThreadID(accountID, key)
	if got == legacy {
		t.Fatalf("versioned stable thread ID = legacy key-only ID %s", legacy)
	}
}

func TestExportMailboxEntStableThreadIDRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	accountID := uuid.MustParse("4ec88cba-fd28-4c8a-ad24-0a530a0bbb0d")
	key := stableThreadIDNonzeroKey()
	var zeroKey mail.ThreadKey
	tests := []struct {
		name      string
		accountID uuid.UUID
		version   mail.ThreadKeyVersion
		key       mail.ThreadKey
	}{
		{name: "zero account", accountID: uuid.Nil, version: mail.ThreadKeyVersionV1, key: key},
		{name: "zero key", accountID: accountID, version: mail.ThreadKeyVersionV1, key: zeroKey},
		{name: "unsupported version", accountID: accountID, version: mail.ThreadKeyVersion(0xff), key: key},
	}

	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			got, err := mail.ExportMailboxEntStableThreadID(testCase.accountID, testCase.version, testCase.key)
			if err == nil {
				t.Fatal("expected malformed input to be rejected")
			}
			if got != uuid.Nil {
				t.Fatalf("rejected input ID = %s, want nil UUID", got)
			}
		})
	}
}

func stableThreadIDNonzeroKey() mail.ThreadKey {
	var key mail.ThreadKey
	for index := range key {
		key[index] = byte(index + 1)
	}
	return key
}

func referenceVersionedThreadID(accountID uuid.UUID, version mail.ThreadKeyVersion, key mail.ThreadKey) uuid.UUID {
	identity := make([]byte, 1+len(key))
	identity[0] = byte(version)
	copy(identity[1:], key[:])
	return referenceFramedThreadUUID([]byte("hub.mail.sync-id.v1"), []byte("thread"), accountID[:], identity)
}

func referenceLegacyKeyOnlyThreadID(accountID uuid.UUID, key mail.ThreadKey) uuid.UUID {
	return referenceFramedThreadUUID([]byte("hub.mail.sync-id.v1"), []byte("thread"), accountID[:], key[:])
}

func referenceFramedThreadUUID(parts ...[]byte) uuid.UUID {
	frame := make([]byte, 0)
	for _, part := range parts {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		frame = append(frame, length[:]...)
		frame = append(frame, part...)
	}
	digest := sha256.Sum256(frame)
	var id uuid.UUID
	copy(id[:], digest[:16])
	id[6] = id[6]&0x0f | 0x80
	id[8] = id[8]&0x3f | 0x80
	return id
}
