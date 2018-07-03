package node

import (
	"crypto/rsa"
	"encoding/binary"
	"sync"
	"github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/util"
	"errors"
	"github.com/rainer37/OnionCoin/records"
)

const (
	FWD              = '0'
	JOIN             = '1'
	FIND             = '2'
	COINREWARD       = '3'
	COINCOSIGN       = '4'
	JOINACK          = '6'
	WELCOME          = '7'
	RAWCOINEXCHANGE  = 'A'
	RAWCOINSIGNED    = 'B'
	REGCOSIGNREQUEST = 'C'
	REGCOSIGNREPLY   = 'D'
	COINFEEDBACK     = 'E'
	TXNRECEIVE       = 'G'
	IPLOOKUP         = 'L'
	IPLOOKUPRP       = 'M'
	TXNAGGRE		 = 'N'
	HASHCMP			 = 'O'
	PUBLISHINGBLOCK  = 'P'
	PUBLISHINGCHECK  = 'Q'
	CHAINSYNC        = 'S'
	CHAINSYNCACK     = 'T'
	CHAINREPAIR      = 'U'
)

var mutex = sync.Mutex{}
var bcCount = 0
/*
	check the OMsg bytes received, verify the sig by senderID,
	 and return [err, opCode, ID, pkEntry, payload]
 */
func (n *Node) syntaxCheck(incoming []byte) (error, rune, string, *records.PKEntry, []byte) {
	omsg, ok := n.unMarshalOMsg(incoming)

	if !ok {
		return errors.New("cannot Unmarshal Msg, discard it." +
			string(len(incoming))), ' ', "", nil, nil
	}

	// print("valid OMsg, continue...")

	senderID := omsg.GetSenderID()
	senderPK := n.getPubRoutingInfo(senderID)

	// check if the sender's id is known, otherwise cannot verify the signature.
	if senderPK == nil {
		return errors.New( "don't know who you are" + senderID), ' ', "", nil, nil
	}

	// verifying the identity of claimed sender by its pk and signature.
	if !verifySig(omsg, &senderPK.Pk) {
		rjmsg := "Cannot verify sig from msg, discard it."
		print(rjmsg)
	}

	// print("verified ID", senderID)

	payload := omsg.GetPayload()

	return nil, omsg.GetOPCode(), senderID, senderPK, payload
}

/*
	Requests regarding blockchain has format:
	000[OPCODE][SENDERI]
 */
func (n *Node) chainRequest(payload []byte) bool {
	return false
}

func (n *Node) chainDispatch(incoming []byte) {

}

/*
	Unmarshal the incoming packet to Omsg and verify the signature.
	Then dispatch the OMsg to its OpCode.
 */
func (n *Node) dispatch(incoming []byte) {

	ok := n.newbieJoin(incoming)

	// if it's a newbie request, it cannot be a OMsg.
	if ok { return }

	ok = n.chainRequest(incoming)

	if ok { n.chainDispatch(incoming) ; return }

	err, opCode, senderID, senderPK, payload := n.syntaxCheck(incoming)
	util.CheckErr(err)

	switch opCode{
	case FWD:
		// print("Forwarding something from", senderID)
		n.forwardProtocol(payload, senderID)
	case JOIN:
		print("Joining", senderID)
		n.joinProtocol(payload)
	case JOINACK:
		// print("JOIN ACK RECEIVED, JOIN SUCCEEDS")
		unmarshalRoutingInfo(payload)
	case WELCOME:
		// print("WELCOME received from", senderID)
		n.welcomeProtocol(payload)
	case FIND:
		print("Finding")
	case COINREWARD:
		//print("Receiving a Coin")
		aye := "N"
		if n.validcoinwrap(payload, senderID) { aye = "Y" }
		n.sendOMsgWithID(COINFEEDBACK, []byte(aye), senderID)
	case COINFEEDBACK:
		// print("Feedback Received", string(payload[0]))
		n.feedbackChan <- rune(payload[0])
	case RAWCOINEXCHANGE:
		//print("COINREWARD Exchange Requesting by", senderID)
		if !n.iamBank() {
			print("SRY IM NOT BANK", &senderPK)
			return
		}
		n.receiveRawCoin(payload, senderID)
	case RAWCOINSIGNED:
		//print("My Signed RawCoin received.", senderID)
		n.receiveNewCoin(payload, senderID)
	case COINCOSIGN:
		//print("Let's make fortune together", "Cosign size", len(payload))
		n.coSignValidCoin(payload)
	case REGCOSIGNREQUEST:
		//print("Helping Registering A New Node")
		n.regCoSignRequest(payload, senderID)
	case REGCOSIGNREPLY:
		//print("Receive Reg CoSign from", senderID)
		n.regChan <- payload
	case TXNRECEIVE:
		bcCount++
		print("A Txn Received from", senderID)
		txn := blockChain.ProduceTxn(payload[1:], rune(payload[0]))
		n.bankProxy.AddTxn(txn)
	case CHAINSYNC:
		bcCount++
		//print("Chain Sync Req Received from", senderID)
		n.chainSyncRequested(payload, senderID)
	case CHAINSYNCACK:
		bcCount++
		mutex.Lock()
		//print("Chain Sync Ack Received from", senderID)
		n.chainSyncAckReceived(payload, senderID)
		mutex.Unlock()
	case CHAINREPAIR:
		bcCount++
		print(senderID, "tries to help repairing my chain")
		n.chainRepairReceived(payload, senderID)
	case PUBLISHINGBLOCK:
		bcCount++
		print(senderID, "is trying to publish a block")
		depth := binary.BigEndian.Uint64(payload[:8])
		print("block depth:", depth, n.chain.Size())
	case PUBLISHINGCHECK:
		bcCount++
		print(senderID, "responded with publishing status")
	case TXNAGGRE:
		bcCount++
		print("Txn Aggregation received from", senderID)
		txns := blockChain.DemuxTxns(payload)
		n.bankProxy.AggreTxns(txns)
		// n.chain.AddNewBlock(blockChain.NewBlock(txns))
		// print(n.chain.GetLastBlock().CurHash)
	case HASHCMP:
		bcCount++
		// print("New Block Hash From", senderID, string(payload))
		if _, ok := HashCmpMap[string(payload)]; !ok{
			HashCmpMap[string(payload)] = 0
		}
		HashCmpMap[string(payload)] += 1
		// print(bank.HashCmpMap)
	case IPLOOKUP:
		print(senderID, "is looking for someone")
		n.handleLookup(payload)
	case IPLOOKUPRP:
		print("IP found")
	default:
		print("Unknown Msg, discard.")
	}
}

/*
	Retrieve the encrypted symmetric key, and decrypt it
	Decrypt the rest of incoming packet, and return it as OMsg
 */
func (n* Node) unMarshalOMsg(incoming []byte) (*records.OMsg, bool) {
	return records.UnmarshalOMsg(incoming, n.sk)
}

/*
	verified the signature with claimed senderID.
 */
func verifySig(oMsg *records.OMsg, pk *rsa.PublicKey) bool {
	return oMsg.VerifySig(pk)
}

/*
	Wrap payload into OMsg and encrypt it with target pk, and send it.
 */
func (n *Node) sendOMsg(opCode rune, payload []byte, pe *records.PKEntry) {
	p := records.MarshalOMsg(opCode, payload, n.ID, n.sk, pe.Pk)
	n.sendActive(p, pe.Port)
}

func (n *Node) sendOMsgWithID(opCode rune, payload []byte, id string) {
	pe := n.getPubRoutingInfo(id)
	if pe == nil {
		return
	}
	n.sendOMsg(opCode, payload, pe)
}