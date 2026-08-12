//go:build hubtest

package database

import dbsql "database/sql"

// ExportApplyPoolLimits wraps applyPoolLimits for hubtest (P1-BE-25).
func ExportApplyPoolLimits(db *dbsql.DB) {
	applyPoolLimits(db)
}

// ExportDefaultMaxOpenConns returns the production MaxOpenConns default.
func ExportDefaultMaxOpenConns() int {
	return defaultMaxOpenConns
}

// ExportEnvBool wraps envBool for hubtest (W4-A4-003 Schema.Create gate).
func ExportEnvBool(key string, fallback bool) bool {
	return envBool(key, fallback)
}
