package store

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/casbin/casbin-mesh/proto/command"

	"github.com/stretchr/testify/assert"
)

func Test_OpenStoreSingleNode(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}

	s.WaitForLeader(10 * time.Second)
	if got, exp := s.LeaderAddr(), s.Addr(); got != exp {
		t.Fatalf("wrong leader address returned, got: %s, exp %s", got, exp)
	}
	id, err := s.LeaderID()
	if err != nil {
		t.Fatalf("failed to retrieve leader ID: %s", err.Error())
	}
	if got, exp := id, s.raftID; got != exp {
		t.Fatalf("wrong leader ID returned, got: %s, exp %s", got, exp)
	}
}

func Test_OpenStoreCloseSingleNode(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	s.WaitForLeader(10 * time.Second)
	if err := s.Close(true); err != nil {
		t.Fatalf("failed to close single-node store: %s", err.Error())
	}
}

func Test_SingleNodeCreateNS(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	err := s.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
}

func Test_SingleNodeSetModel(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	err := s.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s.SetModelFromString(context.TODO(), "default", modelText)
	assert.Equal(t, nil, err)
}

func Test_SingleNodeSetModelFail(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	err := s.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s.SetModelFromString(context.TODO(), "default", incorrectModelText)
	fmt.Println(err)
	assert.NotEqual(t, nil, err)
}

func Test_SingleNodeAddPolices(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())
	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	err := s.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s.SetModelFromString(context.TODO(), "default", modelText)
	assert.Equal(t, nil, err)
	_, err = s.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
		{"alice", "data1", "read"},
	})
	assert.Equal(t, nil, err)
}

func Test_SingleNodeEnforce(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())
	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)
	err := s.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s.SetModelFromString(context.TODO(), "default", modelText)
	assert.Equal(t, nil, err)
	_, err = s.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
	assert.Equal(t, nil, err)
	_, err = s.AddPolicies(context.TODO(), "default", "g", "g", [][]string{
		{"alice", "data2_admin"},
	})
	assert.Equal(t, nil, err)
	for _, set := range RBAC_TEST_SETS {
		r, err := s.Enforce(context.TODO(), "default", 0, 0, set.input...)
		assert.Equal(t, nil, err)
		assert.Equal(t, r, set.expect)
	}
}

func Test_MultiNodeJoinRemove(t *testing.T) {
	s0 := mustNewStore()
	defer os.RemoveAll(s0.Path())
	if err := s0.Open(true); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s0.Close(true)
	s0.WaitForLeader(10 * time.Second)

	s1 := mustNewStore()
	defer os.RemoveAll(s1.Path())
	if err := s1.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s1.Close(true)

	// Get sorted list of cluster nodes.
	storeNodes := []string{s0.ID(), s1.ID()}
	sort.StringSlice(storeNodes).Sort()

	// Join the second node to the first.
	if err := s0.Join(s1.ID(), s1.Addr(), true, nil); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	s1.WaitForLeader(10 * time.Second)

	// Check leader state on follower.
	if got, exp := s1.LeaderAddr(), s0.Addr(); got != exp {
		t.Fatalf("wrong leader address returned, got: %s, exp %s", got, exp)
	}
	id, err := s1.LeaderID()
	if err != nil {
		t.Fatalf("failed to retrieve leader ID: %s", err.Error())
	}
	if got, exp := id, s0.raftID; got != exp {
		t.Fatalf("wrong leader ID returned, got: %s, exp %s", got, exp)
	}

	nodes, err := s0.Nodes()
	if err != nil {
		t.Fatalf("failed to get nodes: %s", err.Error())
	}

	if len(nodes) != len(storeNodes) {
		t.Fatalf("size of cluster is not correct")
	}
	if storeNodes[0] != nodes[0].ID || storeNodes[1] != nodes[1].ID {
		t.Fatalf("cluster does not have correct nodes")
	}

	// Remove a node.
	if err := s0.Remove(s1.ID()); err != nil {
		t.Fatalf("failed to remove %s from cluster: %s", s1.ID(), err.Error())
	}

	nodes, err = s0.Nodes()
	if err != nil {
		t.Fatalf("failed to get nodes post remove: %s", err.Error())
	}
	if len(nodes) != 1 {
		t.Fatalf("size of cluster is not correct post remove")
	}
	if s0.ID() != nodes[0].ID {
		t.Fatalf("cluster does not have correct nodes post remove")
	}
}

