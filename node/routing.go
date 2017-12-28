package node

import "fmt"

type RoutingTable struct {
	table map[string]string
}

func (rt *RoutingTable) InitRT() {
	fmt.Println(rt)
	rt.table = make(map[string]string)
}

func (rt *RoutingTable) insert(id string, address string) {
	rt.table[id] = address
}
func (rt *RoutingTable) remove(id string) {
	delete(rt.table, id)
}
func (rt *RoutingTable) get(id string) string {
	return rt.table[id]
}