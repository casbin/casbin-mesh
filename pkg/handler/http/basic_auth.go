// Copyright 2023 The Casbin Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
