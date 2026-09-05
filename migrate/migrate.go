// Package migrate provides a Laravel-style versioned migration runner on top
// of GORM. Migration files self-register via init() (see GenerateFile); the
// runner then applies pending ones in order and records them in the
// "schema_migrations" table.
//
// The tracking table itself is NOT created here — it must already exist
// (define a SchemaMigration model in the consuming module and include it in
// the slice passed to AutoMigrate, which per convention runs before Run).
package migrate

import (
	"sort"

	"gorm.io/gorm"
)

// Migration is one versioned schema change, self-registered by a generated
// file's init(). Name must be unique within a Schema and sortable
// chronologically (see GenerateFile for the timestamp-prefixed format).
type Migration struct {
	Schema string
	Name   string
	Up     func(db *gorm.DB) error
	// Down reverts Up. Leave nil only for migrations that genuinely cannot be
	// reverted — Rollback then fails loudly instead of silently skipping.
	Down func(db *gorm.DB) error
}

var registry = map[string][]Migration{}

// Register adds a migration to the in-process registry. Called from a
// migration file's init().
func Register(m Migration) {
	registry[m.Schema] = append(registry[m.Schema], m)
}

// Registered returns all migrations for schema, sorted by Name ascending.
func Registered(schema string) []Migration {
	ms := append([]Migration(nil), registry[schema]...)
	sort.Slice(ms, func(i, j int) bool { return ms[i].Name < ms[j].Name })
	return ms
}
