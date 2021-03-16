/*
Copyright The casbind Authors.
@Date: 2021/03/12 20:09
*/

package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	model2 "github.com/casbin/casbin/v2/model"

	"github.com/casbin/casbin/v2"

	"github.com/golang/protobuf/proto"

	"github.com/WenyXu/casbind/proto/command"

	"github.com/hashicorp/raft"
)

type FSMResponse struct {
	error error
}

type FSMEnforceResponse struct {
	ok    bool
	error error
}

var (
	NamespaceNotExist = errors.New("namespace not exist")
	UnmarshalFail     = errors.New("unmarshal failed")
)

func (s *Store) Apply(l *raft.Log) (e interface{}) {
	var cmd command.Command
	err := proto.Unmarshal(l.Data, &cmd)
	if err != nil {
		panic(fmt.Sprintf("failed to unmarshal cluster command: %s",
			err.Error()))
	}

	switch cmd.Type {
	case command.Type_COMMAND_TYPE_ENFORCE_REQUEST:
		var p command.EnforcePayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		var params []interface{}
		for _, b := range p.B {
			var tmp interface{}
			err = json.Unmarshal(b, &tmp)
			if err != nil {
				return &FSMEnforceResponse{error: UnmarshalFail}
			}
			params = append(params, tmp)
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			r, err := enforcer.Enforce(params...)
			if err != nil {
				return &FSMEnforceResponse{error: err}
			}
			return &FSMEnforceResponse{ok: r, error: err}
		}
		return &FSMResponse{error: NamespaceNotExist}
	case command.Type_COMMAND_TYPE_CREATE_NS:
		e, err := casbin.NewDistributedEnforcer()
		if err != nil {
			return &FSMResponse{error: err}
		}
		s.enforcers.Store(cmd.Ns, e)
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_SET_MODEL:
		var p command.SetModelFromString
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			model, err := model2.NewModelFromString(p.Text)
			if err != nil {
				return &FSMResponse{error: err}
			}
			enforcer.SetModel(model)
			log.Println("set model successfully")
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_ADD_POLICIES:
		var p command.AddPoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.AddPoliciesSelf(nil, p.Sec, p.PType, command.ToStringArray(p.Rules))
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_UPDATE_POLICIES:
		var p command.UpdatePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.UpdatePoliciesSelf(nil, p.Sec, p.PType, command.ToStringArray(p.OldRules), command.ToStringArray(p.NewRules))
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_UPDATE_POLICY:
		var p command.UpdatePolicyPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.UpdatePolicySelf(nil, p.Sec, p.PType, p.OldRule, p.NewRule)
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_REMOVE_POLICIES:
		var p command.RemovePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.RemovePoliciesSelf(nil, p.Sec, p.PType, command.ToStringArray(p.Rules))
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_REMOVE_FILTERED_POLICY:
		var p command.RemoveFilteredPolicyPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			panic(fmt.Sprintf("failed to unmarshal add policies payload: %s", err.Error()))
		}
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.RemoveFilteredPolicySelf(nil, p.Sec, p.PType, int(p.FieldIndex), p.FieldValues...)
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_CLEAR_POLICY:
		if e, ok := s.enforcers.Load(cmd.Ns); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			err := enforcer.ClearPolicySelf(nil)
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	default:
		return &FSMResponse{error: fmt.Errorf("unhandled command: %v", cmd.Type)}
	}

}

type fsmSnapshot struct {
	startT    time.Time
	logger    *log.Logger
	enforcers []byte
	meta      []byte
}

func writeUint64(w io.Writer, v uint64) error {
	return binary.Write(w, binary.LittleEndian, v)
}

type persistData struct {
	enforcers []byte
	meta      []byte
}

func (f fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	defer func() {
		f.logger.Printf("snapshot and persist took %s", time.Since(f.startT))
	}()
	err := func() error {
		data, err := json.Marshal(persistData{
			enforcers: f.enforcers,
			meta:      f.meta,
		})
		if err != nil {
			return err
		}
		// Write the cluster enforcers.
		if _, err := sink.Write(data); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}()

	if err != nil {
		sink.Cancel()
		return err
	}

	return nil
}

func (f fsmSnapshot) Release() {
}

func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	enforcers := make(map[interface{}]interface{})
	s.enforcers.Range(func(key, value interface{}) bool {
		enforcers[key] = value
		return true
	})
	var err error
	fsm := &fsmSnapshot{
		startT: time.Now(),
		logger: s.logger,
	}
	fsm.enforcers, err = json.Marshal(enforcers)
	if err != nil {
		s.logger.Printf("failed to encode enforcers for snapshot: %s", err.Error())
		return nil, err
	}
	fsm.meta, err = json.Marshal(s.meta)
	if err != nil {
		s.logger.Printf("failed to encode meta for snapshot: %s", err.Error())
		return nil, err
	}
	return fsm, nil
}

func (s *Store) Restore(closer io.ReadCloser) error {
	var data persistData
	err := json.NewDecoder(closer).Decode(&data)
	if err != nil {
		return err
	}
	var meta map[string]map[string]string
	err = json.Unmarshal(data.meta, &meta)
	if err != nil {
		return err
	}
	var enforcers map[string]casbin.DistributedEnforcer
	err = json.Unmarshal(data.meta, &enforcers)
	if err != nil {
		return err
	}
	s.enforcers = sync.Map{}
	for k, v := range enforcers {
		s.enforcers.Store(k, v)
	}
	return nil
}
