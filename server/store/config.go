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

package store

import (
	"time"

	"github.com/casbin/casbin-mesh/server/auth"
	"go.uber.org/zap"
)

// StoreConfig represents the configuration of the underlying Store.
type StoreConfig struct {
	Dir      string      // The working directory for raft.
	Tn       Transport   // The underlying Transport for raft.
	ID       string      // Node ID.
	Logger   *zap.Logger // The logger to use to log stuff.
	AuthType auth.AuthType
	*auth.CredentialsStore
	AdvAddr string

	RaftLogLevel       string
	ShutdownOnRemove   bool
	SnapshotThreshold  uint64
	SnapshotInterval   time.Duration
	LeaderLeaseTimeout time.Duration
	HeartbeatTimeout   time.Duration
	ElectionTimeout    time.Duration
	ApplyTimeout       time.Duration
}
