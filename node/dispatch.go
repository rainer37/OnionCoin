package node

import (
	"github.com/rainer37/OnionCoin/records"
	"crypto/rsa"
	"encoding/binary"
	"errors"
	"sync"
	"github.com/rainer37/OnionCoin/blockChain"
)

const (
	FWD              = '0'
	JOIN             = '1'
	FIND             = '2'
	COINREWARD       = '3'
	COINCOSIGN       = '4'
	EXPT             = '5'
	JOINACK          = '6'
	WELCOME          = '7'
	RAWCOINEXCHANGE  = 'A'
	RAWCOINSIGNED    = 'B'
	REGCOSIGNREQUEST = 'C'
	REGCOSIGNREPLY   = 'D'
	COINFEEDBACK     = 'E'
	TXNRECEIVE       = 'G'
	REJECT           = 'F'
	CHAINSYNC        = 'S'
	CHAINSYNCACK     = 'T'
	CHAINREPAIR      = 'U'
	CHAINREPAIRREPLY = 'V'
	PUBLISHINGBLOCK  = 'P'
	PUBLISHINGCHECK  = 'Q'
	IPLOOKUP         = 'L'
	IPLOOKUPRP       = 'M'
	TXNAGGRE		 = 'N'
	ADV              = 'G'
)

var mutex = sync.Mutex{}

/*
	check the OMsg bytes received, verify the sig by senderID, and return [err, opCode, ID, pkEntry, payload]
 */
func (n *Node) syntaxCheck(incoming []byte) (error, rune, string, *records.PKEntry, []byte) {
	omsg, ok := n.UnmarshalOMsg(incoming)

	if !ok {
		return errors.New("cannot Unmarshal Msg, discard it." + string(len(incoming))), ' ', "", nil, nil
	}

	// print("valid OMsg, continue...")

	senderID := omsg.GetSenderID()
	senderPK := records.KeyRepo[senderID]

	// check if the sender's id is known, otherwise cannot verify the signature.
	if senderPK == nil {
		return errors.New( "don't know who you are" + senderID), ' ', "", nil, nil
	}

	// verifying the identity of claimed sender by its pk and signature.
	if !verifySig(omsg, &senderPK.Pk) {
		rjmsg := "Cannot verify sig from msg, discard it."
		print(rjmsg)
		n.sendReject(rjmsg, senderPK)
		return errors.New( "cannot verify sig from msg, discard it"), ' ', "", nil, nil
	}

	// print("verified ID", senderID)

	payload := omsg.GetPayload()

	return nil, omsg.GetOPCode(), senderID, senderPK, payload
}

/*
	Unmarshal the incoming packet to Omsg and verify the signature.
	Then dispatch the OMsg to its OpCode.
 */
func (n *Node) dispatch(incoming []byte) {

	ok := n.newbieJoin(incoming)

	// if it's a newbie request, it cannot be a OMsg.
	if ok { return }

	err, opCode, senderID, senderPK, payload := n.syntaxCheck(incoming)
	checkErr(err)

	switch opCode{
	case FWD:
		print("Forwarding something from", senderID)
		n.forwardProtocol(payload, senderID)
	case JOIN:
		print("Joining", senderID)
		ok := n.joinProtocol(payload)
		if ok {
			n.welcomeNewBie(senderID)
		}
	case FIND:
		print("Finding")
	case COINREWARD:
		print("Receiving a Coin")
		spe := n.getPubRoutingInfo(senderID)
		aye := "N"
		if n.ValidateCoin(payload, senderID) {
			aye = "Y"
		}
		p := n.prepareOMsg(COINFEEDBACK, []byte(aye), spe.Pk)
		n.sendActive(p, spe.Port)
	case COINFEEDBACK:
		print("Feedback Received", string(payload[0]))
		n.feedbackChan <- rune(payload[0])
	case JOINACK:
		// print("JOIN ACK RECEIVED, JOIN SUCCEEDS")
		unmarshalRoutingInfo(payload)
	case WELCOME:
		// print("WELCOME received from", senderID)
		welcomeProtocol(payload)
	case RAWCOINEXCHANGE:
		//print("COINREWARD Exchange Requesting by", senderID)
		if !n.iamBank() {
			n.sendReject("SRY IM NOT BANK", senderPK)
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
		print("A Txn Received from", senderID)
		 txn := blockChain.ProduceTxn(payload[1:], rune(payload[0]))
		 n.bankProxy.AddTxn(txn)
	case CHAINSYNC:
		//print("BlockChain Sync Req Received from", senderID)
		n.chainSyncRequested(payload, senderID)
	case CHAINSYNCACK:
		mutex.Lock()
		//print("BlockChain Sync Ack Received from", senderID)
		n.chainSyncAckReceived(payload, senderID)
		mutex.Unlock()
	case CHAINREPAIR:
		print(senderID, "tries to repair its chain")
		n.chainRepairReceived(payload, senderID)
		msgReceived--
	case CHAINREPAIRREPLY:
		print("repair the chain according to", senderID)
		n.repairChain(payload)
	case PUBLISHINGBLOCK:
		print(senderID, "is trying to publish a block")
		depth := binary.BigEndian.Uint64(payload[:8])
		print("block depth:", depth, n.chain.Size())
	case PUBLISHINGCHECK:
		print(senderID, "responded with publishing status")
	case IPLOOKUP:
		print(senderID, "is looking for someone")
		n.handleLookup(payload)
	case IPLOOKUPRP:
		print("IP found")
	case TXNAGGRE:
		print("Txn Aggregation received from", senderID)
		txns := blockChain.DemuxTxns(payload)
		n.bankProxy.AggreTxns(txns)
		// n.chain.AddNewBlock(blockChain.NewBlock(txns))
		// print(n.chain.GetLastBlock().CurHash)
	case REJECT:
		print(string(payload))
	case EXPT:
		//any exception
	default:
		print("Unknown Msg, discard.")
	}
}

/*
	reject some one when exceptional cases came up.
 */
func (n *Node) sendReject(msg string, senderPK *records.PKEntry) {
	rej := n.prepareOMsg(REJECT,[]byte(msg), senderPK.Pk)
	n.sendActive(rej, senderPK.Port)
}

/*
	Wrap payload into OMsg and encrypt it with target pk.
 */
func (n *Node) prepareOMsg(opCode rune, payload []byte, pk rsa.PublicKey) []byte {
	return records.MarshalOMsg(opCode,payload, n.ID, n.sk, pk)
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
func verifySig(oMsg *records.OMsg, pk *rsa.PublicKey) bool {
	return oMsg.VerifySig(pk)
}

