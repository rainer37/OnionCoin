package node

import (
	"net"
	"github.com/rainer37/OnionCoin/records"
)

const (
	REJSTR = "REJECTED"
)

const (
	INVMSGFMT = "INVALID MSG FORMAT"
)

const (
	FWD = '0'
	JOIN = '1'
	FIND = '2'
	FREE = '3'
	COIN = '4'
	EXPT = '5'
	JOINACK = '6'
)
/*
	Unmarshal the incoming packet to Omsg and verify the signature.
	Then dispatch the OMsg to its OpCode.
 */
func (n *Node) dispatch(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	omsg, ok := n.UnmarshalOMsg(incoming)

	if !ok || !n.VerifySig(omsg) {
		print("Terrible Msg, discard it.")
		return
	}

	switch omsg.GetOPCode() {
	case FWD:
		n.forwardProtocol(omsg,con, add)
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
	case JOINACK:
		print("JOIN SUCCESS")
	default:
		print("Unknown Msg, discard.")
	}
}

/*
	Retrieve the encrypted symmetric key, and decrypt it
	Decrypt the rest of incoming packet, and return it as OMsg
 */
func (n* Node) UnmarshalOMsg(incoming []byte) (*records.OMsg, bool) {
	return records.UnmarshalOMsg(incoming, n.sk)
}

func (n* Node) VerifySig(omsg *records.OMsg) bool {
	return omsg.VerifySig(&n.sk.PublicKey)
}