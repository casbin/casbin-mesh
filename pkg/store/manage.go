// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package store

import (
	"context"
	"github.com/casbin/casbin-mesh/proto/command"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
)

type ListNamespacesResponse struct {
	error
	namespace []string
}

func (s *Store) ListNamespace(ctx context.Context) ([]string, error) {
	cmd, err := proto.Marshal(&command.Command{Type: command.Type_COMMAND_TYPE_LIST_NAMESPACES})
	if err != nil {
		return nil, err
	}
	f := s.raft.Apply(cmd, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return nil, ErrNotLeader
		}
		return nil, e.Error()
	}
	r := f.Response().(*ListNamespacesResponse)
	return r.namespace, r.error
}

type ListPoliciesResponse struct {
	policies [][]string
	error
}

func (s *Store) ListPolicies(ctx context.Context, namespace, cursor string, skip, limit int64, reverse bool) ([][]string, error) {
	payload, err := proto.Marshal(&command.ListPoliciesPayload{Cursor: cursor, Skip: skip, Limit: limit, Reverse: reverse})
	if err != nil {
		return nil, err
	}
	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_LIST_POLICIES,
		Namespace: namespace,
		Payload:   payload,
	})
	if err != nil {
		return nil, err
	}
	f := s.raft.Apply(cmd, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return nil, ErrNotLeader
		}
		return nil, e.Error()
	}
	r := f.Response().(*ListPoliciesResponse)
	return r.policies, r.error
}

type PrintModelResponse struct {
	model string
	error
}

func (s *Store) PrintModel(ctx context.Context, namespace string) (string, error) {
	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_PRINT_MODEL,
		Namespace: namespace,
	})
	if err != nil {
		return "", err
	}
	f := s.raft.Apply(cmd, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return "", ErrNotLeader
		}
		return "", e.Error()
	}
	r := f.Response().(*PrintModelResponse)
	return r.model, r.error
}
