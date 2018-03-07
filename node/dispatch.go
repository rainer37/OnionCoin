package node

import (
	"github.com/rainer37/OnionCoin/records"
	"crypto/rsa"
	"encoding/binary"
	"crypto/sha256"
	"github.com/rainer37/OnionCoin/blockChain"
)

const (
	FWD = '0'
	JOIN = '1'
	FIND = '2'
	COIN = '3'
	COSIGN = '4'
	EXPT = '5'
	JOINACK = '6'
	WELCOME = '7'
	RAWCOINEXCHANGE = 'A'
	RAWCOINSIGNED = 'B'
	REGCOSIGNREQUEST = 'C'
	REGCOSIGNREPLY = 'D'
	TXNRECEIVE = 'G'
	REJECT = 'F'
	CHAINSYNC = 'S'
	CHAINSYNCACK = 'T'
	RETURN = 'R'
	ADV = 'G'
)
/*
	Unmarshal the incoming packet to Omsg and verify the signature.
	Then dispatch the OMsg to its OpCode.
 */
func (n *Node) dispatch(incoming []byte) {

	ok := n.newbieJoin(incoming)

	// if it's a newbie request, it cannot be a OMsg.
	if ok { return }

	omsg, ok := n.UnmarshalOMsg(incoming)

	if !ok {
		print("Cannot Unmarshal Msg, discard it.", len(incoming))
		return
	}

	print("valid OMsg, continue...")

	senderID := omsg.GetSenderID()
	senderPK := records.KeyRepo[senderID]

	// check if the sender's id is known, otherwise cannot verify the signature.
	if senderPK == nil {
		rjmsg := "Don't know who you are"
		print(rjmsg, senderID)
		return
	}

	// verifying the identity of claimed sender by its pk and signature.
	if !verifySig(omsg, &senderPK.Pk) {
		rjmsg := "Cannot verify sig from msg, discard it."
		print(rjmsg)
		n.sendReject(rjmsg, senderPK)
		return
	}

	print("verified ID", senderID)

	payload := omsg.GetPayload()

	switch omsg.GetOPCode() {
	case FWD:
		print("Forwarding")
		n.forwardProtocol(payload, senderID)
	case JOIN:
		print("Joining")
		ok := n.joinProtocol(payload)
		if ok {
			n.welcomeNewBie(senderID)
		}
	case FIND:
		print("Finding")
	case COIN:
		print("Receiving a Coin")
	case JOINACK:
		print("JOIN ACK RECEIVED, JOIN SUCCEEDS")
		unmarshalRoutingInfo(payload)
		//n.foo()
	case WELCOME:
		print("WELCOME received from", senderID)
		welcomeProtocol(payload)
	case RAWCOINEXCHANGE:
		print("COIN Exchange Requesting by", senderID)
		if !n.iamBank() {
			n.sendReject("SRY IM NOT BANK", senderPK)
			return
		}
		n.receiveRawCoin(payload, senderID)
	case RAWCOINSIGNED:
		print("My Signed RawCoin received.", senderID)
		n.receiveNewCoin(payload, senderID)
	case COSIGN:
		print("Let's make fortune together")
		print("Cosign size", len(payload))
		counter := binary.BigEndian.Uint16(payload[:2]) // get cosign counter first 2 bytes
		n.coSignValidCoin(payload[2:], counter)
	case REGCOSIGNREQUEST:
		print("Helping Registering A New Node")
		pkHash := sha256.Sum256(payload[:132])
		mySig := n.blindSign(append(pkHash[:], payload[132:]...))
		spk := n.getPubRoutingInfo(senderID)
		p := n.prepareOMsg(REGCOSIGNREPLY, mySig, spk.Pk)
		n.sendActive(p, spk.Port)
	case REGCOSIGNREPLY:
		print("Receive Reg CoSign from", senderID)
		n.regChan <- payload
	case TXNRECEIVE:
		print("A Txn Received from", senderID)
		txn := blockChain.ProduceTxn(payload, blockChain.PK)
		n.bankProxy.AddTxn(txn)
	case CHAINSYNC:
		print("BlockChain Sync Req Received", senderID)
		blockIndex := int64(binary.BigEndian.Uint64(payload))
		for i := blockIndex; i < n.chain.Size(); i++ {
			spk := n.getPubRoutingInfo(senderID)
			blocks := n.chain.GenBlockBytes(i)
			print(len(blocks), "sent")
			p := n.prepareOMsg(CHAINSYNCACK, blocks, spk.Pk)
			n.sendActive(p, spk.Port)
		}
	case CHAINSYNCACK:
		print("BlockChain Sync Ack Received", senderID)
		b := blockChain.DeMuxBlock(payload)
		n.chain.AddOldBlock(b)
	case REJECT:
		print(string(payload))
	case EXPT:
		//any exception
	default:
		print("Unknown Msg, discard.")
	}
}

/*
	format a reject OMsg with msg included.
 */
func (n *Node) formalRejectPacket(msg string, pk rsa.PublicKey) []byte {
	return n.prepareOMsg(REJECT,[]byte(msg),pk)
}

/*
	reject some one when exceptional cases came up.
 */
func (n *Node) sendReject(msg string, senderPK *records.PKEntry) {
	rej := n.formalRejectPacket(msg, senderPK.Pk)
	n.sendActive(rej, senderPK.Port)
}

/*
	Wrap payload into OMsg and encrypt it with target pk.
 */
func (n *Node) prepareOMsg(opcode rune, payload []byte, pk rsa.PublicKey) []byte {
	return records.MarshalOMsg(opcode,payload, n.ID, n.sk, pk)
}

/*
	Retrieve the encrypted symmetric key, and decrypt it
	Decrypt the rest of incoming packet, and return it as OMsg
 */
func (n* Node) UnmarshalOMsg(incoming []byte) (*records.OMsg, bool) {
	return records.UnmarshalOMsg(incoming, n.sk)
}

/*
	verified the signature with claimed senderID.
 */
func verifySig(omsg *records.OMsg, pk *rsa.PublicKey) bool {
	return omsg.VerifySig(pk)
}


/*
	a warm welcome to newbie.
 */
func welcomeProtocol(payload []byte) {
	idLen := binary.BigEndian.Uint32(payload[:4])
	id := string(payload[4:4+idLen])
	print(id, len(records.KeyRepo))

	eLen := binary.BigEndian.Uint32(payload[4+idLen:8+idLen])
	e := records.BytesToPKEntry(payload[8+idLen:8+idLen+eLen])
	records.InsertEntry(id, e.Pk, e.Time, e.IP, e.Port)
}

