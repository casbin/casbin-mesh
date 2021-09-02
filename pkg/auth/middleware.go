/*
@Author: Weny Xu
@Date: 2021/09/02 16:47
*/

package auth

import (
	"context"
	httpWrap "github.com/casbin/casbin-mesh/pkg/handler/http"
	"google.golang.org/grpc"
	"net/http"
)

type HttpMiddleware func(ctx *httpWrap.Context) error
type GrpcMiddleware func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)

type authorMiddleware struct {
	RegistryOptions
}

func getAuthTypeFromRequest(req *http.Request) (string, bool) {
	if v := req.Header.Get(AuthType); v != "" {
		return v, true
	}
	return "", false
}

type AuthorMiddleware interface {
	HttpMiddleware(context *httpWrap.Context) error
	GrpcMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)
}

func (a authorMiddleware) HttpMiddleware(context *httpWrap.Context) error {
	if _, _, ok := context.Request.BasicAuth(); ok {
		if middleware, ok := a.GetHttpMiddleware(Basic); ok {
			return middleware(context)
		}
		return ERRUNAUTHORIZED
	} else {
		if typ, ok := getAuthTypeFromRequest(context.Request); ok {
			if middleware, ok := a.GetHttpMiddleware(typ); ok {
				return middleware(context)
			} else {
				return ERRUNSUPPORTAUTHTYPE
			}
		}

	}
	return ERRUNSUPPORTAUTHTYPE
}

func getAuthTypeFromCtx(ctx context.Context) (string, bool) {
	if v, ok := ctx.Value(AuthType).(string); ok {
		return v, ok
	}
	return "", false
}

func (a authorMiddleware) GrpcMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if typ, ok := getAuthTypeFromCtx(ctx); ok {
		if middleware, ok := a.GetGrpcMiddleware(typ); ok {
			return middleware(ctx, req, info, handler)
		} else {
			return nil, ERRUNSUPPORTAUTHTYPE
		}
	}
	return nil, ERRUNSUPPORTAUTHTYPE
}
