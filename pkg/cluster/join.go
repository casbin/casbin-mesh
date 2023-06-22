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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/casbin/casbin-mesh/pkg/auth"

	"github.com/casbin/casbin-mesh/pkg/utils"
)

var (
	// ErrJoinFailed is returned when a node fails to join a cluster
	ErrJoinFailed = errors.New("failed to join cluster")
)

// Join attempts to join the cluster at one of the addresses given in joinAddr.
// It walks through joinAddr in order, and sets the node ID and Raft address of
// the joining node as id addr respectively. It returns the endpoint successfully
// used to join the cluster.
func Join(srcIP string, joinAddr []string, id, addr string, voter bool, meta map[string]string, numAttempts int,
	attemptInterval time.Duration, tlsConfig *tls.Config, authConfig auth.AuthConfig) (string, error) {
	var err error
	var j string
	logger := log.New(os.Stderr, "[cluster-join] ", log.LstdFlags)
	if tlsConfig == nil {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	for i := 0; i < numAttempts; i++ {
		for _, a := range joinAddr {
			j, err = join(srcIP, a, id, addr, voter, meta, tlsConfig, logger, authConfig)
			if err == nil {
				// Success!
				return j, nil
			}
		}
		logger.Printf("failed to join cluster at %s: %s, sleeping %s before retry", joinAddr, err.Error(), attemptInterval)
		time.Sleep(attemptInterval)
	}
	logger.Printf("failed to join cluster at %s, after %d attempts", joinAddr, numAttempts)
	return "", ErrJoinFailed
}

func join(srcIP, joinAddr, id, addr string, voter bool, meta map[string]string, tlsConfig *tls.Config, logger *log.Logger, authConfig auth.AuthConfig) (string, error) {
	if id == "" {
		return "", fmt.Errorf("node ID not set")
	}
	// The specified source IP is optional
	var dialer *net.Dialer
	dialer = &net.Dialer{}
	if srcIP != "" {
		netAddr := &net.TCPAddr{
			IP:   net.ParseIP(srcIP),
			Port: 0,
		}
		dialer = &net.Dialer{LocalAddr: netAddr}
	}
	// Join using IP address, as that is what Hashicorp Raft works in.
	//resv, err := net.ResolveTCPAddr("tcp", addr)
	//if err != nil {
	//	return "", err
	//}

	// Check for protocol scheme, and insert default if necessary.
	fullAddr := utils.NormalizeAddr(fmt.Sprintf("%s/join", joinAddr))

	// Create and configure the client to connect to the other node.
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
		Dial:            dialer.Dial,
	}
	client := &http.Client{Transport: tr}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for {
		b, err := json.Marshal(map[string]interface{}{
			"id":    id,
			"addr":  addr,
			"voter": voter,
			"meta":  meta,
		})
		if err != nil {
			return "", err
		}

		// Attempt to join.
		req, err := http.NewRequest(http.MethodPost, fullAddr, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application-type/json")
		if err != nil {
			return "", err
		}
		switch authConfig.AuthType {
		case auth.Basic:
			req.SetBasicAuth(authConfig.Username, authConfig.Password)
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return fullAddr, nil
		case http.StatusMovedPermanently:
			fullAddr = resp.Header.Get("location")
			if fullAddr == "" {
				return "", fmt.Errorf("failed to join, invalid redirect received")
			}
			continue
		case http.StatusBadRequest:
			// One possible cause is that the target server is listening for HTTPS, but a HTTP
			// attempt was made. Switch the protocol to HTTPS, and try again. This can happen
			// when using the Disco service, since it doesn't record information about which
			// protocol a registered node is actually using.
			if strings.HasPrefix(fullAddr, "https://") {
				// It's already HTTPS, give up.
				return "", fmt.Errorf("failed to join, node returned: %s: (%s)", resp.Status, string(b))
			}

			logger.Print("join via HTTP failed, trying via HTTPS")
			fullAddr = utils.EnsureHTTPS(fullAddr)
			continue
		default:
			return "", fmt.Errorf("failed to join, node returned: %s: (%s)", resp.Status, string(b))
		}
	}
}
