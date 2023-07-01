// Copyright 2022 The Casbin Mesh Authors.
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

package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/casbin/casbin-mesh/server/auth"
	"github.com/casbin/casbin-mesh/server/cluster"
	"github.com/casbin/casbin-mesh/server/core"
	"github.com/casbin/casbin-mesh/server/store"
	"github.com/casbin/casbin-mesh/server/transport/tcp"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"gopkg.in/yaml.v3"
)

func New(cfg *Config) (close func() error) {
	// Configure logging and pump out initial message.
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
	log.SetPrefix(fmt.Sprintf("[%s] ", name))
	log.Printf("%s, target architecture is %s, operating system target is %s", runtime.Version(), runtime.GOARCH, runtime.GOOS)
	log.Printf("launch command: %s", strings.Join(os.Args, " "))

	if cfg.configPath != "" {
		configFile, err := os.ReadFile(cfg.configPath)
		if err != nil {
			log.Fatalf("Error reading YAML file: %v", err)
		}

		var config Config
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("Error parsing YAML: %v", err)
		}
	}

	// Start requested profiling.
	startProfile(cfg.cpuProfile, cfg.memProfile)

	httpLn, _, err := newListener(cfg.serverHTTPAddress, cfg.serverHTTPAdvertiseAddress, cfg.encrypt, cfg.x509Key, cfg.x509Cert)
	if err != nil {
		log.Fatalf("failed to create HTTP listener: %s", err.Error())
	}

	grpcLn, _, err := newListener(cfg.serverGRPCAddress, cfg.serverGPRCAdvertiseAddress, cfg.encrypt, cfg.x509Key, cfg.x509Cert)
	if err != nil {
		log.Fatalf("failed to create gRPC listener: %s", err.Error())
	}

	raftLn, raftAdv, err := newListener(cfg.raftAddr, cfg.raftAdv, cfg.encrypt, cfg.x509Key, cfg.x509Cert)
	if err != nil {
		log.Fatalf("failed to create raft listener: %s", err.Error())
	}

	var raftTransport *tcp.Transport
	if cfg.encrypt {
		raftTransport = tcp.NewTransportFromListener(raftLn, true, cfg.noVerify, raftAdv)
	} else {
		raftTransport = tcp.NewTransportFromListener(raftLn, false, false, raftAdv)
	}

	// Create and open the store.
	cfg.dataPath, err = filepath.Abs(cfg.dataPath)
	if err != nil {
		log.Fatalf("failed to determine absolute data path: %s", err.Error())
	}

	authType := auth.Noop
	var credentialsStore *auth.CredentialsStore
	if cfg.enableAuth {
		log.Println("auth type basic")
		authType = auth.Basic
		credentialsStore = auth.NewCredentialsStore()
		err = credentialsStore.Add(cfg.rootUsername, cfg.rootPassword)
		if err != nil {
			log.Fatalf("failed to init credentialsStore:%s", err.Error())
		}
	}

	str := store.New(raftTransport, &store.StoreConfig{
		Dir:              cfg.dataPath,
		ID:               idOrRaftAddr(cfg),
		AuthType:         authType,
		CredentialsStore: credentialsStore,
	})

	// Set optional parameters on store.
	str.RaftLogLevel = cfg.raftLogLevel
	str.ShutdownOnRemove = cfg.raftShutdownOnRemove
	str.SnapshotThreshold = cfg.raftSnapThreshold
	str.SnapshotInterval, err = time.ParseDuration(cfg.raftSnapInterval)
	if err != nil {
		log.Fatalf("failed to parse Raft Snapsnot interval %s: %s", cfg.raftSnapInterval, err.Error())
	}
	str.LeaderLeaseTimeout, err = time.ParseDuration(cfg.raftLeaderLeaseTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft Leader lease timeout %s: %s", cfg.raftLeaderLeaseTimeout, err.Error())
	}
	str.HeartbeatTimeout, err = time.ParseDuration(cfg.raftHeartbeatTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft heartbeat timeout %s: %s", cfg.raftHeartbeatTimeout, err.Error())
	}
	str.ElectionTimeout, err = time.ParseDuration(cfg.raftElectionTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft election timeout %s: %s", cfg.raftElectionTimeout, err.Error())
	}
	str.ApplyTimeout, err = time.ParseDuration(cfg.raftApplyTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft apply timeout %s: %s", cfg.raftApplyTimeout, err.Error())
	}

	// Any prexisting node state?
	var enableBootstrap bool
	isNew := store.IsNewNode(cfg.dataPath)
	if isNew {
		log.Printf("no preexisting node state detected in %s, node may be bootstrapping", cfg.dataPath)
		enableBootstrap = true // New node, so we may be bootstrapping
	} else {
		log.Printf("preexisting node state detected in %s", cfg.dataPath)
	}

	// Determine join addresses
	var joins []string
	joins, err = determineJoinAddresses(cfg)
	if err != nil {
		log.Fatalf("unable to determine join addresses: %s", err.Error())
	}

	// Supplying join addresses means bootstrapping a new cluster won't
	// be required.
	if len(joins) > 0 {
		enableBootstrap = false
		log.Println("join addresses specified, node is not bootstrapping")
	} else {
		log.Println("no join addresses set")
	}

	// Join address supplied, but we don't need them!
	if !isNew && len(joins) > 0 {
		log.Println("node is already member of cluster, ignoring join addresses")
	}

	// Now, open store.
	if err := str.Open(enableBootstrap); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	// Prepare metadata for join command.
	apiAdv := raftAdv
	apiProto := "http"
	if cfg.x509Cert != "" {
		apiProto = "https"
	}
	meta := map[string]string{
		"api_addr":  apiAdv,
		"api_proto": apiProto,
	}

	// Execute any requested join operation.
	if len(joins) > 0 && isNew {
		log.Println("join addresses are:", joins)
		advAddr := cfg.raftAddr
		if cfg.raftAdv != "" {
			advAddr = cfg.raftAdv
		}

		joinDur, err := time.ParseDuration(cfg.joinInterval)
		if err != nil {
			log.Fatalf("failed to parse Join interval %s: %s", cfg.joinInterval, err.Error())
		}

		tlsConfig := tls.Config{InsecureSkipVerify: cfg.noVerify}
		if cfg.x509CACert != "" {
			asn1Data, err := ioutil.ReadFile(cfg.x509CACert)
			if err != nil {
				log.Fatalf("ioutil.ReadFile failed: %s", err.Error())
			}
			tlsConfig.RootCAs = x509.NewCertPool()
			ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(asn1Data))
			if !ok {
				log.Fatalf("failed to parse root CA certificate(s) in %q", cfg.x509CACert)
			}
		}

		if j, err := cluster.Join(cfg.joinSrcIP, joins, str.ID(), advAddr, !cfg.raftNonVoter, meta,
			cfg.joinAttempts, joinDur, &tlsConfig, auth.AuthConfig{AuthType: authType, Username: cfg.rootUsername, Password: cfg.rootPassword}); err != nil {
			log.Fatalf("failed to join cluster at %s: %s", joins, err.Error())
		} else {
			log.Println("successfully joined cluster at", j)
		}

	}

	// Wait until the store is in full consensus.
	if err := waitForConsensus(str, cfg); err != nil {
		log.Fatalf(err.Error())
	}
	// Init Auth Enforce
	if isNew && cfg.enableAuth {
		if err := str.InitAuth(context.TODO(), cfg.rootUsername); err != nil {
			log.Printf("failed to init auth: %s", err.Error())
		}
	}
	// This may be a standalone server. In that case set its own metadata.
	if err := str.SetMetadata(meta); err != nil && err != store.ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		log.Fatalf("failed to set store metadata: %s", err.Error())
	}

	c := core.New(str)
	//Start the HTTP API server.
	if err = startHTTPService(c, httpLn); err != nil {
		log.Fatalf("failed to start HTTP server: %s", err.Error())
	}
	var grpcCloser func()
	if grpcCloser, err = startGrpcService(c, grpcLn); err != nil {
		log.Fatalf("failed to start grpc server: %s", err.Error())
	}

	log.Println("node is ready")

	close = func() error {
		if err := str.Close(true); err != nil {
			log.Printf("failed to close store: %s", err.Error())
		}
		_ = httpLn.Close()
		grpcCloser()
		stopProfile()
		log.Println("casbin-mesh server stopped")

		return err
	}
	return close
}

