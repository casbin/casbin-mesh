package store

import (
	"context"

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
