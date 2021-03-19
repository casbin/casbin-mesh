/*
Copyright The casbind Authors.
@Date: 2021/03/10 17:28
*/

package http

import (
	"context"
	"net/http"
)

type Context struct {
	context.Context
	indexHandler   uint8
	handlers       []HandlerFunc
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

// Next run the next handler func until out of range
func (c *Context) Next() (err error) {
	c.indexHandler++
	if c.indexHandler < uint8(len(c.handlers)) {
		err = c.handlers[c.indexHandler](c)
	}
	return err
}

func newContext(ctx context.Context, h ...HandlerFunc) Context {
	return Context{
		Context:      ctx,
		indexHandler: -1,
		handlers:     h,
	}
}
