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

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	runtimePprof "runtime/pprof"
	"strings"
	"time"

	"github.com/casbin/casbin-mesh/server/auth"
	"github.com/casbin/casbin-mesh/server/cluster"
	"github.com/casbin/casbin-mesh/server/core"
	"github.com/casbin/casbin-mesh/server/store"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type Casmesh struct {
	grpcd   *grpc.Server
	str     *store.Store
	pprofLn net.Listener
}

func NewCasmesh(cfg *CmdConfig) (*Casmesh, error) {
	casmesh := &Casmesh{}

	log.SetPrefix("[casmesh]")
	log.Printf("%s, target architecture is %s, operating system target is %s", runtime.Version(), runtime.GOARCH, runtime.GOOS)

	if cfg.ConfigPath != "" {
		configFile, err := os.ReadFile(cfg.ConfigPath)
		if err != nil {
			log.Fatalf("Error reading YAML file: %v", err)
		}

		var config Config
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("Error parsing YAML: %v", err)
		}
	}

	var pprofLn net.Listener
	var err error
	if cfg.PprofAddress != "" {
		pprofLn, err = net.Listen("tcp", cfg.PprofAddress)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			pprofMux := http.NewServeMux()
			pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
			pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			log.Printf("pprof server: http://%s\n", pprofLn.Addr())
			err = http.Serve(pprofLn, pprofMux)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	// Start requested profiling.
	startProfile(cfg.CpuProfile, cfg.MemProfile)

	httpLn, _, err := newListener(cfg.ServerHTTPAddress, cfg.ServerHTTPAdvertiseAddress, cfg.isServerTlsEnabled(), cfg.getServerKeyFile(), cfg.getServerCertFile(), cfg.getServerCAFile(), tls.NoClientCert, true)
	if err != nil {
		log.Fatalf("failed to create HTTP listener: %s", err.Error())
	}

	grpcLn, _, err := newListener(cfg.ServerGRPCAddress, cfg.ServerGPRCAdvertiseAddress, cfg.isServerTlsEnabled(), cfg.getServerKeyFile(), cfg.getServerCertFile(), cfg.getServerCAFile(), tls.NoClientCert, true)
	if err != nil {
		log.Fatalf("failed to create gRPC listener: %s", err.Error())
	}

	var raftAdv string
	raftLn, raftAdv, err := newListener(cfg.RaftAddr, cfg.RaftAdv, cfg.isRaftTlsEnabled(), cfg.getRaftKeyFile(), cfg.getRaftCertFile(), cfg.getRaftCAFile(), tls.RequireAndVerifyClientCert, true)
	if err != nil {
		log.Fatalf("failed to create Raft listener: %s", err.Error())
	}

	var raftTransport *store.TcpTransport
	var raftTlsConfig *tls.Config
	if cfg.isRaftTlsEnabled() {
		raftTlsConfig, err = store.CreateTLSConfig(cfg.getRaftKeyFile(), cfg.getRaftCertFile(), cfg.getRaftCAFile(), false)
		if err != nil {
			log.Fatalf("failed to create TLS config of Raft: %s", err.Error())
		}
	}
	raftTransport = store.NewTransportFromListener(raftLn, raftTlsConfig)

	// Create and open the store.
	cfg.DataPath, err = filepath.Abs(cfg.DataPath)
	if err != nil {
		log.Fatalf("failed to determine absolute data path: %s", err.Error())
	}

	authType := auth.Noop
	var credentialsStore *auth.CredentialsStore
	if cfg.EnableAuth {
		log.Println("auth type basic")
		authType = auth.Basic
		credentialsStore = auth.NewCredentialsStore()
		err = credentialsStore.Add(cfg.RootUsername, cfg.RootPassword)
		if err != nil {
			log.Fatalf("failed to init credentialsStore:%s", err.Error())
		}
	}

	casmesh.str = store.New(raftTransport, &store.StoreConfig{
		Dir:              cfg.DataPath,
		ID:               idOrRaftAddr(cfg),
		AuthType:         authType,
		CredentialsStore: credentialsStore,
	})

	// Set optional parameters on store.
	casmesh.str.RaftLogLevel = cfg.RaftLogLevel
	casmesh.str.ShutdownOnRemove = cfg.RaftShutdownOnRemove
	casmesh.str.SnapshotThreshold = cfg.RaftSnapThreshold
	casmesh.str.SnapshotInterval, err = time.ParseDuration(cfg.RaftSnapInterval)
	if err != nil {
		log.Fatalf("failed to parse Raft Snapsnot interval %s: %s", cfg.RaftSnapInterval, err.Error())
	}
	casmesh.str.LeaderLeaseTimeout, err = time.ParseDuration(cfg.RaftLeaderLeaseTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft Leader lease timeout %s: %s", cfg.RaftLeaderLeaseTimeout, err.Error())
	}
	casmesh.str.HeartbeatTimeout, err = time.ParseDuration(cfg.RaftHeartbeatTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft heartbeat timeout %s: %s", cfg.RaftHeartbeatTimeout, err.Error())
	}
	casmesh.str.ElectionTimeout, err = time.ParseDuration(cfg.RaftElectionTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft election timeout %s: %s", cfg.RaftElectionTimeout, err.Error())
	}
	casmesh.str.ApplyTimeout, err = time.ParseDuration(cfg.RaftApplyTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft apply timeout %s: %s", cfg.RaftApplyTimeout, err.Error())
	}

	// Any prexisting node state?
	var enableBootstrap bool
	isNew := store.IsNewNode(cfg.DataPath)
	if isNew {
		log.Printf("no preexisting node state detected in %s, node may be bootstrapping", cfg.DataPath)
		enableBootstrap = true // New node, so we may be bootstrapping
	} else {
		log.Printf("preexisting node state detected in %s", cfg.DataPath)
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
	if err := casmesh.str.Open(enableBootstrap); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	// Prepare metadata for join command.
	apiAdv := raftAdv
	apiProto := "http"
	if cfg.X509Cert != "" {
		apiProto = "https"
	}
	meta := map[string]string{
		"api_addr":  apiAdv,
		"api_proto": apiProto,
	}

	// Execute any requested join operation.
	if len(joins) > 0 && isNew {
		log.Println("join addresses are:", joins)
		advAddr := cfg.RaftAddr
		if cfg.RaftAdv != "" {
			advAddr = cfg.RaftAdv
		}

		joinDur, err := time.ParseDuration(cfg.JoinInterval)
		if err != nil {
			log.Fatalf("failed to parse Join interval %s: %s", cfg.JoinInterval, err.Error())
		}

		tlsConfig := tls.Config{InsecureSkipVerify: cfg.NoVerify}
		if cfg.X509CACert != "" {
			asn1Data, err := ioutil.ReadFile(cfg.X509CACert)
			if err != nil {
				log.Fatalf("ioutil.ReadFile failed: %s", err.Error())
			}
			tlsConfig.RootCAs = x509.NewCertPool()
			ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(asn1Data))
			if !ok {
				log.Fatalf("failed to parse root CA certificate(s) in %q", cfg.X509CACert)
			}
		}

		if j, err := cluster.Join(cfg.JoinSrcIP, joins, casmesh.str.ID(), advAddr, !cfg.RaftNonVoter, meta,
			cfg.JoinAttempts, joinDur, &tlsConfig, auth.AuthConfig{AuthType: authType, Username: cfg.RootUsername, Password: cfg.RootPassword}); err != nil {
			log.Fatalf("failed to join cluster at %s: %s", joins, err.Error())
		} else {
			log.Println("successfully joined cluster at", j)
		}

	}

	// Wait until the store is in full consensus.
	if err := waitForConsensus(casmesh.str, cfg); err != nil {
		log.Fatalf(err.Error())
	}
	// Init Auth Enforce
	if isNew && cfg.EnableAuth {
		if err := casmesh.str.InitAuth(context.TODO(), cfg.RootUsername); err != nil {
			log.Printf("failed to init auth: %s", err.Error())
		}
	}
	// This may be a standalone server. In that case set its own metadata.
	if err := casmesh.str.SetMetadata(meta); err != nil && err != store.ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		log.Fatalf("failed to set store metadata: %s", err.Error())
	}

	c := core.New(casmesh.str)
	//Start the HTTP API server.
	if err = startHTTPService(c, httpLn); err != nil {
		log.Fatalf("failed to start HTTP server: %s", err.Error())
	}
	if err = casmesh.startGrpcService(c, grpcLn); err != nil {
		log.Fatalf("failed to start grpc server: %s", err.Error())
	}

	log.Println("node is ready")

	return casmesh, nil
}

