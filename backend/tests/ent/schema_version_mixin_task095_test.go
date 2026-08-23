//go:build hubtest

package ent_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mutableNeedsVersion — NFR-24 / ADR-ENT-VERSION-MIXIN-2026 (unclear-12 excluded).
var mutableNeedsVersion = []string{
	"aiprovider", "aiuserpreference", "automationrule", "board", "channel",
	"comment", "component", "crmcontact", "crmdeal", "customfielddefinition",
	"dashboardlayout", "environment", "githubintegration", "gitlabintegration",
	"hubbase", "hubcolumn", "hubrecord", "hubtable", "invoice", "issue",
	"journalentry", "label", "labelgroup", "leaverequest", "ledgeraccount",
	"marketplacelisting", "newspost", "notificationpreference", "organization",
	"poll", "project", "projecttemplate", "role", "savedfilter", "scheduledcall",
	"sprint", "systemsetting", "team", "testcase", "testrun", "testsuite",
	"user", "webhookendpoint", "workflowstatus", "workflowtransition",
}

// unclear12 — owner exclusion без VersionMixin; invite/notification имеют
// отдельный Fields version (A2W8-3-002/006 OL) и не в этом списке.
var unclear12 = []string{
	"apikey", "call", "channelinvite", "coffeeenrollment", "coffeeround",
	"filemetadata", "pushsubscription", "recording",
	"session", "timelog",
}

func TestSchema_TASK095_VersionMixinOnMutable(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	schemaDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")

	mixinPath := filepath.Join(schemaDir, "mixin_version.go")
	mixinSrc, err := os.ReadFile(mixinPath)
	require.NoError(t, err)
	assert.Contains(t, string(mixinSrc), "type VersionMixin struct")
	assert.Contains(t, string(mixinSrc), `field.Int("version")`)
	assert.Contains(t, string(mixinSrc), "Positive()")
	assert.Contains(t, string(mixinSrc), "ADR-ENT-VERSION-MIXIN-2026")

	for _, stem := range mutableNeedsVersion {
		src, err := os.ReadFile(filepath.Join(schemaDir, stem+".go"))
		require.NoError(t, err, stem)
		body := string(src)
		assert.Contains(t, body, "VersionMixin{}", stem+" must use VersionMixin")
		assert.Contains(t, body, "Mixin()", stem+" must define Mixin()")
		// Mutation: duplicate Fields version must not return after Mixin migration.
		assert.NotContains(t, body, `field.Int("version")`, stem+" must not duplicate version in Fields")
	}

	for _, stem := range unclear12 {
		src, err := os.ReadFile(filepath.Join(schemaDir, stem+".go"))
		require.NoError(t, err, stem)
		body := string(src)
		assert.NotContains(t, body, "VersionMixin{}", stem+" unclear exclusion")
		assert.False(t, strings.Contains(body, `field.Int("version")`), stem+" unclear no version field")
	}
}

func TestSchema_TASK095_MigrateSQLExists(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260728_task095_version_mixin.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "ALTER TABLE sprints ADD COLUMN IF NOT EXISTS version")
	assert.Contains(t, body, "ALTER TABLE projects ADD COLUMN IF NOT EXISTS version")
	assert.Contains(t, body, "Unclear-12")
}
