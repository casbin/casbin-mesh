/*
Copyright The casbind Authors.
@Date: 2021/03/10 18:33
*/

package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Context
type Context struct {
	context.Context
	md           metadata.MD
	request      interface{}
	response     interface{}
	indexHandler uint8
	handlers     []HandlerFunc
}

// Response gets read-only response
func (c *Context) Response() interface{} {
	return c.response
}

// SetResponse sets the response to context
func (c *Context) SetResponse(resp interface{}) {
	c.response = resp
}

// Request gets read-only request
func (c *Context) Request() interface{} {
	return c.request
}

// Next run the next handler func until out of range
func (c *Context) Next() (err error) {
	for c.indexHandler < uint8(len(c.handlers)) {
		err = c.handlers[c.indexHandler](c)
		c.indexHandler++
	}
	return err
}

func newContext(ctx context.Context, md metadata.MD, request interface{}, h ...HandlerFunc) Context {
	return Context{
		Context:      ctx,
		request:      request,
		md:           md,
		indexHandler: 0,
		handlers:     h,
	}
}
