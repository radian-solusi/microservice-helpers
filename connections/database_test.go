package connections

import (
	"context"
	"path/filepath"
	"testing"

	helperconfig "github.com/radian-solusi/microservice-helpers/config"
)

type migrationRecord struct {
	ID uint `gorm:"primaryKey"`
}

func TestNewDatabaseSQLiteAndMigration(t *testing.T) {
	cfg := helperconfig.DatabaseConfig{Type: helperconfig.DBTypeSQLite, DBName: filepath.Join(t.TempDir(), "test.db")}
	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(&migrationRecord{}); err != nil {
		t.Fatal(err)
	}
	if !db.DB().Migrator().HasTable(&migrationRecord{}) {
		t.Fatal("table missing")
	}
}

func TestNewDatabaseRejectsUnknownType(t *testing.T) {
	if _, err := NewDatabase(helperconfig.DatabaseConfig{Type: "oracle"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrateRejectsEmptyModels(t *testing.T) {
	cfg := helperconfig.DatabaseConfig{Type: helperconfig.DBTypeSQLite, DBName: filepath.Join(t.TempDir(), "m.db")}
	db, _ := NewDatabase(cfg)
	defer db.Close()
	if err := db.Migrate(); err == nil {
		t.Fatal("expected error")
	}
}
