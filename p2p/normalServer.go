package p2p

import (
	"net"
	"log"
	"fmt"
)

type NServer struct {

}

func (server *NServer) serve(ip string, port int) {
	addr := net.UDPAddr{
		Port: port,
		IP: net.ParseIP(ip),
	}
	pc, err := net.ListenUDP("udp", &addr)

	if err != nil {
		log.Fatal(err)
	}
	defer pc.Close()

	buffer := make([]byte, 2048)

	for {
		n, add, e := pc.ReadFromUDP(buffer)

		if e != nil {
			log.Fatal(err)
		}

		incoming := string(buffer[0:n])

		fmt.Println(P2P_PREFIX,"From",add, n, "bytes:", incoming)

		go dispatch(incoming)

		pc.WriteTo([]byte("Hello World"), add)
	}
}