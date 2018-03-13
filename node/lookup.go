package node

import (
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/records"
	"strings"
)

func (n *Node) LookUpPK(address string) rsa.PublicKey {
	n.sendActive([]byte(REGISTER+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port), address)
	// waiting for the pk request return.
	enPk := <-n.pkChan
	return ocrypto.DecodePK(enPk)
}

/*
	send lookup request with the targetID and my address for reply.
 */
func (n *Node) LookUpIP(id string) {
	for _ ,v := range records.KeyRepo {
		targetID := make([]byte, 16)
		copy(targetID, id)
		p := n.prepareOMsg(IPLOOKUP, append(targetID , []byte(n.addr())...), v.Pk)
		go n.sendActive(p, v.Port)
	}
}

func (n *Node) handleLookup(payload []byte) {
	targetID, origAddr := strings.Trim(string(payload[:16]), "\x00"), payload[16:]
	print("looking for", targetID, "i am at", origAddr)
}