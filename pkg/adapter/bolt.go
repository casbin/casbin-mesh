package adapter

import (
	"github.com/boltdb/bolt"
	"io"
	"io/ioutil"
	"log"
	"sync"
)

type BoltStore struct {
	conn *bolt.DB
	path string
	mu   *sync.Mutex
}

// Restore overwrites the local file
func (b *BoltStore) Restore(reader io.Reader) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// close database connection
	err := b.conn.Close()
	if err != nil {
		log.Println("failed to close database connection", err)
		return err
	}
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println("failed to read data", err)
		return err
	}

	// overwrite or create
	err = ioutil.WriteFile(b.path, buf, dbFileMode)
	if err != nil {
		log.Println("failed to write the snapshot into file", err)
		return err
	}

	// re-open
	b.conn, err = bolt.Open(b.path, dbFileMode, nil)
	if err != nil {
		log.Println("failed to open the database file", err)
		return err
	}

	return nil
}

// Snapshot writes the entire database to a writer.
func (b BoltStore) Snapshot(writer io.Writer) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := b.conn.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(writer)
		return err
	})
	if err != nil {
		log.Println("failed to snapshot the database", err)
		return err
	}
	return nil
}

func (b BoltStore) Foreach(fn func(name []byte, b *bolt.Bucket) error) error {
	return b.conn.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			return fn(name, b)
		})
	})
}

type IBoltStore interface {
	Foreach(fn func(name []byte, b *bolt.Bucket) error) error
	Restore(reader io.Reader) error
	Snapshot(writer io.Writer) error
}

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)

// Options contains all the configuration used to open the BoltDB
type Options struct {
	// Path is the file path to the BoltDB to use
	Path string

	// BoltOptions contains any specific BoltDB options you might
	// want to specify [e.g. open timeout]
	BoltOptions *bolt.Options

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool
}

// NewBoltStore takes a file path and returns a connected Raft backend.
func NewBoltStore(path string) (*BoltStore, error) {
	return New(Options{Path: path})
}

// New uses the supplied options to open the BoltDB and prepare it for use as a raft backend.
func New(options Options) (*BoltStore, error) {
	// Try to connect
	handle, err := bolt.Open(options.Path, dbFileMode, options.BoltOptions)
	if err != nil {
		return nil, err
	}
	handle.NoSync = options.NoSync

	// Create the new store
	store := &BoltStore{
		mu:   &sync.Mutex{},
		conn: handle,
		path: options.Path,
	}

	return store, nil
}
