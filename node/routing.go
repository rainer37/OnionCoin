package node

import (
	"sync"
	"time"
)

type RTEntry interface {
	IDtoIP() RTLessEntry
}

type RTLessEntry string

type RTFullEntry struct {
	addr RTLessEntry
	lastAlive time.Time
}

type RoutingTable struct {
	table map[string]RTEntry
	mutex sync.Mutex
}

func (id RTLessEntry) String() string {
	return string(id)
}

func (id RTLessEntry) IDtoIP() RTLessEntry {
	return id
}

func (entry RTFullEntry) IDtoIP() RTLessEntry {
	return entry.addr
}

func (rt *RoutingTable) InitRT() {
	rt.table = make(map[string]RTEntry)
}

func (rt *RoutingTable) insert(id string, address string) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	rt.table[id] = RTLessEntry(address)
}

func (rt *RoutingTable) remove(id string) {
	rt.mutex.Lock()
	defer rt.mutex.Unlock()
	delete(rt.table, id)
}

func (rt *RoutingTable) get(id string) RTEntry {
	return rt.table[id]
}