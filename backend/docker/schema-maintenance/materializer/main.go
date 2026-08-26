package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	entschema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	_ "github.com/lib/pq"
)

const (
	controlDatabaseURL            = "postgres://schema@schema-maintenance-postgres:5432/hub_schema?sslmode=disable&search_path=public"
	prefeatureTargetDatabaseURL   = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_prefeature?sslmode=disable&search_path=hub_mail_calendar"
	currentTargetDatabaseURL      = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_current?sslmode=disable&search_path=hub_mail_calendar"
	verificationTargetDatabaseURL = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_verification?sslmode=disable&search_path=hub_mail_calendar"
	atlasDiffDevDatabaseURL       = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_diff_dev?sslmode=disable&search_path=hub_mail_calendar"
	atlasValidateDevDatabaseURL   = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_validate_dev?sslmode=disable&search_path=hub_mail_calendar"
	atlasLintDevDatabaseURL       = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_lint_dev?sslmode=disable&search_path=hub_mail_calendar"
	atlasCompareDevDatabaseURL    = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_compare_dev?sslmode=disable&search_path=hub_mail_calendar"
	atlasRevisionsDatabaseURL     = "postgres://schema@schema-maintenance-postgres:5432/hub_schema_verification?sslmode=disable&search_path=atlas_revisions"

	readyFilePath          = "/work/mail-calendar-schema/materializer.lock.ready"
	materializerAdvisoryID int64 = 1296253773

	databaseTimeout  = 20 * time.Second
	migrationTimeout = 2 * time.Minute
	pingInterval     = 5 * time.Second
)

type target struct {
	name         string
	databaseName string
	schemaName   string
	databaseURL  string
	moduleRoot   string
	schemaRoot   string
}

type configuration struct {
	controlURL string
	targets    []target
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "schema materializer:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	config, err := loadConfiguration()
	if err != nil {
		return err
	}

	switch {
	case len(args) == 1 && args[0] == "lock":
		return runLock(config.controlURL)
	case len(args) == 1 && args[0] == "initialize":
		return initializeTargets(config.controlURL, config.targets)
	case len(args) == 2 && args[0] == "materialize":
		if args[1] != "prefeature" && args[1] != "current" {
			return usageError()
		}
		selected, ok := config.target(args[1])
		if !ok {
			return usageError()
		}
		return materializeTarget(selected)
	default:
		return usageError()
	}
}

func usageError() error {
	return errors.New("usage: materializer lock | initialize | materialize <prefeature|current>")
}

func loadConfiguration() (configuration, error) {
	expected := []struct {
		environmentName string
		url             string
		databaseName    string
		searchPath      string
	}{
		{"DATABASE_URL", controlDatabaseURL, "hub_schema", "public"},
		{"PREFEATURE_SCHEMA_URL", prefeatureTargetDatabaseURL, "hub_schema_prefeature", "hub_mail_calendar"},
		{"CURRENT_SCHEMA_URL", currentTargetDatabaseURL, "hub_schema_current", "hub_mail_calendar"},
		{"VERIFICATION_SCHEMA_URL", verificationTargetDatabaseURL, "hub_schema_verification", "hub_mail_calendar"},
		{"ATLAS_DIFF_DEV_URL", atlasDiffDevDatabaseURL, "hub_schema_atlas_diff_dev", "hub_mail_calendar"},
		{"ATLAS_VALIDATE_DEV_URL", atlasValidateDevDatabaseURL, "hub_schema_atlas_validate_dev", "hub_mail_calendar"},
		{"ATLAS_LINT_DEV_URL", atlasLintDevDatabaseURL, "hub_schema_atlas_lint_dev", "hub_mail_calendar"},
		{"ATLAS_COMPARE_DEV_URL", atlasCompareDevDatabaseURL, "hub_schema_atlas_compare_dev", "hub_mail_calendar"},
		{"ATLAS_REVISIONS_URL", atlasRevisionsDatabaseURL, "hub_schema_verification", "atlas_revisions"},
	}
	for _, constraint := range expected {
		value, ok := os.LookupEnv(constraint.environmentName)
		if !ok {
			return configuration{}, fmt.Errorf("%s is required", constraint.environmentName)
		}
		if value != constraint.url {
			return configuration{}, fmt.Errorf("%s must match the fixed schema-maintenance URL", constraint.environmentName)
		}
		if err := validateDatabaseURL(
			constraint.environmentName,
			value,
			constraint.databaseName,
			constraint.searchPath,
		); err != nil {
			return configuration{}, err
		}
	}

	targets := []target{
		{
			name:         "prefeature",
			databaseName: "hub_schema_prefeature",
			schemaName:   "hub_mail_calendar",
			databaseURL:  prefeatureTargetDatabaseURL,
			moduleRoot:   "/opt/hub/prefeature/backend",
			schemaRoot:   "./ent/schema",
		},
		{
			name:         "current",
			databaseName: "hub_schema_current",
			schemaName:   "hub_mail_calendar",
			databaseURL:  currentTargetDatabaseURL,
			moduleRoot:   "/opt/hub/backend",
			schemaRoot:   "./ent/schema",
		},
		{
			name:         "verification",
			databaseName: "hub_schema_verification",
			schemaName:   "hub_mail_calendar",
			databaseURL:  verificationTargetDatabaseURL,
		},
		{
			name:         "atlas_diff_dev",
			databaseName: "hub_schema_atlas_diff_dev",
			schemaName:   "hub_mail_calendar",
			databaseURL:  atlasDiffDevDatabaseURL,
		},
		{
			name:         "atlas_validate_dev",
			databaseName: "hub_schema_atlas_validate_dev",
			schemaName:   "hub_mail_calendar",
			databaseURL:  atlasValidateDevDatabaseURL,
		},
		{
			name:         "atlas_lint_dev",
			databaseName: "hub_schema_atlas_lint_dev",
			schemaName:   "hub_mail_calendar",
			databaseURL:  atlasLintDevDatabaseURL,
		},
		{
			name:         "atlas_compare_dev",
			databaseName: "hub_schema_atlas_compare_dev",
			schemaName:   "hub_mail_calendar",
			databaseURL:  atlasCompareDevDatabaseURL,
		},
	}
	if len(targets) != 7 {
		return configuration{}, errors.New("schema-maintenance requires exactly seven disposable targets")
	}
	seenDatabases := make(map[string]struct{}, len(targets))
	for index, target := range targets {
		if target.databaseName == "" || target.databaseName == "hub_schema" {
			return configuration{}, errors.New("target database must be a fixed non-control database")
		}
		if target.schemaName != "hub_mail_calendar" {
			return configuration{}, errors.New("target schema must be hub_mail_calendar")
		}
		if _, exists := seenDatabases[target.databaseName]; exists {
			return configuration{}, errors.New("target database names must be distinct")
		}
		if index < 2 && (target.moduleRoot == "" || target.schemaRoot == "") {
			return configuration{}, errors.New("materializable target is missing its fixed module or Ent schema root")
		}
		if index >= 2 && (target.moduleRoot != "" || target.schemaRoot != "") {
			return configuration{}, errors.New("non-materializable target must not have a module or Ent schema root")
		}
		seenDatabases[target.databaseName] = struct{}{}
	}

	return configuration{
		controlURL: controlDatabaseURL,
		targets:    targets,
	}, nil
}

