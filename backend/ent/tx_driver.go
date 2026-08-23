package ent

import "entgo.io/ent/dialect"

// Driver returns the query executor bound to this transaction. It deliberately
// exposes neither transaction lifecycle nor a database handle, so callers can
// issue only parameterized statements on the same transaction as Ent writes.
func (tx *Tx) Driver() dialect.ExecQuerier {
	if tx == nil || tx.config.driver == nil {
		return nil
	}
	return tx.config.driver
}
