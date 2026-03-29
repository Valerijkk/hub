# Phase 4: TMS & Allure - Research

**Researched:** 2026-03-21
**Domain:** Test Management System + Allure CI integration + Import/Export
**Confidence:** HIGH

## Summary

Phase 4 builds the Test Management System (TMS) module for Hub, covering test case management, test runs, Allure CI results ingestion, bidirectional issue linking, and import/export capabilities. This is a large domain that combines CRUD-heavy Ent schemas (test cases, suites, folders, runs, results), a defensive Allure JSON parser, dual-store persistence (structured data in PostgreSQL, raw Allure JSON in MongoDB), and cross-module integration with the existing tracker (bidirectional linking, bug creation from failed tests, coverage tracking).

The architecture follows the established modular monolith pattern: `internal/tms/` with handler/service/repository layers, Ent schemas for TMS entities, a CombinedRepository (Ent + Mongo) mirroring the messenger module pattern, NATS events for cross-module communication (test run completion -> messenger channel post, tracker coverage update), and RBAC enforcement using the existing `PermTestCaseCreate`, `PermTestCaseEdit`, `PermTestRunCreate`, `PermTestRunExecute` permissions already defined in the bitmap.

The critical technical challenge is the Allure results parser. The `allure-results` format is a de facto standard with significant variation across framework adapters (JUnit, pytest, JS). The parser must be defensive (every field optional), support multiple test-case matching strategies (annotation > fullName > class+method), store raw JSON in MongoDB for auditability, and extract structured data into PostgreSQL for the TMS views. Import/export adds Excel (excelize), CSV (encoding/csv), XML (encoding/xml), JSON (encoding/json), and PDF (maroto) capabilities.

**Primary recommendation:** Follow the tracker module pattern exactly (handler/service/repository + Ent schemas + outbox events). Use CombinedRepository for dual PG+Mongo persistence. Build the Allure parser as a separate, testable package within tms/ with defensive parsing and multi-strategy matching.

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| TMS-01 | Test cases: classic (steps + expected) + BDD (Given/When/Then) | Ent schema with type enum + JSONB steps field; BDD uses same steps structure with Given/When/Then labels |
| TMS-02 | Folders/sections + tags + components + versioning with rollback | Ent tree (parent_id self-ref) for folders, M2M edges for tags/components, versions table with JSONB snapshots |
| TMS-03 | GitHub linking: @TestCase annotation matching, code display, sync status, webhook | Allure labels parsing for tms links; webhook endpoint for GitHub code changes |
| TMS-04 | Test runs: manual + automatic (Allure) + mixed + re-open + assign to testers | Ent TestRun entity with type enum, M2M tester assignment, status FSM |
| TMS-05 | Screenshot/comment per test result step | FileMetadata edge from TestResult + per-step comment JSONB |
| TMS-06 | Bidirectional issue linking, create bug from failed test, coverage, traceability matrix | Ent edges TestCase <-> Issue, service method for bug creation from failed result |
| TMS-08 | Allure-level reporting: trends, flaky tests, defect categories, run comparison | Aggregation queries on TestResult + flaky detection via history analysis |
| TMS-09 | Import: CSV, JSON, XML, Excel with universal field mapping | excelize for Excel, encoding/csv, encoding/xml, encoding/json + field mapping config |
| TMS-10 | Export: CSV, JSON, XML, PDF | Same libs for export + maroto for PDF generation |
| ALLU-01 | CI Allure parser: read allure-results JSON, POST to TMS API | Defensive Go parser for *-result.json + *-container.json; API endpoint for batch upload |
| ALLU-02 | Matching results to test cases by Allure ID/name | Multi-strategy: annotation label > fullName > class+method; priority-ordered fallback |
| ALLU-03 | Run history, statistics, Telegram notifications | TestRun history query + statistics aggregation + NATS event -> notification module |
| IMEX-02 | Import test cases from Excel | excelize v2.10.x for XLSX parsing with configurable column mapping |
</phase_requirements>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| entgo.io/ent | v0.14.5 | ORM for TMS Ent schemas (test cases, runs, results, folders) | Already used project-wide; graph edges for cross-module linking |
| go.mongodb.org/mongo-driver/v2 | v2.5.0 | Store raw Allure JSON results | Already used by messenger; document store suits variable Allure JSON |
| github.com/gofiber/fiber/v3 | v3.1.0 | HTTP handlers for TMS API | Project standard |
| github.com/nats-io/nats.go | v1.49.0 | Events: test run completed, coverage changed | Project standard event bus |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/xuri/excelize/v2 | v2.10.1 | Excel import/export (XLSX) | TMS-09, TMS-10, IMEX-02: reading/writing Excel test case files |
| github.com/johnfercher/maroto/v2 | v2.3.3 | PDF report generation | TMS-10: export test runs and reports as PDF |
| encoding/csv | stdlib | CSV import/export | TMS-09, TMS-10: CSV test case files |
| encoding/xml | stdlib | XML import/export | TMS-09, TMS-10: XML test case files |
| encoding/json | stdlib | JSON import/export + Allure parsing | Core Allure parser + JSON import/export |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| maroto (PDF) | go-pdf/fpdf | go-pdf/fpdf archived March 2025; maroto v2 wraps it and is actively maintained |
| maroto (PDF) | unipdf | Commercial license required; overkill for simple reports |
| excelize (Excel) | tealeg/xlsx | excelize has 2200+ importers, more active, streaming support for large files |
| Raw Allure JSON in Mongo | Store only parsed data in PG | Loses raw data for debugging parser mismatches; Mongo is already set up |

