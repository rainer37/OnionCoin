package node

import (
	"net"
	"strings"
	"fmt"
	"github.com/rainer37/OnionCoin/records"
)

const NUMNODEINFO = 3

func (n *Node) joinProtocol(payload []byte, con *net.UDPConn, add *net.UDPAddr) {

	addrANDisNew := strings.Split(string(payload), "@")
	if len(addrANDisNew) != 2 {
		fmt.Println("Invalid JOIN format, reject")
		n.send([]byte(REJSTR+" "+INVMSGFMT), con, add)
		return
	}

	address, isNew := addrANDisNew[0], addrANDisNew[1]

	print("starting handle JOIN from", address, isNew)

	//TODO: alternatives on node discovery
	senderID := FAKE_ID+strings.Split(address,":")[1]
	n.insert(senderID, address)

	if isNew == "N" {
		print("Welcome to OnionCon")
		jackPayload := n.gatherRoutingInfo()
		// print(jackPayload)
		tpk := records.GetKeyByID(senderID)
		// handle unknown joiner protocol
		jack := records.MarshalOMsg(JOINACK, jackPayload, n.ID, n.sk, tpk.Pk)
		n.sendActive(string(jack), strings.Split(address,":")[1])
	} else if isNew == "O" {

	} else {
		fmt.Println("Invalid JOIN status, reject")
		n.send([]byte(REJSTR+" "+INVMSGFMT), con, add)
	}
}

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
