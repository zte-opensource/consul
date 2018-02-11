package consul

import (
	"fmt"
	"sync"

	"github.com/hashicorp/consul/agent/metadata"
	"github.com/hashicorp/raft"
)

// ServerLookup encapsulates looking up servers by id and address
type ServerLookup struct {
	lock            sync.RWMutex
	addressToServer map[raft.ServerAddress]*metadata.Server
	idToServer      map[raft.ServerID]*metadata.Server
}

// called by
// agent/consul/server.go/NewServerLogger
func NewServerLookup() *ServerLookup {
    // raft can use this <id, addr>, <addr, id> map to get rpc conn
	return &ServerLookup{
		addressToServer: make(map[raft.ServerAddress]*metadata.Server),
		idToServer:      make(map[raft.ServerID]*metadata.Server),
	}
}

// called by
// agent/consul/server_serf.go/lanNodeJoin
func (sl *ServerLookup) AddServer(server *metadata.Server) {
	sl.lock.Lock()
	defer sl.lock.Unlock()
	sl.addressToServer[raft.ServerAddress(server.Addr.String())] = server
	sl.idToServer[raft.ServerID(server.ID)] = server
}

// called by
// agent/consul/server_serf.go/lanNodeFailed
func (sl *ServerLookup) RemoveServer(server *metadata.Server) {
	sl.lock.Lock()
	defer sl.lock.Unlock()
	delete(sl.addressToServer, raft.ServerAddress(server.Addr.String()))
	delete(sl.idToServer, raft.ServerID(server.ID))
}

// Implements the ServerAddressProvider interface
func (sl *ServerLookup) ServerAddr(id raft.ServerID) (raft.ServerAddress, error) {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	svr, ok := sl.idToServer[id]
	if !ok {
		return "", fmt.Errorf("Could not find address for server id %v", id)
	}
	return raft.ServerAddress(svr.Addr.String()), nil
}

// Server looks up the server by address, returns a boolean if not found
func (sl *ServerLookup) Server(addr raft.ServerAddress) *metadata.Server {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	svr, _ := sl.addressToServer[addr]
	return svr
}

func (sl *ServerLookup) Servers() []*metadata.Server {
	sl.lock.RLock()
	defer sl.lock.RUnlock()
	var ret []*metadata.Server
	for _, svr := range sl.addressToServer {
		ret = append(ret, svr)
	}
	return ret
}
