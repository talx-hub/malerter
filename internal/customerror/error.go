package customerror

import "fmt"

type NotFoundError struct {
	Metric string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Metric %v not found", e.Metric)
}

type WrongTypeError struct {
	RawMetric string
}

func (e *WrongTypeError) Error() string {
	return fmt.Sprintf("Raw metric %v is incorrect", e.RawMetric)
}
