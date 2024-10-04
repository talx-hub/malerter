package customerror

import "fmt"

type NotFoundError struct {
	RawMetric string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Metric %v not found", e.RawMetric)
}

type IvalidArgumentError struct {
	RawMetric string
}

func (e *IvalidArgumentError) Error() string {
	return fmt.Sprintf("Raw metric %v is incorrect", e.RawMetric)
}
