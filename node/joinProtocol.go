package node

import "net"

func (n *Node) joinProtocol(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	ok, id, address := deSegmentJoinMsg(string(incoming[4:]))

	if !ok {
		print(INV_MSG_FMT)
		n.send([]byte(REJ_STR+" "+INV_MSG_FMT), con, add)
	}

	verified := n.verifyID(id)

	if !verified {
		print("Invalidate ID, be aware!")
		n.send([]byte(REJ_STR+" UNABLE TO VERIFY YOUR ID"), con, add)
	}

	n.insert(id, address) //TODO: alternatives on node discovery
	n.sendActive("JACK", address)
	print("Welcome to "+address)
}