// Copyright 2022 The casbin-mesh Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

type Config struct {
	enableAuth             bool
	rootUsername           string
	rootPassword           string
	raftAddr               string
	raftAdv                string
	joinSrcIP              string
	x509CACert             string
	x509Cert               string
	x509Key                string
	nodeID                 string
	joinAddr               string
	joinAttempts           int
	joinInterval           string
	noVerify               bool
	pprofEnabled           bool
	raftLogLevel           string
	raftNonVoter           bool
	raftSnapThreshold      uint64
	raftSnapInterval       string
	raftLeaderLeaseTimeout string
	raftHeartbeatTimeout   string
	raftElectionTimeout    string
	raftApplyTimeout       string
	raftOpenTimeout        string
	raftWaitForLeader      bool
	raftShutdownOnRemove   bool
	compressionSize        int
	compressionBatch       int
	showVersion            bool
	cpuProfile             string
	memProfile             string
	encrypt                bool
	dataPath               string
}

func parseFlags() (cfg *Config) {
	flag.BoolVar(&cfg.enableAuth, "enable-basic", false, "Enable Basic Auth")
	flag.StringVar(&cfg.rootUsername, "root-username", "root", "Root Account Username")
	flag.StringVar(&cfg.rootPassword, "root-password", "root", "Root Account Password")
	flag.StringVar(&cfg.nodeID, "node-id", "", "Unique name for node. If not set, set to hostname")
	flag.StringVar(&cfg.raftAddr, "raft-address", "localhost:4002", "Raft communication bind address, supports multiple addresses by commas")
	flag.StringVar(&cfg.raftAdv, "raft-advertise-address", "", "Advertised Raft communication address. If not set, same as Raft bind")
	flag.StringVar(&cfg.joinSrcIP, "join-source-ip", "", "Set source IP address during Join request")
	flag.BoolVar(&cfg.encrypt, "tls-encrypt", false, "Enable encryption")
	flag.StringVar(&cfg.x509CACert, "endpoint-ca-cert", "", "Path to root X.509 certificate for API endpoint")
	flag.StringVar(&cfg.x509Cert, "endpoint-cert", "", "Path to X.509 certificate for API endpoint")
	flag.StringVar(&cfg.x509Key, "endpoint-key", "", "Path to X.509 private key for API endpoint")
	flag.BoolVar(&cfg.noVerify, "endpoint-no-verify", false, "Skip verification of remote HTTPS cert when joining cluster")
	flag.StringVar(&cfg.joinAddr, "join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	flag.IntVar(&cfg.joinAttempts, "join-attempts", 5, "Number of join attempts to make")
	flag.StringVar(&cfg.joinInterval, "join-interval", "5s", "Period between join attempts")
	flag.BoolVar(&cfg.pprofEnabled, "pprof", true, "Serve pprof data on API server")
	flag.BoolVar(&cfg.showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&cfg.raftNonVoter, "raft-non-voter", false, "Configure as non-voting node")
	flag.StringVar(&cfg.raftHeartbeatTimeout, "raft-timeout", "1s", "Raft heartbeat timeout")
	flag.StringVar(&cfg.raftElectionTimeout, "raft-election-timeout", "1s", "Raft election timeout")
	flag.StringVar(&cfg.raftApplyTimeout, "raft-apply-timeout", "10s", "Raft apply timeout")
	flag.StringVar(&cfg.raftOpenTimeout, "raft-open-timeout", "120s", "Time for initial Raft logs to be applied. Use 0s duration to skip wait")
	flag.BoolVar(&cfg.raftWaitForLeader, "raft-leader-wait", true, "Node waits for a leader before answering requests")
	flag.Uint64Var(&cfg.raftSnapThreshold, "raft-snap", 8192, "Number of outstanding log entries that trigger snapshot")
	flag.StringVar(&cfg.raftSnapInterval, "raft-snap-int", "30s", "Snapshot threshold check interval")
	flag.StringVar(&cfg.raftLeaderLeaseTimeout, "raft-leader-lease-timeout", "0s", "Raft leader lease timeout. Use 0s for Raft default")
	flag.BoolVar(&cfg.raftShutdownOnRemove, "raft-remove-shutdown", false, "Shutdown Raft if node removed")
	flag.StringVar(&cfg.raftLogLevel, "raft-log-level", "INFO", "Minimum log level for Raft module")
	flag.IntVar(&cfg.compressionSize, "compression-size", 150, "Request query size for compression attempt")
	flag.IntVar(&cfg.compressionBatch, "compression-batch", 5, "Request batch threshold for compression attempt")
	flag.StringVar(&cfg.cpuProfile, "cpu-profile", "", "Path to file for CPU profiling information")
	flag.StringVar(&cfg.memProfile, "mem-profile", "", "Path to file for memory profiling information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", desc)
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <data directory>\n", name)
		flag.PrintDefaults()
	}
	flag.Parse()
	if cfg.showVersion {
		msg := fmt.Sprintf("%s %s %s %s (compiler %s)",
			name, runtime.GOOS, runtime.GOARCH, runtime.Version(), runtime.Compiler)
		errorExit(0, msg)
	}

	// Ensure the data path is set.
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "fatal: no data directory set\n")
		os.Exit(1)
	}

	// Ensure no args come after the data directory.
	if flag.NArg() > 1 {
		fmt.Fprintf(os.Stderr, "fatal: arguments after data directory are not accepted\n")
		os.Exit(1)
	}
	cfg.dataPath = flag.Arg(0)

	return
}

func errorExit(code int, msg string) {
	if code != 0 {
		fmt.Fprintf(os.Stderr, "fatal: ")
	}
	fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", msg))
	os.Exit(code)
}
