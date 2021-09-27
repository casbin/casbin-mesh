package store

import (
	"github.com/casbin/casbin-mesh/pkg/auth"
	"log"
)

// StoreConfig represents the configuration of the underlying Store.
type StoreConfig struct {
	Dir      string      // The working directory for raft.
	Tn       Transport   // The underlying Transport for raft.
	ID       string      // Node ID.
	Logger   *log.Logger // The logger to use to log stuff.
	AuthType auth.AuthType
	*auth.CredentialsStore
	AdvAddr string
}
