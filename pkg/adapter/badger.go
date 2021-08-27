package adapter

import (
	"github.com/dgraph-io/badger/v3"
	"time"

	"io"
	"log"
	"sync"
)

var (
	// Prefix names to distingish between namespace and policies
	prefixNamespace = []byte{0x0}
	prefixPolicies  = []byte{0x1}

	// Default value for maxPendingWrites is 256, to minimise memory usage
	// and overall finish time.
	maxPendingWrites = 256
)

// BadgerStore provides access to Badger for Raft to store and retrieve
// log entries. It also provides key/value storage, and can be used as
// a LogStore and StableStore.
type BadgerStore struct {
	// conn is the underlying handle to the db.
	conn *badger.DB

	// The path to the Badger database directory.
	path string

	vlogTicker          *time.Ticker // runs every 1m, check size of vlog and run GC conditionally.
	mandatoryVlogTicker *time.Ticker // runs every 10m, we always run vlog GC.

	mu *sync.Mutex
}

// Restore overwrites the local file
func (b *BadgerStore) Restore(reader io.Reader) error {

	err := b.conn.Load(reader, maxPendingWrites)
	if err != nil {
		log.Println("failed to open the database file", err)
		return err
	}

	return nil
}

// Snapshot writes the entire database to a writer.
func (b BadgerStore) Snapshot(writer io.Writer) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.conn.Backup(writer, 0)
	if err != nil {
		log.Println("failed to snapshot the database", err)
		return err
	}
	return nil
}

type Bucket struct {
	conn      *badger.DB
	txn       *badger.Txn
	namespace []byte
}

type Tx struct {
	txn  *badger.Txn
	conn *badger.DB
}

func (tx *Tx) CreateBucketIfNotExists(name []byte) (*Bucket, error) {
	_, err := tx.txn.Get(name)
	if err != nil {
		switch err {
		// try set new namespace
		case badger.ErrKeyNotFound:
			err := tx.txn.Set(append(prefixNamespace, name...), []byte{})
			// error
			if err != nil {
				return nil, err
			}
		// error
		default:
			return nil, err
		}
	}
	return &Bucket{conn: tx.conn, namespace: name, txn: tx.txn}, nil
}

func (tx *Tx) View(fn func(txn *badger.Txn) error) error {
	txn := tx.conn.NewTransaction(true)
	if err := fn(txn); err != nil {
		return err
	}
	return txn.Commit()
}

func (tx *Tx) Bucket(name []byte) *Bucket {
	return &Bucket{conn: tx.conn, namespace: name, txn: tx.txn}
}

func (bucket *Bucket) ForEach(fn func(key []byte, value []byte) error) error {
	return bucket.conn.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		var value []byte
		prefix := append(prefixPolicies, bucket.namespace...)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			value, err := item.ValueCopy(value)
			if err != nil {
				return err
			}
			// remove prefix
			if err := fn(k[len(prefix):], value); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bucket *Bucket) withPrefix(key []byte) []byte {
	return append(prefixPolicies, append(bucket.namespace, key...)...)
}

func (bucket *Bucket) Exist(key []byte) bool {
	_, err := bucket.txn.Get(bucket.withPrefix(key))
	if err != nil {
		return false
	}
	return true
}

func (bucket *Bucket) Put(key []byte, value []byte) error {
	return bucket.txn.Set(bucket.withPrefix(key), value)
}

func (bucket *Bucket) Delete(key []byte) error {
	return bucket.txn.Delete(bucket.withPrefix(key))
}

