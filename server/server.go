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
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	runtimePprof "runtime/pprof"
	"strings"
	"sync/atomic"
	"time"

	"github.com/casbin/casbin-mesh/server/utils"
	"go.uber.org/zap"

	"github.com/casbin/casbin-mesh/server/auth"
	"github.com/casbin/casbin-mesh/server/cluster"
	"github.com/casbin/casbin-mesh/server/core"
	"github.com/casbin/casbin-mesh/server/store"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

type Server struct {
	cfg       *CmdConfig
	grpcd     *grpc.Server
	str       *store.Store
	strConfig *store.StoreConfig
	pprofLn   net.Listener
	logger    *zap.Logger
	closed    atomic.Bool
}

func NewServer(cfg *CmdConfig) (*Server, error) {
	logger, err := utils.NewZapLogger(zap.InfoLevel)
	if err != nil {
		return nil, err
	}

	server := &Server{
		logger: logger,
		cfg:    cfg,
	}

	if cfg.DataPath == "" {
		return nil, errors.New("no data directory set")
	}

	cfg.DataPath, err = filepath.Abs(cfg.DataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to determine absolute data path: %w", err)
	}

	server.strConfig = &store.StoreConfig{
		Dir:    cfg.DataPath,
		ID:     idOrRaftAddr(cfg),
		Logger: server.logger,
	}
	server.strConfig.SnapshotInterval, err = time.ParseDuration(cfg.RaftSnapInterval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Raft Snapsnot interval: %w", err)
	}
	server.strConfig.LeaderLeaseTimeout, err = time.ParseDuration(cfg.RaftLeaderLeaseTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Raft Leader lease timeout: %w", err)
	}
	server.strConfig.HeartbeatTimeout, err = time.ParseDuration(cfg.RaftHeartbeatTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Raft heartbeat timeout: %w", err)
	}
	server.strConfig.ElectionTimeout, err = time.ParseDuration(cfg.RaftElectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Raft election timeout: %w", err)
	}
	server.strConfig.ApplyTimeout, err = time.ParseDuration(cfg.RaftApplyTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Raft apply timeout: %w", err)
	}

	server.strConfig.RaftLogLevel = cfg.RaftLogLevel
	server.strConfig.ShutdownOnRemove = cfg.RaftShutdownOnRemove
	server.strConfig.SnapshotThreshold = cfg.RaftSnapThreshold

	return server, nil
}

