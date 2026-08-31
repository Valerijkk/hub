package schemamaintenance_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/schemamaintenance/rootfk"
)

const (
	frozenManifestPath = "../../docker/schema-maintenance/rootfk/frozen-external-closure.manifest"
	frozenLineCount    = 82
	frozenByteCount    = 4767
	frozenSHA256       = "889c14127f47bdb73108bfef07573ff4b0bde7fd3b78b47122718a1f5b0a36b7"
)

func TestParseInspectProjectionQuotedCreateAndMultiAdd(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "atlas_quoted_create_multi_add.sql"))
	if err != nil {
		t.Fatalf("read quoted SQL fixture: %v", err)
	}

	got, err := rootfk.ParseInspectProjection(data)
	if err != nil {
		t.Fatalf("parse quoted CREATE and ALTER multi-ADD projection: %v", err)
	}
	want := strings.Join([]string{
		"Organization|calendar_events|organization_id",
		"Organization|mail_accounts|organization_id",
		"User|calendar_events|calendar_owner_id",
		"User|mail_accounts|user_id",
	}, "\n") + "\n"
	if got != want {
		t.Fatalf("quoted CREATE and ALTER multi-ADD projection mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestParseInspectProjectionIgnoresQuotedReferencesIdentifierInCheck(t *testing.T) {
	sql := `CREATE TABLE "public"."mail_messages" (
		"user_id" uuid NOT NULL,
		"references" jsonb NOT NULL,
		CONSTRAINT "mail_messages_references" CHECK (jsonb_typeof("references") = 'array'),
		CONSTRAINT "mail_messages_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id")
	);`

	got, err := rootfk.ParseInspectProjection([]byte(sql))
	if err != nil {
		t.Fatalf("parse quoted references identifier in CHECK: %v", err)
	}
	if want := "User|mail_messages|user_id\n"; got != want {
		t.Fatalf("quoted references CHECK projection mismatch\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestParseInspectProjectionRejectsMalformedRootConstraints(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "inline root reference without foreign key",
			sql: `CREATE TABLE mail_accounts (
				user_id uuid REFERENCES users(id)
			);`,
		},
		{
			name: "root reference targets non-id column",
			sql: `CREATE TABLE mail_accounts (
				CONSTRAINT mail_accounts_user_fkey FOREIGN KEY (user_id) REFERENCES users(legacy_id)
			);`,
		},
		{
			name: "multi-column source key",
			sql: `CREATE TABLE mail_accounts (
				CONSTRAINT mail_accounts_user_fkey FOREIGN KEY (user_id, organization_id) REFERENCES users(id)
			);`,
		},
		{
			name: "root foreign key alter action does not start add",
			sql: `ALTER TABLE mail_accounts ALTER COLUMN user_id REFERENCES users(id);`,
		},
		{
			name: "multiple references in one alter action",
			sql: `ALTER TABLE mail_accounts ADD CONSTRAINT mail_accounts_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) REFERENCES organizations(id);`,
		},
		{
			name: "unterminated create table",
			sql: `CREATE TABLE mail_accounts (
				CONSTRAINT mail_accounts_user_fkey FOREIGN KEY (user_id) REFERENCES users(id);`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := rootfk.ParseInspectProjection([]byte(test.sql)); err == nil {
				t.Fatal("ParseInspectProjection unexpectedly accepted malformed direct root foreign key")
			}
		})
	}
}

func TestParseInspectProjectionRejectsDuplicateTriple(t *testing.T) {
	sql := `CREATE TABLE mail_accounts (
		CONSTRAINT mail_accounts_user_fkey FOREIGN KEY (user_id) REFERENCES users(id)
	);
	CREATE TABLE mail_accounts (
		CONSTRAINT mail_accounts_user_fkey_again FOREIGN KEY (user_id) REFERENCES users(id)
	);`

	if _, err := rootfk.ParseInspectProjection([]byte(sql)); err == nil {
		t.Fatal("ParseInspectProjection unexpectedly accepted a duplicate root-FK triple")
	}
}

func TestFrozenRootFKAssetAndIndependentProjection(t *testing.T) {
	asset := readFixture(t, frozenManifestPath)
	if got := len(asset); got != frozenByteCount {
		t.Fatalf("frozen manifest byte count = %d, want %d", got, frozenByteCount)
	}
	if !strings.HasSuffix(string(asset), "\n") || strings.Contains(string(asset), "\r") {
		t.Fatal("frozen manifest is not LF-only with a terminating newline")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(asset)); got != frozenSHA256 {
		t.Fatalf("frozen manifest SHA-256 = %s, want %s", got, frozenSHA256)
	}

	lines := strings.Split(strings.TrimSuffix(string(asset), "\n"), "\n")
	if got := len(lines); got != frozenLineCount {
		t.Fatalf("frozen manifest line count = %d, want %d", got, frozenLineCount)
	}
	projection := make([]string, 0, len(lines))
	for index, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) != 6 {
			t.Fatalf("frozen manifest line %d has %d fields, want 6", index+1, len(fields))
		}
		projection = append(projection, strings.Join(fields[:3], "|"))
	}
	wantProjection := readFixture(t, filepath.Join("testdata", "frozen_rootfk_projection.txt"))
	if gotProjection := strings.Join(projection, "\n") + "\n"; gotProjection != string(wantProjection) {
		t.Fatalf("frozen manifest projection does not equal the independent fixture\nwant:\n%s\ngot:\n%s", wantProjection, gotProjection)
	}
}

