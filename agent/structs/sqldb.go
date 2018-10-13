package structs

// Result represents the outcome of an operation that changes rows.
type SQLResult struct {
	LastInsertID int64   `json:"last_insert_id,omitempty"`
	RowsAffected int64   `json:"rows_affected,omitempty"`
	Error        string  `json:"error,omitempty"`
	Time         float64 `json:"time,omitempty"`
}

// Rows represents the outcome of an operation that returns query data.
type SQLRows struct {
	Columns []string        `json:"columns,omitempty"`
	Types   []string        `json:"types,omitempty"`
	Values  [][]interface{} `json:"values,omitempty"`
	Error   string          `json:"error,omitempty"`
	Time    float64         `json:"time,omitempty"`
}

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
	Rows []*SQLRows
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
	Results []*SQLResult
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
