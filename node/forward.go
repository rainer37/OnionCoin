package node

import (
	"time"
)

var omsgCount = 0
const FWDTIMEOUT = 3 * 1000

/*
	1. Decrypt the Onion to get nextID, previous coin, and the innerOnion.
	2. forward the innerOnion to nextID.
	3. reply prevCoin to sender.
 */
func (n *Node) forwardProtocol(payload []byte, senderID string) {
	nextID, prevCoin, iOnion := PeelOnion(n.sk, payload)
	n.feedbackChan = make(chan rune)
	defer close(n.feedbackChan)

	// print(nextID, len(prevCoin), string(prevCoin), len(iOnion), len(iOnion))

	// reply the coin to previous peer.
	n.sendOMsgWithID(COINREWARD, prevCoin, senderID)

	// the most innerOnion should have the same ID as receiver ID.
	if nextID == n.ID {
		omsgCount++
		print("   MSG RECEIVED: [", string(iOnion),"] FROM", senderID)
		return
	}

	select {
	case <-time.After(FWDTIMEOUT):
		print("   Time out, no positive feedback, i won't help")
	case feedback := <-n.feedbackChan:
		if feedback == 'Y' {
			n.sendOMsgWithID(FWD, iOnion, nextID)
		} else {
			print("   no positive feedback, i won't help")
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
		//n.CoinExchange(ids[i])
		c := n.Vault.Withdraw(ids[i-1])
		nextID := ids[i+1]
		o = WrapOnion(pe.Pk, nextID, c.Bytes(), o)
	}

	return o
}

/*
	send a onion message toward the path defined by ids.
 */
func (n *Node) SendOninoMsg(ids []string, msg string) {
	print("SENDING ALONG:", ids)
	onion := n.WrapABigOnion([]byte(msg), ids)
	n.sendOMsgWithID(FWD, onion, ids[0])
}