func Test_MultiNodeJoinNonVoterRemove(t *testing.T) {
	s0 := mustNewStore()
	defer os.RemoveAll(s0.Path())
	if err := s0.Open(true); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s0.Close(true)
	s0.WaitForLeader(10 * time.Second)

	s1 := mustNewStore()
	defer os.RemoveAll(s1.Path())
	if err := s1.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s1.Close(true)

	// Get sorted list of cluster nodes.
	storeNodes := []string{s0.ID(), s1.ID()}
	sort.StringSlice(storeNodes).Sort()

	// Join the second node to the first.
	if err := s0.Join(s1.ID(), s1.Addr(), false, nil); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	s1.WaitForLeader(10 * time.Second)

	// Check leader state on follower.
	if got, exp := s1.LeaderAddr(), s0.Addr(); got != exp {
		t.Fatalf("wrong leader address returned, got: %s, exp %s", got, exp)
	}
	id, err := s1.LeaderID()
	if err != nil {
		t.Fatalf("failed to retrieve leader ID: %s", err.Error())
	}
	if got, exp := id, s0.raftID; got != exp {
		t.Fatalf("wrong leader ID returned, got: %s, exp %s", got, exp)
	}

	nodes, err := s0.Nodes()
	if err != nil {
		t.Fatalf("failed to get nodes: %s", err.Error())
	}

	if len(nodes) != len(storeNodes) {
		t.Fatalf("size of cluster is not correct")
	}
	if storeNodes[0] != nodes[0].ID || storeNodes[1] != nodes[1].ID {
		t.Fatalf("cluster does not have correct nodes")
	}

	// Remove the non-voter.
	if err := s0.Remove(s1.ID()); err != nil {
		t.Fatalf("failed to remove %s from cluster: %s", s1.ID(), err.Error())
	}

	nodes, err = s0.Nodes()
	if err != nil {
		t.Fatalf("failed to get nodes post remove: %s", err.Error())
	}
	if len(nodes) != 1 {
		t.Fatalf("size of cluster is not correct post remove")
	}
	if s0.ID() != nodes[0].ID {
		t.Fatalf("cluster does not have correct nodes post remove")
	}
}

func Test_MultiNodeEnforce(t *testing.T) {
	s0 := mustNewStore()
	defer os.RemoveAll(s0.Path())
	if err := s0.Open(true); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s0.Close(true)
	s0.WaitForLeader(10 * time.Second)

	s1 := mustNewStore()
	defer os.RemoveAll(s1.Path())
	if err := s1.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s1.Close(true)

	s2 := mustNewStore()
	defer os.RemoveAll(s2.Path())
	if err := s2.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s2.Close(true)

	// Join the second node to the first as a voting node.
	if err := s0.Join(s1.ID(), s1.Addr(), true, nil); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	// Join the third node to the first as a non-voting node.
	if err := s0.Join(s2.ID(), s2.Addr(), false, nil); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	err := s0.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s0.SetModelFromString(context.TODO(), "default", modelText)
	assert.Equal(t, nil, err)

	// Wait until the 4 log entries have been applied to the voting follower,
	// and then query.
	if err := s1.WaitForAppliedIndex(4, 5*time.Second); err != nil {
		t.Fatalf("error waiting for follower to apply index: %s:", err.Error())
	}
	_, err = s0.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
	assert.Equal(t, nil, err)
	_, err = s0.AddPolicies(context.TODO(), "default", "g", "g", [][]string{
		{"alice", "data2_admin"},
	})
	assert.Equal(t, nil, err)

	// Wait until the 2 log entries have been applied to the voting follower,
	// and then query.
	if err := s1.WaitForAppliedIndex(6, 5*time.Second); err != nil {
		t.Fatalf("error waiting for follower to apply index: %s:", err.Error())
	}

	_, err = s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_WEAK, 0, "alice", "data1", "read")
	if err == nil {
		t.Fatalf("successfully queried non-leader node")
	}
	_, err = s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_STRONG, 0, "alice", "data1", "read")
	if err == nil {
		t.Fatalf("successfully queried non-leader node [strong]")
	}
	r3, err := s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, 0, "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query follower node: %s", err.Error())
	}
	assert.Equal(t, r3, true)
}