func RaftRPCMatcher() cmux.Matcher {
	return func(r io.Reader) bool {
		br := bufio.NewReader(&io.LimitedReader{R: r, N: 1})
		byt, err := br.ReadByte()
		if err != nil {
			log.Printf("Raft RPC Unmatched incoming: %s\n", err)
			return false
		}
		switch byt {
		case 0, 1, 2, 3:
			return true
		}
		return false
	}
}

func determineJoinAddresses(cfg *Config) ([]string, error) {
	//raftAdv := httpAddr
	//if httpAdv != "" {
	//	raftAdv = httpAdv
	//}

	var addrs []string
	if cfg.joinAddr != "" {
		// Explicit join addresses are first priority.
		addrs = strings.Split(cfg.joinAddr, ",")
	}

	return addrs, nil
}

func waitForConsensus(str *store.Store, cfg *Config) error {
	openTimeout, err := time.ParseDuration(cfg.raftOpenTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Raft open timeout %s: %s", cfg.raftOpenTimeout, err.Error())
	}
	if _, err := str.WaitForLeader(openTimeout); err != nil {
		if cfg.raftWaitForLeader {
			return fmt.Errorf("leader did not appear within timeout: %s", err.Error())
		}
		log.Println("ignoring error while waiting for leader")
	}
	if openTimeout != 0 {
		if err := str.WaitForApplied(openTimeout); err != nil {
			return fmt.Errorf("log was not fully applied within timeout: %s", err.Error())
		}
	} else {
		log.Println("not waiting for logs to be applied")
	}
	return nil
}

