package p2p

import (
	"net"
	"fmt"
)

type NServer struct {}

func (server *NServer) serve(ip string, port int) {
	addr := net.UDPAddr{Port: port, IP: net.ParseIP(ip)}
	pc, err := net.ListenUDP("udp", &addr)
	buffer := make([]byte, 2048)

	checkErr(err)

	defer pc.Close()

	for {
		n, add, e := pc.ReadFromUDP(buffer)
		checkErr(e)
		incoming := buffer[0:n]
		fmt.Println(P2P_PREFIX,"From",add, n, "bytes:", incoming)
		go dispatch(incoming, pc, add)
	}
}