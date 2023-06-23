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

package store

import (
	"net"
	"time"

	"github.com/hashicorp/raft"
)

// Listener is the interface expected by the Store for Transports.
type Listener interface {
	net.Listener
	Dial(address string, timeout time.Duration) (net.Conn, error)
}

// Transport is the network service provided to Raft, and wraps a Listener.
type Transport struct {
	ln Listener
}

// NewTransport returns an initialized Transport.
func NewTransport(ln Listener) *Transport {
	return &Transport{
		ln: ln,
	}
}

// Dial creates a new network connection.
func (t *Transport) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return t.ln.Dial(string(addr), timeout)
}

// Accept waits for the next connection.
func (t *Transport) Accept() (net.Conn, error) {
	return t.ln.Accept()
}

// Close closes the transport
func (t *Transport) Close() error {
	return t.ln.Close()
}

// Addr returns the binding address of the transport.
func (t *Transport) Addr() net.Addr {
	return t.ln.Addr()
}
