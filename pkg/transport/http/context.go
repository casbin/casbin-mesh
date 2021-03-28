/*
Copyright The casbind Authors.
@Date: 2021/03/10 17:28
*/

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

func (c *Context) StatusCode(code int) *Context {
	c.ResponseWriter.WriteHeader(code)
	return c
}

func (c *Context) Write(resp interface{}) error {
	c.ResponseWriter.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(c.ResponseWriter).Encode(resp)
}

// Next run the next handler func until out of range
func (c *Context) Next() (err error) {
	c.indexHandler++
	if c.indexHandler < int8(len(c.handlers)) {
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
