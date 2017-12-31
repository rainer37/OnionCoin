package node

import "net"

const (
	REJ_STR = "REJECTED"
)

const (
	INV_MSG_FMT = "INVALID MSG FORMAT"
)

const (
	FWD = "FWD "
	JOIN = "JOIN"
	FIND = "FIND"
	FREE = "FREE"
	COIN = "COIN"
	EXPT = "EXPT"
	JOIN_ACK = "JACK"
)

func (n *Node) dispatch(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	switch string(incoming[:4]) {
	case FWD:
		print("Forwarding")
		n.send([]byte("Fine i will take the coin though."), con, add)
	case JOIN:
		print("Joining "+string(incoming[4:]))
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

	case FIND:
		print("Finding")
	case FREE:
		//receive the free list
	case COIN:
		//receive the coin
	case EXPT:
		//any exception
	case JOIN_ACK:
		print("JOIN SUCCESS")
	default:
		print("Unknown Msg, discard.")
	}
}