**Installation:**
```bash
cd backend && go get github.com/xuri/excelize/v2@v2.10.1 github.com/johnfercher/maroto/v2@v2.3.3
```

## Architecture Patterns

### Recommended Project Structure

```
backend/internal/tms/
+-- handler.go           # Fiber route handlers (test cases, runs, results, import/export)
+-- service.go           # Business logic, orchestration
+-- repository.go        # Repository interface + input structs
+-- ent_repository.go    # Ent/PG implementation
+-- mongo_repository.go  # MongoDB implementation (Allure JSON storage)
+-- combined_repository.go # Combines Ent + Mongo behind single interface
+-- allure_parser.go     # Allure results JSON parser (defensive)
+-- allure_matcher.go    # Test case matching strategies
+-- allure_types.go      # Go structs for Allure JSON format
+-- importer.go          # CSV/JSON/XML/Excel import with field mapping
+-- exporter.go          # CSV/JSON/XML/PDF export
+-- dto.go               # Request/response types
+-- events.go            # NATS event definitions

backend/ent/schema/
+-- testcase.go          # Test case entity
+-- testcaseversion.go   # Versioning snapshots
+-- testsuite.go         # Folders/sections (tree structure)
+-- testrun.go           # Test run entity
+-- testresult.go        # Individual test result in a run
+-- testcaseissuelink.go # Bidirectional test case <-> issue link
+-- component.go         # Test components (M2M with test cases)
```

### Pattern 1: CombinedRepository (PG + MongoDB)

**What:** Single Repository interface backed by two storage implementations, exactly mirroring the messenger module pattern.

**When to use:** TMS needs PG for relational data (test cases, runs, results with edges to issues/users) and MongoDB for raw Allure JSON.

**Example:**
```go
// combined_repository.go
// CombinedRepository объединяет Ent и MongoDB репозитории.
type CombinedRepository struct {
    ent   *EntRepository
    mongo *MongoRepository
}

func NewCombinedRepository(ent *EntRepository, mongo *MongoRepository) *CombinedRepository {
    return &CombinedRepository{ent: ent, mongo: mongo}
}

// Методы PG делегируются в EntRepository
func (r *CombinedRepository) CreateTestCase(ctx context.Context, input CreateTestCaseInput) (*ent.TestCase, error) {
    return r.ent.CreateTestCase(ctx, input)
}

// Методы MongoDB делегируются в MongoRepository
func (r *CombinedRepository) StoreAllureResults(ctx context.Context, runID uuid.UUID, results []AllureResult) error {
    return r.mongo.StoreAllureResults(ctx, runID, results)
}
```

