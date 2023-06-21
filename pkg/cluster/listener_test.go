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

package cluster

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/nettest"
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	ln1, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	ln2, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	l, err := NewListener([]net.Listener{ln1, ln2}, "")
	assert.NoError(t, err)
	assert.NotNil(t, l)

	assert.Equal(t, l.Addr().String(), ln1.Addr().String())

	err = testConnect(l.Addr().String())
	assert.NoError(t, err)

	err = l.Close()
	assert.NoError(t, err)
}

func TestListenerWithAdvertise(t *testing.T) {
	ln1, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	ln2, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	l, err := NewListener([]net.Listener{ln1, ln2}, ln2.Addr().String())
	assert.NoError(t, err)
	assert.NotNil(t, l)

	assert.Equal(t, l.Addr().String(), ln2.Addr().String())

	err = testConnect(l.Addr().String())
	assert.NoError(t, err)

	err = l.Close()
	assert.NoError(t, err)
}

func testConnect(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	_ = conn.Close()
	return nil
}
