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
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casbin/casbin-mesh/pkg/adapter"
	"github.com/casbin/casbin-mesh/pkg/auth"
	"io"
	"io/ioutil"
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
	error         error
	effected      bool
	effectedRules [][]string
}

type FSMEnforceResponse struct {
	ok    bool
	error error
}

var (
	NamespaceExisted = errors.New("namespace already existed")
	ModelUnsetYet    = errors.New("model unset yet")
	// NamespaceNotExist namespace not exist
	NamespaceNotExist = errors.New("namespace not exist")
	// UnmarshalFailed unmarshal failed
	UnmarshalFailed = errors.New("unmarshal failed")
	// Transaction failed
	StateTransactionFailed = errors.New("state transaction failed")
)

var persist = func() bool { return true }

func (s *Store) Apply(l *raft.Log) (e interface{}) {
	var cmd command.Command
	err := proto.Unmarshal(l.Data, &cmd)
	if err != nil {
		return &FSMResponse{error: UnmarshalFailed}
	}
	switch cmd.Type {
	case command.Type_COMMAND_TYPE_LIST_NAMESPACES:
		var ns []string
		err := s.enforcersState.ForEach(func(namespace []byte, bucket *adapter.Bucket) error {
			ns = append(ns, string(namespace))
			return nil
		})
		if err != nil {
			return &ListNamespacesResponse{error: StateTransactionFailed}
		}
		return &ListNamespacesResponse{
			namespace: ns,
		}
	case command.Type_COMMAND_TYPE_PRINT_MODEL:
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			model := enforcer.GetModel()
			if model == nil {
				return &FSMEnforceResponse{error: err}
			}
			return &PrintModelResponse{model: model.ToText()}
		}
		return &PrintModelResponse{error: NamespaceNotExist}
	case command.Type_COMMAND_TYPE_LIST_POLICIES:
		ns := cmd.Namespace
		var policies [][]string
		var payload command.ListPoliciesPayload
		if cmd.Payload != nil {
			err := proto.Unmarshal(cmd.Payload, &payload)
			if err != nil {
				return &ListPoliciesResponse{error: UnmarshalFailed}
			}
		} else {
			//default value
			payload.Limit = 1000
		}
		err = s.enforcersState.View(func(tx *adapter.Tx) error {
			policies, err = tx.Bucket([]byte(ns)).List(payload.Cursor, payload.Skip, payload.Limit, payload.Reverse)
			return err
		})
		if err != nil {
			return &ListPoliciesResponse{error: StateTransactionFailed}
		}
		return &ListPoliciesResponse{
			policies: policies,
		}
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
		return &FSMEnforceResponse{error: NamespaceNotExist}
	case command.Type_COMMAND_TYPE_CREATE_NAMESPACE:
		_, ok := s.enforcers.Load(cmd.Namespace)
		if ok {
			return &FSMResponse{error: NamespaceExisted}
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
			return &FSMResponse{error: UnmarshalFailed}
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
			err = enforcer.InitWithModelAndAdapter(model, a)
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
		var effectedRules [][]string
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			if enforcer.GetModel() == nil {
				return &FSMResponse{error: ModelUnsetYet}
			}
			effectedRules, err = enforcer.AddPoliciesSelf(persist, p.Sec, p.PType, command.ToStringArray(p.Rules))
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{effectedRules: effectedRules}
	case command.Type_COMMAND_TYPE_UPDATE_POLICIES:
		var p command.UpdatePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return &FSMResponse{error: UnmarshalFailed}
		}
		var effected bool
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			effected, err = enforcer.UpdatePoliciesSelf(persist, p.Sec, p.PType, command.ToStringArray(p.OldRules), command.ToStringArray(p.NewRules))
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{effected: effected}
	case command.Type_COMMAND_TYPE_REMOVE_POLICIES:
		var p command.RemovePoliciesPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return &FSMResponse{error: UnmarshalFailed}
		}
		var effectedRules [][]string
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			effectedRules, err = enforcer.RemovePoliciesSelf(persist, p.Sec, p.PType, command.ToStringArray(p.Rules))
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{effectedRules: effectedRules}
	case command.Type_COMMAND_TYPE_REMOVE_FILTERED_POLICY:
		var p command.RemoveFilteredPolicyPayload
		if err = proto.Unmarshal(cmd.Payload, &p); err != nil {
			return &FSMResponse{error: UnmarshalFailed}
		}
		var effectedRules [][]string
		if e, ok := s.enforcers.Load(cmd.Namespace); ok {
			enforcer := e.(*casbin.DistributedEnforcer)
			effectedRules, err = enforcer.RemoveFilteredPolicySelf(persist, p.Sec, p.PType, int(p.FieldIndex), p.FieldValues...)
			if err != nil {
				return &FSMResponse{error: err}
			}
		} else {
			return &FSMResponse{error: NamespaceNotExist}
		}
		return &FSMResponse{effectedRules: effectedRules}
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
			return &FSMResponse{error: UnmarshalFailed}
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
		return &FSMResponse{}
	case command.Type_COMMAND_TYPE_METADATA_DELETE:
		var md command.MetadataDelete
		if err := proto.UnmarshalMerge(cmd.Payload, &md); err != nil {
			return &FSMResponse{error: UnmarshalFailed}
		}
		func() {
			s.metaMu.Lock()
			defer s.metaMu.Unlock()
			delete(s.meta, md.RaftId)
		}()
		return &FSMResponse{}
	default:
		return &FSMResponse{error: fmt.Errorf("unhandled command: %v", cmd.Type)}
	}

}

