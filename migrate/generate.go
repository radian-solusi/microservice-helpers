package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var migrationNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)

// GenerateFile writes a migration stub for schema into outDir (creating it if
// needed) and returns the file path. name is snake_case, e.g.
// "add_status_to_user"; the generated filename is Laravel-style
// "YYYY_MM_DD_HHMMSS_name.go", which also sorts chronologically as a string.
func GenerateFile(schema, name, outDir string) (string, error) {
	if schema != "shared" && schema != "ebus" {
		return "", fmt.Errorf("invalid schema %q (want shared or ebus)", schema)
	}
	if !migrationNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid migration name %q (use snake_case)", name)
	}
	migName := fmt.Sprintf("%s_%s", time.Now().Format("2006_01_02_150405"), name)
	path := filepath.Join(outDir, migName+".go")

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("file already exists: %s", path)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create dir %q: %w", outDir, err)
	}

	content := fmt.Sprintf(migrationTemplate, schema, migName)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write file %q: %w", path, err)
	}
	return path, nil
}

const migrationTemplate = `package migrations

import (
	"github.com/radian-solusi/microservice-helpers/migrate"
	"gorm.io/gorm"
)

func init() {
	migrate.Register(migrate.Migration{
		Schema: %q,
		Name:   %q,
		Up: func(db *gorm.DB) error {
			// TODO: write your migration. Things AutoMigrate never does for you
			// (drop column/index, rename, retype) go here, e.g.:
			//
			//   return db.Migrator().DropColumn(&models.User{}, "old_column")
			//   return db.Migrator().DropIndex(&models.User{}, "idx_old")
			//   return db.Exec(` + "`ALTER TABLE \"user\" RENAME COLUMN old_col TO new_col`" + `).Error
			return nil
		},
		Down: func(db *gorm.DB) error {
			// TODO: revert Up, e.g. re-add the dropped column:
			//
			//   return db.Exec(` + "`ALTER TABLE \"user\" ADD COLUMN old_column varchar(255)`" + `).Error
			//
			// Set Down to nil only if this change genuinely cannot be reverted
			// (dropping a column loses its data — Down restores structure only).
			return nil
		},
	})
}
`
