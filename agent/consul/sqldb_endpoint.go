package consul

import (
	"github.com/hashicorp/raft"
	"github.com/hashicorp/consul/agent/structs"
)

// SQL endpoint is used to manipulate the sql db store
type SQL struct {
	srv *Server
}

// Execute executes queries that return no rows, but do modify the database.
// Changes made to the database through this call are applied via the Raft
// consensus system. The Store must have been opened first. Must be called
// on the leader or an error will we returned. The changes are made using
// the database connection built-in to the Store.
func (s *SQL) Execute(args *structs.SQLExecuteRequest, reply *structs.SQLExecuteResponse) error {
	return s.execute(args, reply)
}

// ExecuteOrAbort executes the requests, but aborts any active transaction
// on the underlying database in the case of any error. Any changes are made
// using the database connection built-in to the Store.
func (s *SQL) ExecuteOrAbort(args *structs.SQLExecuteRequest, reply *structs.SQLExecuteResponse) error {
	return s.executeOrAbort(args, reply)
}

// Query executes queries that return rows, and do not modify the database.
// The queries are made using the database connection built-in to the Store.
// Depending on the read consistency requested, it may or may not need to be
// called on the leader.
func (s *SQL) Query(args *structs.SQLQueryRequest, reply *structs.SQLQueryResponse) error {
	return s.query(args, reply)
}

// Execute executes queries that return no rows, but do modify the database. If connection
// is nil then the utility connection is used.
func (s *SQL) execute(args *structs.SQLExecuteRequest, reply *structs.SQLExecuteResponse) error {
	// TODO: forward
	f := s.srv.raftApply(structs.SQLExecuteRequestType, args)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}

	switch r := f.Response().(type) {
	case *structs.SQLExecuteResponse:
		*reply = *r
		// TODO
		return r.Err
	case error:
		return r
	default:
		panic("unsupported type")
	}
}

// Query executes queries that return rows, and do not modify the database. If
// connection is nil, then the utility connection is used.
func (s *SQL) query(args *structs.SQLQueryRequest, reply *structs.SQLQueryResponse) error {
	// TODO: forward
	f := s.srv.raftApply(structs.SQLQueryRequestType, args)
	if e := f.(raft.Future); e.Error() != nil {
		if e.Error() == raft.ErrNotLeader {
			return ErrNotLeader
		}
		return e.Error()
	}

	switch r := f.Response().(type) {
	case *structs.SQLQueryResponse:
		*reply = *r
		return r.Err
	case error:
		return r
	default:
		panic("unsupported type")
	}
}