func (config configuration) target(name string) (target, bool) {
	for _, candidate := range config.targets {
		if candidate.name == name {
			return candidate, true
		}
	}
	return target{}, false
}

func validateDatabaseURL(environmentName, value, expectedDatabaseName, expectedSearchPath string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("%s is not a valid PostgreSQL URL", environmentName)
	}
	if parsed.Scheme != "postgres" ||
		parsed.User == nil ||
		parsed.User.Username() != "schema" ||
		parsed.Host != "schema-maintenance-postgres:5432" ||
		parsed.Path != "/"+expectedDatabaseName ||
		parsed.RawPath != "" ||
		parsed.Fragment != "" ||
		parsed.Opaque != "" {
		return fmt.Errorf("%s must use the fixed schema-maintenance PostgreSQL endpoint", environmentName)
	}
	if _, hasPassword := parsed.User.Password(); hasPassword {
		return fmt.Errorf("%s must not contain database credentials beyond the fixed user", environmentName)
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil ||
		len(query) != 2 ||
		len(query["sslmode"]) != 1 ||
		len(query["search_path"]) != 1 ||
		query.Get("sslmode") != "disable" ||
		query.Get("search_path") != expectedSearchPath {
		return fmt.Errorf("%s must use the fixed schema-maintenance query parameters", environmentName)
	}
	return nil
}

func runLock(controlURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	db, err := openControlDatabase(ctx, controlURL)
	if err != nil {
		return err
	}
	defer db.Close()

	connection, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open control database connection: %w", err)
	}
	defer connection.Close()

	if _, err := connection.ExecContext(ctx, "SELECT pg_advisory_lock($1)", materializerAdvisoryID); err != nil {
		return fmt.Errorf("acquire schema-maintenance advisory lock: %w", err)
	}
	defer func() {
		unlockContext, unlockCancel := context.WithTimeout(context.Background(), databaseTimeout)
		defer unlockCancel()
		_, _ = connection.ExecContext(unlockContext, "SELECT pg_advisory_unlock($1)", materializerAdvisoryID)
	}()

	if err := createReadyFile(); err != nil {
		return err
	}
	defer os.Remove(readyFilePath)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-signals:
			return nil
		case <-ticker.C:
			pingContext, pingCancel := context.WithTimeout(context.Background(), databaseTimeout)
			err := connection.PingContext(pingContext)
			pingCancel()
			if err != nil {
				return fmt.Errorf("ping control database while holding advisory lock: %w", err)
			}
		}
	}
}

