package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadFromFile[T any](file string, temp *T) error {
	jsonFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		_ = jsonFile.Close()
	}()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = json.Unmarshal(byteValue, temp)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}
