package httperror

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

type HTTPError struct {
	Code    int
	Message string
	Err     error
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

func (e *HTTPError) Render(w http.ResponseWriter, tmpl *template.Template) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(e.Code)
	err := tmpl.ExecuteTemplate(w, "error.html", map[string]any{
		"Code":    e.Code,
		"Message": e.Message,
	})
	if err != nil {
		// Fallback if template rendering fails
		http.Error(w, e.Message, e.Code)
	}
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

func Forbidden(msg string, err error) *HTTPError {
	return NewHttpError(http.StatusForbidden, msg, err)
}