### Pattern 2: Allure Parser with Multi-Strategy Matching

**What:** Defensive parser that handles Allure JSON variations across frameworks with priority-ordered test case matching.

**When to use:** ALLU-01, ALLU-02 -- ingesting CI test results and linking them to TMS test cases.

**Example:**
```go
// allure_types.go
// AllureResult представляет один результат теста из allure-results JSON.
type AllureResult struct {
    UUID          string              `json:"uuid"`
    HistoryID     string              `json:"historyId,omitempty"`
    TestCaseID    string              `json:"testCaseId,omitempty"`
    FullName      string              `json:"fullName,omitempty"`
    Name          string              `json:"name,omitempty"`
    Status        string              `json:"status,omitempty"`
    StatusDetails *AllureStatusDetails `json:"statusDetails,omitempty"`
    Stage         string              `json:"stage,omitempty"`
    Start         *int64              `json:"start,omitempty"`
    Stop          *int64              `json:"stop,omitempty"`
    Steps         []AllureStep        `json:"steps,omitempty"`
    Labels        []AllureLabel       `json:"labels,omitempty"`
    Links         []AllureLink        `json:"links,omitempty"`
    Parameters    []AllureParameter   `json:"parameters,omitempty"`
    Attachments   []AllureAttachment  `json:"attachments,omitempty"`
    Description   string              `json:"description,omitempty"`
}

// AllureLabel представляет метку теста.
type AllureLabel struct {
    Name  string `json:"name"`
    Value string `json:"value"`
}

// allure_matcher.go
// MatchStrategy определяет стратегию сопоставления результата с тест-кейсом.
type MatchStrategy int

const (
    MatchByAnnotation MatchStrategy = iota // @TestCase("HUB-TC-42") -> labels[name="tms", value="HUB-TC-42"]
    MatchByFullName                        // fullName -> test_case.full_name
    MatchByClassName                       // suite+name -> test_case.class_name + method_name
)

// MatchTestCase пытается найти тест-кейс по результату Allure, используя стратегии в порядке приоритета.
func (m *Matcher) MatchTestCase(ctx context.Context, result *AllureResult) (*ent.TestCase, MatchStrategy, error) {
    // 1. Ищем по аннотации @TestCase (labels с type=tms)
    for _, link := range result.Links {
        if link.Type == "tms" && link.Name != "" {
            tc, err := m.repo.FindTestCaseByExternalID(ctx, link.Name)
            if err == nil && tc != nil {
                return tc, MatchByAnnotation, nil
            }
        }
    }
    // Также проверяем labels
    for _, label := range result.Labels {
        if label.Name == "AS_ID" || label.Name == "tms" {
            tc, err := m.repo.FindTestCaseByExternalID(ctx, label.Value)
            if err == nil && tc != nil {
                return tc, MatchByAnnotation, nil
            }
        }
    }

    // 2. Ищем по fullName
    if result.FullName != "" {
        tc, err := m.repo.FindTestCaseByFullName(ctx, result.FullName)
        if err == nil && tc != nil {
            return tc, MatchByFullName, nil
        }
    }

    // 3. Ищем по suite + name
    suite := extractLabelValue(result.Labels, "suite")
    if suite != "" && result.Name != "" {
        tc, err := m.repo.FindTestCaseByClassMethod(ctx, suite, result.Name)
        if err == nil && tc != nil {
            return tc, MatchByClassName, nil
        }
    }

    return nil, -1, nil // Не найдено — результат будет "unmapped"
}
```

### Pattern 3: Test Case Versioning via JSONB Snapshots

**What:** Each test case edit creates a version record with full JSONB snapshot of the test case state, enabling rollback.

**When to use:** TMS-02 -- version history with rollback capability.