func Test_MultiNodeExecuteEnforceFreshness(t *testing.T) {
	s0 := mustNewStore()
	defer os.RemoveAll(s0.Path())
	if err := s0.Open(true); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s0.Close(true)
	s0.WaitForLeader(10 * time.Second)

	s1 := mustNewStore()
	defer os.RemoveAll(s1.Path())
	if err := s1.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s1.Close(true)

	s2 := mustNewStore()
	defer os.RemoveAll(s2.Path())
	if err := s2.Open(false); err != nil {
		t.Fatalf("failed to open node for multi-node test: %s", err.Error())
	}
	defer s2.Close(true)

	// Join the second node to the first as a voting node.
	if err := s0.Join(s1.ID(), s1.Addr(), true, nil); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	// Join the third node to the first as a non-voting node.
	if err := s0.Join(s2.ID(), s2.Addr(), false, nil); err != nil {
		t.Fatalf("failed to join to node at %s: %s", s0.Addr(), err.Error())
	}

	err := s0.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s0.SetModelFromString(context.TODO(), "default", modelText)
	assert.Equal(t, nil, err)

	_, err = s0.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
	assert.Equal(t, nil, err)
	_, err = s0.AddPolicies(context.TODO(), "default", "g", "g", [][]string{
		{"alice", "data2_admin"},
	})
	assert.Equal(t, nil, err)

	// Wait until the 6 log entries have been applied to the voting follower,
	// and then query.
	if err := s1.WaitForAppliedIndex(6, 5*time.Second); err != nil {
		t.Fatalf("error waiting for follower to apply index: %s:", err.Error())
	}
	r, err := s0.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, 0, "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query leader node: %s", err.Error())
	}
	assert.Equal(t, true, r)

	// Wait until the 7 log entries have been applied to the follower,
	// and then query.
	if err := s1.WaitForAppliedIndex(7, 5*time.Second); err != nil {
		t.Fatalf("error waiting for follower to apply index: %s:", err.Error())
	}

	r2, err := s0.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_WEAK, int64(time.Nanosecond), "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query leader node: %s", err.Error())
	}
	assert.Equal(t, true, r2)

	r3, err := s0.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_STRONG, int64(time.Nanosecond), "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query leader node: %s", err.Error())
	}
	assert.Equal(t, true, r3)

	// kill leader
	s0.Close(true)

	// "None" consistency queries should still work.
	r4, err := s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, 0, "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query follower node: %s", err.Error())
	}
	assert.Equal(t, true, r4)

	time.Sleep(time.Second)

	// "None" consistency queries with 1 nanosecond freshness should fail, because at least
	// one nanosecond *should* have passed since leader died (surely!).
	_, err = s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, int64(time.Nanosecond), "alice", "data1", "read")
	if err == nil {
		t.Fatalf("freshness violating query didn't return an error")
	}
	if err != ErrStaleRead {
		t.Fatalf("freshness violating query didn't returned wrong error: %s", err.Error())
	}

	// Freshness of 0 is ignored.
	r5, err := s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, int64(0), "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query follower node: %s", err.Error())
	}
	assert.Equal(t, true, r5)

	// "None" consistency queries with 1 hour freshness should pass, because it should
	// not be that long since the leader died.

	r6, err := s1.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, int64(time.Hour), "alice", "data1", "read")
	if err != nil {
		t.Fatalf("failed to query follower node: %s", err.Error())
	}
	assert.Equal(t, true, r6)

}

