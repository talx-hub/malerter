package retry

import (
	"errors"
	"fmt"
	"time"
)

type Callback func(args ...any) (any, error)
type ErrPredicate func(err error) bool

func Try(cb Callback, pred ErrPredicate, count int) (any, error) {
	data, err := cb()
	if err == nil {
		return data, nil
	}

	const maxAttemptCount = 3
	if count >= maxAttemptCount {
		return nil, errors.New("all attempts to retry are out")
	}
	if pred(err) {
		time.Sleep((time.Duration(count*2 + 1)) * time.Second) // count: 0 1 2 -> seconds: 1 3 5.
		return Try(cb, pred, count+1)
	}
	return nil, fmt.Errorf(
		"on attempt #%d error occurred: %w", count, err)
}