**Example:**
```go
// Ent schema: TestCaseVersion
// При каждом обновлении тест-кейса сервис создаёт snapshot:
func (s *TMSService) UpdateTestCase(ctx context.Context, id uuid.UUID, userID uuid.UUID, req UpdateTestCaseRequest) (*TestCaseResponse, error) {
    // Получаем текущую версию для snapshot
    current, err := s.repo.GetTestCase(ctx, id)
    if err != nil {
        return nil, err
    }

    // Создаём snapshot текущей версии перед обновлением
    snapshot := map[string]interface{}{
        "title":       current.Title,
        "steps":       current.Steps,
        "precondition": current.Precondition,
        "priority":    current.Priority,
        // ... все поля
    }
    _, err = s.repo.CreateTestCaseVersion(ctx, id, userID, current.Version, snapshot)
    if err != nil {
        slog.Error("не удалось создать версию тест-кейса", "error", err)
    }

    // Обновляем тест-кейс
    return s.repo.UpdateTestCase(ctx, id, input)
}
```

### Pattern 4: Test Run State Machine

**What:** Test runs follow a strict status flow: created -> in_progress -> completed/aborted. Re-open transitions completed -> in_progress.

**When to use:** TMS-04 -- manual, automatic, and mixed test runs.

```
created -> in_progress -> completed
                       -> aborted
completed -> in_progress (re-open)
```

### Anti-Patterns to Avoid

- **Storing steps as separate rows:** Steps are tightly coupled to the test case and versioned together. Use JSONB array, not a steps table. Separate rows create orphan cleanup and ordering problems.
- **Single matching strategy for Allure:** Never rely solely on `@TestCase` annotations -- teams forget them. Always fall back to fullName and class+method.
- **Parsing Allure JSON with strict struct mapping:** Fields are genuinely optional across framework adapters. Use `json:",omitempty"` and nil pointers for every non-trivial field.
- **Synchronous cross-module calls:** Use NATS events for tracker coverage updates and messenger notifications, never direct imports from tms to tracker/messenger packages.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Excel parsing | Custom XLSX parser | excelize v2 | OOXML format is complex (ZIP with XML inside); excelize handles 2222+ edge cases |
| PDF generation | Raw PDF byte construction | maroto v2 | PDF spec is enormous; maroto provides Bootstrap-like layout API |
| CSV parsing | Custom delimiter handling | encoding/csv | RFC 4180 compliance, quoted fields, multiline values -- all handled |
| XML parsing | Custom XML parser | encoding/xml | Namespaces, entity escaping, CDATA -- standard library handles it |
| UUID generation | Custom ID scheme | github.com/google/uuid | Already used project-wide |
| File uploads for screenshots | Custom multipart handling | Existing storage module | Presigned URL pattern already works for file attachments |

**Key insight:** Import/export formats have deceptive complexity (Excel = ZIP of XML, CSV = RFC 4180 with quoting rules, PDF = binary format spec). Standard libraries exist for every format this phase needs.

## Common Pitfalls

### Pitfall 1: Allure Parser Brittleness

**What goes wrong:** Parser breaks on results from different framework adapters (JUnit 5 vs pytest vs JS). Fields that are present in Java adapter output are absent in pytest output.
**Why it happens:** Allure results format is a de facto standard, not a strict specification. Each adapter has quirks.
**How to avoid:**
- Every field in Go structs uses pointer types or omitempty
- Parse defensively: `json.Unmarshal` into `AllureResult` struct where all fields are optional
- Log warnings for unexpected shapes, never panic
- Unit test parser against sample files from at least 3 frameworks (JUnit 5, pytest, Jest/Mocha)
- Store raw JSON in MongoDB BEFORE parsing so failed parses don't lose data
**Warning signs:** "Unknown" test results appearing in runs; different CI jobs producing inconsistent results.

### Pitfall 2: Cross-Module Linking Complexity

**What goes wrong:** Test case <-> issue bidirectional edges create circular dependencies between TMS and tracker modules.
**Why it happens:** Ent schemas are centralized (both TestCase and Issue in ent/schema/), but service logic must not import across module boundaries.
**How to avoid:**
- Define the Ent edge in TestCase schema pointing to Issue (Ent handles bidirectional edges in the schema layer)
- TMS service creates/reads links via its own repository methods
- When a bug is created from failed test, TMS publishes a NATS event; a listener in TMS module calls tracker service indirectly (or TMS repo creates the issue directly since Ent client is shared)
- Coverage tracking: TMS publishes `tms.coverage.updated` event; dashboard module consumes it
**Warning signs:** Import cycles between internal/tms and internal/tracker.

