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
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
)

// TcpTransport is the network layer for Raft communications.
type TcpTransport struct {
	ln              net.Listener
	advAddr         net.Addr
	clientTlsConfig *tls.Config
	logger          *zap.Logger
}

// NewTransportFromListener returns an initialized TcpTransport
func NewTransportFromListener(logger *zap.Logger, ln net.Listener, clientTlsConfig *tls.Config) *TcpTransport {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &TcpTransport{ln: ln, clientTlsConfig: clientTlsConfig, advAddr: ln.Addr(), logger: logger}
}

// Dial opens a network connection.
func (t *TcpTransport) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	var dialer *net.Dialer
	dialer = &net.Dialer{Timeout: timeout}
	if t.clientTlsConfig == nil {
		return dialer.Dial("tcp", addr)
	} else {
		return tls.DialWithDialer(dialer, "tcp", addr, t.clientTlsConfig)
	}
}

// Accept waits for the next connection.
func (t *TcpTransport) Accept() (net.Conn, error) {
	c, err := t.ln.Accept()
	if err != nil {
		t.logger.Error("err accepting", zap.Error(err))
	}
	return c, err
}

// Close closes the transport
func (t *TcpTransport) Close() error {
	if t.ln != nil {
		return t.ln.Close()
	}
	return nil
}

// Addr returns the binding address of the transport.
func (t *TcpTransport) Addr() net.Addr {
	return t.advAddr
}

// createTLSConfig returns a TLS config from the given cert and key.
func createTLSConfig(certFile, keyFile, caFile string, isServer bool) (*tls.Config, error) {
	var err error
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	if caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		if isServer {
			config.ClientCAs = caCertPool
		} else {
			config.RootCAs = caCertPool
		}
	}

	return config, nil
}

var CreateTLSConfig = createTLSConfig
