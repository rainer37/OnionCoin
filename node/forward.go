package node

import (
	"time"
)

var omsgCount = 0

/*
	1. Decrypt the Onion to get nextID, previous coin, and the innerOnion.
	2. forward the innerOnion to nextID.
	3. reply prevCoin to sender.
 */
func (n *Node) forwardProtocol(payload []byte, senderID string) {
	nextID, prevCoin, iOnion := PeelOnion(n.sk, payload)
	n.feedbackChan = make(chan rune)
	// print(nextID, len(prevCoin), string(prevCoin), len(iOnion), len(iOnion))

	spe := n.getPubRoutingInfo(senderID)

	if spe == nil {
		print("this is impossible, the sender is not verified ?")
		return
	}

	// reply the coin to previous peer.
	pm := n.prepareOMsg(COINREWARD, prevCoin, spe.Pk)
	go n.sendActive(pm, spe.Port)

	// the most innerOnion should have the same ID as receiver ID.
	if nextID == n.ID {
		//print("destination reached")
		omsgCount++
		print("   MSG RECEIVED: [", string(iOnion),"]")
		return
	}

	npe := n.getPubRoutingInfo(nextID)

	if npe == nil {
		print("cannot verify next hop id")
		return
	}

	select {
	case <-time.After(5 * time.Second):
		print("no positive feedback, i won't help")
	case feedback := <-n.feedbackChan:
		if feedback == 'Y' {
			m := n.prepareOMsg(FWD,iOnion,npe.Pk)
			n.sendActive(m, npe.Port)
		}
	}
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
		n.CoinExchange(ids[i])
		c := n.Vault.Withdraw(ids[i-1])
		// c := coin.NewCoin(ids[i-1], []byte(""), []string{})
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