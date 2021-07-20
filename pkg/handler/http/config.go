package http

import (
	"context"
	"encoding/json"
	"net/http"
)

type Config struct {
	ErrorHandler func(ctx context.Context, err error, w http.ResponseWriter)
}

func DefaultConfig() Config {
	return Config{ErrorHandler: DefaultErrorHandler}
}

type errorWrapper struct {
	Error string `json:"error"`
}

func err2code(err error) int {
	return http.StatusInternalServerError
}

func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	_ = json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

var (
	DefaultErrorHandler = ErrorEncoder
)
