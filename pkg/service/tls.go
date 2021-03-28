/*
Copyright 2020 RS4
@Author: Weny Xu
@Date: 2021/03/28 22:39
*/

package service

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// CreateTLSConfig returns a TLS config from the given cert and key.
func CreateTLSConfig(certFile, keyFile, caCertFile string, tls1011 bool) (*tls.Config, error) {
	var err error

	var minTls = uint16(tls.VersionTLS12)
	if tls1011 {
		minTls = tls.VersionTLS10
	}

	config := &tls.Config{
		NextProtos: []string{"h2", "http/1.1"},
		MinVersion: minTls,
	}
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	if caCertFile != "" {
		asn1Data, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			return nil, err
		}
		config.RootCAs = x509.NewCertPool()
		ok := config.RootCAs.AppendCertsFromPEM([]byte(asn1Data))
		if !ok {
			return nil, fmt.Errorf("failed to parse root certificate(s) in %q", caCertFile)
		}
	}
	return config, nil
}