func (c *Casmesh) Close() {
	if c.str != nil {
		if err := c.str.Close(true); err != nil {
			log.Printf("failed to close store: %s", err.Error())
		}
	}

	if c.grpcd != nil {
		c.grpcd.Stop()
	}

	if c.pprofLn != nil {
		_ = c.pprofLn.Close()
	}

	stopProfile()

	log.Println("casbin-mesh server stopped")
}

func determineJoinAddresses(cfg *CmdConfig) ([]string, error) {
	//raftAdv := httpAddr
	//if httpAdv != "" {
	//	raftAdv = httpAdv
	//}

	var addrs []string
	if cfg.JoinAddr != "" {
		// Explicit join addresses are first priority.
		addrs = strings.Split(cfg.JoinAddr, ",")
	}

	return addrs, nil
}

func waitForConsensus(str *store.Store, cfg *CmdConfig) error {
	openTimeout, err := time.ParseDuration(cfg.RaftOpenTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Raft open timeout %s: %s", cfg.RaftOpenTimeout, err.Error())
	}
	if _, err := str.WaitForLeader(openTimeout); err != nil {
		if cfg.RaftWaitForLeader {
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

func (c *Casmesh) startGrpcService(co core.Core, ln net.Listener) (err error) {
	grpcd := core.NewGrpcService(co)
	go func() {
		err := grpcd.Serve(ln)
		if err != nil {
			log.Println("Grpc service Serve() returned:", err.Error())
		}
	}()
	c.grpcd = grpcd
	return err
}

func idOrRaftAddr(cfg *CmdConfig) string {
	if cfg.NodeID != "" {
		return cfg.NodeID
	}
	if cfg.RaftAdv == "" {
		return cfg.RaftAddr
	}
	return cfg.RaftAdv
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
		runtimePprof.StartCPUProfile(prof.cpu)
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
		runtimePprof.StopCPUProfile()
		prof.cpu.Close()
		log.Println("CPU profiling stopped")
	}
	if prof.mem != nil {
		runtimePprof.Lookup("heap").WriteTo(prof.mem, 0)
		prof.mem.Close()
		log.Println("memory profiling stopped")
	}
}

func newListener(address string, advertiseAddress string, encrypt bool, keyFile string, certFile string, caFile string, clientAuthType tls.ClientAuthType, isServer bool) (net.Listener, string, error) {
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
			cfg, err := store.CreateTLSConfig(certFile, keyFile, caFile, isServer)
			if err != nil {
				log.Fatalf("failed to create tls config: %s", err.Error())
			}
			cfg.ClientAuth = clientAuthType
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
