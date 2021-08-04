package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/casbin/casbin-mesh/pkg/adapter"
	"io"
	"log"
	"sync"
	"time"

	model2 "github.com/casbin/casbin/v2/model"

	"github.com/casbin/casbin/v2"

	"github.com/golang/protobuf/proto"

	"github.com/casbin/casbin-mesh/proto/command"

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
	NamespaceExisted = errors.New("namespace already existed")

	// NamespaceNotExist namespace not exist
	NamespaceNotExist = errors.New("namespace not exist")
	// UnmarshalFailed unmarshal failed
	UnmarshalFailed = errors.New("unmarshal failed")
)

var persist = func() bool { return true }

func (s *Store) Apply(l *raft.Log) (e interface{}) {
	var cmd command.Command
	err := proto.Unmarshal(l.Data, &cmd)
	if err != nil {
		return &FSMEnforceResponse{error: UnmarshalFailed}
	}

	switch cmd.Type {
	case command.Type_COMMAND_TYPE_ENFORCE_REQUEST:
		var p command.EnforcePayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		var params []interface{}
		for _, b := range p.B {
			var tmp interface{}
			err = json.Unmarshal(b, &tmp)
			if err != nil {
				return &FSMEnforceResponse{error: UnmarshalFailed}
			}
			params = append(params, tmp)
		}
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			r, err := enforcer.Enforce(params...)
			if err != nil {
				return &FSMEnforceResponse{error: err}
			}
			return &FSMEnforceResponse{ok: r, error: err}
		}
		return &FSMResponse{error: NamespaceNotExist}
	case command.Type_COMMAND_TYPE_CREATE_NAMESPACE:
		_, ok := s.enforcers.Load(cmd.Namespace)
		if ok {
			return &FSMResponse{NamespaceExisted}
		}
		e, err := casbin.NewDistributedEnforcer()
		if err != nil {
			return &FSMResponse{error: err}
		}

		s.enforcers.Store(cmd.Namespace, e)
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_SET_MODEL:
		var p command.SetModelFromString
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			a, err := adapter.NewAdapter(s.enforcersState, cmd.Namespace, "")
			if err != nil {
				return &FSMResponse{error: err}
			}
			model, err := model2.NewModelFromString(p.Text)
			if err != nil {
				return &FSMResponse{error: err}
			}
			// load existed policies
			err = a.LoadPolicy(model)
			if err != nil {
				return &FSMResponse{error: err}
			}
			err = enforcer.InitWithModelAndAdapter(model, a)
			if err != nil {
				return &FSMResponse{error: err}
			}
			// rebuild role links
			err = enforcer.BuildRoleLinks()
			if err != nil {
				return &FSMResponse{error: err}
			}
			log.Println("set model successfully")
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_ADD_POLICIES:
		var p command.AddPoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return &FSMResponse{error: NamespaceNotExist}
		}
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.AddPoliciesSelf(persist, p.Sec, p.PType, command.ToStringArray(p.Rules))
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
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.UpdatePoliciesSelf(persist, p.Sec, p.PType, command.ToStringArray(p.OldRules), command.ToStringArray(p.NewRules))
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
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.RemovePoliciesSelf(persist, p.Sec, p.PType, command.ToStringArray(p.Rules))
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
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			_, err := enforcer.RemoveFilteredPolicySelf(persist, p.Sec, p.PType, int(p.FieldIndex), p.FieldValues...)
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_CLEAR_POLICY:
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			err := enforcer.ClearPolicySelf(persist)
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_METADATA_SET:
		var ms command.MetadataSet
		if err := proto.UnmarshalMerge(cmd.Payload, &ms); err != nil {
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		func() {
			s.metaMu.Lock()
			defer s.metaMu.Unlock()
			if _, ok := s.meta[ms.RaftId]; !ok {
				s.meta[ms.RaftId] = make(map[string]string)
			}
			for k, v := range ms.Data {
				s.meta[ms.RaftId][k] = v
			}
		}()
		return &FSMEnforceResponse{}
	case command.Type_COMMAND_TYPE_METADATA_DELETE:
		var md command.MetadataDelete
		if err := proto.UnmarshalMerge(cmd.Payload, &md); err != nil {
			return &FSMEnforceResponse{error: UnmarshalFailed}
		}
		func() {
			s.metaMu.Lock()
			defer s.metaMu.Unlock()
			delete(s.meta, md.RaftId)
		}()
		return &FSMEnforceResponse{}
	default:
		return &FSMResponse{error: fmt.Errorf("unhandled command: %v", cmd.Type)}
	}

}

