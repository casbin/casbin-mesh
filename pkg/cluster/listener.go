package cluster

import (
	"errors"
	"fmt"
	"net"

	"github.com/hashicorp/go-multierror"
)

// Listener handles the connection of multiple net.Listener, provides listening and accepting new connection.
type Listener struct {
	net.Listener

	listeners []net.Listener
	connCh    chan net.Conn

	advertise net.Addr

	shutdownCh chan struct{}
}

// NewListener creates a Listener.
// If the advertise is not empty, it will be set as the listener address.
func NewListener(listeners []net.Listener, advertise string) (net.Listener, error) {
	if len(listeners) == 0 {
		return nil, errors.New("listeners cannot empty")
	}

	n := &Listener{
		listeners:  listeners,
		connCh:     make(chan net.Conn),
		shutdownCh: make(chan struct{}),
	}

	if len(advertise) != 0 {
		advertiseAddress, err := net.ResolveTCPAddr("tcp", advertise)
		if err != nil {
			return nil, err
		}
		n.advertise = advertiseAddress
	}

	for i := range n.listeners {
		go n.handle(n.listeners[i])
	}

	return n, nil
}

// handle handles connection from net.Listener.
func (l *Listener) handle(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		select {
		case l.connCh <- conn:
		case <-l.shutdownCh:
			_ = conn.Close()
			return
		}
	}
}

// Accept waits and returns the next incoming connection.
func (l *Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connCh:
		return conn, nil
	case <-l.shutdownCh:
		return nil, fmt.Errorf("listener is closed")
	}
}

// Addr returns a network address from advertise if the advertise is not empty, otherwise from first listener.
func (l *Listener) Addr() net.Addr {
	if l.advertise != nil {
		return l.advertise
	}

	return l.listeners[0].Addr()
}

// Close closes all listeners provided by NewListener, and stops accept the new connection.
func (l *Listener) Close() error {
	var result error

	select {
	case <-l.shutdownCh:
		break
	default:
		close(l.shutdownCh)
		for _, listener := range l.listeners {
			err := listener.Close()
			if err != nil && !errors.Is(err, net.ErrClosed) {
				result = multierror.Append(result, err)
			}
		}
	}

	l.listeners = nil

	return result
}
