package node

import (
	"net"
	"strconv"
	"github.com/rainer37/OnionCoin/records"
	"time"
	"github.com/rainer37/OnionCoin/bank"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/blockChain"
	"crypto/sha256"
	"os"
)

const BUFSIZE = 4096 * 2
const LOCALHOST = "127.0.0.1"
const NEWBIEMARKER = "100000"

var msgSendCount = 0
var msgReceived = 0

func (n *Node) SelfInit() {
	print("PeerNet Initiated.")

	n.bankProxy = bank.InitBank(n.sk, n.chain)

	if n.iamBank() {
		print("My Turn To be Bank!")
		n.bankProxy.SetStatus(true)
	} else {
		n.bankProxy.SetStatus(false)
	}

	records.InsertEntry(n.ID, n.sk.PublicKey, time.Now().Unix(), n.IP, n.Port)

	if n.ID == "FAKEID1338" && n.chain.Size() < 2{
		superPK := ocrypto.EncodePK(n.sk.PublicKey)
		pkHash := sha256.Sum256(superPK)
		sig1 := n.blindSign(append(pkHash[:], []byte(n.ID)...))
		sig2 := n.blindSign(append(pkHash[:], []byte(n.ID)...))
		signers := []string{n.ID, n.ID}
		txn := blockChain.NewPKRTxn(n.ID, n.sk.PublicKey, append(sig1, sig2...), signers)
		n.bankProxy.AddTxn(txn)
	}

	p,err := strconv.Atoi(n.Port)
	checkErr(err)
	// go n.syncBlockChain()
	//if !n.iamBank() {
		go n.random_exchg()
		// go n.random_msg()
	//}
	go n.epochTimer()
	//go n.bankStatusDetection()
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
	if status == 0 {
		payload := []byte(n.IP + ":" + n.Port + "@" + NEWBIEMARKER)
		pe := n.getPubRoutingInfo(JID)
		if pe == nil {
			print("Cannot Join On this Unregistered Peer")
			os.Exit(1)
			return
		}
		p := n.prepareOMsg(JOIN, payload, pe.Pk)
		n.sendActive(p, address)
	} else if status == 1 {
		print("Im Newbie")
		if !n.isBank(JID) {
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


	defer con.Close()
	print("Serving ["+addr.String()+"]")
	for {
		buffer := make([]byte, BUFSIZE)
		msgReceived++
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
	msgSendCount++
	// fmt.Println(len(msg))
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