### Pitfall 3: Folder Tree Performance

**What goes wrong:** Recursive queries for folder tree with test case counts become slow as tree depth grows.
**Why it happens:** Naive approach queries each level separately (N+1 for tree depth).
**How to avoid:**
- Use materialized path or adjacency list with CTE (Common Table Expression) for tree queries
- Ent supports recursive CTE via raw SQL for tree operations
- Cache folder tree structure (rarely changes) with invalidation on folder CRUD
- Limit tree depth (10 levels max is practical)
**Warning signs:** Test case list page slowing down with deep folder hierarchies.

### Pitfall 4: Import Field Mapping Complexity

**What goes wrong:** Every team's CSV/Excel has different column names for the same concepts (e.g., "Test Name", "Title", "Название теста").
**Why it happens:** No universal standard for test case data exchange formats.
**How to avoid:**
- Implement universal field mapping: user selects which source column maps to which TMS field
- Frontend sends mapping config with the import request: `{ "column_0": "title", "column_1": "steps", ... }`
- Provide sensible auto-detection (case-insensitive match on known field names)
- Validate mapped data before committing (dry-run mode)
**Warning signs:** Users unable to import their existing test suites; confusing error messages.

### Pitfall 5: Test Run Result Aggregation Performance

**What goes wrong:** Computing statistics (pass/fail/skip counts, flaky detection, trends) via live queries is slow.
**Why it happens:** Each query aggregates across all results in a run, and trend queries span many runs.
**How to avoid:**
- Store pre-computed statistics on the TestRun entity: total, passed, failed, skipped, broken counts
- Update counters when results are added/modified
- Flaky detection: compare last N results for same test case, mark as flaky if status alternates
- Trends: store daily snapshots or compute from run-level aggregates, not individual results
**Warning signs:** Dashboard "test trends" widget taking 5+ seconds to load.

## Code Examples

### Ent Schema: TestCase

```go
// backend/ent/schema/testcase.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

// TestCase хранит определение тест-кейса.
type TestCase struct {
    ent.Schema
}

func (TestCase) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Int("case_number").Positive().Immutable(),
        field.String("case_key").Unique().Immutable().
            Comment("Ключ: HUB-TC-42"),
        field.String("title").NotEmpty().MaxLen(500),
        field.Text("description").Optional(),
        field.Text("precondition").Optional(),
        field.Text("postcondition").Optional(),
        field.Enum("case_type").Values("classic", "bdd").Default("classic"),
        field.Enum("priority").Values("critical", "high", "medium", "low", "trivial").Default("medium"),
        field.Enum("status").Values("draft", "active", "deprecated").Default("draft"),
        field.JSON("steps", []TestStep{}).Optional().
            Comment("JSONB массив шагов"),
        field.String("full_name").Optional().
            Comment("Allure fullName для матчинга"),
        field.String("class_name").Optional().
            Comment("Класс теста для матчинга"),
        field.String("method_name").Optional().
            Comment("Метод теста для матчинга"),
        field.String("external_id").Optional().
            Comment("ID из аннотации @TestCase для матчинга"),
        field.Int("version").Default(1),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (TestCase) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("test_cases").Unique().Required(),
        edge.From("suite", TestSuite.Type).Ref("test_cases").Unique(),
        edge.From("created_by", User.Type).Ref("created_test_cases").Unique(),
        edge.To("versions", TestCaseVersion.Type),
        edge.To("results", TestResult.Type),
        edge.To("labels", Label.Type),           // Переиспользуем метки из трекера
        edge.To("linked_issues", Issue.Type),     // Двусторонняя связь с задачами
        edge.To("components", Component.Type),
    }
}

func (TestCase) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("case_key"),
        index.Fields("external_id"),
        index.Fields("full_name"),
        index.Fields("class_name", "method_name"),
        index.Fields("status"),
        index.Fields("priority"),
    }
}

// TestStep описывает один шаг тест-кейса (хранится в JSONB).
type TestStep struct {
    Order          int    `json:"order"`
    Action         string `json:"action"`
    ExpectedResult string `json:"expected_result"`
    // Для BDD:
    Keyword        string `json:"keyword,omitempty"` // Given, When, Then, And
}
```

