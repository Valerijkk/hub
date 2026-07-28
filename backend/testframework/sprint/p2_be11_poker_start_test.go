package sprint_test

// P2-BE-11 / TASK-041: poker CAS/TTL/RBAC + StartSprint race + UpdateSprint dates.
// Mutations: vote after reveal→reject; stranger session→403; two Start→one active; end<start→400.
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/platform/middleware"
	"github.com/valerii/hub/backend/internal/rbac"
	"github.com/valerii/hub/backend/internal/sprint"
)

// TestPoker_VoteAfterReveal_Rejected — W1-A4-002 mutation: без CAS голос после reveal проходит.
func TestPoker_VoteAfterReveal_Rejected(t *testing.T) {
	rdb, _ := newPokerTestRedis(t)
	poker := sprint.NewPokerService(rdb, nil)
	ctx := context.Background()
	userID := uuid.New()

	session, err := poker.StartPokerSession(ctx, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.NoError(t, poker.CastVote(ctx, session.SessionID, userID, "5"))

	_, err = poker.RevealVotes(ctx, session.SessionID)
	require.NoError(t, err)

	err = poker.CastVote(ctx, session.SessionID, uuid.New(), "8")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "раскрыты")
}

// TestPoker_VotesKeyTTL_InitializedOnStart — W1-A4-003: votes key + TTL при Start.
func TestPoker_VotesKeyTTL_InitializedOnStart(t *testing.T) {
	rdb, mr := newPokerTestRedis(t)
	poker := sprint.NewPokerService(rdb, nil)
	ctx := context.Background()

	session, err := poker.StartPokerSession(ctx, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	votesKey := "poker:" + session.SessionID + ":votes"
	assert.True(t, mr.Exists(votesKey), "votes hash must exist after StartPokerSession")
	ttl := mr.TTL(votesKey)
	assert.Greater(t, ttl, time.Duration(0), "votes key must have TTL")
}

// TestUpdateSprint_EndBeforeStart_BadRequest — W1-A4-007 mutation: без проверки PUT принимает end<start.
func TestUpdateSprint_EndBeforeStart_BadRequest(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()
	sprintID := uuid.New()
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "S", Status: sprintEnt.StatusPlanning,
		StartDate: start, EndDate: end,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = projectID

	badEnd := start.Add(-24 * time.Hour)
	_, err := svc.UpdateSprint(ctx, projectID, sprintID, sprint.UpdateSprintRequest{EndDate: &badEnd})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "дата окончания")
	assert.Equal(t, end, repo.sprints[sprintID].EndDate, "invalid Update must not mutate EndDate")
}

// TestStartSprint_Concurrent_OnlyOneActive — W1-A4-009: два Start → один active (mock mutex).
func TestStartSprint_Concurrent_OnlyOneActive(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	aID, bID := uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{aID, bID} {
		repo.sprints[id] = &ent.Sprint{
			ID: id, Name: "S-" + id.String()[:4], Status: sprintEnt.StatusPlanning,
			StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
			CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}
		repo.sprintProject[id] = projectID
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := svc.StartSprint(ctx, projectID, aID)
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, err := svc.StartSprint(ctx, projectID, bID)
		errs <- err
	}()
	wg.Wait()
	close(errs)

	var okCount, failCount int
	for err := range errs {
		if err == nil {
			okCount++
		} else {
			failCount++
			assert.Contains(t, err.Error(), "активный спринт")
		}
	}
	assert.Equal(t, 1, okCount, "ровно один Start должен пройти")
	assert.Equal(t, 1, failCount, "второй Start — conflict")

	active := 0
	for _, s := range repo.sprints {
		if s.Status == sprintEnt.StatusActive {
			active++
		}
	}
	assert.Equal(t, 1, active, "в проекте ровно один active sprint")
}

// TestStartSprintAtomic_Ent_Concurrent — guard на реальном Ent+SQLite (row-lock project).
func TestStartSprintAtomic_Ent_Concurrent(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	a, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "A", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	b, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "B", StartDate: start.AddDate(0, 0, 14), EndDate: end.AddDate(0, 0, 14),
	})
	require.NoError(t, err)

	svc := sprint.NewSprintService(env.repo, nil, nil)
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, e := svc.StartSprint(env.ctx, env.projectID, a.ID)
		errs <- e
	}()
	go func() {
		defer wg.Done()
		_, e := svc.StartSprint(env.ctx, env.projectID, b.ID)
		errs <- e
	}()
	wg.Wait()
	close(errs)

	var okCount int
	for e := range errs {
		if e == nil {
			okCount++
		}
	}
	assert.Equal(t, 1, okCount)

	found, err := env.repo.FindActiveSprint(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Equal(t, sprintEnt.StatusActive, found.Status)

	all, err := env.repo.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	active := 0
	for _, s := range all {
		if s.Status == sprintEnt.StatusActive {
			active++
		}
	}
	assert.Equal(t, 1, active)
}

