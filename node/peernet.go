package node

import (
	"net"
	"strconv"
	"github.com/rainer37/OnionCoin/records"
	"crypto/rsa"
	"time"
	"github.com/rainer37/OnionCoin/bank"
)

const BUFSIZE = 2048
const LOCALHOST = "127.0.0.1"

func (n *Node) SelfInit() {
	print("PeerNet Initiated.")

	if n.iamBank() {
		print("My Turn To be Bank!")
		n.bankProxy = bank.InitBank(n.sk)
	}

	p,err := strconv.Atoi(n.Port)
	checkErr(err)
	go n.syncBlockChain()
	n.Serve(LOCALHOST, p)
}

/*
	Joining Routine:
		0. request public from target node if public key unknown
		1. send JOIN request to the joining node with [ip:port, isNew]
		2. into joinProtocol.

 */
func (n *Node) IniJoin(address string) {
	go n.SelfInit()

	// request public key from target node if its pk is unknown
	pk := records.GetKeyByID(FAKEID +address)
	var tPk rsa.PublicKey
	if pk == nil {
		print("No Known Pub-Key Stored, Looking-UP")
		tPk = n.LookUpPK(address)
		records.InsertEntry(FAKEID+address, tPk, time.Now().Unix(), LOCALHOST, address)
	} else {
		tPk = pk.Pk
	}

	isNew := NEWBIE
	// TODO: check if it's old client.
	if isNew != "N" {
		isNew = OLDBIE
	}

	payload := []byte(n.IP+":"+n.Port+"@"+isNew)
	joinMsg := n.prepareOMsg(JOIN, payload, tPk)
	n.sendActive(joinMsg, address)
	select {}
}

/*
	Start listening on [ip:port].
	Handle all incoming packets with dispatch procedure
 */
func (n *Node) Serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	con, err := net.ListenUDP("udp", &addr)
	buffer := make([]byte, BUFSIZE)

	checkErr(err)

	defer con.Close()
	print("Serving ["+addr.String()+"]")
	for {
		l, add, e := con.ReadFromUDP(buffer)
		checkErr(e)
		incoming := buffer[0:l]
		if l < 50 {
			print("From", add, l, "bytes : [", string(incoming), "]")
		}
		go n.dispatch(incoming)
	}
}

/*
	msg: data as bytes to send
	con: local udp connection
	add: remote destination address
	PROBABLY USELESS
*/
func (n *Node) send(msg []byte, con *net.UDPConn, add *net.UDPAddr) {
	_, err := con.WriteTo(msg, add)
	checkErr(err)
}

/*
	build a udp connection and send msg to add.
*/
func (n *Node) sendActive(msg []byte, add string) {
	con, err := net.Dial("udp", ":"+add)
	checkErr(err)
	_, err = con.Write(msg)
	checkErr(err)
	con.Close()
}

/*
	check if n.ID is one of current bank ids.
 */
func (n *Node) iamBank() bool {
	return checkBankStatus(n.ID)
}

func isBank(id string) bool {
	return checkBankStatus(id)
}

func checkBankStatus(id string) bool {
	banks := bank.GetBankIDSet()
	for _,bid := range banks {
		if bid == id {
			return true
		}
	}
	return false
}