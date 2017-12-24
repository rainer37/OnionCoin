package p2p

import "net"

const (
	FWD = "FWD "
	JOIN = "JOIN"
	FIND = "FIND"
)

func dispatch(incoming []byte, con *net.UDPConn, add *net.UDPAddr) {
	switch string(incoming[:4]) {
	case FWD:
		con.WriteTo([]byte("Hello World"), add)
	case JOIN:
	case FIND:
	default:
		print("Unknown Msg, discard.")
	}
}
