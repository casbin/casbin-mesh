// Copyright 2022 The casbin-neo Authors. All Rights Reserved.
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

package hashicorp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"log"
	"net"
	"os"
	"time"
)

const (
	raftTimeout = 10 * time.Second
)

type NodeOptions struct {
	NodeID    string
	Addr      string
	bootstrap bool
}

type Node interface {
	MockGet(string) (string, error)
	MockLeaderGet(string) (string, error)
	MockSet(k, v string) error
	Join(nodeId, add string) error
}

type node struct {
	Raft *raft.Raft
	FSM  FSM
}

func (n node) Join(nodeID, addr string) error {
	log.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := n.Raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				log.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := n.Raft.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := n.Raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	log.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

func (n node) MockGet(s string) (string, error) {
	log.Printf("[get] applied idx:%d,last idx:%d\n", n.Raft.AppliedIndex(), n.Raft.LastIndex())
	//log.Println(n.Raft.Stats())
	return n.FSM.Get(s)
}

func (n node) MockLeaderGet(s string) (string, error) {
	c := &command{
		Op:  "get",
		Key: s,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	f := n.Raft.Apply(b, 0)
	if f.Error() != nil {
		return "", f.Error()
	}
	return f.Response().(string), nil
}

func (n node) MockSet(k, v string) error {
	c := &command{
		Op:    "set",
		Key:   k,
		Value: v,
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	f := n.Raft.Apply(b, 0)
	if err = f.Error(); err != nil {
		return err
	}
	_ = f.Response()
	// Only uses in testing
	// it issues a command that blocks until all preceding operations have been applied to the FSM.
	// It can be used to ensure the FSM reflects all queued writes.
	fu := n.Raft.Barrier(0)
	fu.Error()
	log.Printf("[set] applied idx:%d,last idx:%d\n", n.Raft.AppliedIndex(), n.Raft.LastIndex())
	return nil
}

func NewNode(options NodeOptions) (Node, error) {
	f := NewFSM()
	r, err := NewRaft(nil, options.NodeID, options.Addr, f)
	if err != nil {
		return nil, err
	}
	if options.bootstrap {
		cfg := raft.Configuration{Servers: []raft.Server{
			{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(options.NodeID),
				Address:  raft.ServerAddress(options.Addr),
			},
		}}
		fu := r.BootstrapCluster(cfg)
		if fu.Error() != nil {
			return nil, err
		}
	}
	return &node{r, f}, nil
}

func NewRaft(ctx context.Context, nodeId, bind string, fsm raft.FSM) (*raft.Raft, error) {
	c := raft.DefaultConfig()
	c.CommitTimeout = 20 * time.Millisecond
	c.LocalID = raft.ServerID(nodeId)
	ldb := raft.NewInmemStore()
	sdb := raft.NewInmemStore()
	snapdb := raft.NewInmemSnapshotStore()
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return nil, err
	}
	t, err := raft.NewTCPTransport(bind, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, err
	}
	r, err := raft.NewRaft(c, fsm, ldb, sdb, snapdb, t)
	if err != nil {
		return nil, err
	}
	return r, nil
}
