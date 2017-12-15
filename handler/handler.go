package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// FlowError represents an error with an associated HTTP status code.
type FlowError struct {
	Code   int
	Type   string
	Detail string
	Err    error
}

type Response struct {
	Result  interface{}    `json:"result,omitempty"`
	Success bool           `json:"success"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Type    string `json:"type,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// Allows FlowError to satisfy the error interface.
func (se FlowError) Error() string {
	return se.Err.Error()
}

// Status returns our HTTP status code.
func (se FlowError) Status() int {
	return se.Code
}

// Env is a (simple) example of our application-wide configuration.
type Env struct {
}

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type Handler struct {
	*Env
	H func(e *Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case FlowError:
			w.Header().Set("Content-type", "application/json")
			json.NewEncoder(w).Encode(Response{Success: false, Error: &ResponseError{Code: e.Status(), Message: e.Error(), Type: e.Type, Detail: e.Detail}})
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}

func LogginHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

func RecoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v\n", err)
				http.Error(w, http.StatusText(500), 500)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
