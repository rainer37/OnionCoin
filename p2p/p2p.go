package p2p

import(
	"fmt"
	"strconv"
	"log"
)

const P2P_PREFIX = "[PEER]"

type OCServer interface {
	serve(string, int)
}

type NetInfo struct {
	IP string
	Port int
}

func checkErr(err error){
	if err != nil {log.Fatal(err)}
}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", P2P_PREFIX, str)
	case string:
		fmt.Println(P2P_PREFIX, str.(string))
	default:

	}
}

func P2PInit(ip string, port string) {
	print("p2p net initiated.")
	p,err := strconv.Atoi(port)

	if err != nil {
		print("ERROR on translate port #")
		return
	}

	new(NServer).serve(ip, p)
}
