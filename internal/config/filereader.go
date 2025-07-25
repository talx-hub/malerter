package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
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

func ReplaceValues[T any](source T, target T) {
	srcVal := reflect.ValueOf(source).Elem()
	tgtVal := reflect.ValueOf(target).Elem()

	for i := range srcVal.NumField() {
		srcField := srcVal.Field(i)
		tgtField := tgtVal.Field(i)

		if !tgtField.CanSet() {
			continue
		}

		if !isZeroValue(srcField) {
			tgtField.Set(srcField)
		}
	}
}

func isZeroValue(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}
