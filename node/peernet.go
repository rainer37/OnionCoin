package node

import (
	"net"
	"fmt"
	"strconv"
	"strings"
)

func (n *Node) SelfInit() {
	print("PeerNet Initiated.")
	p,err := strconv.Atoi(n.Port)
	checkErr(err)
	n.Serve(":", p)
}

func (n *Node) Serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	con, err := net.ListenUDP("udp", &addr)
	buffer := make([]byte, 2048)

	checkErr(err)

	defer con.Close()
	print("Serving ["+addr.String()+"]")
	for {
		l, add, e := con.ReadFromUDP(buffer)
		checkErr(e)
		incoming := buffer[0:l]
		fmt.Println(NODE_PREFIX,"From", add, l, "bytes : [", string(incoming),"]")
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
	checkErr(err)
	_, err = con.Write([]byte(msg))
	checkErr(err)
	con.Close()
}

func (n *Node) verifyID(id string) bool { return true }

func deSegmentJoinMsg(msg string) (bool, string, string) {
	segs := strings.Split(msg, "@")
	if len(segs) != 2 {
		return false,"",""
	}
	return true, segs[0], segs[1]
}