func TestVerifyFilesAcceptsFrozenProjectionRenderedAsQuotedAlterSQL(t *testing.T) {
	asset := readFixture(t, frozenManifestPath)
	projection := string(readFixture(t, filepath.Join("testdata", "frozen_rootfk_projection.txt")))
	sql := quotedAlterSQL(t, projection)
	dir := t.TempDir()
	expectedPath := writeTempFile(t, dir, "frozen-external-closure.manifest", asset)
	desiredPath := writeTempFile(t, dir, "desired.sql", []byte(sql))
	replayPath := writeTempFile(t, dir, "replay.sql", []byte(sql))

	evidence, err := rootfk.VerifyFiles(expectedPath, desiredPath, replayPath)
	if err != nil {
		t.Fatalf("VerifyFiles returned error for exact quoted ALTER replay: %v", err)
	}
	if evidence.ExpectedCount != frozenLineCount || evidence.ExpectedBytes != frozenByteCount || evidence.ExpectedSHA256 != frozenSHA256 {
		t.Fatalf("expected evidence = %+v, want frozen 82/4767/%s", evidence, frozenSHA256)
	}
	if evidence.DesiredCount != frozenLineCount || evidence.ReplayCount != frozenLineCount {
		t.Fatalf("desired/replay counts = %d/%d, want %d/%d", evidence.DesiredCount, evidence.ReplayCount, frozenLineCount, frozenLineCount)
	}
	if evidence.DesiredSHA256 != evidence.ReplaySHA256 || evidence.DesiredBytes != evidence.ReplayBytes {
		t.Fatalf("desired/replay evidence diverged: %+v", evidence)
	}
}

func TestVerifyFilesRejectsDesiredReplayMismatch(t *testing.T) {
	asset := readFixture(t, frozenManifestPath)
	projection := string(readFixture(t, filepath.Join("testdata", "frozen_rootfk_projection.txt")))
	dir := t.TempDir()
	expectedPath := writeTempFile(t, dir, "frozen-external-closure.manifest", asset)
	desiredPath := writeTempFile(t, dir, "desired.sql", []byte(quotedAlterSQL(t, projection)))
	replayPath := writeTempFile(t, dir, "replay.sql", []byte(quotedAlterSQL(t, replaceProjectionLine(t, projection, "User|user_blocks|blocked_id", "User|user_blocks|replay_blocked_id"))))

	if _, err := rootfk.VerifyFiles(expectedPath, desiredPath, replayPath); err == nil {
		t.Fatal("VerifyFiles unexpectedly accepted different desired and replay projections")
	}
}

func TestVerifyFilesRejectsNoncanonicalFrozenAssets(t *testing.T) {
	asset := readFixture(t, frozenManifestPath)
	projection := string(readFixture(t, filepath.Join("testdata", "frozen_rootfk_projection.txt")))
	sql := []byte(quotedAlterSQL(t, projection))

	tests := []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{
			name: "CRLF",
			mutate: func(asset []byte) []byte {
				return []byte(strings.Replace(string(asset), "\n", "\r\n", 1))
			},
		},
		{
			name: "missing terminal newline",
			mutate: func(asset []byte) []byte {
				return append([]byte(nil), asset[:len(asset)-1]...)
			},
		},
		{
			name: "same-length mutation",
			mutate: func(asset []byte) []byte {
				mutated := append([]byte(nil), asset...)
				mutated[0] = 'X'
				return mutated
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			expectedPath := writeTempFile(t, dir, "frozen-external-closure.manifest", test.mutate(asset))
			desiredPath := writeTempFile(t, dir, "desired.sql", sql)
			replayPath := writeTempFile(t, dir, "replay.sql", sql)

			if _, err := rootfk.VerifyFiles(expectedPath, desiredPath, replayPath); err == nil {
				t.Fatal("VerifyFiles unexpectedly accepted a noncanonical frozen manifest")
			}
		})
	}
}

func readFixture(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %q: %v", path, err)
	}
	return data
}

func writeTempFile(t *testing.T, directory, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write temporary fixture %q: %v", path, err)
	}
	return path
}

func quotedAlterSQL(t *testing.T, projection string) string {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(projection, "\n"), "\n")
	if len(lines) != frozenLineCount {
		t.Fatalf("projection line count = %d, want %d", len(lines), frozenLineCount)
	}

	var output strings.Builder
	for index, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) != 3 {
			t.Fatalf("projection line %d has %d fields, want 3", index+1, len(fields))
		}
		target := "users"
		if fields[0] == "Organization" {
			target = "organizations"
		} else if fields[0] != "User" {
			t.Fatalf("projection line %d root %q is unsupported", index+1, fields[0])
		}
		fmt.Fprintf(&output, "ALTER TABLE ONLY \"public\".\"%s\" ADD CONSTRAINT \"root_fk_%03d\" FOREIGN KEY (\"%s\") REFERENCES \"public\".\"%s\" (\"id\");\n", fields[1], index+1, fields[2], target)
	}
	return output.String()
}

func replaceProjectionLine(t *testing.T, projection, old, replacement string) string {
	t.Helper()
	if !strings.Contains(projection, old+"\n") {
		t.Fatalf("projection fixture does not contain %q", old)
	}
	return strings.Replace(projection, old+"\n", replacement+"\n", 1)
}