// TestStartSprintAtomic_Ent_SecondActiveConflicts — P2-BE-11 / TASK-041.
// Последовательный второй Start при уже active → *ErrActiveSprintConflict.
// Мутация: убрать conflict-check → B становится active → RED (не зависит от SQLite tx race).
func TestStartSprintAtomic_Ent_SecondActiveConflicts(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	a, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "First", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	b, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Second", StartDate: start.AddDate(0, 0, 14), EndDate: end.AddDate(0, 0, 14),
	})
	require.NoError(t, err)

	_, err = env.repo.StartSprintAtomic(env.ctx, env.projectID, a.ID)
	require.NoError(t, err)

	_, err = env.repo.StartSprintAtomic(env.ctx, env.projectID, b.ID)
	require.Error(t, err)
	var conflict *sprint.ErrActiveSprintConflict
	require.ErrorAs(t, err, &conflict)
	assert.Equal(t, "First", conflict.Name)

	all, err := env.repo.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	active := 0
	for _, s := range all {
		if s.Status == sprintEnt.StatusActive {
			active++
		}
	}
	assert.Equal(t, 1, active)
}

// TestPoker_RBAC_StrangerForbidden — W1-A4-005: outsider с sessionId → 403; member vote OK.
func TestPoker_RBAC_StrangerForbidden(t *testing.T) {
	dsn := "file:sprpoker_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	member, err := client.User.Create().SetEmail("pm-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	outsider, err := client.User.Create().SetEmail("po-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("Poker Deny").SetKey("PD").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).SetProjectID(proj.ID).SetProjectRole("owner").Save(ctx)
	require.NoError(t, err)

	todo, err := client.WorkflowStatus.Create().
		SetName("Todo").SetCategory(workflowstatus.CategoryTodo).SetColor("#6B7280").
		SetSortOrder(0).SetIsDefault(true).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)

	iss, err := client.Issue.Create().
		SetTitle("Est").SetIssueNumber(1).SetIssueKey("PD-1").
		SetIssueType(issue.IssueTypeTask).SetProjectID(proj.ID).
		SetReporterID(member.ID).SetStatusID(todo.ID).Save(ctx)
	require.NoError(t, err)

	repo := sprint.NewEntRepository(client)
	svc := sprint.NewSprintService(repo, rdb, nil)
	poker := sprint.NewPokerService(rdb, nil)
	h := sprint.NewHandler(svc, poker)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	start := time.Now().UTC()
	sp, err := repo.CreateSprint(ctx, proj.ID, sprint.CreateSprintRequest{
		Name: "S", StartDate: start, EndDate: start.AddDate(0, 0, 14),
	})
	require.NoError(t, err)

	session, err := poker.StartPokerSession(ctx, proj.ID, sp.ID, iss.ID)
	require.NoError(t, err)

	voteBody := `{"value":"5"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sprints/poker/"+session.SessionID+"/vote", bytes.NewReader([]byte(voteBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", outsider.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode, "stranger vote must be 403")

	req = httptest.NewRequest(http.MethodPost, "/api/sprints/poker/"+session.SessionID+"/reveal", nil)
	req.Header.Set("X-Test-User", outsider.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode, "stranger reveal must be 403")

	req = httptest.NewRequest(http.MethodGet, "/api/sprints/poker/"+session.SessionID, nil)
	req.Header.Set("X-Test-User", outsider.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode, "stranger GET must be 403")

	req = httptest.NewRequest(http.MethodPost, "/api/sprints/poker/"+session.SessionID+"/vote", bytes.NewReader([]byte(voteBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", member.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestStartPoker_ForeignIssue_BadRequest — W1-A4-006.
func TestStartPoker_ForeignIssue_BadRequest(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	projectID := uuid.New()

	err := svc.EnsureIssueInProject(context.Background(), projectID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не принадлежит проекту")

	app, repo2, projectID2, _ := setupSprintHandlerApp(t)
	app = fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	// Remount with ErrorHandler for HTTP status assert.
	rdb, _ := newPokerTestRedis(t)
	h := sprint.NewHandler(sprint.NewSprintService(repo2, nil, nil), sprint.NewPokerService(rdb, nil))
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	}
	sprints := api.Group("/projects/:projectId/sprints", authMw)
	sprints.Post("/", h.CreateSprint)
	sprints.Post("/:sprintId/poker", h.StartPokerSession)

	base := "/api/projects/" + projectID2.String() + "/sprints"
	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"S","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	pokerBody := fmt.Sprintf(`{"issue_id":%q}`, uuid.New().String())
	req = httptest.NewRequest(http.MethodPost, base+"/"+sp.ID+"/poker", bytes.NewReader([]byte(pokerBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
