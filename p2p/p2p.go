package p2p

import(
	"fmt"
)

const P2P_PREFIX string = "[PEER]"

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", P2P_PREFIX, str)
	case string:
		fmt.Println(P2P_PREFIX, str.(string))
	default:

	}
}

func Init_p2p_net() {
	print("p2p net initiated.")
}