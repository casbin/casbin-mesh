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

package log

import (
	"fmt"

	raftbadgerdb "github.com/BBVA/raft-badger"
	"github.com/hashicorp/raft"
)

// Log is an object that can return information about the Raft log.
type Log struct {
	*raftbadgerdb.BadgerStore
}

// NewLog returns an instantiated Log object.
func NewLog(path string) (*Log, error) {
	bs, err := raftbadgerdb.NewBadgerStore(path)
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}
	return &Log{bs}, nil
}

// Indexes returns the first and last indexes.
func (l *Log) Indexes() (uint64, uint64, error) {
	fi, err := l.FirstIndex()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get first index: %s", err)
	}
	li, err := l.LastIndex()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get last index: %s", err)
	}
	return fi, li, nil
}

// LastCommandIndex returns the index of the last Command
// log entry written to the Raft log. Returns an index of
// zero if no such log exists.
func (l *Log) LastCommandIndex() (uint64, error) {
	fi, li, err := l.Indexes()
	if err != nil {
		return 0, fmt.Errorf("get indexes: %s", err)
	}

	// Check for empty log.
	if li == 0 {
		return 0, nil
	}

	var rl raft.Log
	for i := li; i >= fi; i-- {
		if err := l.GetLog(i, &rl); err != nil {
			return 0, fmt.Errorf("get log at index %d: %s", i, err)
		}
		if rl.Type == raft.LogCommand {
			return i, nil
		}
	}
	return 0, nil
}
