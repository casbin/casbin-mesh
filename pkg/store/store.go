/*
Copyright The casbind Authors.
@Date: 2021/03/12 19:47
*/

package store

import (
	"errors"
	"expvar"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/WenyXu/casbind/proto/command"

	rlog "github.com/WenyXu/casbind/pkg/log"
	"github.com/hashicorp/raft"
)

var (
	// ErrNotLeader is returned when a node attempts to execute a leader-only
	// operation.
	ErrNotLeader = errors.New("not leader")

	// ErrStaleRead is returned if the executing the query would violate the
	// requested freshness.
	ErrStaleRead = errors.New("stale read")

	// ErrOpenTimeout is returned when the Store does not apply its initial
	// logs within the specified time.
	ErrOpenTimeout = errors.New("timeout waiting for initial logs application")

	// ErrInvalidBackupFormat is returned when the requested backup format
	// is not valid.
	ErrInvalidBackupFormat = errors.New("invalid backup format")
)

const (
	raftDBPath          = "default-raft.db" // Changing this will break backwards compatibility.
	retainSnapshotCount = 2
	applyTimeout        = 10 * time.Second
	openTimeout         = 120 * time.Second
	leaderWaitDelay     = 100 * time.Millisecond
	appliedWaitDelay    = 100 * time.Millisecond
	connectionPoolCount = 5
	connectionTimeout   = 10 * time.Second
	raftLogCacheSize    = 512
	trailingScale       = 1.25
)

const (
	numSnaphots             = "num_snapshots"
	numBackups              = "num_backups"
	numRestores             = "num_restores"
	numUncompressedCommands = "num_uncompressed_commands"
	numCompressedCommands   = "num_compressed_commands"
	numLegacyCommands       = "num_legacy_commands"
)

// BackupFormat represents the format of database backup.
type BackupFormat int

const (
	// BackupBinary is a file backup format.
	BackupBinary = iota
)

// stats captures stats for the Store.
var stats *expvar.Map

func init() {
	stats = expvar.NewMap("store")
	stats.Add(numSnaphots, 0)
	stats.Add(numBackups, 0)
	stats.Add(numRestores, 0)
	stats.Add(numUncompressedCommands, 0)
	stats.Add(numCompressedCommands, 0)
	stats.Add(numLegacyCommands, 0)
}

// ClusterState defines the possible Raft states the current node can be in
type ClusterState int

// Represents the Raft cluster states
const (
	Leader ClusterState = iota
	Follower
	Candidate
	Shutdown
	Unknown
)

// Store is casbin memory data, where all changes are made via Raft consensus.
type Store struct {
	raftDir string

	raft   *raft.Raft // The consensus mechanism.
	ln     Listener
	raftTn *raft.NetworkTransport
	raftID string // Node ID.

	raftLog    raft.LogStore    // Persistent log store.
	raftStable raft.StableStore // Persistent k-v store.
	boltStore  *rlog.Log        // Physical store.

	//onDiskCreated        bool      // On disk database actually created?
	snapsExistOnOpen     bool      // Any snaps present when store opens?
	firstIdxOnOpen       uint64    // First index on log when Store opens.
	lastIdxOnOpen        uint64    // Last index on log when Store opens.
	lastCommandIdxOnOpen uint64    // Last command index on log when Store opens.
	firstLogAppliedT     time.Time // Time first log is applied
	appliedOnOpen        uint64    // Number of logs applied at open.
	openT                time.Time // Timestamp when Store opens.

	numNoops int // For whitebox testing

	txMu    sync.RWMutex // Sync between snapshots and query-level transactions.
	queryMu sync.RWMutex // Sync queries generally with other operations.

	metaMu    sync.RWMutex
	meta      map[string]map[string]string
	enforcers sync.Map
	logger    *log.Logger

	ShutdownOnRemove   bool
	SnapshotThreshold  uint64
	SnapshotInterval   time.Duration
	LeaderLeaseTimeout time.Duration
	HeartbeatTimeout   time.Duration
	ElectionTimeout    time.Duration
	ApplyTimeout       time.Duration
	RaftLogLevel       string

	numTrailingLogs uint64
}

// IsNewNode returns whether a node using raftDir would be a brand new node.
// It also means that the window this node joining a different cluster has passed.
func IsNewNode(raftDir string) bool {
	// If there is any pre-existing Raft state, then this node
	// has already been created.
	return !pathExists(filepath.Join(raftDir, raftDBPath))
}

