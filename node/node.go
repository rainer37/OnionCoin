package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
	"github.com/rainer37/OnionCoin/coin"
	"log"
)

const NODE_PREFIX = "[NODE]"

type PKPair struct {
	pk []byte
	sk []byte
}

type Node struct {
	IP string
	Port string
	*vault.Vault
	*PKPair
	*RoutingTable
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
	n.Vault = new(vault.Vault)
	n.RoutingTable = new(RoutingTable)
	n.InitRT()
	n.InitVault()
	return n
}

func (n *Node) GetBalance() int {
	return n.Len()
}

func (n *Node) Deposit(coin *coin.Coin) error {
	return n.Deposit(coin)
}

func (n *Node) Withdraw(rid string) *coin.Coin {
	return n.Withdraw(rid)
}

func (n *Node) Join(address string) {
	n.sendActive("JOIN"+n.Port, address)
}


