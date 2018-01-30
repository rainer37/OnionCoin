package node

import (
	"github.com/rainer37/OnionCoin/records"
	"net"
)

func (n *Node) forwardProtocol(omsg *records.OMsg, con *net.UDPConn, add *net.UDPAddr) {
	print("Forwarding")
	n.send([]byte("FWD ACK"), con, add)
	payload := omsg.GetPayload(omsg.GetLenPayload())
	print("Payload:", payload)
}

func deSegOnion(onion []byte) string {
	return ""
}