### Ent Schema: TestRun

```go
// backend/ent/schema/testrun.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

// TestRun хранит прогон тестов.
type TestRun struct {
    ent.Schema
}

func (TestRun) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.String("title").NotEmpty().MaxLen(500),
        field.Text("description").Optional(),
        field.Enum("run_type").Values("manual", "automatic", "mixed").Default("manual"),
        field.Enum("status").Values("created", "in_progress", "completed", "aborted").Default("created"),
        // Предвычисленная статистика
        field.Int("total_count").Default(0),
        field.Int("passed_count").Default(0),
        field.Int("failed_count").Default(0),
        field.Int("skipped_count").Default(0),
        field.Int("broken_count").Default(0),
        field.Int("unknown_count").Default(0),
        // Allure-специфичные поля
        field.String("allure_run_id").Optional().
            Comment("ID прогона из Allure для дедупликации"),
        field.Time("started_at").Optional().Nillable(),
        field.Time("completed_at").Optional().Nillable(),
        field.Time("created_at").Default(time.Now).Immutable(),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (TestRun) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("test_runs").Unique().Required(),
        edge.From("created_by", User.Type).Ref("created_runs").Unique(),
        edge.To("results", TestResult.Type),
        edge.To("assignees", User.Type), // Назначенные тестировщики (M2M)
    }
}

func (TestRun) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("status"),
        index.Fields("run_type"),
        index.Fields("allure_run_id"),
        index.Fields("created_at"),
    }
}
```

### Ent Schema: TestResult

```go
// backend/ent/schema/testresult.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
    "time"
)

// TestResult хранит результат выполнения одного тест-кейса в прогоне.
type TestResult struct {
    ent.Schema
}

func (TestResult) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
        field.Enum("status").Values("passed", "failed", "broken", "skipped", "unknown").Default("unknown"),
        field.Text("comment").Optional(),
        field.JSON("step_results", []StepResult{}).Optional().
            Comment("Результаты по каждому шагу"),
        field.Int64("duration_ms").Optional().
            Comment("Длительность выполнения"),
        // Allure-специфичные поля
        field.String("allure_uuid").Optional().
            Comment("UUID из allure-results"),
        field.Text("error_message").Optional(),
        field.Text("error_trace").Optional(),
        field.Bool("is_flaky").Default(false),
        field.Enum("match_strategy").Values("annotation", "fullname", "classmethod", "unmapped").
            Default("unmapped").
            Comment("Как был сопоставлен с тест-кейсом"),
        field.Time("started_at").Optional().Nillable(),
        field.Time("completed_at").Optional().Nillable(),
        field.Time("created_at").Default(time.Now).Immutable(),
    }
}

func (TestResult) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("run", TestRun.Type).Ref("results").Unique().Required(),
        edge.From("test_case", TestCase.Type).Ref("results").Unique(),
        edge.From("executed_by", User.Type).Ref("executed_results").Unique(),
        edge.To("files", FileMetadata.Type), // Скриншоты, вложения
    }
}

func (TestResult) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("status"),
        index.Fields("allure_uuid"),
        index.Fields("is_flaky"),
    }
}

// StepResult хранит результат одного шага (в JSONB).
type StepResult struct {
    StepOrder int    `json:"step_order"`
    Status    string `json:"status"` // passed, failed, skipped
    Comment   string `json:"comment,omitempty"`
    // FileIDs ссылаются на FileMetadata (скриншоты)
    FileIDs   []string `json:"file_ids,omitempty"`
}
```

### MongoDB: Allure Raw Storage

