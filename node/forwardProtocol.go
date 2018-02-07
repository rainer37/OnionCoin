package node

import (
	"net"
)

func (n *Node) forwardProtocol(payload []byte, con *net.UDPConn, add *net.UDPAddr) {
	print("Forwarding")
	n.send([]byte("FWD ACK"), con, add)
	print("Payload:", payload)
}

func deSegOnion(onion []byte) string {
	return ""
}