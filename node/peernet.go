package node

import (
	"net"
	"fmt"
	"strconv"
	"strings"
)

const (
	FWD = "FWD "
	JOIN = "JOIN"
	FIND = "FIND"
	FREE = "FREE"
	COIN = "COIN"
	EXPT = "EXPT"
)

func (n *Node) SelfInit() {
	print("PeerNet Initiated.")
	p,err := strconv.Atoi(n.Port)
	checkErr(err)
	n.Serve(":", p)
}

func (n *Node) dispatch(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	switch string(incoming[:4]) {
	case FWD:
		print("Forwarding")
		n.send([]byte("Fine i will take the coin though."), con, add)
	case JOIN:
		print("Joining "+string(incoming[4:]))
		ok, id, address := deSegementJoinMsg(string(incoming[4:]))

		if !ok {
			print("Invalidate message format, rejected!")
			n.send([]byte("INVALID MSG FORMAT, REJECTED"), con, add)
		}

		verified := n.verifyID(id)

		if !verified {
			print("Invalidate ID, be aware!")
			n.send([]byte("UNABLE TO VERIFY YOUR ID, REJECTED"), con, add)
		}

		n.insert(id, address)

	case FIND:
		print("Finding")
	case FREE:
		//receive the free list
	case COIN:
		//receive the coin
	case EXPT:
		//any exception
	default:
		print("Unknown Msg, discard.")
	}
}

func (n *Node) Serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	con, err := net.ListenUDP("udp", &addr)
	buffer := make([]byte, 2048)

	checkErr(err)

	defer con.Close()

	for {
		len, add, e := con.ReadFromUDP(buffer)
		checkErr(e)
		incoming := buffer[0:len]
		fmt.Println(NODE_PREFIX,"From", add, len, "bytes:[", string(incoming),"]")
		//TODO: verify authenticity of msg
		go n.dispatch(incoming, con, add)
	}
}

func (n *Node) send(msg []byte, con *net.UDPConn, add *net.UDPAddr) {
	_, err := con.WriteTo(msg, add)
	checkErr(err)
}

func (n *Node) sendActive(msg string, add string) {
	con, err := net.Dial("udp", add)
	defer con.Close()
	checkErr(err)
	_, err = con.Write([]byte(msg))
	checkErr(err)
}

func (n *Node) verifyID(id string) bool { return true }

func deSegementJoinMsg(msg string) (bool,string, string) {
	segs := strings.Split(msg, "@")
	if len(segs) != 2 {
		return false,"",""
	}
	return true, segs[0], segs[1]
}