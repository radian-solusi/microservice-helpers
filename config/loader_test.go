package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	got, err := FindProjectRoot(nested)
	if err != nil {
		t.Fatal(err)
	}
	if got != root {
		t.Fatalf("got %q want %q", got, root)
	}
}

func TestFindProjectRootRejectsMissingModule(t *testing.T) {
	_, err := FindProjectRoot(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "go.mod") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "config"), 0o700); err != nil {
		t.Fatal(err)
	}
	body := "[app]\napp_key = \"12345678901234567890123456789012\"\nlimit_data = 77\n"
	if err := os.WriteFile(filepath.Join(root, "config", "test.toml"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	var got MainConfig
	err := LoadFromEnvironment(root, func(key string) string {
		if key == "FILE_CONFIG" {
			return "test.toml"
		}
		return ""
	}, &got)
	if err != nil {
		t.Fatal(err)
	}
	if got.App.LimitData != 77 {
		t.Fatalf("got %d", got.App.LimitData)
	}
}

func TestLoadFromEnvironmentValidation(t *testing.T) {
	tests := []struct {
		name      string
		target    any
		env, want string
	}{
		{"nil", nil, "test.toml", "must not be nil"},
		{"value", MainConfig{}, "test.toml", "non-nil pointer"},
		{"nil pointer", (*MainConfig)(nil), "test.toml", "non-nil pointer"},
		{"missing env", &MainConfig{}, "", "FILE_CONFIG"},
		{"absolute", &MainConfig{}, "/tmp/secret.toml", "file name"},
		{"traversal", &MainConfig{}, "../secret.toml", "file name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LoadFromEnvironment(t.TempDir(), func(string) string { return tt.env }, tt.target)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v want %q", err, tt.want)
			}
		})
	}
}

func TestFindProjectRootDevelopment(t *testing.T) {
	t.Setenv("GIN_MODE", "debug")

	root := t.TempDir()
	nested := filepath.Join(root, "internal", "services")

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "go.mod"),
		[]byte("module example.test\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	result, err := FindProjectRoot(nested)
	if err != nil {
		t.Fatal(err)
	}

	if result != root {
		t.Fatalf("expected root %q, got %q", root, result)
	}
}

func TestFindProjectRootProduction(t *testing.T) {
	t.Setenv("GIN_MODE", "release")

	root := t.TempDir()
	nested := filepath.Join(root, "internal", "services")

	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		filepath.Join(root, "app"),
		[]byte("binary-placeholder"),
		0o755,
	); err != nil {
		t.Fatal(err)
	}

	result, err := FindProjectRoot(nested)
	if err != nil {
		t.Fatal(err)
	}

	if result != root {
		t.Fatalf("expected root %q, got %q", root, result)
	}
}
