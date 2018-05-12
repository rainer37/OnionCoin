package node

import (
	"strings"
)

/*
	send lookup request with the targetID and my address for reply.
 */
func (n *Node) LookUpIP(id string) {
	//for _ ,v := range records.keyRepo {
	//	targetID := make([]byte, util.IDLEN)
	//	copy(targetID, id)
	//	p := n.prepareOMsg(IPLOOKUP, append(targetID , []byte(n.addr())...), v.Pk)
	//	go n.sendActive(p, v.Port)
	//}
}

func (n *Node) handleLookup(payload []byte) {
	targetID, origAddr := strings.Trim(string(payload[:16]), "\x00"), payload[16:]
	print("looking for", targetID, "i am at", origAddr)
}