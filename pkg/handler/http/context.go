package http

import (
	"context"
	"encoding/json"
	"net/http"
)

type Context struct {
	context.Context
	indexHandler   int8
	handlers       []HandlerFunc
	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

func (c *Context) Clone() *Context {
	return &Context{
		Context:        c.Context,
		indexHandler:   c.indexHandler,
		handlers:       c.handlers,
		ResponseWriter: c.ResponseWriter,
		Request:        c.Request.Clone(context.TODO()),
	}

}

func (c *Context) StatusCode(code int) *Context {
	c.ResponseWriter.WriteHeader(code)
	return c
}

func (c *Context) JSON(resp interface{}) error {
	return json.NewEncoder(c.ResponseWriter).Encode(resp)
}

// Next runs the next handler func until out of range
func (c *Context) Next() (err error) {
	c.indexHandler++
	for c.indexHandler < int8(len(c.handlers)) {
		err = c.handlers[c.indexHandler](c)
		if err != nil {
			return err
		}
		c.indexHandler++
	}
	return nil
}

func newContext(ctx context.Context, h ...HandlerFunc) Context {
	return Context{
		Context:      ctx,
		indexHandler: -1,
		handlers:     h,
	}
}
