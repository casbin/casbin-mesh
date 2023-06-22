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
	"context"

	"github.com/casbin/casbin-mesh/pkg/adapter"
	"github.com/casbin/casbin-mesh/proto/command"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
)

// AddPolicies implements the casbin.Adapter interface.
func (s *Store) AddPolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) ([][]string, error) {
	payload, err := proto.Marshal(&command.AddPoliciesPayload{
		Sec:   sec,
		PType: pType,
		Rules: command.NewStringArray(rules),
	})
	if err != nil {
		return nil, err
	}

	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_ADD_POLICIES,
		Namespace: ns,
		Payload:   payload,
		Metadata:  nil,
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
	r := f.Response().(*FSMResponse)
	return r.effectedRules, r.error
}

// RemovePolicies implements the casbin.Adapter interface.
func (s *Store) RemovePolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) ([][]string, error) {
	payload, err := proto.Marshal(&command.RemovePoliciesPayload{
		Sec:   sec,
		PType: pType,
		Rules: command.NewStringArray(rules),
	})
	if err != nil {
		return nil, err
	}

	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_REMOVE_POLICIES,
		Namespace: ns,
		Payload:   payload,
		Metadata:  nil,
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
	r := f.Response().(*FSMResponse)
	return r.effectedRules, r.error
}

// RemoveFilteredPolicy implements the casbin.Adapter interface.
func (s *Store) RemoveFilteredPolicy(ctx context.Context, ns string, sec string, pType string, fi int32, fv []string) ([][]string, error) {
	payload, err := proto.Marshal(&command.RemoveFilteredPolicyPayload{
		Sec:         sec,
		PType:       pType,
		FieldIndex:  fi,
		FieldValues: fv,
	})
	if err != nil {
		return nil, err
	}

	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_REMOVE_FILTERED_POLICY,
		Namespace: ns,
		Payload:   payload,
		Metadata:  nil,
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
	r := f.Response().(*FSMResponse)
	return r.effectedRules, r.error
}

// UpdatePolicies implements the casbin.Adapter interface.
func (s *Store) UpdatePolicies(ctx context.Context, ns string, sec string, pType string, nr, or [][]string) (bool, error) {
	payload, err := proto.Marshal(&command.UpdatePoliciesPayload{
		Sec:      sec,
		PType:    pType,
		NewRules: command.NewStringArray(nr),
		OldRules: command.NewStringArray(or),
	})
	if err != nil {
		return false, err
	}

	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_UPDATE_POLICIES,
		Namespace: ns,
		Payload:   payload,
		Metadata:  nil,
	})
	if err != nil {
		return false, err
	}

	f := s.raft.Apply(cmd, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return false, ErrNotLeader
		}
		return false, e.Error()
	}
	r := f.Response().(*FSMResponse)
	return r.effected, r.error
}

// ClearPolicy implements the casbin.Adapter interface.
func (s *Store) ClearPolicy(ctx context.Context, ns string) error {
	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_CLEAR_POLICY,
		Namespace: ns,
		Payload:   nil,
		Metadata:  nil,
	})
	if err != nil {
		return err
	}

	f := s.raft.Apply(cmd, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}
	r := f.Response().(*FSMResponse)
	return r.error
}

type ListPoliciesOptions struct {
	Cursor  string
	Limit   int64
	Skip    int64
	Reverse bool
}

var defaultListPoliciesOptions = ListPoliciesOptions{
	Limit: 1000,
}

// Policies list policies
func (s *Store) Policies(ctx context.Context, ns string, options *ListPoliciesOptions) ([][]string, error) {
	if options == nil {
		options = &defaultListPoliciesOptions
	}
	var (
		policies [][]string
		err      error
	)
	err = s.enforcersState.View(func(tx *adapter.Tx) error {
		bucket := tx.Bucket([]byte(ns))
		policies, err = bucket.List(options.Cursor, options.Skip, options.Limit, options.Reverse)
		return err
	})
	return policies, err

}
