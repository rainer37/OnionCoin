package node

import (
	"net"
	"github.com/rainer37/OnionCoin/records"
)

const (
	REJ_STR = "REJECTED"
)

const (
	INV_MSG_FMT = "INVALID MSG FORMAT"
)

const (
	FWD = '0'
	JOIN = '1'
	FIND = '2'
	FREE = '3'
	COIN = '4'
	EXPT = '5'
	JOIN_ACK = '6'
)
/*
	Unmarshal the incoming packet to Omsg and verify the signature.
	Then dispatch the OMsg to its OpCode.
 */
func (n *Node) dispatch(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	omsg, ok := UnmarshalOMsg(incoming)

	if !ok || !omsg.VerifySig() {
		print("Terrible Msg, discard it.")
		return
	}

	switch omsg.GetOPCode() {
	case FWD:
		print("Forwarding")
		n.send([]byte("Fine i will take the coin though."), con, add)
	case JOIN:
		print("Joining "+string(incoming[4:]))
		n.joinProtocol(incoming, con, add)
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

func UnmarshalOMsg(msg []byte) (records.OnionMsg, bool) {
	return nil, true
}