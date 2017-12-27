package node

import (
	"net"
	"fmt"
	"strconv"
)

const (
	FWD = "FWD "
	JOIN = "JOIN"
	FIND = "FIND"
)

func (n *Node) PeerNetInit(port string) {
	print("Peernet Initiated.")
	p,e := strconv.Atoi(port)
	checkErr(e)
	n.Serve("127.0.0.1", p)
}

func (n *Node) dispatch(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	switch string(incoming[:4]) {
	case FWD:
		con.WriteTo([]byte("Hello World"), add)
	case JOIN:
	case FIND:
	default:
		print("Unknown Msg, discard.")
	}
}

func (n *Node) Serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	pc, err := net.ListenUDP("udp", &addr)
	buffer := make([]byte, 2048)

	checkErr(err)

	defer pc.Close()

	for {
		len, add, e := pc.ReadFromUDP(buffer)
		checkErr(e)
		incoming := buffer[0:len]
		fmt.Println(NODE_PREFIX,"From",add, len, "bytes:", incoming)

		go n.dispatch(incoming, pc, add)
	}
}

func (n *Node) send() {}
func (n *Node) receive() {}