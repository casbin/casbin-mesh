/*
Copyright The casbind Authors.
@Date: 2021/03/10 18:33
*/

package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type Context struct {
	context.Context
	md           metadata.MD
	request      interface{}
	response     interface{}
	indexHandler uint8
	handlers     []HandlerFunc
}

func (c *Context) Response() interface{} {
	return c.response
}

func (c *Context) SetResponse(resp interface{}) {
	c.response = resp
}

func (c *Context) Request() interface{} {
	return c.request
}

func (c *Context) Next() (err error) {
	c.indexHandler++
	if c.indexHandler < uint8(len(c.handlers)) {
		err = c.handlers[c.indexHandler](c)
	}
	return err
}

func newContext(ctx context.Context, md metadata.MD, request interface{}, h ...HandlerFunc) Context {
	return Context{
		Context:      ctx,
		request:      request,
		md:           md,
		indexHandler: -1,
		handlers:     h,
	}
}