func createReadyFile() error {
	file, err := os.OpenFile(readyFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create materializer ready file: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		_ = os.Remove(readyFilePath)
		return fmt.Errorf("set materializer ready file permissions: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(readyFilePath)
		return fmt.Errorf("close materializer ready file: %w", err)
	}
	return nil
}

func initializeTargets(controlURL string, targets []target) error {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	control, err := openControlDatabase(ctx, controlURL)
	if err != nil {
		return err
	}
	defer control.Close()

	for _, target := range targets {
		var count int
		if err := control.QueryRowContext(
			ctx,
			"SELECT count(*) FROM pg_database WHERE datname = $1",
			target.databaseName,
		).Scan(&count); err != nil {
			return fmt.Errorf("check %s target database before initialization: %w", target.name, err)
		}
		if count != 0 {
			return fmt.Errorf("%s target database already exists; initialization refuses to modify it", target.name)
		}
	}
	for _, target := range targets {
		if _, err := control.ExecContext(ctx, "CREATE DATABASE "+quoteIdentifier(target.databaseName)); err != nil {
			return fmt.Errorf("create %s target database: %w", target.name, err)
		}
	}
	for _, target := range targets {
		if err := createTargetSchemas(ctx, target); err != nil {
			return err
		}
	}
	return nil
}

func createTargetSchemas(ctx context.Context, target target) error {
	db, err := openDatabase(ctx, target.databaseURL, target.name)
	if err != nil {
		return err
	}
	defer db.Close()

	var databaseName string
	if err := db.QueryRowContext(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		return fmt.Errorf("verify %s target database identity: %w", target.name, err)
	}
	if databaseName != target.databaseName {
		return fmt.Errorf("%s target database identity does not match its fixed target", target.name)
	}
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA "+quoteIdentifier(target.schemaName)); err != nil {
		return fmt.Errorf("create %s target schema: %w", target.name, err)
	}
	if target.name == "verification" || target.name == "atlas_compare_dev" {
		if _, err := db.ExecContext(ctx, `CREATE SCHEMA "atlas_revisions"`); err != nil {
			return fmt.Errorf("create %s atlas_revisions schema: %w", target.name, err)
		}
	}
	return nil
}

func materializeTarget(target target) error {
	if target.moduleRoot == "" || target.schemaRoot == "" {
		return fmt.Errorf("%s target has no fixed module or Ent schema root", target.name)
	}
	if err := requireEmptyTarget(target); err != nil {
		return err
	}
	if err := os.Chdir(target.moduleRoot); err != nil {
		return fmt.Errorf("change to %s target module root: %w", target.name, err)
	}

	graph, err := entc.LoadGraph(target.schemaRoot, &gen.Config{})
	if err != nil {
		return fmt.Errorf("load %s Ent schema graph: %w", target.name, err)
	}
	tables, err := graph.Tables()
	if err != nil {
		return fmt.Errorf("derive %s Ent schema tables: %w", target.name, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancel()
	db, err := openTargetDatabase(ctx, target)
	if err != nil {
		return err
	}
	defer db.Close()

	driver := entsql.OpenDB(dialect.Postgres, db)
	migrator, err := entschema.NewMigrate(driver)
	if err != nil {
		return fmt.Errorf("construct %s schema migrator: %w", target.name, err)
	}
	if err := migrator.Create(ctx, tables...); err != nil {
		return fmt.Errorf("materialize %s target schema: %w", target.name, err)
	}
	return nil
}

func requireEmptyTarget(target target) error {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	db, err := openTargetDatabase(ctx, target)
	if err != nil {
		return err
	}
	defer db.Close()

	var tableCount int
	if err := db.QueryRowContext(
		ctx,
		"SELECT count(*) FROM information_schema.tables WHERE table_schema = $1 AND table_type = 'BASE TABLE'",
		target.schemaName,
	).Scan(&tableCount); err != nil {
		return fmt.Errorf("check %s target base tables: %w", target.name, err)
	}
	if tableCount != 0 {
		return fmt.Errorf("%s target schema must be empty before materialization", target.name)
	}
	return nil
}

func openControlDatabase(ctx context.Context, controlURL string) (*sql.DB, error) {
	db, err := openDatabase(ctx, controlURL, "control")
	if err != nil {
		return nil, err
	}

	var databaseName, schemaName string
	if err := db.QueryRowContext(ctx, "SELECT current_database(), COALESCE(current_schema(), '')").Scan(&databaseName, &schemaName); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("verify control database identity: %w", err)
	}
	if databaseName != "hub_schema" || schemaName != "public" {
		_ = db.Close()
		return nil, errors.New("control database identity must be hub_schema with public current schema")
	}
	return db, nil
}

func openTargetDatabase(ctx context.Context, target target) (*sql.DB, error) {
	db, err := openDatabase(ctx, target.databaseURL, target.name)
	if err != nil {
		return nil, err
	}

	var databaseName, schemaName string
	if err := db.QueryRowContext(ctx, "SELECT current_database(), COALESCE(current_schema(), '')").Scan(&databaseName, &schemaName); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("verify %s target identity: %w", target.name, err)
	}
	if databaseName != target.databaseName || schemaName != target.schemaName {
		_ = db.Close()
		return nil, fmt.Errorf("%s target identity does not match its fixed database or schema", target.name)
	}
	return db, nil
}

func openDatabase(ctx context.Context, databaseURL, role string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", role, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping %s database: %w", role, err)
	}
	return db, nil
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
