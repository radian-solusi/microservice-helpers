package strutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

func IsFileComplete(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && len(data) > 0 && json.Valid(data)
}

func ReadJSONFile(path string, target any, attempts int, delay time.Duration) error {
	if attempts <= 0 {
		return errors.New("attempts must be positive")
	}
	for i := 0; i < attempts; i++ {
		if IsFileComplete(path) {
			break
		}
		if i+1 < attempts {
			time.Sleep(delay)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read JSON file %q: %w", path, err)
	}
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON in file: %s", path)
	}
	return JSONToStruct(data, target)
}
