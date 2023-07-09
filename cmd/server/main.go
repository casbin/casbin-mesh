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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/spf13/cobra"
)

const name = `casmesh`
const desc = `casmesh is a lightweight, distributed casbin service, which uses casbin as its engine.`

type Config struct {
	enableAuth                 bool
	rootUsername               string
	rootPassword               string
	raftAddr                   string
	raftAdv                    string
	joinSrcIP                  string
	x509CACert                 string
	x509Cert                   string
	x509Key                    string
	nodeID                     string
	joinAddr                   string
	joinAttempts               int
	joinInterval               string
	noVerify                   bool
	pprofEnabled               bool
	raftLogLevel               string
	raftNonVoter               bool
	raftSnapThreshold          uint64
	raftSnapInterval           string
	raftLeaderLeaseTimeout     string
	raftHeartbeatTimeout       string
	raftElectionTimeout        string
	raftApplyTimeout           string
	raftOpenTimeout            string
	raftWaitForLeader          bool
	raftShutdownOnRemove       bool
	compressionSize            int
	compressionBatch           int
	showVersion                bool
	cpuProfile                 string
	memProfile                 string
	encrypt                    bool
	dataPath                   string
	configPath                 string
	serverHTTPAddress          string
	serverHTTPAdvertiseAddress string
	serverGRPCAddress          string
	serverGPRCAdvertiseAddress string
	raftCAFile                 string
	raftKeyFile                string
	raftCertFile               string
	raftTLSEnabled             bool
	serverCAFile               string
	serverKeyFile              string
	serverCertFile             string
	serverTLSEnabled           bool
}

func (c *Config) getRaftCAFile() string {
	if c.raftCAFile != "" {
		return c.raftCAFile
	}
	return c.x509CACert
}

func (c *Config) getRaftKeyFile() string {
	if c.raftKeyFile != "" {
		return c.raftKeyFile
	}
	return c.x509Key
}

func (c *Config) getRaftCertFile() string {
	if c.raftCertFile != "" {
		return c.raftCertFile
	}
	return c.x509Cert
}

func (c *Config) isRaftTlsEnabled() bool {
	return c.raftTLSEnabled || c.encrypt
}

func (c *Config) getServerCAFile() string {
	if c.serverCAFile != "" {
		return c.serverCAFile
	}
	return c.x509CACert
}

func (c *Config) getServerKeyFile() string {
	if c.serverKeyFile != "" {
		return c.serverKeyFile
	}
	return c.x509Key
}

func (c *Config) getServerCertFile() string {
	if c.serverCertFile != "" {
		return c.serverCertFile
	}
	return c.x509Cert
}

func (c *Config) isServerTlsEnabled() bool {
	return c.serverTLSEnabled || c.encrypt
}

