//go:build hubtest

package tms

import "github.com/valerii/hub/backend/ent"

// ExportEntRepositoryClient возвращает ent.Client из EntRepository для hubtest.
func ExportEntRepositoryClient(r *EntRepository) *ent.Client {
	return r.client
}
