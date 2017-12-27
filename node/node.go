package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
	"github.com/rainer37/OnionCoin/coin"
	"log"
)

const NODE_PREFIX = "[NODE]"


type Node struct {
	vault *vault.Vault
}

func checkErr(err error){
	if err != nil {log.Fatal(err)}
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
	n.vault = new(vault.Vault)
	n.vault.InitVault()
	return n
}

func (n *Node) GetBalance() int {
	return n.vault.Len()
}

func (n *Node) Deposit(coin *coin.Coin) error {
	return n.vault.Deposit(coin)
}

func (n *Node) Withdraw(rid string) *coin.Coin {
	return n.vault.Withdraw(rid)
}



