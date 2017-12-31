package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
	"github.com/rainer37/OnionCoin/coin"
	"log"
)

const NODE_PREFIX = "[NODE]"
const FAKE_ID = "FAKEID"

type PKPair struct {
	pk []byte
	sk []byte
}

type Node struct {
	ID string
	IP string
	Port string
	*vault.Vault
	*PKPair
	*RoutingTable
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	//switch str.(type) {
	//case int, uint, uint64:
	//	fmt.Printf("%s %d\n", NODE_PREFIX, str)
	//case string:
	//	fmt.Println(NODE_PREFIX, str)
	//default:
	//
	//}
	//fmt.Println(str)
	fmt.Printf("%s %v\n", NODE_PREFIX, str)

}

func NewNode() *Node {
	print("Create a new node.")
	n := new(Node)
	n.ID = FAKE_ID+n.Port
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

