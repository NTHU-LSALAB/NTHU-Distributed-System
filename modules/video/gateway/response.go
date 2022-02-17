package gateway

type StatusCoder interface {
	StatusCode() int
}

type responseError struct {
	Message    string `json:"message"`
	Err        error  `json:"error"`
	statusCode int    `json:"-"`
}

func NewResponseError(statusCode int, message string, err error) *responseError {
	return &responseError{
		Message:    message,
		Err:        err,
		statusCode: statusCode,
	}
}

func (re *responseError) StatusCode() int {
	return re.statusCode
}
