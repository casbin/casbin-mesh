/*
@Author: Weny Xu
@Date: 2021/09/01 0:04
*/

package middleware

import (
	"errors"
	"fmt"
	httpWrap "github.com/casbin/casbin-mesh/pkg/handler/http"
	"net/http"
)

var ERRUNAUTHORIZED = errors.New("UNAUTHORIZED")

func BasicAuthor(author func(username, password string) bool) httpWrap.HandlerFunc {
	return func(c *httpWrap.Context) error {
		username, password, ok := c.Request.BasicAuth()
		// UNAUTHORIZED
		if !ok || !author(username, password) {
			unauthorized(c.ResponseWriter, username)
			return ERRUNAUTHORIZED
		}
		// AUTHORIZED
		return nil
	}
}

func unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