func (s *Server) Start() error {
	cfg := s.cfg
	s.logger.Info("runtime info", zap.String("go", runtime.Version()), zap.String("os", runtime.GOOS), zap.String("arch", runtime.GOARCH))
	s.logger.Debug("server config", zap.Any("config", cfg))

	if cfg.ConfigPath != "" {
		configFile, err := os.ReadFile(cfg.ConfigPath)
		if err != nil {
			return err
		}

		var config Config
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			return err
		}
	}

	var pprofLn net.Listener
	var err error
	if cfg.PprofAddress != "" {
		pprofLn, err = net.Listen("tcp", cfg.PprofAddress)
		if err != nil {
			return err
		}

		go func() {
			pprofMux := http.NewServeMux()
			pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
			pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			s.logger.Info("pprof server", zap.String("address", pprofLn.Addr().String()))
			err = http.Serve(pprofLn, pprofMux)
			if err != nil {
				s.logger.Fatal("failed to start pprof server", zap.Error(err))
			}
		}()
	}

	// Start requested profiling.
	err = s.startProfile(cfg.CpuProfile, cfg.MemProfile)
	if err != nil {
		return err
	}

	httpLn, err := s.newListener(cfg.ServerHTTPAddress, cfg.ServerHTTPAdvertiseAddress, cfg.isServerTlsEnabled(), cfg.getServerKeyFile(), cfg.getServerCertFile(), cfg.getServerCAFile(), tls.NoClientCert, true)
	if err != nil {
		return fmt.Errorf("failed to create HTTP listener: %w", err)
	}

	grpcLn, err := s.newListener(cfg.ServerGRPCAddress, cfg.ServerGPRCAdvertiseAddress, cfg.isServerTlsEnabled(), cfg.getServerKeyFile(), cfg.getServerCertFile(), cfg.getServerCAFile(), tls.NoClientCert, true)
	if err != nil {
		return fmt.Errorf("failed to create gPRC listener: %w", err)
	}

	var raftAdv string
	raftLn, err := s.newListener(cfg.RaftAddr, cfg.RaftAdv, cfg.isRaftTlsEnabled(), cfg.getRaftKeyFile(), cfg.getRaftCertFile(), cfg.getRaftCAFile(), tls.RequireAndVerifyClientCert, true)
	if err != nil {
		return fmt.Errorf("failed to create Raft listener: %w", err)
	}

	var raftTransport *store.TcpTransport
	var raftTlsConfig *tls.Config
	if cfg.isRaftTlsEnabled() {
		raftTlsConfig, err = store.CreateTLSConfig(cfg.getRaftKeyFile(), cfg.getRaftCertFile(), cfg.getRaftCAFile(), false)
		if err != nil {
			return fmt.Errorf("failed to create TLS config of Raft: %w", err)
		}
	}
	raftTransport = store.NewTransportFromListener(s.logger, raftLn, raftTlsConfig)

	authType := auth.Noop
	var credentialsStore *auth.CredentialsStore
	if cfg.EnableAuth {
		s.logger.Info("auth type basic")
		authType = auth.Basic
		credentialsStore = auth.NewCredentialsStore()
		err = credentialsStore.Add(cfg.RootUsername, cfg.RootPassword)
		if err != nil {
			return fmt.Errorf("failed to init credentialsStore: %w", err)
		}
	}

	s.strConfig.AuthType = authType
	s.strConfig.CredentialsStore = credentialsStore
	s.str = store.New(raftTransport, s.strConfig)

	// Any prexisting node state?
	var enableBootstrap bool
	isNew := store.IsNewNode(cfg.DataPath)
	if isNew {
		s.logger.Info("no preexisting node state detected, node may be bootstrapping")
		enableBootstrap = true // New node, so we may be bootstrapping
	} else {
		s.logger.Info("preexisting node state detected")
	}

	// Determine join addresses
	var joins []string
	joins, err = determineJoinAddresses(cfg)
	if err != nil {
		s.logger.Info("unable to determine join addresses", zap.Error(err))
	}

	// Supplying join addresses means bootstrapping a new cluster won't
	// be required.
	if len(joins) > 0 {
		enableBootstrap = false
		s.logger.Info("join addresses specified, node is not bootstrapping")
	} else {
		s.logger.Info("no join addresses set")
	}

	// Join address supplied, but we don't need them!
	if !isNew && len(joins) > 0 {
		s.logger.Info("node is already member of cluster, ignoring join addresses")
	}

	// Now, open store.
	if err := s.str.Open(enableBootstrap); err != nil {
		return fmt.Errorf("failed to open store: %w", err)
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
		s.logger.Info("join addresses", zap.Any("join-address", joins))
		advAddr := cfg.RaftAddr
		if cfg.RaftAdv != "" {
			advAddr = cfg.RaftAdv
		}

		joinDur, err := time.ParseDuration(cfg.JoinInterval)
		if err != nil {
			return fmt.Errorf("failed to parse Join interval: %w", err)
		}

		tlsConfig := tls.Config{InsecureSkipVerify: cfg.NoVerify}
		if cfg.X509CACert != "" {
			asn1Data, err := ioutil.ReadFile(cfg.X509CACert)
			if err != nil {
				return fmt.Errorf("ioutil.ReadFile failed: %w", err)
			}
			tlsConfig.RootCAs = x509.NewCertPool()
			ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(asn1Data))
			if !ok {
				return errors.New("failed to parse root CA certificate(s)")
			}
		}

		if j, err := cluster.Join(s.logger, cfg.JoinSrcIP, joins, s.str.ID(), advAddr, !cfg.RaftNonVoter, meta,
			cfg.JoinAttempts, joinDur, &tlsConfig, auth.AuthConfig{AuthType: authType, Username: cfg.RootUsername, Password: cfg.RootPassword}); err != nil {
			return fmt.Errorf("failed to join cluster: %w", err)
		} else {
			s.logger.Info("successfully joined cluster", zap.Any("join-address", j))
		}

	}

	// Wait until the store is in full consensus.
	if err := s.waitForConsensus(s.str, cfg); err != nil {
		return fmt.Errorf("waitForConsensus error: %w", err)
	}
	// Init Auth Enforce
	if isNew && cfg.EnableAuth {
		if err := s.str.InitAuth(context.TODO(), cfg.RootUsername); err != nil {
			return fmt.Errorf("failed to init auth: %w", err)
		}
	}
	// This may be a standalone s. In that case set its own metadata.
	if err := s.str.SetMetadata(meta); err != nil && err != store.ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		return fmt.Errorf("failed to set store metadata: %w", err)
	}

	c := core.New(s.str)
	s.startHTTPService(c, httpLn)
	s.startGrpcService(c, grpcLn)

	s.logger.Info("node is ready", zap.Error(err))
	return nil
}

