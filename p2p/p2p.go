package p2p

import(
	"fmt"
)

const P2P_PREFIX = "[PEER]"

type NetInfo struct {
	IP string
	Port int

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

func P2PInit() {
	print("p2p net initiated.")
}