package httperr

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPStatus(t *testing.T) {
	a := assert.New(t)

	normalError := errors.New("standard error")

	a.Equal(0, HTTPStatus(nil))
	a.Equal(http.StatusInternalServerError, HTTPStatus(normalError))
	a.Equal(400, HTTPStatus(WithStatus(normalError, 400)))
}

func TestStdHandler(t *testing.T) {

	t.Run("standard error", func(t *testing.T) {
		a := assert.New(t)
		handler := func(w http.ResponseWriter, _ *http.Request) error {
			return errors.New("standard error")
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rw := httptest.NewRecorder()

		StdHandler(handler)(rw, req)
		a.Equal(http.StatusInternalServerError, rw.Result().StatusCode)
		a.Equal("text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		a.Equal("standard error\n", rw.Body.String())
	})

	t.Run("error with status code", func(t *testing.T) {
		a := assert.New(t)
		handler := func(w http.ResponseWriter, _ *http.Request) error {
			return WithStatus(errors.New("standard error"), http.StatusBadRequest)
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rw := httptest.NewRecorder()

		StdHandler(handler)(rw, req)
		a.Equal(http.StatusBadRequest, rw.Result().StatusCode)
		a.Equal("text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		a.Equal("standard error\n", rw.Body.String())
	})

	t.Run("without error", func(t *testing.T) {
		a := assert.New(t)
		handler := func(w http.ResponseWriter, _ *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK\n"))
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rw := httptest.NewRecorder()

		StdHandler(handler)(rw, req)
		a.Equal(http.StatusOK, rw.Result().StatusCode)
		a.Equal("OK\n", rw.Body.String())
	})
}

func TestStdErrorWriter(t *testing.T) {
	a := assert.New(t)
	err := errors.New("standard error")
	rw := httptest.NewRecorder()

	StdErrorWriter(rw, err, http.StatusInternalServerError)
	a.Equal(http.StatusInternalServerError, rw.Result().StatusCode)
	a.Equal("text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
	a.Equal("nosniff", rw.Header().Get("X-Content-Type-Options"))
	a.Equal("standard error\n", rw.Body.String())
}

func TestJsonErrorWriter(t *testing.T) {
	a := assert.New(t)
	err := errors.New("standard error")
	rw := httptest.NewRecorder()

	JsonErrorWriter(rw, err, http.StatusInternalServerError)
	a.Equal(http.StatusInternalServerError, rw.Result().StatusCode)
	a.Equal("application/json; charset=utf-8", rw.Header().Get("Content-Type"))
	a.Equal("nosniff", rw.Header().Get("X-Content-Type-Options"))
	a.Equal(`{"error":"standard error","status":500}`+"\n", rw.Body.String())
}

func TestErrorServeMux(t *testing.T) {
	mux := NewErrorServeMux(StdErrorWriter)
	mux.HandleFunc("GET /standard", func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("standard error")
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) error {
		return New("custom error", http.StatusBadRequest)
	})
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
		return nil
	})

	t.Run("standard error", func(t *testing.T) {
		a := assert.New(t)
		req := httptest.NewRequest(http.MethodGet, "/standard", nil)
		rw := httptest.NewRecorder()

		mux.ServeHTTP(rw, req)
		a.Equal(http.StatusInternalServerError, rw.Result().StatusCode)
		a.Equal("text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		a.Equal("standard error\n", rw.Body.String())
	})

	t.Run("error with status code", func(t *testing.T) {
		a := assert.New(t)
		req := httptest.NewRequest(http.MethodGet, "/status", nil)
		rw := httptest.NewRecorder()

		mux.ServeHTTP(rw, req)
		a.Equal(http.StatusBadRequest, rw.Result().StatusCode)
		a.Equal("text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		a.Equal("custom error\n", rw.Body.String())
	})

	t.Run("without error", func(t *testing.T) {
		a := assert.New(t)
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		rw := httptest.NewRecorder()

		mux.ServeHTTP(rw, req)
		a.Equal(http.StatusOK, rw.Result().StatusCode)
		a.Equal("OK\n", rw.Body.String())
	})

	t.Run("endpoint not found", func(t *testing.T) {
		a := assert.New(t)
		req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
		rw := httptest.NewRecorder()

		mux.ServeHTTP(rw, req)
		a.Equal(http.StatusNotFound, rw.Result().StatusCode)
	})
}