func startHTTPService(c core.Core, ln net.Listener) error {
	httpd := core.NewHttpService(c)
	handler := cors.AllowAll().Handler(httpd)
	go func() {
		err := http.Serve(ln, handler)
		if err != nil {
			log.Println("HTTP service Serve() returned:", err.Error())
		}
	}()

	return nil
}

func startGrpcService(c core.Core, ln net.Listener) (close func(), err error) {
	grpcd := core.NewGrpcService(c)
	go func() {
		err := grpcd.Serve(ln)
		if err != nil {
			log.Println("Grpc service Serve() returned:", err.Error())
		}
	}()
	close = func() {
		grpcd.Stop()
	}
	return close, nil
}

func idOrRaftAddr(cfg *Config) string {
	if cfg.nodeID != "" {
		return cfg.nodeID
	}
	if cfg.raftAdv == "" {
		return cfg.raftAddr
	}
	return cfg.raftAdv
}

// prof stores the file locations of active profiles.
var prof struct {
	cpu *os.File
	mem *os.File
}

// startProfile initializes the CPU and memory profile, if specified.
func startProfile(cpuprofile, memprofile string) {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatalf("failed to create CPU profile file at %s: %s", cpuprofile, err.Error())
		}
		log.Printf("writing CPU profile to: %s\n", cpuprofile)
		prof.cpu = f
		pprof.StartCPUProfile(prof.cpu)
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatalf("failed to create memory profile file at %s: %s", cpuprofile, err.Error())
		}
		log.Printf("writing memory profile to: %s\n", memprofile)
		prof.mem = f
		runtime.MemProfileRate = 4096
	}
}

// stopProfile closes the CPU and memory profiles if they are running.
func stopProfile() {
	if prof.cpu != nil {
		pprof.StopCPUProfile()
		prof.cpu.Close()
		log.Println("CPU profiling stopped")
	}
	if prof.mem != nil {
		pprof.Lookup("heap").WriteTo(prof.mem, 0)
		prof.mem.Close()
		log.Println("memory profiling stopped")
	}
}

func newListener(address string, advertiseAddress string, encrypt bool, keyFile string, certFile string) (net.Listener, string, error) {
	listenerAddresses := strings.Split(address, ",")
	if len(listenerAddresses) == 0 {
		log.Fatal("fatal: bind-address cannot empty")
	}

	advAddr := listenerAddresses[0]
	if advertiseAddress != "" {
		advAddr = advertiseAddress
	}

	// Create peer communication network layer.
	var lns []net.Listener
	for _, address := range listenerAddresses {
		if encrypt {
			log.Printf("enabling encryption with cert: %s, key: %s", certFile, keyFile)
			cfg, err := tcp.CreateTLSConfig(certFile, keyFile)
			if err != nil {
				log.Fatalf("failed to create tls config: %s", err.Error())
			}
			ln, err := tls.Listen("tcp", address, cfg)
			if err != nil {
				log.Fatalf("failed to open internode network layer: %s", err.Error())
			}
			lns = append(lns, ln)
		} else {
			ln, err := net.Listen("tcp", advAddr)
			if err != nil {
				log.Fatalf("failed to open internode network layer: %s", err.Error())
			}
			lns = append(lns, ln)
		}
	}

	ln, err := cluster.NewListener(lns, advAddr)
	return ln, advAddr, err
}
