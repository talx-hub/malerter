package customerror

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return "not found: " + e.Message
}

type InvalidArgumentError struct {
	Info string
}

func (e *InvalidArgumentError) Error() string {
	return "incorrect request: " + e.Info
}
