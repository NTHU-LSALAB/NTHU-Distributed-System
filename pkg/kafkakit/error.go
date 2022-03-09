package kafkakit

import "errors"

var errUnknown = errors.New("unknown error")

type HandlerError struct {
	Err   error
	Retry bool
}

func (e HandlerError) Error() string {
	if e.Err == nil {
		e.Err = errUnknown
	}

	return e.Err.Error()
}
