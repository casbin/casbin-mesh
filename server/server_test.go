package server

import (
	"crypto/tls"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewListener(t *testing.T) {
	cfg := &CmdConfig{}
	s, err := NewServer(cfg)
	require.NoError(t, err)

	listener, err := s.newListener("127.0.0.1:0", "", false, "", "", "", tls.NoClientCert, false)
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer func() {
		_ = listener.Close()
	}()

	listener, err = s.newListener("127.0.0.1:0", "", false, "", "", "", tls.NoClientCert, true)
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer func() {
		_ = listener.Close()
	}()
}

func TestNewListenerWithTls(t *testing.T) {
	cfg := &CmdConfig{}
	s, err := NewServer(cfg)
	require.NoError(t, err)

	caPath := "../test/testdata/certificate-authority/ca-key.pem"
	raftClientKeyPath := "../test/testdata/certificate-authority/raft-client-key.pem"
	raftClientCertPath := "../test/testdata/certificate-authority/raft-client.pem"
	listener, err := s.newListener("127.0.0.1:0", "", true, raftClientKeyPath, raftClientCertPath, caPath, tls.RequestClientCert, false)
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer func() {
		_ = listener.Close()
	}()

	raftServerKeyPath := "../test/testdata/certificate-authority/raft-server-key.pem"
	raftServerCertPath := "../test/testdata/certificate-authority/raft-server.pem"
	listener, err = s.newListener("127.0.0.1:0", "", true, raftServerKeyPath, raftServerCertPath, caPath, tls.RequestClientCert, true)
	require.NoError(t, err)
	require.NotNil(t, listener)
	defer func() {
		_ = listener.Close()
	}()
}
