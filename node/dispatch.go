package node

import (
	"net"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"time"
)

const (
	PKRQLEN = 132
	REJSTR = "REJECTED"
	INVMSGFMT = "INVALID MSG FORMAT"
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

	// if the newbie is joining, special protocol is invoked.
	if string(incoming[:4]) == PKREQUEST {
		spk := ocrypto.DecodePK(incoming[4:4+PKRQLEN])
		senderAddr := string(incoming[4+PKRQLEN:])
		records.InsertEntry(FAKE_ID+senderAddr, spk, time.Now())
		// PKAK | EncodedPK | PortListening
		n.sendActive(PKRQACK+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port, senderAddr)
		return
	} else if string(incoming[:4]) == PKRQACK {
		// return the pk to the requesting node to finish the join protocol.
		print("thank you for the pub-key")
		n.pkChan <- incoming[4:4+PKRQLEN]
		return
	}

	omsg, ok := n.UnmarshalOMsg(incoming)

	if !ok {
		print("Cannot Unmarshal Msg, discard it.")
		return
	}

	print("valid OMsg, continue...")

	senderPK := records.KeyRepo[omsg.GetSenderID()] // TODO: check if there is no known pk.

	// verifying the identity of claimed sender by its pk and signature.
	if !n.VerifySig(omsg, &senderPK.Pk) {
		print("Cannot verify sig from msg, discard it.")
		return
	}

	print("verified ID", omsg.GetSenderID())

	payload := omsg.GetPayload()

	switch omsg.GetOPCode() {
	case FWD:
		print("Forwarding")
		n.forwardProtocol(payload, con, add)
	case JOIN:
		print("Joining")
		n.joinProtocol(payload, con, add)
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

func (n* Node) VerifySig(omsg *records.OMsg, pk *rsa.PublicKey) bool {
	return omsg.VerifySig(pk)
}