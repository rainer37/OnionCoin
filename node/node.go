package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
)

const NODE_PREFIX string = "[NODE]"

type Node struct {
	Vault *vault.Vault
}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", NODE_PREFIX, str)
	case string:
		fmt.Println(NODE_PREFIX, str.(string))
	default:

	}
}

func New_Node() *Node {
	print("Create a new node.")
	n := new(Node)
	n.Vault = new(vault.Vault)
	n.Vault.Init_vault()
	return n
}