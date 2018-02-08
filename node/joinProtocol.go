package node

import (
	"strings"
	"fmt"
	"github.com/rainer37/OnionCoin/records"
)

const NUMNODEINFO = 3

func (n *Node) joinProtocol(payload []byte) bool {

	addrANDisNew := strings.Split(string(payload), "@")

	if len(addrANDisNew) != 2 {
		fmt.Println("Invalid JOIN format, reject")
		return false
	}

	address, isNew := addrANDisNew[0], addrANDisNew[1]

	print("starting handle JOIN from", address, isNew)

	//TODO: alternatives on node discovery
	senderID := FAKE_ID+strings.Split(address,":")[1]
	senderPort := strings.Split(address,":")[1]
	//n.insert(senderID, address)

	if isNew == "N" {
		print("Welcome to OnionCon")
		jackPayload := n.gatherRoutingInfo()
		// print(jackPayload)
		tpk := records.GetKeyByID(senderID)
		// TODO: handle unknown joiner protocol
		jack := records.MarshalOMsg(JOINACK, jackPayload, n.ID, n.sk, tpk.Pk)
		records.KeyRepo[senderID].Port = senderPort
		n.sendActive(string(jack), senderPort)
	} else if isNew == "O" {

	} else {
		fmt.Println("Invalid JOIN status, reject")
		return false
	}

	return true
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
	 		wpayload := records.MarshalOMsg(WELCOME,payload,n.ID,n.sk,v.Pk)
			n.sendActive(string(wpayload), v.Port)
		}
	 }
 }