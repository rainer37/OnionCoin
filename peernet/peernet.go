package peernet

/*
	p2p network building and cooperatting.
*/

import(
	"fmt"
	"net"
	"os"
)

func Peernet_Init() {
	fmt.Println("Joining Onion Coin Peernet...")
}

/*
	Peernet server listen for requests from peers.
*/
func Serve() {
	go func() {

		lnn, err := net.Listen("tcp", ":8331")
		if err != nil { fmt.Println("Cannot Listen on porrt 8331"); os.Exit(1) }

		defer lnn.Close()

		fmt.Println("Deamon started")

		for {
			conn, err := lnn.Accept()

			if err != nil { fmt.Println("Error on accepting connection") }

			go receiving(conn)
		}
	
	}()
}

/*
	sample handler. 
	echo.
*/
func receiving(conn net.Conn) {
	buf := make([]byte, 512)
	_, err := conn.Read(buf)
	if err != nil { fmt.Println("Error reading external msg") }
	fmt.Println(string(buf))
	conn.Close()
}

func Query_for_peers() {

}

func load_local_state() {
	
}