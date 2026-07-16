package helpers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/radian-solusi/microservice-helpers/web"
)

func writeConfig(t *testing.T, extra string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "config"), 0o700); err != nil {
		t.Fatal(err)
	}
	body := "[app]\napp_key=\"12345678901234567890123456789012\"\nlimit_data=99\n" + extra
	if err := os.WriteFile(filepath.Join(root, "config", "c.toml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func fileConfigLookup(k string) string {
	if k == "FILE_CONFIG" {
		return "c.toml"
	}
	return ""
}

func TestOptionsApply(t *testing.T) {
	h := &Helpers{}
	WithConfigStart("/tmp/app")(h)
	WithMigrationModels(&struct{}{})(h)
	if h.configStart != "/tmp/app" || len(h.migrationModels) != 1 {
		t.Fatalf("options not applied: %+v", h)
	}
}

func TestLoadConfigAndPureDelegations(t *testing.T) {
	root := writeConfig(t, "")
	h := NewHelpers(WithConfigStart(root), WithEnvLookup(fileConfigLookup))
	if h.GetDefaultLimitData() != 99 || h.GetMainConfig().App.AppKey == "" {
		t.Fatal("config not loaded")
	}
	if h.Slugify("Hello World") != "hello-world" || h.ConvertStringToInt64("42") != 42 || !h.FindString("b", []string{"a", "b"}) {
		t.Fatal("pure delegation")
	}
	key := "12345678901234567890123456789012"
	enc, err := h.Encrypt([]byte("secret"), &key)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := h.Decrypt(enc, &key)
	if err != nil || string(dec) != "secret" {
		t.Fatalf("crypto: %q %v", dec, err)
	}
}

func TestStateAndJWT(t *testing.T) {
	root := writeConfig(t, "")
	h := NewHelpers(WithConfigStart(root), WithEnvLookup(fileConfigLookup))
	h.SetTokenActive("tok")
	h.SetUserActive(web.PayloadAuthorization{UserID: "u1"})
	if h.GetTokenActive() != "tok" || h.GetUserActive().UserID != "u1" {
		t.Fatal("state")
	}
	token, err := h.GenerateJWTToken(map[string]string{"user_id": "u1"}, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := h.ParsingJWT(token, &got); err != nil || got["user_id"] != "u1" {
		t.Fatalf("JWT: %v %v", got, err)
	}
}

type migrationModel struct {
	ID uint `gorm:"primaryKey"`
}

func TestInitializeSystemDatabaseOnly(t *testing.T) {
	root := writeConfig(t, "[database]\ntype=\"sqlite\"\ndb_name=\""+filepath.Join(t.TempDir(), "t.db")+"\"\n")
	h := NewHelpers(WithConfigStart(root), WithEnvLookup(fileConfigLookup), WithMigrationModels(&migrationModel{}))
	if err := h.InitializeSystem(); err != nil {
		t.Fatal(err)
	}
	defer h.GetDatabase().Close()
	defer h.GetRedisClient().Close()
	if err := h.RunMigration(); err != nil {
		t.Fatal(err)
	}
	if !h.GetDatabase().DB().Migrator().HasTable(&migrationModel{}) {
		t.Fatal("table missing")
	}
	if err := h.GetDatabase().Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestUninitializedDependenciesReturnErrors(t *testing.T) {
	h := NewHelpers()
	if err := h.RunMigration(); err == nil {
		t.Fatal("migration expected error")
	}
	if err := h.SetCache("x", "y", 1); err == nil {
		t.Fatal("cache expected error")
	}
}

func TestSetupLoggingDevelopmentDoesNotPanic(t *testing.T) {
	t.Setenv("GIN_MODE", "debug")
	NewHelpers().SetupLogging()
}
