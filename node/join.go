package node

import (
	"strings"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/util"
)

const NUMNODEINFO = 10
const REGISTER = "PKRQ"
const PKRQACK  = "PKAK"
const PDLEN = 4
const PKRQLEN = util.CIPHERKEYLEN + 4

func (n *Node) joinProtocol(payload []byte) {

	IDAndAddr := strings.Split(string(payload), "@")

	if len(IDAndAddr) != 2 {
		print("Invalid JOIN format, reject")
		return
	}

	senderID := IDAndAddr[0]
	senderIP := strings.Split(IDAndAddr[1],":")[1]
	senderPort := strings.Split(IDAndAddr[1],":")[1]

	// very new node joining the system.
	ackPayload := n.gatherRoutingInfo()
	tpe := n.getPubRoutingInfo(senderID)
	if tpe == nil {
		return
	}

	n.sendOMsg(JOINACK, ackPayload, tpe)

	n.recordPE(senderID, tpe.Pk, senderIP, senderPort)

	// broadcast newbie's info to other peers.
	n.welcomeNewBie(senderID)
}

/*
	Handle join from a new node to the system.
	1. retrieve the pub-key and id from the REGISTER request.
	2. start cosign protocol to verify the nodes with multiple banks.
	3. return my pub-key and the confirmation once registration is done.
 */
func (n* Node) newbieJoin(b []byte) bool {
	// if the newbie is joining, special protocol is invoked.
	if string(b[:PDLEN]) == REGISTER {
		newBiePk := ocrypto.DecodePK(b[PDLEN: PDLEN + PKRQLEN])
		senderID := string(b[PDLEN + PKRQLEN:])
		senderAddr := senderID[6:]

		// TODO: remove this cond
		if senderID != "FAKEID1339" {
			n.registerCoSign(newBiePk, senderID)
		} else {
			n.defaultBankRoutine(newBiePk, senderID)
		}

		n.recordPE(senderID, newBiePk, util.LOCALHOST, senderAddr)

		// PKAK | EncodedPK | PortListening
		ackPayload := append([]byte(PKRQACK), ocrypto.EncodePK(n.sk.PublicKey)...)
		n.sendActive(ackPayload, senderAddr)
		return true
	} else if string(b[:PDLEN]) == PKRQACK {
		// return the pk to the requesting node to finish the join protocol.
		confirmBytes := b[PDLEN: PDLEN + PKRQLEN]
		n.pkChan <- confirmBytes
		return true
	}
	return false
}