```go
// mongo_repository.go
// StoreAllureResults сохраняет сырой JSON из allure-results в MongoDB.
// Каждый результат хранится как отдельный документ для быстрого поиска.
func (r *MongoRepository) StoreAllureResults(ctx context.Context, runID uuid.UUID, results []AllureResult) error {
    coll := r.db.Collection("allure_results")
    docs := make([]interface{}, 0, len(results))
    for _, result := range results {
        docs = append(docs, bson.M{
            "run_id":     runID.String(),
            "allure_uuid": result.UUID,
            "raw":        result, // Сохраняем как есть
            "created_at": time.Now(),
        })
    }
    if len(docs) == 0 {
        return nil
    }
    _, err := coll.InsertMany(ctx, docs)
    return err
}

// EnsureIndexes создаёт индексы для коллекции allure_results.
func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
    coll := r.db.Collection("allure_results")
    _, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "run_id", Value: 1}, {Key: "created_at", Value: -1}}},
        {Keys: bson.D{{Key: "allure_uuid", Value: 1}}},
    })
    return err
}
```

### NATS Events for Cross-Module Communication

```go
// events.go
const (
    EventTestRunCompleted = "tms.run.completed"
    EventTestRunCreated   = "tms.run.created"
    EventCoverageUpdated  = "tms.coverage.updated"
    EventTestCaseLinked   = "tms.case.linked"
)

// TestRunCompletedEvent публикуется при завершении прогона.
// Подписчики: notification (Telegram), messenger (пост в канал), dashboard.
type TestRunCompletedEvent struct {
    RunID       string `json:"run_id"`
    ProjectID   string `json:"project_id"`
    Title       string `json:"title"`
    RunType     string `json:"run_type"`
    Total       int    `json:"total"`
    Passed      int    `json:"passed"`
    Failed      int    `json:"failed"`
    Skipped     int    `json:"skipped"`
    Broken      int    `json:"broken"`
    CompletedBy string `json:"completed_by"`
}
```

### Route Registration (Following Tracker Pattern)

