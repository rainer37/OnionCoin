package node

import "net"

func (n *Node) joinProtocol(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	ok, id, address := deSegmentJoinMsg(string(incoming[4:]))

	if !ok {
		print(INVMSGFMT)
		n.send([]byte(REJSTR+" "+INVMSGFMT), con, add)
	}

	verified := n.verifyID(id)

	if !verified {
		print("Invalidate ID, be aware!")
		n.send([]byte(REJSTR+" UNABLE TO VERIFY YOUR ID"), con, add)
	}

	n.insert(id, address) //TODO: alternatives on node discovery
	n.sendActive("JACK", address)
	print("Welcome to "+address)
}