package httperror

import "fmt"

type HTTPError struct {
	Code int
	Message string
	Err error
}

func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("http error: %d - %s | %v", e.Code, e.Message, e.Err)
	}
	
	return fmt.Sprintf("http error: %d - %s", e.Code, e.Message)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

