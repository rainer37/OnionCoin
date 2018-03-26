package node

import "github.com/rainer37/OnionCoin/coin"

var omsgCount = 0

/*
	1. Decrypt the Onion to get nextID, previous coin, and the innerOnion.
	2. forward the innerOnion to nextID.
	3. reply prevCoin to sender.
 */
func (n *Node) forwardProtocol(payload []byte, senderID string) {
	nextID, prevCoin, iOnion := PeelOnion(n.sk, payload)

	// print(nextID, len(prevCoin), string(prevCoin), len(iOnion), len(iOnion))

	// the most innerOnion should have the same ID as receiver ID.
	if nextID == n.ID {
		//print("destination reached")
		omsgCount++
		print("   MSG RECEIVED: ", string(iOnion))
		return
	}

	npe := n.getPubRoutingInfo(nextID)

	if npe == nil {
		print("cannot verify next hop id")
		return
	}

	go func() {
		m := n.prepareOMsg(FWD,iOnion,npe.Pk)
		n.sendActive(m, npe.Port)
	}()

	spe := n.getPubRoutingInfo(senderID)

	if spe == nil {
		print("this is impossible, the sender is not verified ?")
		return
	}

	// reply the coin to previous peer.
	pm := n.prepareOMsg(COINREWARD, prevCoin, spe.Pk)
	n.sendActive(pm, spe.Port)
}

/*
	Given a list of ids of nodes on the path, create a onion wrapping the message to send.
	ids : [s -> n0 -> n1 -> ... -> r]
*/
func (n *Node) WrapABigOnion(msg []byte, ids []string) []byte {
	ids = append(ids, ids[len(ids)-1])
	ids = append([]string{ids[0]}, ids...)

	o := msg

	for i:=len(ids)-2; i > 0; i-- {
		pe := n.getPubRoutingInfo(ids[i])
		// c := n.Vault.Withdraw(ids[i-1])
		c := coin.NewCoin(ids[i-1], []byte(""))
		nextID := ids[i+1]
		o = WrapOnion(pe.Pk, nextID, c.Bytes(), o)
	}

	return o
}

/*
	send a onion message toward the path defined by ids.
 */
func (n *Node) SendOninoMsg(ids []string, msg string) {
	m := n.WrapABigOnion([]byte(msg), ids)
	npe := n.getPubRoutingInfo(ids[0])
	m = n.prepareOMsg(FWD, m, npe.Pk)
	n.sendActive(m, npe.Port)
}