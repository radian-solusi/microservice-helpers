package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/BurntSushi/toml"
)

func FindProjectRoot(start string) (string, error) {
	if start == "" {
		var err error

		start, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
	}

	current, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve start path: %w", err)
	}

	// Development searches for go.mod.
	// Production searches for the compiled application binary.
	projectMarker := "go.mod"

	if os.Getenv("GIN_MODE") == "release" {
		projectMarker = "app"
	}

	for {
		markerPath := filepath.Join(current, projectMarker)

		if _, err := os.Stat(markerPath); err == nil {
			return current, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf(
				"inspect project marker %q: %w",
				markerPath,
				err,
			)
		}

		parent := filepath.Dir(current)

		if parent == current {
			return "", fmt.Errorf(
				"project marker %q not found from %q",
				projectMarker,
				start,
			)
		}

		current = parent
	}
}

func Load(path string, target any) error {
	if target == nil {
		return errors.New("config target must not be nil")
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("config target must be a non-nil pointer")
	}

	if _, err := toml.DecodeFile(path, target); err != nil {
		return fmt.Errorf("decode TOML %q: %w", path, err)
	}

	return nil
}

func LoadFromEnvironment(
	start string,
	lookupEnv func(string) string,
	target any,
) error {
	if target == nil {
		return errors.New("config target must not be nil")
	}

	value := reflect.ValueOf(target)

	if value.Kind() != reflect.Pointer || value.IsNil() {
		return errors.New("config target must be a non-nil pointer")
	}

	name := lookupEnv("FILE_CONFIG")

	if name == "" {
		return errors.New("missing FILE_CONFIG environment variable")
	}

	if filepath.IsAbs(name) || filepath.Base(name) != name {
		return errors.New(
			"FILE_CONFIG must be a file name without path components",
		)
	}

	root, err := FindProjectRoot(start)
	if err != nil {
		return fmt.Errorf("find project root: %w", err)
	}

	configPath := filepath.Join(root, "config", name)

	return Load(configPath, target)
}