func main() {
	cfg := Config{}

	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.showVersion {
				fmt.Printf("%s %s %s (compiler %s)\n", runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.Compiler)
				os.Exit(0)
				return
			}

			if len(args) < 1 {
				fmt.Printf("fatal: no data directory set\n")
				os.Exit(1)
				return
			}

			// Ensure no args come after the data directory.
			if len(args) > 1 {
				fmt.Printf("fatal: arguments after data directory are not accepted\n")
				os.Exit(1)
				return
			}

			cfg.dataPath = args[0]

			closer := New(&cfg)

			// Block until signalled.
			terminate := make(chan os.Signal, 1)
			signal.Notify(terminate, os.Interrupt)
			<-terminate
			_ = closer()
		},
	}

	cmd.Flags().BoolVar(&cfg.enableAuth, "enable-basic", false, "Enable Basic Auth")
	cmd.Flags().StringVar(&cfg.rootUsername, "root-username", "root", "Root Account Username")
	cmd.Flags().StringVar(&cfg.rootPassword, "root-password", "root", "Root Account Password")
	cmd.Flags().StringVar(&cfg.nodeID, "node-id", "", "Unique name for node. If not set, set to hostname")
	cmd.Flags().StringVar(&cfg.raftAddr, "raft-address", "localhost:5300", "Raft communication bind address, supports multiple addresses by commas")
	cmd.Flags().StringVar(&cfg.raftAdv, "raft-advertise-address", "", "Advertised Raft communication address. If not set, same as Raft bind")
	cmd.Flags().StringVar(&cfg.joinSrcIP, "join-source-ip", "", "Set source IP address during Join request")
	cmd.Flags().BoolVar(&cfg.noVerify, "endpoint-no-verify", false, "Skip verification of remote HTTPS cert when joining cluster")
	cmd.Flags().StringVar(&cfg.joinAddr, "join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	cmd.Flags().IntVar(&cfg.joinAttempts, "join-attempts", 5, "Number of join attempts to make")
	cmd.Flags().StringVar(&cfg.joinInterval, "join-interval", "5s", "Period between join attempts")
	cmd.Flags().BoolVar(&cfg.pprofEnabled, "pprof", true, "Serve pprof data on API server")
	cmd.Flags().BoolVar(&cfg.showVersion, "version", false, "Show version information and exit")
	cmd.Flags().BoolVar(&cfg.raftNonVoter, "raft-non-voter", false, "Configure as non-voting node")
	cmd.Flags().StringVar(&cfg.raftHeartbeatTimeout, "raft-timeout", "1s", "Raft heartbeat timeout")
	cmd.Flags().StringVar(&cfg.raftElectionTimeout, "raft-election-timeout", "1s", "Raft election timeout")
	cmd.Flags().StringVar(&cfg.raftApplyTimeout, "raft-apply-timeout", "10s", "Raft apply timeout")
	cmd.Flags().StringVar(&cfg.raftOpenTimeout, "raft-open-timeout", "120s", "Time for initial Raft logs to be applied. Use 0s duration to skip wait")
	cmd.Flags().BoolVar(&cfg.raftWaitForLeader, "raft-leader-wait", true, "Node waits for a leader before answering requests")
	cmd.Flags().Uint64Var(&cfg.raftSnapThreshold, "raft-snap", 8192, "Number of outstanding log entries that trigger snapshot")
	cmd.Flags().StringVar(&cfg.raftSnapInterval, "raft-snap-int", "30s", "Snapshot threshold check interval")
	cmd.Flags().StringVar(&cfg.raftLeaderLeaseTimeout, "raft-leader-lease-timeout", "0s", "Raft leader lease timeout. Use 0s for Raft default")
	cmd.Flags().BoolVar(&cfg.raftShutdownOnRemove, "raft-remove-shutdown", false, "Shutdown Raft if node removed")
	cmd.Flags().StringVar(&cfg.raftLogLevel, "raft-log-level", "INFO", "Minimum log level for Raft module")
	cmd.Flags().IntVar(&cfg.compressionSize, "compression-size", 150, "Request query size for compression attempt")
	cmd.Flags().IntVar(&cfg.compressionBatch, "compression-batch", 5, "Request batch threshold for compression attempt")
	cmd.Flags().StringVar(&cfg.cpuProfile, "cpu-profile", "", "Path to file for CPU profiling information")
	cmd.Flags().StringVar(&cfg.memProfile, "mem-profile", "", "Path to file for memory profiling information")
	cmd.Flags().StringVar(&cfg.configPath, "config", "", "Path to a configuration file")
	cmd.Flags().StringVar(&cfg.serverHTTPAddress, "server-http-address", "localhost:5200", "HTTP communication bind address, supports multiple addresses by commas")
	cmd.Flags().StringVar(&cfg.serverHTTPAdvertiseAddress, "server-http-advertise-address", "", "Advertised HTTP communication address. If not set, same as HTTP bind")
	cmd.Flags().StringVar(&cfg.serverGRPCAddress, "server-grpc-address", "localhost:5201", "gRPC communication bind address, supports multiple addresses by commas")
	cmd.Flags().StringVar(&cfg.serverGPRCAdvertiseAddress, "server-grpc-advertise-address", "", "gRPC API communication address. If not set, same as gRPC bind")
	cmd.Flags().StringVar(&cfg.raftCAFile, "raft-ca-file", "", "Path to root certificate for Raft server and client")
	cmd.Flags().StringVar(&cfg.raftCertFile, "raft-cert-file", "", "Path to X.509 certificate for Raft server and client")
	cmd.Flags().StringVar(&cfg.raftKeyFile, "raft-key-file", "", "Path to X.509 private key for Raft server and client")
	cmd.Flags().BoolVar(&cfg.raftTLSEnabled, "raft-tls-encrypt", false, "Enable TLS encryption for Raft server and client")
	cmd.Flags().StringVar(&cfg.serverCAFile, "server-ca-file", "", "Path to root certificate for HTTP and gRPC server")
	cmd.Flags().StringVar(&cfg.serverCertFile, "server-cert-file", "", "Path to X.509 certificate for HTTP and gRPC server")
	cmd.Flags().StringVar(&cfg.serverKeyFile, "server-key-file", "", "Path to X.509 private key for HTTP and gRPC server")
	cmd.Flags().BoolVar(&cfg.serverTLSEnabled, "server-tls-encrypt", false, "Enable TLS encryption for HTTP and gRPC server")

	cmd.Flags().BoolVar(&cfg.encrypt, "tls-encrypt", false, "Enable encryption")
	cmd.Flags().StringVar(&cfg.x509CACert, "endpoint-ca-cert", "", "Path to root X.509 certificate for API endpoint")
	cmd.Flags().StringVar(&cfg.x509Cert, "endpoint-cert", "", "Path to X.509 certificate for API endpoint")
	cmd.Flags().StringVar(&cfg.x509Key, "endpoint-key", "", "Path to X.509 private key for API endpoint")
	_ = cmd.Flags().MarkDeprecated("tls-encrypt", "use --raft-tls-encrypt and --server-tls-encrypt instead")
	_ = cmd.Flags().MarkDeprecated("endpoint-ca-cert", "use --raft-ca-file and --server-ca-file instead")
	_ = cmd.Flags().MarkDeprecated("endpoint-cert", "use --raft-cert-file and --server-cert-file instead")
	_ = cmd.Flags().MarkDeprecated("endpoint-key", "use --raft-key-file and --server-key-file instead")

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
