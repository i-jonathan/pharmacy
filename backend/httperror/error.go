package httperror

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPError struct {
	Code int
	Message string
	Err error
}


func NewHttpError(code int, message string, err error) *HTTPError {
	return &HTTPError{Code: code, Message: message, Err: err}
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

func (e *HTTPError) JSONRespond(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)
	json.NewEncoder(w).Encode(map[string]string{"error": e.Message})
}

func BadRequest(msg string, err error) *HTTPError {
	return NewHttpError(http.StatusBadRequest, msg, err)
}

func ServerError(msg string, err error) *HTTPError {
	return NewHttpError(http.StatusInternalServerError, msg, err)
}

func NotFound(msg string, err error) *HTTPError {
	return NewHttpError(http.StatusNotFound, msg, err)
}

func Unauthorized(msg string, err error) *HTTPError {
	return NewHttpError(http.StatusUnauthorized, msg, err)
}