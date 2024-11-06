package customerror

import "fmt"

type NotFoundError struct {
	MetricURL string
	Info      string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("metric %v not found: %s", e.MetricURL, e.Info)
}

type InvalidArgumentError struct {
	MetricURL string
	Info      string
}

func (e *InvalidArgumentError) Error() string {
	return fmt.Sprintf("metric %v is incorrect: %s", e.MetricURL, e.Info)
}
