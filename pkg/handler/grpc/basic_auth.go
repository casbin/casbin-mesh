/*
@Author: Weny Xu
@Date: 2021/09/01 0:04
*/

package grpc

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"net/http"
	"strings"
)

var ERRUNAUTHORIZED = errors.New("UNAUTHORIZED")

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	// Case insensitive prefix match. See Issue 22736.
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func getBasicAuthFormContext(ctx context.Context) (username, password string, ok bool) {
	if auth, ok := ctx.Value("Authorization").(string); ok {
		return parseBasicAuth(auth)
	}
	return
}

func BasicAuthor(author func(username, password string) bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		username, password, ok := getBasicAuthFormContext(ctx)
		//UNAUTHORIZED
		if !ok || !author(username, password) {
			return nil, ERRUNAUTHORIZED
		}
		// AUTHORIZED
		return handler(ctx, req)
	}
}

func unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
