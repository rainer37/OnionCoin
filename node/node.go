package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
	"github.com/rainer37/OnionCoin/coin"
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

func NewNode() *Node {
	print("Create a new node.")
	n := new(Node)
	n.Vault = new(vault.Vault)
	n.Vault.InitVault()
	return n
}

func (n *Node) GetBalance() int {
	return n.Vault.Len()
}

func (n *Node) Deposit(coin *coin.Coin) error {
	return n.Vault.Deposit(coin)
}

func (n *Node) Withdraw(rid string) *coin.Coin {
	return n.Vault.Withdraw(rid)
}