package node

import (
	"github.com/rainer37/OnionCoin/util"
)

/*
	send lookup request with the targetID and my address for reply.
 */
func (n *Node) LookUpIP(id string) {
	//for _ ,v := range records.keyRepo {
	//	targetID := make([]byte, util.IDLEN)
	//	copy(targetID, id)
	//	p := n.prepareOMsg(IPLOOKUP, append(targetID ,
	// []byte(n.addr())...), v.Pk)
	//	go n.sendActive(p, v.Port)
	//}
}

func (n *Node) handleLookup(payload []byte) {
	targetID, origAddr := util.Strip(payload[:16]), payload[16:]
	print("looking for", targetID, "i am at", origAddr)
}