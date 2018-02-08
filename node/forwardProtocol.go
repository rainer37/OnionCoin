package node

func (n *Node) forwardProtocol(payload []byte) {
	print("Forwarding")
	//n.send([]byte("FWD ACK"), con, add)
	print("Payload:", payload)
}

func deSegOnion(onion []byte) string {
	return ""
}