func Test_SingleNodeSnapshotOnDisk(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)

	err := s.CreateNamespace(context.TODO(), "default")
	assert.Equal(t, nil, err)
	err = s.SetModelFromString(context.TODO(), "default", modelText)
	assert.Equal(t, nil, err)
	_, err = s.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
	assert.Equal(t, nil, err)
	_, err = s.AddPolicies(context.TODO(), "default", "g", "g", [][]string{
		{"alice", "data2_admin"},
	})
	assert.Equal(t, nil, err)

	// Snap the node and write to disk.
	f, err := s.Snapshot()
	if err != nil {
		t.Fatalf("failed to snapshot node: %s", err.Error())
	}

	snapDir := mustTempDir()
	defer os.RemoveAll(snapDir)
	snapFile, err := os.Create(filepath.Join(snapDir, "snapshot"))
	if err != nil {
		t.Fatalf("failed to create snapshot file: %s", err.Error())
	}
	sink := &mockSnapshotSink{snapFile}
	if err := f.Persist(sink); err != nil {
		t.Fatalf("failed to persist snapshot to disk: %s", err.Error())
	}

	// Check restoration.
	snapFile, err = os.Open(filepath.Join(snapDir, "snapshot"))
	if err != nil {
		t.Fatalf("failed to open snapshot file: %s", err.Error())
	}
	if err := s.Restore(snapFile); err != nil {
		t.Fatalf("failed to restore snapshot from disk: %s", err.Error())
	}

	for _, set := range RBAC_TEST_SETS {
		fmt.Println(set.input)
		r, err := s.Enforce(context.TODO(), "default", command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE, 0, set.input...)
		assert.Equal(t, nil, err)
		assert.Equal(t, set.expect, r)
	}
}

func Test_IsLeader(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)

	if !s.IsLeader() {
		t.Fatalf("single node is not leader!")
	}
}

func Test_State(t *testing.T) {
	s := mustNewStore()
	defer os.RemoveAll(s.Path())

	if err := s.Open(true); err != nil {
		t.Fatalf("failed to open single-node store: %s", err.Error())
	}
	defer s.Close(true)
	s.WaitForLeader(10 * time.Second)

	state := s.State()
	if state != Leader {
		t.Fatalf("single node returned incorrect state (not Leader): %v", s)
	}
}

type mockSnapshotSink struct {
	*os.File
}

func (m *mockSnapshotSink) ID() string {
	return "1"
}

func (m *mockSnapshotSink) Cancel() error {
	return nil
}

type EnforceData struct {
	input  []interface{}
	expect bool
}

var (
	RBAC_TEST_SETS = []EnforceData{
		{
			[]interface{}{"alice", "data1", "read"}, true,
		},
		{
			[]interface{}{"alice", "data1", "write"}, false,
		},
		{
			[]interface{}{"alice", "data2", "read"}, true,
		},
		{
			[]interface{}{"alice", "data2", "write"}, true,
		},
		{
			[]interface{}{"bob", "data1", "read"}, false,
		},
		{
			[]interface{}{"bob", "data1", "write"}, false,
		},
		{
			[]interface{}{"bob", "data2", "read"}, false,
		},
		{
			[]interface{}{"bob", "data2", "write"}, true,
		},
	}
)
var (
	modelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	incorrectModelText = `
[request_definition
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
)

type mockListener struct {
	ln net.Listener
}

func (m *mockListener) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, timeout)
}

func (m *mockListener) Accept() (net.Conn, error) { return m.ln.Accept() }

func (m *mockListener) Close() error { return m.ln.Close() }

func (m *mockListener) Addr() net.Addr { return m.ln.Addr() }

func mustMockLister(addr string) Listener {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic("failed to create new listner")
	}
	return &mockListener{ln}
}

func mustTempDir() string {
	var err error
	path, err := ioutil.TempDir("", "casbin-mesh-test-")
	if err != nil {
		panic("failed to create temp dir")
	}
	return path
}

func mustNewStoreAtPath(path string) *Store {
	s := New(mustMockLister("localhost:0"), &StoreConfig{
		Dir: path,
		ID:  path, // Could be any unique string.
	})
	if s == nil {
		panic("failed to create new store")
	}
	return s
}

func mustNewStore() *Store {
	return mustNewStoreAtPath(mustTempDir())
}
