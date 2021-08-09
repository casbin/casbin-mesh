package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/casbin/casbin-mesh/pkg/cluster"
	"github.com/casbin/casbin-mesh/pkg/core"
	"github.com/casbin/casbin-mesh/pkg/store"
	"github.com/casbin/casbin-mesh/pkg/transport/tcp"
	"github.com/soheilhy/cmux"
)

const name = `Casbin-mesh`
const desc = `casbin-mesh is a lightweight, distributed casbin service, which uses casbin as its
engine.`

var (
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
)

func init() {
	flag.StringVar(&nodeID, "node-id", "", "Unique name for node. If not set, set to hostname")
	flag.StringVar(&raftAddr, "raft-address", "localhost:4002", "Raft communication bind address")
	flag.StringVar(&raftAdv, "raft-advertise-address", "", "Advertised Raft communication address. If not set, same as Raft bind")
	flag.StringVar(&joinSrcIP, "join-source-ip", "", "Set source IP address during Join request")
	flag.BoolVar(&encrypt, "tls-encrypt", false, "Enable encryption")
	flag.StringVar(&x509CACert, "endpoint-ca-cert", "", "Path to root X.509 certificate for API endpoint")
	flag.StringVar(&x509Cert, "endpoint-cert", "", "Path to X.509 certificate for API endpoint")
	flag.StringVar(&x509Key, "endpoint-key", "", "Path to X.509 private key for API endpoint")
	flag.BoolVar(&noVerify, "endpoint-no-verify", false, "Skip verification of remote HTTPS cert when joining cluster")
	flag.StringVar(&joinAddr, "join", "", "Comma-delimited list of nodes, through which a cluster can be joined (proto://host:port)")
	flag.IntVar(&joinAttempts, "join-attempts", 5, "Number of join attempts to make")
	flag.StringVar(&joinInterval, "join-interval", "5s", "Period between join attempts")
	flag.BoolVar(&pprofEnabled, "pprof", true, "Serve pprof data on API server")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&raftNonVoter, "raft-non-voter", false, "Configure as non-voting node")
	flag.StringVar(&raftHeartbeatTimeout, "raft-timeout", "1s", "Raft heartbeat timeout")
	flag.StringVar(&raftElectionTimeout, "raft-election-timeout", "1s", "Raft election timeout")
	flag.StringVar(&raftApplyTimeout, "raft-apply-timeout", "10s", "Raft apply timeout")
	flag.StringVar(&raftOpenTimeout, "raft-open-timeout", "120s", "Time for initial Raft logs to be applied. Use 0s duration to skip wait")
	flag.BoolVar(&raftWaitForLeader, "raft-leader-wait", true, "Node waits for a leader before answering requests")
	flag.Uint64Var(&raftSnapThreshold, "raft-snap", 8192, "Number of outstanding log entries that trigger snapshot")
	flag.StringVar(&raftSnapInterval, "raft-snap-int", "30s", "Snapshot threshold check interval")
	flag.StringVar(&raftLeaderLeaseTimeout, "raft-leader-lease-timeout", "0s", "Raft leader lease timeout. Use 0s for Raft default")
	flag.BoolVar(&raftShutdownOnRemove, "raft-remove-shutdown", false, "Shutdown Raft if node removed")
	flag.StringVar(&raftLogLevel, "raft-log-level", "INFO", "Minimum log level for Raft module")
	flag.IntVar(&compressionSize, "compression-size", 150, "Request query size for compression attempt")
	flag.IntVar(&compressionBatch, "compression-batch", 5, "Request batch threshold for compression attempt")
	flag.StringVar(&cpuProfile, "cpu-profile", "", "Path to file for CPU profiling information")
	flag.StringVar(&memProfile, "mem-profile", "", "Path to file for memory profiling information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s\n\n", desc)
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <data directory>\n", name)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
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

	dataPath := flag.Arg(0)
	// Configure logging and pump out initial message.
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stderr)
	log.SetPrefix(fmt.Sprintf("[%s] ", name))
	log.Printf("%s, target architecture is %s, operating system target is %s", runtime.Version(), runtime.GOARCH, runtime.GOOS)
	log.Printf("launch command: %s", strings.Join(os.Args, " "))

	// Start requested profiling.
	startProfile(cpuProfile, memProfile)

	// Create internode network layer.
	var ln net.Listener
	if encrypt {
		log.Printf("enabling encryption with cert: %s, key: %s", x509Cert, x509Key)
		cfg, err := tcp.CreateTLSConfig(x509Cert, x509Key)
		if err != nil {
			log.Fatalf("failed to create tls config: %s", err.Error())
		}
		ln, err = tls.Listen("tcp", raftAddr, cfg)
		if err != nil {
			log.Fatalf("failed to open internode network layer: %s", err.Error())
		}
	} else {
		var err error
		ln, err = net.Listen("tcp", raftAddr)
		if err != nil {
			log.Fatalf("failed to open internode network layer: %s", err.Error())
		}
	}
	mux := cmux.New(ln)
	grpcLn := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpLn := mux.Match(cmux.HTTP1())
	raftLnBase := mux.Match(cmux.Any())
	go mux.Serve()
	var raftLn *tcp.Transport
	if encrypt {
		raftLn = tcp.NewTransportFromListener(raftLnBase, true, noVerify)
	} else {
		raftLn = tcp.NewTransportFromListener(raftLnBase, false, false)
	}

	// Create and open the store.
	dataPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Fatalf("failed to determine absolute data path: %s", err.Error())
	}

	str := store.New(raftLn, &store.StoreConfig{
		Dir: dataPath,
		ID:  idOrRaftAddr(),
	})

	// Set optional parameters on store.
	str.RaftLogLevel = raftLogLevel
	str.ShutdownOnRemove = raftShutdownOnRemove
	str.SnapshotThreshold = raftSnapThreshold
	str.SnapshotInterval, err = time.ParseDuration(raftSnapInterval)
	if err != nil {
		log.Fatalf("failed to parse Raft Snapsnot interval %s: %s", raftSnapInterval, err.Error())
	}
	str.LeaderLeaseTimeout, err = time.ParseDuration(raftLeaderLeaseTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft Leader lease timeout %s: %s", raftLeaderLeaseTimeout, err.Error())
	}
	str.HeartbeatTimeout, err = time.ParseDuration(raftHeartbeatTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft heartbeat timeout %s: %s", raftHeartbeatTimeout, err.Error())
	}
	str.ElectionTimeout, err = time.ParseDuration(raftElectionTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft election timeout %s: %s", raftElectionTimeout, err.Error())
	}
	str.ApplyTimeout, err = time.ParseDuration(raftApplyTimeout)
	if err != nil {
		log.Fatalf("failed to parse Raft apply timeout %s: %s", raftApplyTimeout, err.Error())
	}

	// Any prexisting node state?
	var enableBootstrap bool
	isNew := store.IsNewNode(dataPath)
	if isNew {
		log.Printf("no preexisting node state detected in %s, node may be bootstrapping", dataPath)
		enableBootstrap = true // New node, so we may be bootstrapping
	} else {
		log.Printf("preexisting node state detected in %s", dataPath)
	}

	// Determine join addresses
	var joins []string
	joins, err = determineJoinAddresses()
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
	apiAdv := raftAddr
	apiProto := "http"
	if x509Cert != "" {
		apiProto = "https"
	}
	meta := map[string]string{
		"api_addr":  apiAdv,
		"api_proto": apiProto,
	}

	// Execute any requested join operation.
	if len(joins) > 0 && isNew {
		log.Println("join addresses are:", joins)
		advAddr := raftAddr
		if raftAdv != "" {
			advAddr = raftAdv
		}

		joinDur, err := time.ParseDuration(joinInterval)
		if err != nil {
			log.Fatalf("failed to parse Join interval %s: %s", joinInterval, err.Error())
		}

		tlsConfig := tls.Config{InsecureSkipVerify: noVerify}
		if x509CACert != "" {
			asn1Data, err := ioutil.ReadFile(x509CACert)
			if err != nil {
				log.Fatalf("ioutil.ReadFile failed: %s", err.Error())
			}
			tlsConfig.RootCAs = x509.NewCertPool()
			ok := tlsConfig.RootCAs.AppendCertsFromPEM([]byte(asn1Data))
			if !ok {
				log.Fatalf("failed to parse root CA certificate(s) in %q", x509CACert)
			}
		}

		if j, err := cluster.Join(joinSrcIP, joins, str.ID(), advAddr, !raftNonVoter, meta,
			joinAttempts, joinDur, &tlsConfig); err != nil {
			log.Fatalf("failed to join cluster at %s: %s", joins, err.Error())
		} else {
			log.Println("successfully joined cluster at", j)
		}

	}

	// TODO
	// Wait until the store is in full consensus.
	if err := waitForConsensus(str); err != nil {
		log.Fatalf(err.Error())
	}

	// This may be a standalone server. In that case set its own metadata.
	if err := str.SetMetadata(meta); err != nil && err != store.ErrNotLeader {
		// Non-leader errors are OK, since metadata will then be set through
		// consensus as a result of a join. All other errors indicate a problem.
		log.Fatalf("failed to set store metadata: %s", err.Error())
	}
	c := core.New(str)
	// Start the HTTP API server.
	if err = startHTTPService(c, httpLn); err != nil {
		log.Fatalf("failed to start HTTP server: %s", err.Error())
	}
	if err = startGrpcService(c, grpcLn); err != nil {
		log.Fatalf("failed to start grpc server: %s", err.Error())
	}

	log.Println("node is ready")

	// Block until signalled.
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	if err := str.Close(true); err != nil {
		log.Printf("failed to close store: %s", err.Error())
	}
	stopProfile()
	log.Println("casbin-mesh server stopped")
}

