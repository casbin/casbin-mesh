/*
Copyright The casbind Authors.
@Date: 2021/03/10 17:41
*/

package http

import (
	"net/http"
)

type server struct {
	*http.ServeMux
	cfg         Config
	middlewares []HandlerFunc
}

func New() Server {
	return &server{cfg: DefaultConfig(), ServeMux: http.NewServeMux()}
}

// Options modify server's Config
func (s *server) Options(f func(*Config)) {
	f(&s.cfg)
}

// Use set global middlewares
func (s *server) Use(middlewares ...HandlerFunc) {
	s.middlewares = middlewares
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (s *server) Handle(pattern string, handlers ...HandlerFunc) {
	s.ServeMux.Handle(pattern, CombineHandlers(s.cfg, append(s.middlewares, handlers...)...))
}

type Server interface {
	http.Handler
	Use(middleware ...HandlerFunc)
	Handle(pattern string, handlers ...HandlerFunc)
}
