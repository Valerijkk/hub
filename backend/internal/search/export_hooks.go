//go:build hubtest

package search

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/predicate"
"github.com/valerii/hub/backend/internal/messenger"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func ExportGetUserID(c fiber.Ctx) (uuid.UUID, error) {
	return getUserID(c)
}

func ExportSavedFilterToDTO(f *ent.SavedFilter) *SavedFilterDTO {
	return savedFilterToDTO(f)
}

func ExportNewSearchServiceSQL(sqlDB *sql.DB) *SearchService {
	return &SearchService{sqlDB: sqlDB}
}

func ExportNewSearchServiceRepoOnly(repo Repository) *SearchService {
	return &SearchService{repo: repo}
}

func ExportNewSearchServiceRepoSQL(repo Repository, sqlDB *sql.DB) *SearchService {
	return &SearchService{repo: repo, sqlDB: sqlDB}
}

func ExportNewSearchServiceFull(repo Repository, sqlDB *sql.DB, mongoDB *mongo.Database) *SearchService {
	return &SearchService{repo: repo, sqlDB: sqlDB, mongoDB: mongoDB}
}

func ExportNewSearchServiceWithAuthorizedMessageSearchScopeTest(
	resolver messenger.AuthorizedMessageSearchScopeResolver,
	executor AuthorizedMessageSearchExecutor,
	starter AuthorizedMessageSearchTaskStarter,
) *SearchService {
	return newSearchServiceWithAuthorizedMessageSearchDependencies(
		nil, nil, nil, resolver, executor, starter,
	)
}
func ExportSetSearchServiceRepo(s *SearchService, repo Repository) {
	s.repo = repo
}

func ExportSearchIssues(s *SearchService, ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	return s.searchIssues(ctx, query, projectIDs)
}

func ExportSearchTestCases(s *SearchService, ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	return s.searchTestCases(ctx, query, projectIDs)
}

func ExportSearchUsers(s *SearchService, ctx context.Context, query string, callerID uuid.UUID) ([]SearchResult, error) {
	return s.searchUsers(ctx, query, callerID)
}

func ExportSearchUsersUnscoped(s *SearchService, ctx context.Context, query string) ([]SearchResult, error) {
	return s.searchUsersUnscoped(ctx, query)
}

func ExportSearchProjects(s *SearchService, ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	return s.searchProjects(ctx, query, projectIDs)
}

func ExportSearchFiles(s *SearchService, ctx context.Context, query string, projectIDs []uuid.UUID) ([]SearchResult, error) {
	return s.searchFiles(ctx, query, projectIDs)
}

func ExportSearchChannels(s *SearchService, ctx context.Context, query string, projectIDs []uuid.UUID, userID uuid.UUID) ([]SearchResult, error) {
	return s.searchChannels(ctx, query, projectIDs, userID)
}

func ExportSearchMessages(s *SearchService, ctx context.Context, query string, userID uuid.UUID) ([]SearchResult, error) {
	return s.searchAuthorizedMessagesForUser(ctx, query, userID)
}

func ExportGetUserChannelIDs(s *SearchService, ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.getUserChannelIDs(ctx, userID)
}

func ExportGetUserProjectIDs(s *SearchService, ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.getUserProjectIDs(ctx, userID)
}

func ExportIsSystemAdmin(s *SearchService, ctx context.Context, userID uuid.UUID) bool {
	return s.isSystemAdmin(ctx, userID)
}

func ExportUUIDSliceToPgArray(ids []uuid.UUID) string {
	return uuidSliceToPgArray(ids)
}

func ExportDateFieldPredicate(value, op, field string) predicate.Issue {
	return dateFieldPredicate(value, op, field)
}

func ExportIssuesIndex() string {
	return issuesIndex
}

func ExportNewMeiliIndexerTest(host, apiKey string, client *http.Client) *MeiliIndexer {
	return &MeiliIndexer{host: strings.TrimRight(host, "/"), apiKey: apiKey, client: client}
}

// ExportValidateMeiliHost — P2-BE-18 SSRF guard для hubtest.
func ExportValidateMeiliHost(host string) error {
	return validateMeiliHost(host)
}

// ExportTruncateRunes — P3-BE-04 / TASK-060: rune-safe truncate для hubtest.
func ExportTruncateRunes(s string, maxRunes int) string {
	return truncateRunes(s, maxRunes)
}

func ExportMeiliIndexerHost(m *MeiliIndexer) string {
	if m == nil {
		return ""
	}
	return m.host
}

// ExportMeiliIndexerClient — TASK-092: SafeDial client для hubtest.
func ExportMeiliIndexerClient(m *MeiliIndexer) *http.Client {
	if m == nil {
		return nil
	}
	return m.client
}

// ExportMeiliAllowsOperatorPrivate — A3W5-4-001 hubtest.
func ExportMeiliAllowsOperatorPrivate(raw string) bool {
	return meiliAllowsOperatorPrivate(raw)
}

// ExportSetMeiliIndexBatchSize — A3W5-4-003: уменьшить batch в hubtest (mutation pagination).
// Возвращает previous; restore через defer.
func ExportSetMeiliIndexBatchSize(n int) int {
	prev := meiliIndexBatchSize
	meiliIndexBatchSize = n
	return prev
}

// ExportValidateJQLCondition — A3W5-4-002 hubtest.
func ExportValidateJQLCondition(c JQLCondition) error {
	return validateJQLCondition(c)
}

type ExportMeiliIssueDoc struct {
	ID        string
	IssueKey  string
	Title     string
	ProjectID string
}

func ExportMeiliUpsertDocuments(m *MeiliIndexer, ctx context.Context, index string, docs []ExportMeiliIssueDoc) error {
	if m == nil {
		return nil
	}
	internal := make([]meiliIssueDoc, len(docs))
	for i, d := range docs {
		internal[i] = meiliIssueDoc{
			ID:        d.ID,
			IssueKey:  d.IssueKey,
			Title:     d.Title,
			ProjectID: d.ProjectID,
		}
	}
	return m.upsertDocuments(ctx, index, internal)
}
