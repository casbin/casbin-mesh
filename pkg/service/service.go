/*
Copyright The casbind Authors.
@Date: 2021/03/10 15:24
*/

package service

import (
	"context"

	"github.com/WenyXu/casbind/proto/command"

	"github.com/WenyXu/casbind/pkg/store"
)

type service struct {
	store *store.Store
}

func (s service) Join(ctx context.Context, id, addr string, voter bool, metadata map[string]string) error {
	return s.store.Join(id, addr, voter, metadata)
}

func (s service) Remove(ctx context.Context, id string) error {
	return s.store.Remove(id)
}

func (s service) CreateNamespace(ctx context.Context, ns string) error {
	return s.store.CreateNamespace(ctx, ns)
}

func (s service) SetModelFromString(ctx context.Context, ns string, text string) error {
	return s.store.SetModelFromString(ctx, ns, text)
}

func (s service) Enforce(ctx context.Context, ns string, level int32, freshness int64, params ...interface{}) (bool, error) {
	return s.store.Enforce(ctx, ns, command.EnforcePayload_Level(level), freshness, params...)
}

func (s service) AddPolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error {
	return s.store.AddPolicies(ctx, ns, sec, pType, rules)
}

func (s service) RemovePolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error {
	return s.store.RemovePolicies(ctx, ns, sec, pType, rules)
}

func (s service) RemoveFilteredPolicy(ctx context.Context, ns string, sec string, pType string, fi int32, fv []string) error {
	return s.store.RemoveFilteredPolicy(ctx, ns, sec, pType, fi, fv)
}

func (s service) UpdatePolicy(ctx context.Context, ns string, sec string, pType string, nr, or []string) error {
	return s.store.UpdatePolicy(ctx, ns, sec, pType, nr, or)
}

func (s service) UpdatePolicies(ctx context.Context, ns string, sec string, pType string, nr, or [][]string) error {
	return s.store.UpdatePolicies(ctx, ns, sec, pType, nr, or)
}

func (s service) ClearPolicy(ctx context.Context, ns string) error {
	return s.store.ClearPolicy(ctx, ns)
}

func (s service) Stats(ctx context.Context) (map[string]interface{}, error) {
	return s.store.Stats()
}

type Service interface {
	Stats(ctx context.Context) (map[string]interface{}, error)
	CreateNamespace(ctx context.Context, ns string) error
	SetModelFromString(ctx context.Context, ns string, text string) error
	Enforce(ctx context.Context, ns string, level int32, freshness int64, params ...interface{}) (bool, error)
	AddPolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error
	RemovePolicies(ctx context.Context, ns string, sec string, pType string, rules [][]string) error
	RemoveFilteredPolicy(ctx context.Context, ns string, sec string, pType string, fi int32, fv []string) error
	UpdatePolicy(ctx context.Context, ns string, sec string, pType string, nr, or []string) error
	UpdatePolicies(ctx context.Context, ns string, sec string, pType string, nr, or [][]string) error
	ClearPolicy(ctx context.Context, ns string) error
	Join(ctx context.Context, id, addr string, voter bool, metadata map[string]string) error
	Remove(ctx context.Context, id string) error
}

func New(store *store.Store) Service {
	return &service{store}
}
