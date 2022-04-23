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
