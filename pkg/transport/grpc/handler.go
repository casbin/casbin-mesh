/*
Copyright The casbind Authors.
@Date: 2021/03/10 18:30
*/

package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type Handler interface {
	ServeGRPC(ctx context.Context, request interface{}) (interface{}, error)
}

type HandlerFunc func(*Context) error

type handler struct {
	handlers []HandlerFunc
	cfg      Config
}

// ServeGRPC the endpoint of handle request
func (h handler) ServeGRPC(ctx context.Context, request interface{}) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	c := newContext(ctx, md, request, h.handlers...)
	err := c.Next()
	if err != nil {
		return nil, err
	}
	return c.Response(), nil
}

// CombineHandlers return a grpc Handler
func CombineHandlers(cfg Config, h ...HandlerFunc) Handler {
	return handler{cfg: cfg, handlers: h}
}
