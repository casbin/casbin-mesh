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
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/casbin/casbin-mesh/server"
	"github.com/spf13/cobra"
)

const name = `casmesh`
const desc = `casmesh is a lightweight, distributed casbin service, which uses casbin as its engine.`

func main() {
	cfg := server.CmdConfig{}

	// Configure logging and pump out initial message.
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)

	cmd := &cobra.Command{
		Use:   name,
		Short: desc,
		Run: func(cmd *cobra.Command, args []string) {
			if cfg.ShowVersion {
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

			cfg.DataPath = args[0]

			s, err := server.NewServer(&cfg)
			if err != nil {
				fmt.Printf("fatal: faild to new casmesh: %v\n", err)
				os.Exit(1)
				return
			}
			err = s.Start()
			if err != nil {
				fmt.Printf("fatal: faild to start casmesh: %v\n", err)
				os.Exit(1)
				return
			}

			// Block until signalled.
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			<-ctx.Done()
			s.Close()
		},
	}

	cmd.Flags().BoolVar(&cfg.EnableAuth, "enable-basic", false, "Enable Basic Auth")
	cmd.Flags().StringVar(&cfg.RootUsername, "root-username", "root", "Root Account Username")
	cmd.Flags().StringVar(&cfg.RootPassword, "root-password", "root", "Root Account Password")
	cmd.Flags().StringVar(&cfg.NodeID, "node-id", "", "Unique name for node. If not set, set to hostname")
	cmd.Flags().StringVar(&cfg.RaftAddr, "raft-address", "localhost:5300", "Raft communication bind address, supports multiple addresses by commas")
	cmd.Flags().StringVar(&cfg.RaftAdv, "raft-advertise-address", "", "Advertised Raft communication address. If not set, same as Raft bind")
	cmd.Flags().StringVar(&cfg.JoinSrcIP, "join-source-ip", "", "Set source IP address during Join request")
	cmd.Flags().BoolVar(&cfg.NoVerify, "endpoint-no-verify", false, "Skip verification of remote HTTPS cert when joining cluster")
	cmd.Flags().StringVar(&cfg.JoinAddr, "join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	cmd.Flags().IntVar(&cfg.JoinAttempts, "join-attempts", 5, "Number of join attempts to make")
	cmd.Flags().StringVar(&cfg.JoinInterval, "join-interval", "5s", "Period between join attempts")
	cmd.Flags().BoolVar(&cfg.ShowVersion, "version", false, "Show version information and exit")
	cmd.Flags().BoolVar(&cfg.RaftNonVoter, "raft-non-voter", false, "Configure as non-voting node")
	cmd.Flags().StringVar(&cfg.RaftHeartbeatTimeout, "raft-timeout", "1s", "Raft heartbeat timeout")
	cmd.Flags().StringVar(&cfg.RaftElectionTimeout, "raft-election-timeout", "1s", "Raft election timeout")
	cmd.Flags().StringVar(&cfg.RaftApplyTimeout, "raft-apply-timeout", "10s", "Raft apply timeout")
	cmd.Flags().StringVar(&cfg.RaftOpenTimeout, "raft-open-timeout", "120s", "Time for initial Raft logs to be applied. Use 0s duration to skip wait")
	cmd.Flags().BoolVar(&cfg.RaftWaitForLeader, "raft-leader-wait", true, "Node waits for a leader before answering requests")
	cmd.Flags().Uint64Var(&cfg.RaftSnapThreshold, "raft-snap", 8192, "Number of outstanding log entries that trigger snapshot")
	cmd.Flags().StringVar(&cfg.RaftSnapInterval, "raft-snap-int", "30s", "Snapshot threshold check interval")
	cmd.Flags().StringVar(&cfg.RaftLeaderLeaseTimeout, "raft-leader-lease-timeout", "0s", "Raft leader lease timeout. Use 0s for Raft default")
	cmd.Flags().BoolVar(&cfg.RaftShutdownOnRemove, "raft-remove-shutdown", false, "Shutdown Raft if node removed")
	cmd.Flags().StringVar(&cfg.RaftLogLevel, "raft-log-level", "INFO", "Minimum log level for Raft module")
	cmd.Flags().IntVar(&cfg.CompressionSize, "compression-size", 150, "Request query size for compression attempt")
	cmd.Flags().IntVar(&cfg.CompressionBatch, "compression-batch", 5, "Request batch threshold for compression attempt")
	cmd.Flags().StringVar(&cfg.ConfigPath, "config", "", "Path to a configuration file")
	cmd.Flags().StringVar(&cfg.ServerHTTPAddress, "server-http-address", "localhost:5200", "HTTP communication bind address, supports multiple addresses by commas")
	cmd.Flags().StringVar(&cfg.ServerHTTPAdvertiseAddress, "server-http-advertise-address", "", "Advertised HTTP communication address. If not set, same as HTTP bind")
	cmd.Flags().StringVar(&cfg.ServerGRPCAddress, "server-grpc-address", "localhost:5201", "gRPC communication bind address, supports multiple addresses by commas")
	cmd.Flags().StringVar(&cfg.ServerGPRCAdvertiseAddress, "server-grpc-advertise-address", "", "gRPC API communication address. If not set, same as gRPC bind")
	cmd.Flags().StringVar(&cfg.RaftCAFile, "raft-ca-cert", "", "Path to root certificate for Raft server and client")
	cmd.Flags().StringVar(&cfg.RaftCertFile, "raft-cert", "", "Path to X.509 certificate for Raft server and client")
	cmd.Flags().StringVar(&cfg.RaftKeyFile, "raft-key", "", "Path to X.509 private key for Raft server and client")
	cmd.Flags().BoolVar(&cfg.RaftTLSEnabled, "raft-tls-encrypt", false, "Enable TLS encryption for Raft server and client")
	cmd.Flags().StringVar(&cfg.ServerCAFile, "server-ca-cert", "", "Path to root certificate for HTTP and gRPC server")
	cmd.Flags().StringVar(&cfg.ServerCertFile, "server-cert", "", "Path to X.509 certificate for HTTP and gRPC server")
	cmd.Flags().StringVar(&cfg.ServerKeyFile, "server-key", "", "Path to X.509 private key for HTTP and gRPC server")
	cmd.Flags().BoolVar(&cfg.ServerTLSEnabled, "server-tls-encrypt", false, "Enable TLS encryption for HTTP and gRPC server")
	cmd.Flags().StringVar(&cfg.PprofAddress, "http-pprof", "", "Start the pprof server on HTTP")

	// tls
	cmd.Flags().BoolVar(&cfg.Encrypt, "tls-encrypt", false, "Enable encryption")
	cmd.Flags().StringVar(&cfg.X509CACert, "endpoint-ca-cert", "", "Path to root X.509 certificate for API endpoint")
	cmd.Flags().StringVar(&cfg.X509Cert, "endpoint-cert", "", "Path to X.509 certificate for API endpoint")
	cmd.Flags().StringVar(&cfg.X509Key, "endpoint-key", "", "Path to X.509 private key for API endpoint")
	_ = cmd.Flags().MarkDeprecated("tls-encrypt", "use --raft-tls-encrypt and --server-tls-encrypt instead")
	_ = cmd.Flags().MarkDeprecated("endpoint-ca-cert", "use --raft-ca-cert and --server-ca-cert instead")
	_ = cmd.Flags().MarkDeprecated("endpoint-cert", "use --raft-cert and --server-cert instead")
	_ = cmd.Flags().MarkDeprecated("endpoint-key", "use --raft-key and --server-key instead")

	// pprof
	cmd.Flags().StringVar(&cfg.CpuProfile, "cpu-profile", "", "Path to file for CPU profiling information")
	cmd.Flags().StringVar(&cfg.MemProfile, "mem-profile", "", "Path to file for memory profiling information")
	cmd.Flags().BoolVar(&cfg.PprofEnabled, "pprof", true, "Serve pprof data on API server")
	_ = cmd.Flags().MarkDeprecated("cpu-profile", "use --http-pprof instead")
	_ = cmd.Flags().MarkDeprecated("mem-profile", "use --http-pprof instead")
	_ = cmd.Flags().MarkDeprecated("pprof", "use --http-pprof instead")

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
