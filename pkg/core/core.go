package core

import (
	"context"

	"github.com/casbin/casbin-mesh/pkg/store"
	"github.com/casbin/casbin-mesh/proto/command"
)

type core struct {
	store *store.Store
}

func (s core) Join(ctx context.Context, id, addr string, voter bool, metadata map[string]string) error {
	return s.store.Join(id, addr, voter, metadata)
}

func (s core) Remove(ctx context.Context, id string) error {
	return s.store.Remove(id)
}

func (s core) CreateNamespace(ctx context.Context, ns string) error {
	return s.store.CreateNamespace(ctx, ns)
}

func (s core) SetModelFromString(ctx context.Context, ns string, text string) error {
	return s.store.SetModelFromString(ctx, ns, text)
}

func (s core) Enforce(ctx context.Context, ns string, level int32, freshness int64, params ...interface{}) (bool, error) {
	return s.store.Enforce(ctx, ns, command.EnforcePayload_Level(level), freshness, params...)
}

func (s core) AddPolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error {
	return s.store.AddPolicies(ctx, ns, sec, pType, rules)
}

func (s core) RemovePolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error {
	return s.store.RemovePolicies(ctx, ns, sec, pType, rules)
}

func (s core) RemoveFilteredPolicy(ctx context.Context, ns string, sec string, pType string, fi int32, fv []string) error {
	return s.store.RemoveFilteredPolicy(ctx, ns, sec, pType, fi, fv)
}

func (s core) UpdatePolicies(ctx context.Context, ns string, sec string, pType string, nr, or [][]string) error {
	return s.store.UpdatePolicies(ctx, ns, sec, pType, nr, or)
}

func (s core) ClearPolicy(ctx context.Context, ns string) error {
	return s.store.ClearPolicy(ctx, ns)
}

func (s core) Stats(ctx context.Context) (map[string]interface{}, error) {
	return s.store.Stats()
}

func (s core) IsLeader(ctx context.Context) bool {
	return s.store.IsLeader()
}

func (s core) LeaderAddr(ctx context.Context) string {
	return s.store.LeaderAddr()
}

// LeaderAPIAddr returns the API address of the leader, as known by this node.
func (s core) LeaderAPIAddr() string {
	id, err := s.store.LeaderID()
	if err != nil {
		return ""
	}
	return s.store.Metadata(id, "api_addr")
}

// LeaderAPIProto returns the protocol used by the leader, as known by this node.
func (s core) LeaderAPIProto() string {
	id, err := s.store.LeaderID()
	if err != nil {
		return "http"
	}

	p := s.store.Metadata(id, "api_proto")
	if p == "" {
		return "http"
	}
	return p
}

type Core interface {
	LeaderAPIProto() string
	LeaderAPIAddr() string
	IsLeader(ctx context.Context) bool
	LeaderAddr(ctx context.Context) string
	Stats(ctx context.Context) (map[string]interface{}, error)
	CreateNamespace(ctx context.Context, ns string) error
	SetModelFromString(ctx context.Context, ns string, text string) error
	Enforce(ctx context.Context, ns string, level int32, freshness int64, params ...interface{}) (bool, error)
	AddPolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error
	RemovePolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error
	RemoveFilteredPolicy(ctx context.Context, ns string, sec string, pType string, fi int32, fv []string) error
	UpdatePolicies(ctx context.Context, ns string, sec string, pType string, nr, or [][]string) error
	ClearPolicy(ctx context.Context, ns string) error
	Join(ctx context.Context, id, addr string, voter bool, metadata map[string]string) error
	Remove(ctx context.Context, id string) error
}

func New(store *store.Store) Core {
	return &core{store}
}