// New returns a new Store.
func New(ln Listener, c *StoreConfig) *Store {
	logger := c.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "[store] ", log.LstdFlags)
	}

	return &Store{
		ln:           ln,
		raftDir:      c.Dir,
		raftID:       c.ID,
		meta:         make(map[string]map[string]string),
		logger:       logger,
		ApplyTimeout: applyTimeout,
	}
}

// Open opens the Store. If enableBootstrap is set, then this node becomes a
// standalone node. If not set, then the calling layer must know that this
// node has pre-existing state, or the calling layer will trigger a join
// operation after opening the Store.
func (s *Store) Open(enableBootstrap bool) error {
	s.openT = time.Now()
	s.logger.Printf("opening store with node ID %s", s.raftID)

	s.logger.Printf("ensuring directory at %s exists", s.raftDir)
	err := os.MkdirAll(s.raftDir, 0755)
	if err != nil {
		return err
	}

	// Create Raft-compatible network layer.
	s.raftTn = raft.NewNetworkTransport(NewTransport(s.ln), connectionPoolCount, connectionTimeout, nil)

	// Don't allow control over trailing logs directly, just implement a policy.
	s.numTrailingLogs = uint64(float64(s.SnapshotThreshold) * trailingScale)

	config := s.raftConfig()
	config.LocalID = raft.ServerID(s.raftID)

	// Create the snapshot store. This allows Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}
	snaps, err := snapshots.List()
	if err != nil {
		return fmt.Errorf("list snapshots: %s", err)
	}
	s.logger.Printf("%d preexisting snapshots present", len(snaps))
	s.snapsExistOnOpen = len(snaps) > 0

	// Create the log store and stable store.
	s.boltStore, err = rlog.NewLog(filepath.Join(s.raftDir, raftDBPath))
	if err != nil {
		return fmt.Errorf("new log store: %s", err)
	}
	s.raftStable = s.boltStore
	s.raftLog, err = raft.NewLogCache(raftLogCacheSize, s.boltStore)
	if err != nil {
		return fmt.Errorf("new cached store: %s", err)
	}

	// Get some info about the log, before any more entries are committed.
	if err := s.setLogInfo(); err != nil {
		return fmt.Errorf("set log info: %s", err)
	}
	s.logger.Printf("first log index: %d, last log index: %d, last command log index: %d:",
		s.firstIdxOnOpen, s.lastIdxOnOpen, s.lastCommandIdxOnOpen)

	//// If an on-disk database has been requested, and there are no snapshots, and
	//// there are no commands in the log, then this is the only opportunity to
	//// create that on-disk database file before Raft initializes.
	//if !s.dbConf.Memory && !s.snapsExistOnOpen && s.lastCommandIdxOnOpen == 0 {
	//	s.db, err = s.openOnDisk(nil)
	//	if err != nil {
	//		return fmt.Errorf("failed to open on-disk database")
	//	}
	//	s.onDiskCreated = true
	//} else {
	//	// We need an in-memory database, at least for bootstrapping purposes.
	//	s.db, err = s.openInMemory(nil)
	//	if err != nil {
	//		return fmt.Errorf("failed to open in-memory database")
	//	}
	//}

	// Instantiate the Raft system.
	ra, err := raft.NewRaft(config, s, s.raftLog, s.raftStable, snapshots, s.raftTn)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}

	if enableBootstrap {
		s.logger.Printf("executing new cluster bootstrap")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: s.raftTn.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	} else {
		s.logger.Printf("no cluster bootstrap requested")
	}

	s.raft = ra

	return nil
}

// raftConfig returns a new Raft config for the store.
func (s *Store) raftConfig() *raft.Config {
	config := raft.DefaultConfig()
	config.ShutdownOnRemove = s.ShutdownOnRemove
	config.LogLevel = s.RaftLogLevel
	if s.SnapshotThreshold != 0 {
		config.SnapshotThreshold = s.SnapshotThreshold
		config.TrailingLogs = s.numTrailingLogs
	}
	if s.SnapshotInterval != 0 {
		config.SnapshotInterval = s.SnapshotInterval
	}
	if s.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = s.LeaderLeaseTimeout
	}
	if s.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = s.HeartbeatTimeout
	}
	if s.ElectionTimeout != 0 {
		config.ElectionTimeout = s.ElectionTimeout
	}
	return config
}

