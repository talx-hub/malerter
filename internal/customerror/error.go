package customerror

import "fmt"

type NotFoundError struct {
	Metric string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("Metric %v not found", e.Metric)
}