type fsmSnapshot struct {
	startT          time.Time
	logger          *log.Logger
	models          []byte
	state           []byte
	meta            []byte
	credentialStore []byte
}

type persistData struct {
	Models          []byte
	State           []byte
	Meta            []byte
	CredentialStore []byte
}

// SnapshotHdr is used to identify the snapshot protocol version.
// length 8 bytes
func SnapshotHdr() []byte {
	hdr := [8]byte{}
	// protocol version
	binary.LittleEndian.PutUint16(hdr[0:], uint16(1))
	return hdr[:]
}

// Persist implements persistence of states
func (f fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	defer func() {
		f.logger.Printf("snapshot and persist took %s", time.Since(f.startT))
	}()
	err := func() error {
		data, err := json.Marshal(persistData{
			State:           f.state,
			Models:          f.models,
			Meta:            f.meta,
			CredentialStore: f.credentialStore,
		})
		if err != nil {
			return err
		}
		// JSON the cluster Enforcers.
		if _, err := sink.Write(append(SnapshotHdr(), data...)); err != nil {
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
	if s.authCredStore != nil {
		credStoreWriter := new(bytes.Buffer)
		if err := s.authCredStore.Snapshot(credStoreWriter); err != nil {
			s.logger.Println("failed to snapshot authCredStore", err)
		} else {
			fsm.credentialStore = credStoreWriter.Bytes()
		}
	}
	return fsm, nil
}

// Restore restores form a preexisted states
func (s *Store) Restore(closer io.ReadCloser) error {
	var err error
	var data persistData
	snap, err := ioutil.ReadAll(closer)
	err = json.Unmarshal(snap[8:], &data)
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
	err = s.enforcersState.ForEach(func(name []byte, bucket *adapter.Bucket) error {
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
			err = enforcer.InitWithModelAndAdapter(model, a)
			if err != nil {
				s.logger.Println("failed to init enforcer", err)
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
	if data.CredentialStore != nil {
		s.authCredStore = auth.NewCredentialsStore()
		err := s.authCredStore.Load(bytes.NewReader(data.CredentialStore))
		if err != nil {
			s.logger.Println("failed to load authCredStore ", err)
			return err
		}
	}
	return nil
}
