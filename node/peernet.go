package node

import (
	"net"
	"strconv"
	"github.com/rainer37/OnionCoin/records"
	"time"
	"github.com/rainer37/OnionCoin/bank"
	"github.com/rainer37/OnionCoin/ocrypto"
)

const BUFSIZE = 2048
const LOCALHOST = "127.0.0.1"
const NEWBIEMARKER = "100000"

func (n *Node) SelfInit() {
	print("PeerNet Initiated.")

	if n.iamBank() {
		print("My Turn To be Bank!")
		n.bankProxy = bank.InitBank(n.sk, n.chain)
	}

	records.InsertEntry(n.ID, n.sk.PublicKey, time.Now().Unix(), n.IP, n.Port)

	p,err := strconv.Atoi(n.Port)
	checkErr(err)
	// go n.syncBlockChain()
	n.Serve(LOCALHOST, p)
}

/*
	Joining Routine:
		A. No local SK found, considered as new node in the system.
			0. generate a sk.
			1. connect to one node in system.
			2. send REGISTER request. [PK:Proposed ID:IP:Port]
		B. Has a local SK, oldbie then.
			0. retrieve routing information from keys dir.
			1. connect to one of node.
			2. send JOIN request. [IP:Port]

 */
func (n *Node) IniJoin(address string, status int) {
	go n.SelfInit()

	JID := FAKEID + address // FOR NOW ONLY USE FAKEID + PORT AS THE ID.
	print(status)
	if status == 0 {
		payload := []byte(n.IP + ":" + n.Port + "@" + NEWBIEMARKER)
		pe := n.getPubRoutingInfo(JID)
		if pe == nil {
			print("Cannot Join On this Unregistered Peer")
			return
		}
		p := n.prepareOMsg(JOIN, payload, pe.Pk)
		n.sendActive(p, address)
	} else if status == 1 {
		print("Im Newbie")
		if !isBank(JID) {
			print("NewBie Please Join On The Bank Node For Registration.")
			return
		}

		n.sendActive([]byte(REGISTER+string(ocrypto.EncodePK(n.sk.PublicKey))+n.Port), address)

		enPk := <-n.pkChan // waiting for registration cosign finish.

		print("Good! I'm Now Registered")

		talkingPK := ocrypto.DecodePK(enPk)

		records.InsertEntry(JID, talkingPK, time.Now().Unix(), LOCALHOST, address)

		payload := []byte(n.IP+":"+n.Port+"@"+NEWBIEMARKER)
		joinMsg := n.prepareOMsg(JOIN, payload, talkingPK)
		n.sendActive(joinMsg, address)
	}

	select{}
}

/*
	Start listening on [ip:port].
	Handle all incoming packets with dispatch procedure
 */
func (n *Node) Serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	con, err := net.ListenUDP("udp", &addr)
	checkErr(err)

	buffer := make([]byte, BUFSIZE)

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
	build a udp connection and send msg to add.
*/
func (n *Node) sendActive(msg []byte, add string) {
	con, err := net.Dial("udp", ":"+add)
	if err != nil {
		print(err)
		return
	}
	defer con.Close()
	_, err = con.Write(msg)
	if err != nil {
		print(err)
		return
	}
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

/*
	Check if the id given is a current bank.
 */
func checkBankStatus(id string) bool {
	banks := bank.GetBankIDSet()
	for _,bid := range banks {
		if bid == id {
			return true
		}
	}
	return false
}