package customerror

import "fmt"

type NotFoundError struct {
	RawMetric string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("metric %v not found", e.RawMetric)
}

type InvalidArgumentError struct {
	RawMetric string
}

func (e *InvalidArgumentError) Error() string {
	return fmt.Sprintf("metric %v is incorrect", e.RawMetric)
}
