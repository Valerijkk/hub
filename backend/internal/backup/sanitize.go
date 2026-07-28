package backup

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// sanitizeBackupErrorMessage — W7-A4-007: в ErrorMessage только безопасные
// категории без URI/host/DSN из обёрток pg_dump/mongodump.
func sanitizeBackupErrorMessage(err error) string {
	if err == nil {
		return "backup failed"
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "pg_dump"):
		return "pg_dump failed"
	case strings.Contains(msg, "mongodump"):
		return "mongodump failed"
	case strings.Contains(msg, "pg_restore"):
		return "pg_restore failed"
	case strings.Contains(msg, "mongorestore"):
		return "mongorestore failed"
	case strings.Contains(msg, "s3"):
		return "s3 upload failed"
	case strings.Contains(msg, "zip"):
		return "zip failed"
	default:
		return "backup failed"
	}
}

// pgDumpArgsEnv — W7-A4-013: пароль только в PGPASSWORD, не в argv URI.
// Возвращает args для pg_dump/pg_restore (--dbname= без userinfo) и env с PGPASSWORD.
func pgDumpArgsEnv(pgURL string, extraArgs ...string) (args []string, env []string, err error) {
	u, err := url.Parse(pgURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse postgres url: %w", err)
	}
	password := ""
	if u.User != nil {
		password, _ = u.User.Password()
		u.User = url.User(u.User.Username())
	}
	safeURL := u.String()
	args = append(extraArgs, "--dbname="+safeURL)
	env = append(os.Environ(), "PGPASSWORD="+password)
	return args, env, nil
}

// writeMongoConfig — W7-A4-013: URI/password в temp YAML для --config, не в argv.
func writeMongoConfig(dir, mongoURL string) (configPath string, cleanup func(), err error) {
	u, err := url.Parse(mongoURL)
	if err != nil {
		return "", nil, fmt.Errorf("parse mongo url: %w", err)
	}
	password := ""
	if u.User != nil {
		password, _ = u.User.Password()
		u.User = url.User(u.User.Username())
	}
	safeURI := u.String()

	f, err := os.CreateTemp(dir, "hub-mongo-*.yaml")
	if err != nil {
		return "", nil, fmt.Errorf("create mongo config: %w", err)
	}
	path := f.Name()
	cleanup = func() { _ = os.Remove(path) }

	var b strings.Builder
	b.WriteString("uri: ")
	b.WriteString(safeURI)
	b.WriteByte('\n')
	if password != "" {
		b.WriteString("password: ")
		b.WriteString(password)
		b.WriteByte('\n')
	}
	if _, err := f.WriteString(b.String()); err != nil {
		_ = f.Close()
		cleanup()
		return "", nil, fmt.Errorf("write mongo config: %w", err)
	}
	if err := f.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("close mongo config: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("chmod mongo config: %w", err)
	}
	return path, cleanup, nil
}

// mongoConfigArg возвращает единственный argv-фрагмент с путём к config (без секрета).
func mongoConfigArg(configPath string) string {
	return "--config=" + filepath.Clean(configPath)
}
