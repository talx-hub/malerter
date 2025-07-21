package retry

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// errTemporary — пример временной ошибки, которая допускает повтор.
var errTemporary = errors.New("temporary error")

// errFatal — ошибка, при которой повтор не выполняется.
var errFatal = errors.New("fatal error")

func TestTry_Success(t *testing.T) {
	cb := func(args ...any) (any, error) {
		return "ok", nil
	}
	pred := func(err error) bool {
		return true
	}

	result, err := Try(cb, pred, 0)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected result 'ok', got: %v", result)
	}
}

func TestTry_RetrySuccess(t *testing.T) {
	attempts := 0
	cb := func(args ...any) (any, error) {
		if attempts < 2 {
			attempts++
			return nil, errTemporary
		}
		return "retried", nil
	}
	pred := func(err error) bool {
		return errors.Is(err, errTemporary)
	}

	start := time.Now()
	result, err := Try(cb, pred, 0)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result != "retried" {
		t.Fatalf("expected result 'retried', got: %v", result)
	}
	if elapsed < 4*time.Second { // 1s + 3s sleep
		t.Errorf("expected retries with delays, but too fast: %s", elapsed)
	}
}

func TestTry_ExhaustedRetries(t *testing.T) {
	cb := func(args ...any) (any, error) {
		return nil, errTemporary
	}
	pred := func(err error) bool {
		return errors.Is(err, errTemporary)
	}

	start := time.Now()
	_, err := Try(cb, pred, 0)
	elapsed := time.Since(start)

	if err == nil || err.Error() != "all attempts to retry are out" {
		t.Fatalf("expected retry exhaustion error, got: %v", err)
	}
	if elapsed < 9*time.Second { // 1s + 3s + 5s = 9s total delay
		t.Errorf("expected full retry delay, got: %s", elapsed)
	}
}

func TestTry_NoRetryOnFatalError(t *testing.T) {
	cb := func(args ...any) (any, error) {
		return nil, errFatal
	}
	pred := func(err error) bool {
		return false
	}

	result, err := Try(cb, pred, 0)
	if err == nil || result != nil {
		t.Fatal("expected immediate failure, but got success")
	}
	expected := fmt.Sprintf("on attempt #0 error occurred: %v", errFatal)
	if err.Error() != expected {
		t.Errorf("unexpected error message: got %q, want %q", err.Error(), expected)
	}
}
