package config_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/talx-hub/malerter/internal/config"
)

type TestConfig struct {
	StringField string        `json:"string_field"`
	IntField    int           `json:"int_field"`
	BoolField   bool          `json:"bool_field"`
	TimeField   time.Duration `json:"time_field"`
}

func TestReadFromFile_Success(t *testing.T) {
	content := `{
		"string_field": "test",
		"int_field": 42,
		"bool_field": true,
		"time_field": 300000000000
	}`

	tmpFile, err := os.CreateTemp("", "test_config_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	_, _ = tmpFile.WriteString(content)
	_ = tmpFile.Close()

	var cfg TestConfig
	err = config.ReadFromFile(tmpFile.Name(), &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := TestConfig{
		StringField: "test",
		IntField:    42,
		BoolField:   true,
		TimeField:   300 * time.Second,
	}

	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("expected %+v, got %+v", expected, cfg)
	}
}

func TestReadFromFile_FileNotExist(t *testing.T) {
	var cfg TestConfig
	err := config.ReadFromFile("non_existing_file.json", &cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadFromFile_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_json_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	_, _ = tmpFile.WriteString("{invalid json")
	_ = tmpFile.Close()

	var cfg TestConfig
	err = config.ReadFromFile(tmpFile.Name(), &cfg)
	if err == nil {
		t.Fatal("expected error on invalid JSON, got nil")
	}
}
