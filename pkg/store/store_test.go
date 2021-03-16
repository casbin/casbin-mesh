/*
Copyright The casbind Authors.
@Date: 2021/03/16 17:59
*/

package store

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

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
	err = s.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
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
	err = s.AddPolicies(context.TODO(), "default", "p", "p", [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
	assert.Equal(t, nil, err)
	err = s.AddPolicies(context.TODO(), "default", "g", "g", [][]string{
		{"alice", "data2_admin"},
	})
	assert.Equal(t, nil, err)
	for _, set := range RBAC_TEST_SETS {
		r, err := s.Enforce(context.TODO(), "default", 0, 0, set.input...)
		assert.Equal(t, nil, err)
		assert.Equal(t, r, set.expect)
	}

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
	path, err := ioutil.TempDir("", "casbind-test-")
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
