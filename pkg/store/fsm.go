/*
Copyright The casbind Authors.
@Date: 2021/03/12 20:09
*/

package store

import (
	"io"

	"github.com/hashicorp/raft"
)

func (s *Store) Apply(l *raft.Log) interface{} {
	panic("implement me")
}

func (s *Store) Snapshot() (raft.FSMSnapshot, error) {
	panic("implement me")
}

func (s *Store) Restore(closer io.ReadCloser) error {
	panic("implement me")
}
