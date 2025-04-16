package customerror

type NotFoundError struct {
	Info string
}

func (e *NotFoundError) Error() string {
	return "not found: " + e.Info
}

type InvalidArgumentError struct {
	Info string
}

func (e *InvalidArgumentError) Error() string {
	return "incorrect request: " + e.Info
}
