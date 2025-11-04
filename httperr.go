package httperr

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorHandlerFunc extends the standard http.HandlerFunc with an error return type.
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorWriter defines a function which is reply to the request with a specific error message.
type ErrorWriter func(w http.ResponseWriter, err error, code int)

// HttpStatusError is an error type which embeds HTTP status information for responding.
type HttpStatusError interface {
	error
	StatusCode() int
}

// StdHandler converts an ErrorHandlerFunc into the standard library http.HandlerFunc.
func StdHandler(f ErrorHandlerFunc) http.HandlerFunc {
	return StdHandlerWithError(f, StdErrorWriter)
}

// StdHandlerWithError converts an ErrorHandlerFunc into the standard library http.HandlerFunc.
// Whilst also given the freedom to write the error result as preferred.
func StdHandlerWithError(f ErrorHandlerFunc, errorWriter ErrorWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			errorWriter(w, err, HTTPStatus(err))
		}
	}
}

// StdErrorWriter is the default http.Error implementation that will be used to write the error.
func StdErrorWriter(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

type jsonErrorPayload struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// JsonErrorWriter will write a JSON error response
func JsonErrorWriter(w http.ResponseWriter, err error, code int) {
	h := w.Header()
	h.Del("Content-Length")
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(jsonErrorPayload{
		Error:  err.Error(),
		Status: code,
	})
}

type statusError struct {
	error
	status int
}

func (e statusError) StatusCode() int {
	return e.status
}

// WithStatus embeds an HTTP status code to the original error.
// When caught by the handler it will use the HTTP status in the response writing.
func WithStatus(err error, code int) error {
	return &statusError{err, code}
}

// New creates a new error with a custom HTTP status code.
// When caught by the handler it will use the HTTP status in the response writing.
func New(err string, code int) error {
	return &statusError{errors.New(err), code}
}

// HTTPStatus extracts an HTTP status code from err, if available.
// If err implements HttpStatusError we will return the embedded HTTP status code.
// Otherwise, http.StatusInternalServerError is returned.
func HTTPStatus(err error) int {
	if err == nil {
		return 0
	}

	var hse HttpStatusError
	if errors.As(err, &hse) {
		return hse.StatusCode()
	}

	return http.StatusInternalServerError
}

// ErrorServeMux embeds an http.ServeMux with functional handler compatible
// with the ErrorHandlerFunc.
type ErrorServeMux struct {
	mux         *http.ServeMux
	errorWriter ErrorWriter
}

// NewErrorServeMux allocates and returns a new [ErrorServeMux].
func NewErrorServeMux(ew ErrorWriter) *ErrorServeMux {
	return &ErrorServeMux{
		mux:         http.NewServeMux(),
		errorWriter: ew,
	}
}

// HandleFunc registers the handler function for the given pattern.
// If the given pattern conflicts, with one that is already registered, HandleFunc panics.
// It will register the ErrorHandlerFunc with the ErrorWriter configured in the ErrorServeMux.
func (m *ErrorServeMux) HandleFunc(pattern string, handler ErrorHandlerFunc) {
	m.mux.HandleFunc(pattern, StdHandlerWithError(handler, m.errorWriter))
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (m *ErrorServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}