// setLogInfo records some key indexs about the log.
func (s *Store) setLogInfo() error {
	var err error
	s.firstIdxOnOpen, err = s.boltStore.FirstIndex()
	if err != nil {
		return fmt.Errorf("failed to get last index: %s", err)
	}
	s.lastIdxOnOpen, err = s.boltStore.LastIndex()
	if err != nil {
		return fmt.Errorf("failed to get last index: %s", err)
	}
	s.lastCommandIdxOnOpen, err = s.boltStore.LastCommandIndex()
	if err != nil {
		return fmt.Errorf("failed to get last command index: %s", err)
	}
	return nil
}

// Close closes the store. If wait is true, waits for a graceful shutdown.
func (s *Store) Close(wait bool) error {
	f := s.raft.Shutdown()
	if wait {
		if e := f.(raft.Future); e.Error() != nil {
			return e.Error()
		}
	}
	// Only shutdown Bolt and SQLite when Raft is done.
	//if err := s.db.Close(); err != nil {
	//	return err
	//}
	if err := s.boltStore.Close(); err != nil {
		return err
	}
	return nil
}

// WaitForApplied waits for all Raft log entries to to be applied to the
// underlying database.
func (s *Store) WaitForApplied(timeout time.Duration) error {
	if timeout == 0 {
		return nil
	}
	s.logger.Printf("waiting for up to %s for application of initial logs", timeout)
	if err := s.WaitForAppliedIndex(s.raft.LastIndex(), timeout); err != nil {
		return ErrOpenTimeout
	}
	return nil
}

// IsLeader is used to determine if the current node is cluster leader
func (s *Store) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

// State returns the current node's Raft state
func (s *Store) State() ClusterState {
	state := s.raft.State()
	switch state {
	case raft.Leader:
		return Leader
	case raft.Candidate:
		return Candidate
	case raft.Follower:
		return Follower
	case raft.Shutdown:
		return Shutdown
	default:
		return Unknown
	}
}

// Path returns the path to the store's storage directory.
func (s *Store) Path() string {
	return s.raftDir
}

// Addr returns the address of the store.
func (s *Store) Addr() string {
	return string(s.raftTn.LocalAddr())
}

// ID returns the Raft ID of the store.
func (s *Store) ID() string {
	return s.raftID
}

// LeaderAddr returns the address of the current leader. Returns a
// blank string if there is no leader.
func (s *Store) LeaderAddr() string {
	return string(s.raft.Leader())
}

// LeaderID returns the node ID of the Raft leader. Returns a
// blank string if there is no leader, or an error.
func (s *Store) LeaderID() (string, error) {
	addr := s.LeaderAddr()
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.logger.Printf("failed to get raft configuration: %v", err)
		return "", err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.Address == raft.ServerAddress(addr) {
			return string(srv.ID), nil
		}
	}
	return "", nil
}

// Nodes returns the slice of nodes in the cluster, sorted by ID ascending.
func (s *Store) Nodes() ([]*Server, error) {
	f := s.raft.GetConfiguration()
	if f.Error() != nil {
		return nil, f.Error()
	}

	rs := f.Configuration().Servers
	servers := make([]*Server, len(rs))
	for i := range rs {
		servers[i] = &Server{
			ID:   string(rs[i].ID),
			Addr: string(rs[i].Address),
		}
	}

	sort.Sort(Servers(servers))
	return servers, nil
}

// WaitForLeader blocks until a leader is detected, or the timeout expires.
func (s *Store) WaitForLeader(timeout time.Duration) (string, error) {
	tck := time.NewTicker(leaderWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			l := s.LeaderAddr()
			if l != "" {
				return l, nil
			}
		case <-tmr.C:
			return "", fmt.Errorf("timeout expired")
		}
	}
}

//// SetRequestCompression allows low-level control over the compression threshold
//// for the request marshaler.
//func (s *Store) SetRequestCompression(batch, size int) {
//	s.reqMarshaller.BatchThreshold = batch
//	s.reqMarshaller.SizeThreshold = size
//}

// WaitForAppliedIndex blocks until a given log index has been applied,
// or the timeout expires.
func (s *Store) WaitForAppliedIndex(idx uint64, timeout time.Duration) error {
	tck := time.NewTicker(appliedWaitDelay)
	defer tck.Stop()
	tmr := time.NewTimer(timeout)
	defer tmr.Stop()

	for {
		select {
		case <-tck.C:
			if s.raft.AppliedIndex() >= idx {
				return nil
			}
		case <-tmr.C:
			return fmt.Errorf("timeout expired")
		}
	}
}

