package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/casbin/casbin/v2"

	"github.com/casbin/casbin-mesh/proto/command"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
)

// CreateNamespace creates a new namespace.
func (s *Store) CreateNamespace(ctx context.Context, ns string) error {
	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_CREATE_NAMESPACE,
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

// SetModelFromString sets casbin model from string.
func (s *Store) SetModelFromString(ctx context.Context, ns string, text string) error {
	payload, err := proto.Marshal(&command.SetModelFromString{
		Text: text,
	})
	if err != nil {
		return err
	}

	cmd, err := proto.Marshal(&command.Command{
		Type:      command.Type_COMMAND_TYPE_SET_MODEL,
		Namespace: ns,
		Payload:   payload,
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

// Enforce executes enforcement.
func (s *Store) Enforce(ctx context.Context, ns string, level command.EnforcePayload_Level, freshness int64, params ...interface{}) (bool, error) {
	if level == command.EnforcePayload_QUERY_REQUEST_LEVEL_STRONG {
		var B [][]byte
		for _, p := range params {
			b, err := json.Marshal(p)
			if err != nil {
				return false, err
			}
			B = append(B, b)
		}

		payload, err := proto.Marshal(&command.EnforcePayload{
			B:         B,
			Level:     level,
			Freshness: freshness,
		})
		if err != nil {
			return false, err
		}

		cmd, err := proto.Marshal(&command.Command{
			Type:      command.Type_COMMAND_TYPE_ENFORCE_REQUEST,
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
		r := f.Response().(*FSMEnforceResponse)
		return r.ok, r.error
	}
	if level == command.EnforcePayload_QUERY_REQUEST_LEVEL_WEAK && s.raft.State() != raft.Leader {
		return false, ErrNotLeader
	}
	if level == command.EnforcePayload_QUERY_REQUEST_LEVEL_NONE &&
		freshness > 0 &&
		s.raft.State() != raft.Leader &&
		time.Since(s.raft.LastContact()).Nanoseconds() > freshness {
		return false, ErrStaleRead
	}
	if e, ok := s.enforcers.Load(ns); ok {
		enforcer := e.(*casbin.DistributedEnforcer)
		r, err := enforcer.Enforce(params...)
		return r, err
	} else {
		return false, NamespaceNotExist
	}
}

// SetMetadata adds the metadata md to any existing metadata for
// this node.
func (s *Store) SetMetadata(md map[string]string) error {
	return s.setMetadata(s.raftID, md)
}

// setMetadata adds the metadata md to any existing metadata for
// the given node ID.
func (s *Store) setMetadata(id string, md map[string]string) error {
	// Check local data first.
	if func() bool {
		s.metaMu.RLock()
		defer s.metaMu.RUnlock()
		if _, ok := s.meta[id]; ok {
			for k, v := range md {
				if s.meta[id][k] != v {
					return false
				}
			}
			return true
		}
		return false
	}() {
		// Local data is same as data being pushed in,
		// nothing to do.
		return nil
	}

	ms := &command.MetadataSet{
		RaftId: id,
		Data:   md,
	}
	bms, err := proto.Marshal(ms)
	if err != nil {
		return err
	}

	c := &command.Command{
		Type:    command.Type_COMMAND_TYPE_METADATA_SET,
		Payload: bms,
	}
	bc, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(bc, s.ApplyTimeout)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}

	return nil
}
