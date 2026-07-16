package strutil

import (
	"encoding/json"
	"errors"
	"fmt"
)

func JSONToStruct(data []byte, target any) error {
	if target == nil {
		return errors.New("JSON target must not be nil")
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("unmarshal JSON: %w", err)
	}
	return nil
}

func StructToJSON(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}
	return data, nil
}

func InterfaceToStruct(value, target any) error {
	if value == nil {
		return errors.New("data cannot be nil")
	}
	data, err := StructToJSON(value)
	if err != nil {
		return err
	}
	return JSONToStruct(data, target)
}
