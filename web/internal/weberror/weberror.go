package weberror

import "fmt"

type WebError struct {
	HTTPCode int
	Err      error
}

func (e *WebError) Error() string {
	return fmt.Sprintf("%d %v", e.HTTPCode, e.Err)
}

func (e *WebError) Unwrap() error {
	return e.Err
}

func NewWebError(httpCode int, err error) *WebError {
	return &WebError{
		HTTPCode: httpCode,
		Err:      err,
	}
}
