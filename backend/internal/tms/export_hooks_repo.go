//go:build hubtest

package tms

import (
	"context"

	"github.com/valerii/hub/backend/ent"
)

// ExportEntRepositoryClient возвращает ent.Client из EntRepository для hubtest.
func ExportEntRepositoryClient(r *EntRepository) *ent.Client {
	return r.client
}

// ExportSetAfterExistingRunBatchHook инъектирует mid-fail после batch CreateBulk
// внутри ImportResultsToExistingRun tx (A3W2-2-003 / TASK-139).
func ExportSetAfterExistingRunBatchHook(fn func(ctx context.Context) error) {
	afterExistingRunBatchHook = fn
}

// ExportMaxListTestResults — A3W2-3-001 list/nested results cap.
func ExportMaxListTestResults() int { return maxListTestResults }

// ExportSetMaxListTestResultsForTest временно снижает cap (mutation на Limit).
func ExportSetMaxListTestResultsForTest(n int) (restore func()) {
	old := maxListTestResults
	maxListTestResults = n
	return func() { maxListTestResults = old }
}

// ExportMaxListEnvironments — A3W2-3-001 env list cap.
func ExportMaxListEnvironments() int { return maxListEnvironments }

// ExportSetMaxListEnvironmentsForTest временно снижает env cap.
func ExportSetMaxListEnvironmentsForTest(n int) (restore func()) {
	old := maxListEnvironments
	maxListEnvironments = n
	return func() { maxListEnvironments = old }
}

// ExportMaxListProjectIssues — A3W2-3-001 issues-with-TC list cap.
func ExportMaxListProjectIssues() int { return maxListProjectIssues }

// ExportSetMaxListProjectIssuesForTest временно снижает issues cap.
func ExportSetMaxListProjectIssuesForTest(n int) (restore func()) {
	old := maxListProjectIssues
	maxListProjectIssues = n
	return func() { maxListProjectIssues = old }
}

// ExportMaxBatchCreateTestResults — A3W2-3-001 AddCases/BatchCreate err threshold.
func ExportMaxBatchCreateTestResults() int { return maxBatchCreateTestResults }

// ExportSetMaxBatchCreateTestResultsForTest временно снижает batch cap.
func ExportSetMaxBatchCreateTestResultsForTest(n int) (restore func()) {
	old := maxBatchCreateTestResults
	maxBatchCreateTestResults = n
	return func() { maxBatchCreateTestResults = old }
}
