package structs

import "github.com/hashicorp/consul/agent/consul/state/sqlite"

// QueryRequest represents a query that returns rows, and does not modify
// the database.
type SQLQueryRequest struct {
	Queries []string
	Timings bool
	Atomic  bool
	Lvl     ConsistencyLevel

	// WriteRequest is a common struct containing ACL tokens and other
	// write-related common elements for requests.
	WriteRequest
}

// QueryResponse encapsulates a response to a query.
type SQLQueryResponse struct {
	Rows []*sqlite.Rows
	Time float64
	Err error
}

// ExecuteRequest represents a query that returns now rows, but does modify
// the database.
type SQLExecuteRequest struct {
	Queries []string
	Timings bool
	Atomic  bool

	// WriteRequest is a common struct containing ACL tokens and other
	// write-related common elements for requests.
	WriteRequest
}

// ExecuteResponse encapsulates a response to an execute.
type SQLExecuteResponse struct {
	Results []*sqlite.Result
	Time    float64
	Err error
}

// ConsistencyLevel represents the available read consistency levels.
type ConsistencyLevel int

// Represents the available consistency levels.
const (
	None ConsistencyLevel = iota
	Weak
	Strong
)