func (s *Server) Close() {
	if !s.closed.CompareAndSwap(false, true) {
		return
	}

	if s.str != nil {
		if err := s.str.Close(true); err != nil {
			s.logger.Error("failed to close store", zap.Error(err))
		}
	}

	if s.grpcd != nil {
		s.grpcd.Stop()
	}

	if s.pprofLn != nil {
		_ = s.pprofLn.Close()
	}

	if prof.cpu != nil {
		runtimePprof.StopCPUProfile()
		_ = prof.cpu.Close()
		s.logger.Info("CPU profiling stopped")
	}
	if prof.mem != nil {
		_ = runtimePprof.Lookup("heap").WriteTo(prof.mem, 0)
		_ = prof.mem.Close()
		s.logger.Info("memory profiling stopped")
	}

	s.logger.Info("casbin-mesh server stopped")
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

func (s *Server) waitForConsensus(str *store.Store, cfg *CmdConfig) error {
	openTimeout, err := time.ParseDuration(cfg.RaftOpenTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Raft open timeout %s: %s", cfg.RaftOpenTimeout, err.Error())
	}
	if _, err := str.WaitForLeader(openTimeout); err != nil {
		if cfg.RaftWaitForLeader {
			return fmt.Errorf("leader did not appear within timeout: %s", err.Error())
		}
		s.logger.Info("ignoring error while waiting for leader")
	}
	if openTimeout != 0 {
		if err := str.WaitForApplied(openTimeout); err != nil {
			return fmt.Errorf("log was not fully applied within timeout: %s", err.Error())
		}
	} else {
		s.logger.Info("not waiting for logs to be applied")
	}
	return nil
}

func (s *Server) startHTTPService(c core.Core, ln net.Listener) {
	httpd := core.NewHttpService(c)
	handler := cors.AllowAll().Handler(httpd)
	go func() {
		err := http.Serve(ln, handler)
		if err != nil {
			s.logger.Fatal("HTTP service Serve() returned", zap.Error(err))
		}
	}()
}

func (s *Server) startGrpcService(c core.Core, ln net.Listener) {
	grpcd := core.NewGrpcService(c)
	go func() {
		err := grpcd.Serve(ln)
		if err != nil {
			s.logger.Fatal("Grpc service Serve() returned", zap.Error(err))
		}
	}()
	s.grpcd = grpcd
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
func (s *Server) startProfile(cpuprofile, memprofile string) error {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			return fmt.Errorf("failed to create CPU profile file: %w", err)
		}
		s.logger.Info("writing CPU profile to path", zap.String("path", cpuprofile))
		prof.cpu = f
		_ = runtimePprof.StartCPUProfile(prof.cpu)
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			return fmt.Errorf("failed to create memory profile: %w", err)
		}
		s.logger.Info("writing memory profile to path", zap.String("path", memprofile))
		prof.mem = f
		runtime.MemProfileRate = 4096
	}

	return nil
}

func (s *Server) newListener(address string, advertiseAddress string, encrypt bool, keyFile string, certFile string, caFile string, clientAuthType tls.ClientAuthType, isServer bool) (net.Listener, error) {
	listenerAddresses := strings.Split(address, ",")
	if len(listenerAddresses) == 0 {
		return nil, errors.New("bind-address cannot empty")
	}

	// Create peer communication network layer.
	var lns []net.Listener
	var ln net.Listener
	var err error
	for _, address := range listenerAddresses {
		if encrypt {
			s.logger.Info("enabling encryption", zap.String("cert", certFile), zap.String("key", keyFile))
			cfg, err := store.CreateTLSConfig(certFile, keyFile, caFile, isServer)
			if err != nil {
				return nil, fmt.Errorf("failed to create tls config: %w", err)
			}
			cfg.ClientAuth = clientAuthType
			ln, err = tls.Listen("tcp", address, cfg)
		} else {
			ln, err = net.Listen("tcp", address)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to open internode network layer: %w", err)
		}
		lns = append(lns, ln)
	}

	return cluster.NewListener(lns, advertiseAddress)
}
