package state

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/hashicorp/consul/agent/consul/state/sqlite"
)

const (
	sqliteFile           = "db.sqlite"
)

// SQLDB is a SQLite database, where all changes are made via Raft consensus.
type SQLDB struct {
	dbPath string    // Path to underlying SQLite file, if not in-memory.
	dsn    string // Any custom DSN
	memory bool   // Whether the database is in-memory only.

	db      *sqlite.DB                // The underlying SQLite database.
	dbConn  *sqlite.Conn              // Hidden connection to underlying SQLite database.

	wg   sync.WaitGroup
	done chan struct{}

	closedMu sync.Mutex
	closed   bool // Has the store been closed?

	restoreMu sync.RWMutex // Restore needs exclusive access to database.

	logger *log.Logger
}

// NewStore returns a new Store.
func NewSQLDB(dir string, dsn string, memory bool, logger *log.Logger) (*SQLDB, error) {
	if logger == nil {
		logger = log.New(os.Stderr, "[store] ", log.LstdFlags)
	}

	sqldb := &SQLDB{
		dbPath:         filepath.Join(dir, sqliteFile),
		done:           make(chan struct{}, 1),
		logger:         logger,
	}
	err := sqldb.Open()
	if err != nil {
		return nil, err
	}
	return sqldb, nil
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
func (s *SQLDB) Open() error {
	s.closedMu.Lock()
	defer s.closedMu.Unlock()
	if s.closed {
		return ErrStoreInvalidState
	}

	// Create underlying in-memory or file-based database.
	var db *sqlite.DB
	var err error
	if !s.memory {
		// as it will be rebuilt from (possibly) a snapshot and committed log entries.
		if err := os.Remove(s.dbPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		db, err = sqlite.NewDB(s.dbPath, s.dsn, false)
		if err != nil {
			return err
		}
		s.logger.Println("SQLite database opened at", s.dbPath)
	} else {
		db, err = sqlite.NewDB(s.dbPath, s.dsn, true)
		if err != nil {
			return err
		}
		s.logger.Println("SQLite in-memory database opened")
	}
	s.db = db

	// Get utility connection to database.
	conn, err := s.db.Connect()
	if err != nil {
		return err
	}
	s.dbConn = conn

	return nil
}

// Close closes the store. If wait is true, waits for a graceful shutdown.
// Once closed, a Store may not be re-opened.
func (s *SQLDB) Close(wait bool) error {
	s.closedMu.Lock()
	defer s.closedMu.Unlock()
	if s.closed {
		return nil
	}
	defer func() {
		s.closed = true
	}()

	close(s.done)
	s.wg.Wait()

	if err := s.dbConn.Close(); err != nil {
		return err
	}
	s.dbConn = nil
	s.db = nil

	return nil
}

// Execute applies a Raft log entry to the database.
func (s *SQLDB) Execute(queries []string, atomic bool) interface{} {
	s.restoreMu.RLock()
	defer s.restoreMu.RUnlock()

	r, err := s.dbConn.Execute(queries, atomic)
	return &structs.SQLExecuteResponse{results: r, error: err}
}

// Query applies a Raft log entry to the database.
func (s *SQLDB) Query(queries []string, atomic bool) interface{} {
	s.restoreMu.RLock()
	defer s.restoreMu.RUnlock()

	r, err := s.dbConn.Query(queries, atomic)
	return &structs.SQLQueryResponse{rows: r, error: err}
}

// Restore restores the node to a previous state.
func (s *SQLDB) Restore(src []byte) error {
	s.restoreMu.Lock()
	defer s.restoreMu.Unlock()

	f, err := ioutil.TempFile("", "rqlilte-snap-")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	if _, err := f.Write(src); err != nil {
		return err
	}

	// Create new database from file, connect, and load
	// existing database from that.
	db, err := sqlite.NewDB(f.Name(), "", false)
	if err != nil {
		return err
	}
	conn, err := db.Connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := s.dbConn.Load(conn); err != nil {
		return err
	}

	stats.Add(numRestores, 1)
	return nil
}

// Database copies contents of the underlying SQLite file to dst
func (s *SQLDB) Backup(dst *bytes.Buffer) error {
	f, err := ioutil.TempFile("", "rqlilte-snap-")
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	os.Remove(f.Name())

	db, err := sqlite.NewDB(f.Name(), "", false)
	if err != nil {
		return err
	}
	conn, err := db.Connect()
	if err != nil {
		return err
	}

	if err := s.dbConn.Backup(conn); err != nil {
		return err
	}
	if err := conn.Close(); err != nil {
		return err
	}

	of, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	defer of.Close()

	_, err = io.Copy(dst, of)
	return err
}

func (s *Snapshot) SQLDB() *SQLDB {
	return s.store.sqldb
}

func (r *Restore) SQLDB() *SQLDB {
	return r.store.sqldb
}
