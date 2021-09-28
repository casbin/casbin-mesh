// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package http

import (
	"net/http"
)

type server struct {
	*http.ServeMux
	cfg         Config
	middlewares []HandlerFunc
}

func New() Server {
	return &server{cfg: DefaultConfig(), ServeMux: http.NewServeMux()}
}

// Options modify server's Config
func (s *server) Options(f func(*Config)) {
	f(&s.cfg)
}

// Use set global middlewares
func (s *server) Use(middlewares ...HandlerFunc) {
	s.middlewares = append(s.middlewares, middlewares...)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (s *server) Handle(pattern string, handlers ...HandlerFunc) {
	s.ServeMux.Handle(pattern, CombineHandlers(s.cfg, append(append([]HandlerFunc{}, s.middlewares...), handlers...)...))
}

type Server interface {
	http.Handler
	Use(middleware ...HandlerFunc)
	Handle(pattern string, handlers ...HandlerFunc)
}
