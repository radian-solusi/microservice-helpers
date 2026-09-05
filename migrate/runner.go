package migrate

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Run applies every registered migration for schema not yet recorded in the
// "schema_migrations" table, each inside its own transaction. Call after
// AutoMigrate (which must already have created "schema_migrations"). Safe to
// call repeatedly — already-applied migrations are skipped.
func Run(db *gorm.DB, schema string) (applied []string, err error) {
	done, err := appliedNames(db)
	if err != nil {
		return nil, err
	}

	var batch int
	if err := db.Table("schema_migrations").Select("COALESCE(MAX(batch), 0)").Scan(&batch).Error; err != nil {
		return nil, fmt.Errorf("read batch: %w", err)
	}
	batch++

	for _, m := range Registered(schema) {
		if done[m.Name] {
			continue
		}
		txErr := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Up(tx); err != nil {
				return err
			}
			return tx.Table("schema_migrations").Create(map[string]any{
				"migration":  m.Name,
				"batch":      batch,
				"executed_at": time.Now(),
			}).Error
		})
		if txErr != nil {
			return applied, fmt.Errorf("migration %q: %w", m.Name, txErr)
		}
		applied = append(applied, m.Name)
	}
	return applied, nil
}

// Rollback reverts the last `batches` batches of applied migrations for schema,
// newest first, each inside its own transaction (Down runs, then the
// schema_migrations row is deleted). batches < 1 is treated as 1.
//
// It refuses to start if any migration in range has no Down or is missing from
// the registry, so a partial rollback never leaves a half-reverted batch.
func Rollback(db *gorm.DB, schema string, batches int) (reverted []string, err error) {
	if batches < 1 {
		batches = 1
	}

	var maxBatch int
	if err := db.Table("schema_migrations").Select("COALESCE(MAX(batch), 0)").Scan(&maxBatch).Error; err != nil {
		return nil, fmt.Errorf("read batch: %w", err)
	}
	if maxBatch == 0 {
		return nil, nil
	}
	minBatch := maxBatch - batches + 1

	var names []string
	if err := db.Table("schema_migrations").
		Where("batch >= ?", minBatch).
		Order("batch DESC, migration DESC").
		Pluck("migration", &names).Error; err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}

	byName := make(map[string]Migration, len(names))
	for _, m := range Registered(schema) {
		byName[m.Name] = m
	}

	// Pre-flight: every migration in range must be revertible.
	targets := make([]Migration, 0, len(names))
	for _, name := range names {
		m, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("migration %q is recorded but not registered for schema %q", name, schema)
		}
		if m.Down == nil {
			return nil, fmt.Errorf("migration %q has no Down and cannot be rolled back", name)
		}
		targets = append(targets, m)
	}

	for _, m := range targets {
		txErr := db.Transaction(func(tx *gorm.DB) error {
			if err := m.Down(tx); err != nil {
				return err
			}
			return tx.Exec("DELETE FROM schema_migrations WHERE migration = ?", m.Name).Error
		})
		if txErr != nil {
			return reverted, fmt.Errorf("rollback %q: %w", m.Name, txErr)
		}
		reverted = append(reverted, m.Name)
	}
	return reverted, nil
}

// Status is one registered migration and whether it has been applied.
type Status struct {
	Name       string
	Applied    bool
	Revertible bool
}

// Statuses reports every registered migration for schema in order, marking
// which are already applied.
func Statuses(db *gorm.DB, schema string) ([]Status, error) {
	done, err := appliedNames(db)
	if err != nil {
		return nil, err
	}
	all := Registered(schema)
	out := make([]Status, 0, len(all))
	for _, m := range all {
		out = append(out, Status{Name: m.Name, Applied: done[m.Name], Revertible: m.Down != nil})
	}
	return out, nil
}

func appliedNames(db *gorm.DB) (map[string]bool, error) {
	var names []string
	if err := db.Table("schema_migrations").Pluck("migration", &names).Error; err != nil {
		return nil, fmt.Errorf("read schema_migrations: %w", err)
	}
	done := make(map[string]bool, len(names))
	for _, n := range names {
		done[n] = true
	}
	return done, nil
}