func determineJoinAddresses() ([]string, error) {
	//raftAdv := httpAddr
	//if httpAdv != "" {
	//	raftAdv = httpAdv
	//}

	var addrs []string
	if joinAddr != "" {
		// Explicit join addresses are first priority.
		addrs = strings.Split(joinAddr, ",")
	}

	return addrs, nil
}

func waitForConsensus(str *store.Store) error {
	openTimeout, err := time.ParseDuration(raftOpenTimeout)
	if err != nil {
		return fmt.Errorf("failed to parse Raft open timeout %s: %s", raftOpenTimeout, err.Error())
	}
	if _, err := str.WaitForLeader(openTimeout); err != nil {
		if raftWaitForLeader {
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
	go func() {
		err := http.Serve(ln, httpd)
		if err != nil {
			log.Println("HTTP service Serve() returned:", err.Error())
		}
	}()
	return nil
}

func startGrpcService(c core.Core, ln net.Listener) error {
	grpcd := core.NewGrpcService(c)
	go func() {
		err := grpcd.Serve(ln)
		if err != nil {
			log.Println("Grpc service Serve() returned:", err.Error())
		}
	}()
	return nil
}

func idOrRaftAddr() string {
	if nodeID != "" {
		return nodeID
	}
	if raftAdv == "" {
		return raftAddr
	}
	return raftAdv
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
