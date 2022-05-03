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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"sync"
)

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type FSM interface {
	Get(key string) (string, error)
	raft.FSM
}

type fsm struct {
	mu sync.RWMutex
	m  map[string]string
}

var (
	ErrKeyNotExits = errors.New("key not exits")
)

func (f *fsm) Get(key string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if v, ok := f.m[key]; ok {
		return v, nil
	}
	return "", ErrKeyNotExits
}

func (f *fsm) Apply(log *raft.Log) interface{} {
	var c command
	if err := json.Unmarshal(log.Data, &c); err != nil {
		panic(fmt.Sprintf("failed to unmarshal command: %s", err.Error()))
	}

	switch c.Op {
	case "get":
		v, _ := f.Get(c.Key)
		return v
	case "set":
		return f.applySet(c.Key, c.Value)
	case "delete":
		return f.applyDelete(c.Key)
	default:
		panic(fmt.Sprintf("unrecognized command op: %s", c.Op))
	}
}

func (f *fsm) applySet(key, value string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.m[key] = value
	return nil
}

func (f *fsm) applyDelete(key string) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.m, key)
	return nil
}

func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clone the map.
	o := make(map[string]string)
	for k, v := range f.m {
		o[k] = v
	}
	return &fsmSnapshot{store: o}, nil
}

func (f *fsm) Restore(snapshot io.ReadCloser) error {
	o := make(map[string]string)
	if err := json.NewDecoder(snapshot).Decode(&o); err != nil {
		return err
	}

	// Set the state from the snapshot, no lock required according to
	// Hashicorp docs.
	f.m = o
	return nil
}

func NewFSM() FSM {
	return &fsm{m: make(map[string]string)}
}
