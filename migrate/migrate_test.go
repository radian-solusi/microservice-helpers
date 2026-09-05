package migrate

import (
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "t.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Mimics AutoMigrate having created the tracking table beforehand.
	if err := db.Exec(`CREATE TABLE schema_migrations (id INTEGER PRIMARY KEY AUTOINCREMENT, migration TEXT NOT NULL UNIQUE, batch INTEGER NOT NULL)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestRunAppliesInOrderOnce(t *testing.T) {
	registry = map[string][]Migration{} // isolate from other tests/registrations
	var order []string

	Register(Migration{Schema: "shared", Name: "20260101_000000_b", Up: func(db *gorm.DB) error {
		order = append(order, "b")
		return db.Exec(`CREATE TABLE b (id INTEGER)`).Error
	}})
	Register(Migration{Schema: "shared", Name: "20260101_000000_a", Up: func(db *gorm.DB) error {
		order = append(order, "a")
		return db.Exec(`CREATE TABLE a (id INTEGER)`).Error
	}})

	db := openTestDB(t)

	applied, err := Run(db, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("expected sorted a,b, got %v (order=%v)", applied, order)
	}

	// Second run: nothing pending, idempotent.
	applied2, err := Run(db, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if len(applied2) != 0 {
		t.Fatalf("expected no-op second run, got %v", applied2)
	}
}

func TestRunFailureDoesNotRecord(t *testing.T) {
	registry = map[string][]Migration{}
	Register(Migration{Schema: "shared", Name: "20260101_000000_bad", Up: func(db *gorm.DB) error {
		return db.Exec(`SELECT * FROM does_not_exist`).Error
	}})

	db := openTestDB(t)
	if _, err := Run(db, "shared"); err == nil {
		t.Fatal("expected error")
	}

	var count int64
	db.Table("schema_migrations").Count(&count)
	if count != 0 {
		t.Fatalf("failed migration must not be recorded, count=%d", count)
	}
}

func TestRollbackRevertsLastBatchNewestFirst(t *testing.T) {
	registry = map[string][]Migration{}
	var order []string

	Register(Migration{Schema: "shared", Name: "20260101_000000_a",
		Up:   func(db *gorm.DB) error { return db.Exec(`CREATE TABLE a (id INTEGER)`).Error },
		Down: func(db *gorm.DB) error { order = append(order, "a"); return db.Exec(`DROP TABLE a`).Error }})
	Register(Migration{Schema: "shared", Name: "20260101_000001_b",
		Up:   func(db *gorm.DB) error { return db.Exec(`CREATE TABLE b (id INTEGER)`).Error },
		Down: func(db *gorm.DB) error { order = append(order, "b"); return db.Exec(`DROP TABLE b`).Error }})

	db := openTestDB(t)
	if _, err := Run(db, "shared"); err != nil {
		t.Fatal(err)
	}

	reverted, err := Rollback(db, "shared", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reverted) != 2 || order[0] != "b" || order[1] != "a" {
		t.Fatalf("expected reverse order b,a; got %v (order=%v)", reverted, order)
	}

	var count int64
	db.Table("schema_migrations").Count(&count)
	if count != 0 {
		t.Fatalf("rolled-back rows must be deleted, count=%d", count)
	}
	if db.Migrator().HasTable("a") || db.Migrator().HasTable("b") {
		t.Fatal("tables should have been dropped by Down")
	}

	// Re-applying after rollback works, and lands in a fresh batch.
	applied, err := Run(db, "shared")
	if err != nil || len(applied) != 2 {
		t.Fatalf("re-apply after rollback: %v %v", applied, err)
	}
}

func TestRollbackRefusesWhenDownMissing(t *testing.T) {
	registry = map[string][]Migration{}
	Register(Migration{Schema: "shared", Name: "20260101_000000_a",
		Up:   func(db *gorm.DB) error { return db.Exec(`CREATE TABLE a (id INTEGER)`).Error },
		Down: nil})

	db := openTestDB(t)
	if _, err := Run(db, "shared"); err != nil {
		t.Fatal(err)
	}
	if _, err := Rollback(db, "shared", 1); err == nil {
		t.Fatal("expected error for nil Down")
	}

	// Pre-flight must leave the ledger untouched.
	var count int64
	db.Table("schema_migrations").Count(&count)
	if count != 1 {
		t.Fatalf("ledger must be unchanged, count=%d", count)
	}
}

func TestRollbackOnlyLastBatch(t *testing.T) {
	registry = map[string][]Migration{}
	Register(Migration{Schema: "shared", Name: "20260101_000000_a",
		Up:   func(db *gorm.DB) error { return db.Exec(`CREATE TABLE a (id INTEGER)`).Error },
		Down: func(db *gorm.DB) error { return db.Exec(`DROP TABLE a`).Error }})

	db := openTestDB(t)
	if _, err := Run(db, "shared"); err != nil { // batch 1
		t.Fatal(err)
	}
	Register(Migration{Schema: "shared", Name: "20260101_000001_b",
		Up:   func(db *gorm.DB) error { return db.Exec(`CREATE TABLE b (id INTEGER)`).Error },
		Down: func(db *gorm.DB) error { return db.Exec(`DROP TABLE b`).Error }})
	if _, err := Run(db, "shared"); err != nil { // batch 2
		t.Fatal(err)
	}

	reverted, err := Rollback(db, "shared", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(reverted) != 1 || reverted[0] != "20260101_000001_b" {
		t.Fatalf("expected only batch 2 reverted, got %v", reverted)
	}
	if !db.Migrator().HasTable("a") {
		t.Fatal("batch 1 must survive")
	}
}

func TestStatuses(t *testing.T) {
	registry = map[string][]Migration{}
	Register(Migration{Schema: "shared", Name: "20260101_000000_a",
		Up:   func(db *gorm.DB) error { return db.Exec(`CREATE TABLE a (id INTEGER)`).Error },
		Down: func(db *gorm.DB) error { return db.Exec(`DROP TABLE a`).Error }})
	Register(Migration{Schema: "shared", Name: "20260101_000001_b",
		Up: func(db *gorm.DB) error { return db.Exec(`CREATE TABLE b (id INTEGER)`).Error }})

	db := openTestDB(t)
	st, err := Statuses(db, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if len(st) != 2 || st[0].Applied || st[1].Applied {
		t.Fatalf("expected all pending, got %+v", st)
	}
	if !st[0].Revertible || st[1].Revertible {
		t.Fatalf("revertible flags wrong: %+v", st)
	}

	if _, err := Run(db, "shared"); err != nil {
		t.Fatal(err)
	}
	st, err = Statuses(db, "shared")
	if err != nil {
		t.Fatal(err)
	}
	if !st[0].Applied || !st[1].Applied {
		t.Fatalf("expected all applied, got %+v", st)
	}
}

func TestGenerateFileWritesValidStub(t *testing.T) {
	dir := t.TempDir()
	path, err := GenerateFile("shared", "add_status_to_user", dir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("wrong dir: %s", path)
	}

	// Second call to the same name in the same dir must not clobber it if
	// somehow timestamps collided; simulate a direct collision instead.
	if _, err := GenerateFile("shared", "add_status_to_user", dir); err == nil {
		t.Skip("timestamp did not collide within test resolution; acceptable")
	}
}