```go
// handler.go
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
    tms := api.Group("/projects/:projectId/tms", authMw)

    // Тест-кейсы
    cases := tms.Group("/cases")
    cases.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateTestCase)
    cases.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestCases)
    cases.Get("/:caseKey", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestCase)
    cases.Put("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateTestCase)
    cases.Delete("/:caseKey", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteTestCase)
    cases.Get("/:caseKey/versions", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestCaseVersions)
    cases.Post("/:caseKey/rollback/:versionId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.RollbackTestCase)
    cases.Post("/:caseKey/links", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.LinkTestCaseToIssue)
    cases.Delete("/:caseKey/links/:linkId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UnlinkTestCaseFromIssue)

    // Папки/разделы
    suites := tms.Group("/suites")
    suites.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateSuite)
    suites.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListSuites)
    suites.Put("/:suiteId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.UpdateSuite)
    suites.Delete("/:suiteId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteSuite)

    // Компоненты
    components := tms.Group("/components")
    components.Post("/", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.CreateComponent)
    components.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListComponents)
    components.Delete("/:componentId", rbac.RequirePermission(rbac.PermTestCaseEdit, rbacSvc), h.DeleteComponent)

    // Прогоны
    runs := tms.Group("/runs")
    runs.Post("/", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.CreateTestRun)
    runs.Get("/", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestRuns)
    runs.Get("/:runId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestRun)
    runs.Put("/:runId/status", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.UpdateRunStatus)
    runs.Post("/:runId/reopen", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.ReopenTestRun)

    // Результаты внутри прогона
    runs.Post("/:runId/results", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.AddTestResult)
    runs.Put("/:runId/results/:resultId", rbac.RequirePermission(rbac.PermTestRunExecute, rbacSvc), h.UpdateTestResult)
    runs.Get("/:runId/results", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListTestResults)

    // Allure интеграция
    runs.Post("/:runId/allure", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.ImportAllureResults)
    // Создание прогона + импорт в одном запросе (для CI)
    tms.Post("/allure/import", rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc), h.CreateRunFromAllure)

    // Импорт/экспорт
    tms.Post("/import", rbac.RequirePermission(rbac.PermTestCaseCreate, rbacSvc), h.ImportTestCases)
    tms.Get("/export", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ExportTestCases)
    runs.Get("/:runId/export", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ExportTestRunReport)

    // Отчёты
    tms.Get("/reports/trends", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetTestTrends)
    tms.Get("/reports/flaky", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetFlakyTests)
    tms.Get("/reports/coverage", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetCoverageReport)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| gofpdf (jung-kurt) | go-pdf/fpdf -> maroto v2 (wraps phpdave11/gofpdf) | gofpdf archived; go-pdf/fpdf archived Mar 2025 | Use maroto v2.3.3 which internally switched to phpdave11 |
| Store all test data in PG only | Hybrid: PG for structured + MongoDB for raw Allure JSON | Project architecture decision | Preserves raw data, enables re-parsing if parser improves |
| Single Allure matching strategy | Multi-strategy with priority fallback | Industry learning from Allure TestOps | Handles real-world annotation gaps |

**Deprecated/outdated:**
- `gofpdf` (jung-kurt): archived, replaced by `go-pdf/fpdf` which is also archived; use `maroto v2`
- `go-pdf/fpdf`: archived March 2025; maroto v2.3.3 switched to `phpdave11/gofpdf` internally

## Open Questions

1. **Test case key format**
   - What we know: Issues use `HUB-1` (project key + counter). Test cases need a similar scheme.
   - Recommendation: Use `{PROJECT_KEY}-TC-{counter}` format (e.g., `HUB-TC-42`). Requires a `test_case_counter` field on Project schema (similar to existing `task_counter`).

2. **Bug creation from failed test: which module owns it?**
   - What we know: Creating a bug is a tracker operation. TMS needs to trigger it.
   - Recommendation: TMS service calls tracker repository directly (shared Ent client) to create the issue, then creates the bidirectional link. Publish NATS event `tms.bug.created` for notifications. This avoids import cycles since both modules use the same Ent client.

3. **Allure attachments storage**
   - What we know: Allure results include attachment file references. CI would need to upload these alongside JSON.
   - Recommendation: CI uploads attachments to SeaweedFS via presigned URLs (existing storage module), includes file IDs in the Allure import request. Parser links attachments to test results.

4. **Traceability matrix complexity**
   - What we know: TMS-06 requires requirement coverage tracking. No formal "requirements" entity exists -- issues serve as requirements.
   - Recommendation: Traceability = TestCase <-> Issue links. Matrix view is a frontend query aggregation: for each linked issue, show which test cases cover it and their last result status. No separate entity needed.

## Sources

### Primary (HIGH confidence)
- [Allure Report - Test result file](https://allurereport.org/docs/how-it-works-test-result-file/) - complete JSON schema for allure-results
- [Allure Report - Container file](https://allurereport.org/docs/how-it-works-container-file/) - fixture/container format
- [excelize - Go XLSX library](https://pkg.go.dev/github.com/xuri/excelize/v2) - v2.10.1, 2222+ importers
- [maroto v2 - PDF generation](https://github.com/johnfercher/maroto) - v2.3.3, Bootstrap-like layout
- [Go encoding/csv](https://pkg.go.dev/encoding/csv) - RFC 4180 compliant
- [Go encoding/xml](https://pkg.go.dev/encoding/xml) - standard library XML
- Existing codebase: tracker handler.go, service.go, repository.go, issue.go schema patterns

### Secondary (MEDIUM confidence)
- [Allure TestOps labels/metadata](https://docs.qameta.io/allure-testops/briefly/test-cases/labels/) - label conventions across frameworks
- [go-pdf/fpdf archived notice](https://github.com/go-pdf/fpdf) - confirmed archived March 2025

### Tertiary (LOW confidence)
- Allure format stability across framework adapters - de facto standard, not strict spec; variations confirmed but exact differences per adapter need testing with real CI output

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries verified, versions confirmed against Go module registry and GitHub
- Architecture: HIGH - follows established patterns from tracker and messenger modules already built
- Pitfalls: MEDIUM-HIGH - Allure parser brittleness confirmed by research and project pitfalls doc; import/export complexity well-documented
- Allure JSON format: HIGH - official docs provide complete schema; matching strategies derived from TestOps documentation

**Research date:** 2026-03-21
**Valid until:** 2026-04-21 (stable domain, libraries mature)
