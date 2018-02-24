package node

import (
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/records"
)

/*
	1. Decrypt the Onion to get nextID, previous coin, and the innerOnion.
	2. forward the innerOnion to nextID.
	3. reply prevCoin to sender.
 */
func (n *Node) forwardProtocol(payload []byte, senderID string) {
	nextID, prevCoin, iOnion := ocrypto.PeelOnion(n.sk, payload)

	print(nextID, len(prevCoin), string(prevCoin), len(iOnion), string(iOnion))

	// the most innerOnion should have the same ID as receiver ID.
	if nextID == n.ID {
		print("destination reached")
		return
	}

	npe := records.GetKeyByID(nextID)

	if npe == nil {
		print("cannot verify next hop id")
		return
	}

	go func() {
		m := n.prepareOMsg(FWD,iOnion,npe.Pk)
		n.sendActive(m, npe.Port)
	}()

	spe := records.GetKeyByID(senderID)

	if spe == nil {
		print("this is impossible, the sender is not verified ?")
		return
	}

	pm := n.prepareOMsg(COIN,prevCoin,spe.Pk)
	n.sendActive(pm, spe.Port)
}