type fsmSnapshot struct {
	startT time.Time
	logger *log.Logger
	models []byte
	state  []byte
	meta   []byte
}

type persistData struct {
	Models []byte
	State  []byte
	Meta   []byte
}

// Persist implements persistence of states
func (f fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	defer func() {
		f.logger.Printf("snapshot and persist took %s", time.Since(f.startT))
	}()
	err := func() error {
		data, err := json.Marshal(persistData{
			State:  f.state,
			Models: f.models,
			Meta:   f.meta,
		})
		if err != nil {
			return err
		}
		// Write the cluster Enforcers.
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

// Snapshot creates a persistable state for application
func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	var err error
	fsm := &fsmSnapshot{
		startT: time.Now(),
		logger: s.logger,
	}
	writer := new(bytes.Buffer)
	err = s.enforcersState.Snapshot(writer)
	if err != nil {
		s.logger.Printf("failed to encode enforcerState: %s", err.Error())
		return nil, err
	}
	models := make(map[string]string)
	s.enforcers.Range(func(key, value interface{}) bool {
		if e, ok := value.(*casbin.DistributedEnforcer); ok {
			models[key.(string)] = e.GetModel().ToText()
		}
		return true
	})
	fsm.state = writer.Bytes()
	fsm.meta, err = json.Marshal(s.meta)
	if err != nil {
		s.logger.Printf("failed to encode Meta: %s", err.Error())
		return nil, err
	}
	fsm.models, err = json.Marshal(models)
	if err != nil {
		s.logger.Printf("failed to encode Meta: %s", err.Error())
		return nil, err
	}
	return fsm, nil
}

// Restore restores form a preexisted states
func (s *Store) Restore(closer io.ReadCloser) error {
	var err error
	var data persistData
	err = json.NewDecoder(closer).Decode(&data)
	if err != nil {
		s.logger.Println("failed to decode restore data", err)
		return err
	}

	err = s.enforcersState.Restore(bytes.NewReader(data.State))
	if err != nil {
		s.logger.Println("failed to restore enforcer state", err)
		return err
	}
	s.enforcers = sync.Map{}
	models := make(map[string]string)
	err = json.Unmarshal(data.Models, &models)
	if err != nil {
		s.logger.Println("failed to unmarshal models state", err)
		return err
	}
	err = s.enforcersState.Foreach(func(name []byte, b *bolt.Bucket) error {
		if model, ok := models[string(name)]; ok {
			enforcer, err := casbin.NewDistributedEnforcer()
			if err != nil {
				s.logger.Println("failed to create enforcer", err)
				return err
			}
			a, err := adapter.NewAdapter(s.enforcersState, string(name), "")
			if err != nil {
				s.logger.Println("failed to create adapter", err)
				return err
			}
			model, err := model2.NewModelFromString(model)
			if err != nil {
				s.logger.Println("failed to create model", err)
				return err
			}
			err = a.LoadPolicy(model)
			if err != nil {
				s.logger.Println("failed to load policy", err)
				return err
			}
			err = enforcer.InitWithModelAndAdapter(model, a)
			if err != nil {
				s.logger.Println("failed to init enforcer", err)
				return err
			}
			err = enforcer.BuildRoleLinks()
			if err != nil {
				s.logger.Println("failed to rebuild role links", err)
				return err
			}
			s.enforcers.Store(string(name), enforcer)
		} else {
			s.logger.Printf("%s namespace is not existing a valid model\n", string(name))
		}
		return nil
	})
	if err != nil {
		s.logger.Println("failed to restore enforcer ", err)
		return err
	}
	return nil
}
