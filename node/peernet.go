package node

import (
	"net"
	"strconv"
	"github.com/rainer37/OnionCoin/ocrypto"
	"os"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/blockChain"
	"crypto/rsa"
)

const BUFSIZE = 4096 * 10

var msgSendCount = 0

func (n *Node) SelfInit() {

	n.bankProxy = InitBank(n.sk, n.chain)

	n.bankProxy.SetStatus(false)
	if n.iamBank() {
		n.bankProxy.SetStatus(true)
	}

	n.recordPE(n.ID, n.sk.PublicKey, n.IP, n.Port)

	if n.ID == "FAKEID1338" && n.chain.Size() < 2 {
		n.defaultBankRoutine(n.sk.PublicKey, n.ID)
	}

	p,err := strconv.Atoi(n.Port)
	util.CheckErr(err)

	// go n.random_exchg()
	go n.random_msg()
	go n.epochTimer()

	n.Serve(util.LOCALHOST, p)
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
func (n *Node) IniJoin(targetAddress string, status int) {
	go n.SelfInit()

	JID := FAKEID + targetAddress // FOR NOW ONLY USE FAKEID + PORT AS THE ID.

	// if i am new to system, register first.
	if status == 1 {
		print("Im Newbie")
		if !n.isBank(JID) {
			print("I need to join on banks")
			return
		}

		/*
			REGISTER | EncodedPK(RSAKEYLEN / 8) | MyID
		*/
		encodedPK := ocrypto.EncodePK(n.sk.PublicKey)
		regPayload := []byte(REGISTER + string(encodedPK) + n.ID)
		n.sendActive(regPayload, targetAddress)

		encodedPk := <-n.pkChan // waiting for registration cosign finish.

		print("Good! I'm In The System")
		targetPK := ocrypto.DecodePK(encodedPk)

		// insert target PE into repo
		n.recordPE(JID, targetPK, util.LOCALHOST, targetAddress)
	}

	pe := n.getPubRoutingInfo(JID)
	if pe == nil {
		print("Cannot Join On this Unregistered Peer")
		os.Exit(1)
	}

	// ID @ IP:port
	payload := []byte(n.ID + "@" + n.IP + ":" + n.Port)
	n.sendOMsg(JOIN, payload, pe)

	select{}
}

/*
	Start listening on [ip:port].
	Handle all incoming packets with dispatch procedure
 */
func (n *Node) Serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	con, err := net.ListenUDP("udp", &addr)
	util.CheckErr(err)

	defer con.Close()
	print("Serving [" + addr.String() + "]")
	for {
		buffer := make([]byte, BUFSIZE)
		l, _, e := con.ReadFromUDP(buffer)
		util.CheckErr(e)
		incoming := buffer[0:l]
		go n.dispatch(incoming)
	}
}

/*
	build a udp connection and send msg to add.
*/
func (n *Node) sendActive(msg []byte, add string) {
	msgSendCount++
	// fmt.Println(len(msg))
	con, err := net.Dial("udp", ":" + add)
	util.CheckErr(err)
	defer con.Close()
	_, err = con.Write(msg)
	util.CheckErr(err)
}

func (n *Node) defaultBankRoutine(pk rsa.PublicKey, id string) {
	superPK := ocrypto.EncodePK(pk)
	pkHash := util.ShaHash(superPK)
	sig1 := n.blindSign(append(pkHash[:], []byte(id)...))
	sig2 := n.blindSign(append(pkHash[:], []byte(id)...))
	signers := []string{n.ID, n.ID}
	txn := blockChain.NewPKRTxn(id, pk, append(sig1, sig2...), signers)
	n.bankProxy.AddTxn(txn)
}