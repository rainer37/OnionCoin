package node

import (
	"sync"
)

type RoutingTable struct {
	table map[string]string
	mutex sync.Mutex
}

func (rt *RoutingTable) InitRT() {
	rt.table = make(map[string]string)
}

func (rt *RoutingTable) insert(id string, address string) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	rt.table[id] = address
}
func (rt *RoutingTable) remove(id string) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	delete(rt.table, id)
}
func (rt *RoutingTable) get(id string) string {
	return rt.table[id]
}