func (b BadgerStore) ForEach(fn func(namespace []byte, bucket *Bucket) error) error {
	return b.conn.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := prefixNamespace
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			name := k[len(prefix):]
			// remove prefix
			if err := fn(name, &Bucket{conn: b.conn, namespace: name, txn: txn}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (b BadgerStore) View(fn func(tx *Tx) error) error {
	txn := b.conn.NewTransaction(true)
	tx := &Tx{conn: b.conn, txn: txn}
	err := fn(tx)
	if err != nil {
		txn.Discard()
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

func (b BadgerStore) Update(fn func(tx *Tx) error) error {
	txn := b.conn.NewTransaction(true)
	tx := &Tx{conn: b.conn, txn: txn}
	err := fn(tx)
	if err != nil {
		txn.Discard()
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

type IBoltStore interface {
	Restore(reader io.Reader) error
	Snapshot(writer io.Writer) error
}

const (
	// Permissions to use on the db file. This is only used if the
	// database file does not exist and needs to be created.
	dbFileMode = 0600
)

// Options contains all the configuration used to open the Badger db
type Options struct {
	// Path is the directory path to the Badger db to use.
	Path string

	// BadgerOptions contains any specific Badger options you might
	// want to specify.
	BadgerOptions *badger.Options

	// NoSync causes the database to skip fsync calls after each
	// write to the log. This is unsafe, so it should be used
	// with caution.
	NoSync bool

	// ValueLogGC enables a periodic goroutine that does a garbage
	// collection of the value log while the underlying Badger is online.
	ValueLogGC bool

	// GCInterval is the interval between conditionally running the garbage
	// collection process, based on the size of the vlog. By default, runs every 1m.
	GCInterval time.Duration

	// GCInterval is the interval between mandatory running the garbage
	// collection process. By default, runs every 10m.
	MandatoryGCInterval time.Duration

	// GCThreshold sets threshold in bytes for the vlog size to be included in the
	// garbage collection cycle. By default, 1GB.
	GCThreshold int64
}

// NewBadgerStore takes a file path and returns a connected Raft backend.
func NewBadgerStore(path string) (*BadgerStore, error) {
	return New(Options{Path: path})
}

// New uses the supplied options to open the Badger db and prepare it for
// use as a raft backend.
func New(options Options) (*BadgerStore, error) {

	// build badger options
	if options.BadgerOptions == nil {
		defaultOpts := badger.DefaultOptions(options.Path)
		options.BadgerOptions = &defaultOpts
	}
	options.BadgerOptions.SyncWrites = !options.NoSync

	// Try to connect
	handle, err := badger.Open(*options.BadgerOptions)
	if err != nil {
		return nil, err
	}

	// Create the new store
	store := &BadgerStore{
		conn: handle,
		path: options.Path,
		mu:   &sync.Mutex{},
	}

	// Start GC routine
	if options.ValueLogGC {

		var gcInterval time.Duration
		var mandatoryGCInterval time.Duration
		var threshold int64

		if gcInterval = 1 * time.Minute; options.GCInterval != 0 {
			gcInterval = options.GCInterval
		}
		if mandatoryGCInterval = 10 * time.Minute; options.MandatoryGCInterval != 0 {
			mandatoryGCInterval = options.MandatoryGCInterval
		}
		if threshold = int64(1 << 30); options.GCThreshold != 0 {
			threshold = options.GCThreshold
		}

		store.vlogTicker = time.NewTicker(gcInterval)
		store.mandatoryVlogTicker = time.NewTicker(mandatoryGCInterval)
		go store.runVlogGC(handle, threshold)
	}

	return store, nil
}

func (b *BadgerStore) runVlogGC(db *badger.DB, threshold int64) {
	// Get initial size on start.
	_, lastVlogSize := db.Size()

	runGC := func() {
		var err error
		for err == nil {
			// If a GC is successful, immediately run it again.
			err = db.RunValueLogGC(0.7)
		}
		_, lastVlogSize = db.Size()
	}

	for {
		select {
		case <-b.vlogTicker.C:
			_, currentVlogSize := db.Size()
			if currentVlogSize < lastVlogSize+threshold {
				continue
			}
			runGC()
		case <-b.mandatoryVlogTicker.C:
			runGC()
		}
	}
}
