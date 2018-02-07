package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
	"github.com/rainer37/OnionCoin/coin"
	"log"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
)

const NODE_PREFIX = "[NODE]"
const FAKE_ID = "FAKEID"
const NEWBIE = "N"
const OLDBIE = "O"
const PKREQUEST  = "PKRQ"
const PKRQACK  = "PKAK"

type Node struct {
	ID string
	IP string
	Port string
	*vault.Vault
	*RoutingTable
	sk *rsa.PrivateKey
	pkChan chan []byte // for pk lookup await
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	fmt.Print(NODE_PREFIX+" ")
	fmt.Println(str...)
}

func NewNode() *Node {
	print("Create a new node.")
	n := new(Node)
	n.Vault = new(vault.Vault)
	n.RoutingTable = new(RoutingTable)
	n.InitRT()
	n.InitVault()
	n.sk = ocrypto.RSAKeyGen()
	n.pkChan = make(chan []byte)
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

/*
	Retrieve the encrypted symmetric key, and decrypt it
	Decrypt the rest of incoming packet, and return it as OMsg
 */

func (n *Node) DecryptOMsg(incoming []byte) *records.OMsg {
	ckey := ocrypto.PKDecrypt(n.sk, incoming[:ocrypto.SYM_KEY_LEN])
	omsg := new(records.OMsg)
	b, err := ocrypto.AESDecrypt(ckey, incoming[ocrypto.SYM_KEY_LEN:])
	if err == nil { return nil }
	omsg.B = b
	return omsg
}
