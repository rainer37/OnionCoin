package node

import (
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"time"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/coin"
)

const (
	PKRQLEN = 132
	FWD = '0'
	JOIN = '1'
	FIND = '2'
	FREE = '3'
	COIN = '4'
	EXPT = '5'
	JOINACK = '6'
	WELCOME = '7'
	COINEXCHANGE = 'A'
	REJECT = 'F'
)
/*
	Unmarshal the incoming packet to Omsg and verify the signature.
	Then dispatch the OMsg to its OpCode.
 */
func (n *Node) dispatch(incoming []byte) {

	// if the newbie is joining, special protocol is invoked.
	if string(incoming[:4]) == PKREQUEST {
		spk := ocrypto.DecodePK(incoming[4:4+PKRQLEN])
		senderAddr := string(incoming[4+PKRQLEN:])
		records.InsertEntry(FAKE_ID+senderAddr, spk, time.Now().Unix(), LOCALHOST, senderAddr)
		// PKAK | EncodedPK | PortListening
		n.sendActive([]byte(PKRQACK+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port), senderAddr)
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

	// check if the sender's id is known, otherwise cannot verify the signature.
	if senderPK == nil {
		rjmsg := "Don't know who you are"
		print(rjmsg, omsg.GetSenderID())
		return
	}

	// verifying the identity of claimed sender by its pk and signature.
	if !n.VerifySig(omsg, &senderPK.Pk) {
		rjmsg := "Cannot verify sig from msg, discard it."
		print(rjmsg)
		n.sendReject(rjmsg, senderPK)
		return
	}

	print("verified ID", omsg.GetSenderID())

	payload := omsg.GetPayload()

	switch omsg.GetOPCode() {
	case FWD:
		print("Forwarding")
		n.forwardProtocol(payload)
	case JOIN:
		print("Joining")
		ok := n.joinProtocol(payload)
		print(records.KeyRepo)
		if ok {
			n.welcomeNewBie(omsg.GetSenderID())
		}
	case FIND:
		print("Finding")
	case FREE:
		//receive the free list
	case COIN:
		//receive the coin
	case EXPT:
		//any exception
	case JOINACK:
		print("JOIN ACK RECEIVED, JOIN SUCCEEDS")
		unmarshalRoutingInfo(payload)
		//n.foo()
	case WELCOME:
		print("WELCOME received from", omsg.GetSenderID())
		idLen := binary.BigEndian.Uint32(payload[:4])
		id := string(payload[4:4+idLen])
		print(id, idLen)

		eLen := binary.BigEndian.Uint32(payload[4+idLen:8+idLen])
		e := records.BytesToPKEntry(payload[8+idLen:8+idLen+eLen])
		records.InsertEntry(id, e.Pk, e.Time, e.IP, e.Port)
		print(records.KeyRepo)
	case COINEXCHANGE:
		if !n.iamBank() {
			//rej := n.formalRejectPacket("SRY IM NOT BANK", senderPK.Pk)
			//n.sendActive(rej, senderPK.Port)
			n.sendReject("SRY IM NOT BANK", senderPK)
			return
		}

		print("Make a wish")
	case REJECT:
		print(string(omsg.GetPayload()))
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

/*
	return len(b):b in bytes
 */
func makeBytesLen(b []byte) []byte {
	lb := make([]byte, 4)
	binary.BigEndian.PutUint32(lb,uint32(len(b)))
	return append(lb, b...)
}

/*
	decode received routing info, and update routing table
 */
func unmarshalRoutingInfo(b []byte) {
	cur := 0
	for cur < len(b) {
		idLen := binary.BigEndian.Uint32(b[cur:cur+4])
		id := string(b[cur+4:cur+4+int(idLen)])
		print(id, idLen)
		cur += int(idLen) + 4

		eLen := binary.BigEndian.Uint32(b[cur:cur+4])
		e := records.BytesToPKEntry(b[cur+4:cur+4+int(eLen)])
		cur += int(eLen) + 4

		records.InsertEntry(id, e.Pk, e.Time, e.IP, e.Port)
	}
}

func (n *Node) foo() {
	if n.Port != "1339" {
		return
	}
	pk := records.GetKeyByID("FAKEID1338")
	pk2 := records.GetKeyByID("FAKEID1340")

	payload := ocrypto.WrapOnion(pk2.Pk, "myHome", new(coin.Coin).Bytes(), []byte("msg received"))
	p2 := ocrypto.WrapOnion(pk.Pk, "FAKEID1340", new(coin.Coin).Bytes(), payload)
	m := records.MarshalOMsg(FWD,p2,n.ID,n.sk,pk.Pk)
	n.sendActive(m,"1338")
}

func (n *Node) formalRejectPacket(msg string, pk rsa.PublicKey) []byte {
	return records.MarshalOMsg(REJECT,[]byte(msg),n.ID,n.sk,pk)
}

func (n *Node) sendReject(msg string, senderPK *records.PKEntry) {
	rej := n.formalRejectPacket(msg, senderPK.Pk)
	n.sendActive(rej, senderPK.Port)
}