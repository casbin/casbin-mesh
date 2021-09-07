/*
@Author: Weny Xu
@Date: 2021/09/01 0:04
*/

package http

import (
	"errors"
	"fmt"
	"net/http"
)

var ErrUnauthorized = errors.New("unauthorized")

func BasicAuthor(author func(username, password string) bool) HandlerFunc {
	return func(c *Context) error {
		username, password, ok := c.Request.BasicAuth()
		// UNAUTHORIZED
		if !ok || !author(username, password) {
			unauthorized(c.ResponseWriter, username)
			return ErrUnauthorized
		}
		// AUTHORIZED
		return nil
	}
}

func unauthorized(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}
