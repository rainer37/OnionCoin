package node

import (
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/records"
	"encoding/json"
	"crypto/rsa"
	"time"
)

/*
	advertise the new comer to random other nodes
 */
func (n *Node) welcomeNewBie(newbieID string) {
	pe := n.getPubRoutingInfo(newbieID)
	if pe == nil {
		print("Cannot find pk by id")
		return
	}
	idb := util.NewBytes(util.IDLEN, []byte(newbieID))
	payload := append(idb, pe.Bytes()...)
	for _, v := range records.AllPEs([]string{n.ID, newbieID}) {
		n.sendOMsg(WELCOME, payload, v)
	}
}

/*
	a warm welcome to newbie.
 */
func (n *Node) welcomeProtocol(payload []byte) {
	id := util.Strip(payload[:util.IDLEN])
	e := new(records.PKEntry)
	err := json.Unmarshal(payload[util.IDLEN:], e)
	util.CheckErr(err)
	n.recordPE(id, e.Pk, e.IP, e.Port)
}

func (n *Node) recordPE(id string, pk rsa.PublicKey, ip string, port string) {
	records.InsertEntry(id, pk, time.Now().Unix(), ip, port)
}

/*
	generate bytes encoding PKEntries.
 */
func (n *Node) gatherRoutingInfo() []byte {
	return records.PackPEs(NUMNODEINFO)
}

/*
	decode received routing info, and update routing table
 */
func unmarshalRoutingInfo(b []byte) {
	records.UnpackPEs(b)
}