// Stats returns stats for the store.
func (s *Store) Stats() (map[string]interface{}, error) {
	//fkEnabled, err := s.db.FKConstraints()
	//if err != nil {
	//	return nil, err
	//}

	//dbSz, err := s.db.Size()
	//if err != nil {
	//	return nil, err
	//}
	//dbStatus := map[string]interface{}{
	//	"dsn":            s.dbConf.DSN,
	//	"fk_constraints": enabledFromBool(fkEnabled),
	//	"version":        sql.DBVersion,
	//	"db_size":        dbSz,
	//}
	//if s.dbConf.Memory {
	//	dbStatus["path"] = ":memory:"
	//} else {
	//	dbStatus["path"] = s.dbPath
	//	if s.onDiskCreated {
	//		if dbStatus["size"], err = s.db.FileSize(); err != nil {
	//			return nil, err
	//		}
	//	}
	//}

	nodes, err := s.Nodes()
	if err != nil {
		return nil, err
	}
	leaderID, err := s.LeaderID()
	if err != nil {
		return nil, err
	}

	// Perform type-conversion to actual numbers where possible.
	raftStats := make(map[string]interface{})
	for k, v := range s.raft.Stats() {
		if s, err := strconv.ParseInt(v, 10, 64); err != nil {
			raftStats[k] = v
		} else {
			raftStats[k] = s
		}
	}
	raftStats["log_size"], err = s.logSize()
	if err != nil {
		return nil, err
	}

	dirSz, err := dirSize(s.raftDir)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"node_id": s.raftID,
		"raft":    raftStats,
		"addr":    s.Addr(),
		"leader": map[string]string{
			"node_id": leaderID,
			"addr":    s.LeaderAddr(),
		},
		"apply_timeout":      s.ApplyTimeout.String(),
		"heartbeat_timeout":  s.HeartbeatTimeout.String(),
		"election_timeout":   s.ElectionTimeout.String(),
		"snapshot_threshold": s.SnapshotThreshold,
		"snapshot_interval":  s.SnapshotInterval,
		"trailing_logs":      s.numTrailingLogs,
		"metadata":           s.meta,
		"nodes":              nodes,
		"dir":                s.raftDir,
		"dir_size":           dirSz,
	}
	return status, nil
}

// Join joins a node, identified by id and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (s *Store) Join(id, addr string, voter bool, metadata map[string]string) error {
	s.logger.Printf("received request to join node at %s", addr)
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		s.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(id) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, the no
			// join is actually needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(id) {
				s.logger.Printf("node %s at %s already member of cluster, ignoring join request", id, addr)
				return nil
			}

			if err := s.remove(id); err != nil {
				s.logger.Printf("failed to remove node: %v", err)
				return err
			}
		}
	}

	var f raft.IndexFuture
	if voter {
		f = s.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)
	} else {

		f = s.raft.AddNonvoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)
	}
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}

	if err := s.setMetadata(id, metadata); err != nil {
		return err
	}

	s.logger.Printf("node at %s joined successfully as %s", addr, prettyVoter(voter))
	return nil
}

// Remove removes a node from the store, specified by ID.
func (s *Store) Remove(id string) error {
	s.logger.Printf("received request to remove node %s", id)
	if err := s.remove(id); err != nil {
		s.logger.Printf("failed to remove node %s: %s", id, err.Error())
		return err
	}

	s.logger.Printf("node %s removed successfully", id)
	return nil
}

// remove removes the node, with the given ID, from the cluster.
func (s *Store) remove(id string) error {
	if s.raft.State() != raft.Leader {
		return ErrNotLeader
	}

	f := s.raft.RemoveServer(raft.ServerID(id), 0, 0)
	if f.Error() != nil {
		if f.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return f.Error()
	}

	md := command.MetadataDelete{
		RaftId: id,
	}
	p, err := proto.Marshal(&md)
	if err != nil {
		return err
	}

	c := &command.Command{
		Type:    command.Type_COMMAND_TYPE_METADATA_DELETE,
		Payload: p,
	}
	bc, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	f = s.raft.Apply(bc, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		e.Error()
	}

	return nil
}
