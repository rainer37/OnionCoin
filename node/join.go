package node

import (
	"strings"
	"fmt"
	"github.com/rainer37/OnionCoin/records"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
)

const NUMNODEINFO = 3
const PKREQUEST  = "PKRQ"
const PKRQACK  = "PKAK"
const PKRQCODELEN = 4

func (n *Node) joinProtocol(payload []byte) bool {

	addrANDisNew := strings.Split(string(payload), "@")

	if len(addrANDisNew) != 2 {
		fmt.Println("Invalid JOIN format, reject")
		return false
	}

	address, isNew := addrANDisNew[0], addrANDisNew[1]

	print("starting handle JOIN from", address, isNew)

	//TODO: alternatives on node discovery
	senderID := FAKEID +strings.Split(address,":")[1]
	senderPort := strings.Split(address,":")[1]
	//n.insert(senderID, address)

	if isNew == "N" {
		print("Welcome to OnionCon")
		jackPayload := n.gatherRoutingInfo()
		// print(jackPayload)
		tpk := records.GetKeyByID(senderID)
		if tpk == nil {
			// TODO: handle unknown joiner protocol
			return false
		}
		jack := n.prepareOMsg(JOINACK, jackPayload, tpk.Pk)
		records.KeyRepo[senderID].Port = senderPort
		n.sendActive(jack, senderPort)
	} else if isNew == "O" {
		return false
	} else {
		fmt.Println("Invalid JOIN status, reject")
		return false
	}

	return true
}

func (n* Node) newbieJoin(incoming []byte) bool {
	// if the newbie is joining, special protocol is invoked.
	if string(incoming[:PKRQCODELEN]) == PKREQUEST {
		spk := ocrypto.DecodePK(incoming[PKRQCODELEN:PKRQCODELEN+PKRQLEN])
		senderAddr := string(incoming[PKRQCODELEN+PKRQLEN:])
		records.InsertEntry(FAKEID+senderAddr, spk, time.Now().Unix(), LOCALHOST, senderAddr)
		// PKAK | EncodedPK | PortListening
		n.sendActive([]byte(PKRQACK+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port), senderAddr)
		return true
	} else if string(incoming[:PKRQCODELEN]) == PKRQACK {
		// return the pk to the requesting node to finish the join protocol.
		print("thank you for the pub-key")
		n.pkChan <- incoming[PKRQCODELEN:PKRQCODELEN+PKRQLEN]
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
			result = append(result, makeBytesLen([]byte(i))...)
			result = append(result, makeBytesLen(v.Bytes())...)
			count++
		} else {
			break
		}
	}
	return result
}

/*
	advertise the new comer to random other nodes
 */
 func (n *Node) welcomeNewBie(newbieID string) {
	 pe := records.GetKeyByID(newbieID)
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