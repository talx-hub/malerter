// Package retry предоставляет механизм повторных попыток (retry) выполнения операции,
// если выполнение завершилось ошибкой, соответствующей заданному предикату.
//
// Позволяет настроить повтор с экспоненциальной задержкой и максимальным числом попыток.
package retry

import (
	"errors"
	"fmt"
	"time"
)

// Callback — тип функции, которую можно повторно вызывать.
//
// Она принимает произвольное количество аргументов и возвращает результат или ошибку.
type Callback func(args ...any) (any, error)

// ErrPredicate — предикат, проверяющий, подлежит ли ошибка повтору.
//
// Если возвращает true, то выполнение будет повторено.
type ErrPredicate func(err error) bool

// Try выполняет Callback функцию с возможностью повторения при ошибке.
//
// Параметры:
//   - cb: функция, которую нужно вызвать;
//   - pred: функция, определяющая, нужно ли повторять при ошибке;
//   - count: текущая попытка (обычно передаётся 0 при первом вызове).
//
// Повтор происходит, если `pred(err)` возвращает true. Задержка между повторами: 1s, 3s, 5s.
// Максимальное количество попыток: 3. Если все попытки исчерпаны — возвращается ошибка.
//
// Пример:
//
//	data, err := retry.Try(
//	    myFunc,
//	    func(err error) bool { return errors.Is(err, io.ErrUnexpectedEOF) },
//	    0,
//	)
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
