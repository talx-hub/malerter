package customerror

import "fmt"

type ErrNotFound struct {
	Message string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found: %s", e.Message)
}

type ErrInvalidArgument struct {
	Info string
}

func (e *ErrInvalidArgument) Error() string {
	return fmt.Sprintf("incorrect request: %s", e.Info)
}
