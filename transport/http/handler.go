/*
Copyright The casbind Authors.
@Date: 2021/03/10 17:30
*/

package http

import (
	"context"
	"net/http"
)

type HandlerFunc func(*Context) error

type Handler struct {
	handlers []HandlerFunc
	cfg      Config
}

func CombineHandlers(cfg Config, h ...HandlerFunc) http.Handler {
	return &Handler{cfg: cfg, handlers: h}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(context.Background(), h.handlers...)
	c.Request = r
	c.ResponseWriter = w
	err := c.Next()
	if err != nil {
		h.cfg.ErrorHandler(c.Context, err, w)
	}
	return
}
