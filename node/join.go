package node

import (
	"strings"
	"github.com/rainer37/OnionCoin/records"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"crypto/rsa"
	"crypto/sha256"
	"github.com/rainer37/OnionCoin/blockChain"
	"bytes"
)

const NUMNODEINFO = 10
const REGISTER = "PKRQ"
const PKRQACK  = "PKAK"
const PKRQCODELEN = 4
const PKRQLEN = 132
const NUMREGCOSIGNER = 2

func (n *Node) joinProtocol(payload []byte) bool {

	addrANDisNew := strings.Split(string(payload), "@")

	if len(addrANDisNew) != 2 {
		print("Invalid JOIN format, reject")
		return false
	}

	address, isNew := addrANDisNew[0], addrANDisNew[1]

	//print("starting handle JOIN from", address, isNew)

	senderID := FAKEID +strings.Split(address,":")[1]
	senderPort := strings.Split(address,":")[1]
	//n.insert(senderID, address)

	if isNew == NEWBIEMARKER {
		// very new node joining the system.
		// print("Welcome to OnionCon", senderID)
		jackPayload := n.gatherRoutingInfo()
		tpk := records.GetKeyByID(senderID)
		if tpk == nil {
			return false
		}
		jack := n.prepareOMsg(JOINACK, jackPayload, tpk.Pk)
		records.KeyRepo[senderID].Port = senderPort
		n.sendActive(jack, senderPort)
		// print("JOIN ACK sent", len(jack))
	} else {
		print("Invalid JOIN status, reject")
		return false
	}

	return true
}

/*
	Handle join from a new node to the system.
	1. retrieve the pub-key and id from the REGISTER request.
	2. start cosign protocol to verify the nodes with multiple banks.
	3. return my pub-key and the confirmation once registration is done.
 */
func (n* Node) newbieJoin(incoming []byte) bool {
	// if the newbie is joining, special protocol is invoked.
	if string(incoming[:PKRQCODELEN]) == REGISTER {
		newBiePk := ocrypto.DecodePK(incoming[PKRQCODELEN:PKRQCODELEN+PKRQLEN])
		senderAddr := string(incoming[PKRQCODELEN+PKRQLEN:])

		// TODO: remove this cond
		if senderAddr[:4] != "1339" {
			n.registerCoSign(newBiePk, FAKEID+senderAddr)
		} else {
			newbieID := "FAKEID"+senderAddr[:4]
			superPK := ocrypto.EncodePK(newBiePk)
			pkHash := sha256.Sum256(superPK)
			sig1 := n.blindSign(append(pkHash[:], []byte(newbieID)...))
			sig2 := n.blindSign(append(pkHash[:], []byte(newbieID)...))
			signers := []string{n.ID, n.ID}
			txn := blockChain.NewPKRTxn(newbieID, newBiePk, append(sig1, sig2...), signers)
			n.bankProxy.AddTxn(txn)
		}

		records.InsertEntry(FAKEID+senderAddr, newBiePk, time.Now().Unix(), LOCALHOST, senderAddr)
		// PKAK | EncodedPK | PortListening
		n.sendActive([]byte(PKRQACK+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port), senderAddr)
		return true
	} else if string(incoming[:PKRQCODELEN]) == PKRQACK {
		// return the pk to the requesting node to finish the join protocol.
		//print("thank you for the pub-key, you are now registered")
		confirmBytes := incoming[PKRQCODELEN:PKRQCODELEN + PKRQLEN]
		n.pkChan <- confirmBytes
		return true
	}
	return false
}

/*
	generate bytes encoding PKEntries.
 */
func (n *Node) gatherRoutingInfo() []byte {
	result := make([]byte, 0)
	count := 1
	for i,v := range records.KeyRepo {
		if count <= NUMNODEINFO {
			result = bytes.Join([][]byte{result, makeBytesLen([]byte(i)), makeBytesLen(v.Bytes())}, []byte{})
			count++
		} else {
			break
		}
	}
	return result
}

/*
	decode received routing info, and update routing table
 */
func unmarshalRoutingInfo(b []byte) {
	cur := 0
	for cur < len(b) {
		idLen := binary.BigEndian.Uint32(b[cur:cur+4])
		id := string(b[cur+4:cur+4+int(idLen)])
		// print(id, len(records.KeyRepo))
		cur += int(idLen) + 4

		eLen := binary.BigEndian.Uint32(b[cur:cur+4])
		e := records.BytesToPKEntry(b[cur+4:cur+4+int(eLen)])
		cur += int(eLen) + 4

		records.InsertEntry(id, e.Pk, e.Time, e.IP, e.Port)
		// print("inserting", id)
	}
}

/*
	advertise the new comer to random other nodes
 */
 func (n *Node) welcomeNewBie(newbieID string) {
	 pe := records.GetKeyByID(newbieID)
	 if pe == nil {
	 	print("Cannot find pk by id")
	 	return
	 }
	 payload := append(makeBytesLen([]byte(newbieID)), makeBytesLen(pe.Bytes())...)
	 for id, v := range records.KeyRepo {
	 	if newbieID != id && n.ID != id {
			wpayload := n.prepareOMsg(WELCOME,payload,v.Pk)
			n.sendActive(wpayload, v.Port)
		}
	 }
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
	CoSigning protocol to get the newBie registered into blockChain.
	pk, id = newBie'pub key and its id.
 */
func (n *Node) registerCoSign(pk rsa.PublicKey, id string){
	//print("Starting Registration CoSign Protocol")
	// banks := n.chain.GetCurBankIDSet()
	banks := currentBanks
	counter := 1
	newBieInfo := append(ocrypto.EncodePK(pk), []byte(id)...)

	// first sign it by myself.
	pkHash := sha256.Sum256(ocrypto.EncodePK(pk))
	mySig := n.blindSign(append(pkHash[:], []byte(id)...))
	regBytes := mySig
	signers := []string{n.ID}

	for _, b := range banks {
		if counter == NUMREGCOSIGNER {
			break
		}
		if b != n.ID {
			// print("sending REGCOSIGNRQ to", b)
			bpe := n.getPubRoutingInfo(b)
			p := n.prepareOMsg(REGCOSIGNREQUEST,newBieInfo,bpe.Pk)
			n.sendActive(p, bpe.Port)

			var rBytes []byte

			select{
			case reply := <-n.regChan:
				print("cosigned pk received from", b)
				rBytes = reply
			case <-time.After(COSIGNTIMEOUT * time.Second):
				print(b, "reg cosign no response, try next bank")
				counter++
				continue
			}

			regBytes = append(regBytes, rBytes...)
			signers = append(signers, b)
			counter++
		}
	}

	print("Enough Signing Received, Register Node", id, "by", len(signers), "Signer:", signers)

	txn := blockChain.NewPKRTxn(id, pk, regBytes, signers)
	ok := n.bankProxy.AddTxn(txn)
	if ok {
		// n.publishBlock()
	}

	// TODO: sync this?
	go n.broadcastTxn(txn, blockChain.PK)
}

/*
	Upon received register request, sign the pk, and reply it.
 */
func (n *Node) regCoSignRequest(payload []byte, senderID string) {
	pkHash := sha256.Sum256(payload[:PKRQLEN])
	mySig := n.blindSign(append(pkHash[:], payload[PKRQLEN:]...))
	spk := n.getPubRoutingInfo(senderID)
	p := n.prepareOMsg(REGCOSIGNREPLY, mySig, spk.Pk)
	n.sendActive(p, spk.Port)
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

