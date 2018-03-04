package node

import(
	"fmt"
	"log"
	"crypto/rsa"
	"crypto/x509"
	"os"
	"io/ioutil"
	"time"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/bank"
	bc "github.com/rainer37/OnionCoin/blockChain"
)

const NODEPREFIX = "[NODE]"
const FAKEID = "FAKEID"
const SELFSKEYPATH = "self.sk"

type Node struct {
	ID string
	IP string
	Port string
	*coin.Vault
	sk *rsa.PrivateKey
	pkChan chan []byte // for pk lookup await when joining
	bankProxy *bank.Bank
	regChan chan []byte
	chain *bc.BlockChain
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	fmt.Print(NODEPREFIX+" ")
	fmt.Println(str...)
}

func NewNode(port string) *Node {
	print("Create a new node.")
	n := new(Node)
	n.Vault = new(coin.Vault)
	n.Port = port
	n.pkChan = make(chan []byte)
	n.regChan = make(chan []byte)
	n.sk = produceSK()
	n.InitVault()
	n.chain = bc.InitBlockChain()
	return n
}

func (n *Node) GetBalance() int {
	return n.Vault.Len()
}

func (n *Node) Deposit(coin *coin.Coin)  {
	n.Vault.Deposit(coin)
}

func (n *Node) Withdraw(rid string) *coin.Coin {
	return n.Vault.Withdraw(rid)
}

/*
	Try retrieving the pub-key and routing information with given id.
	The PK must be retrieved from blockchain.
	The current net addr of associated node may need to be looked up if the routing info is outdated.
 */
func (n *Node) getPubRoutingInfo(id string) *records.PKEntry {
	pe := records.GetKeyByID(id)
	return pe
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

/*
	Check if there are sk stored locally, if not create one.
	AND change dir into personal dir.
 */
func produceSK() *rsa.PrivateKey {
	dat, err := ioutil.ReadFile(SELFSKEYPATH)
	checkErr(err)
	sk, err := x509.ParsePKCS1PrivateKey(dat)
	checkErr(err)
	return sk
}

/*
	Try sync block chain with peers.
 */
func (n *Node) syncBlockChain() {
	ticker := time.NewTicker(time.Millisecond * 10000)
	for t := range ticker.C {
		fmt.Println("Tick at", t.Unix())
	}
}

/*
	Given a list of ids of nodes on the path, create a onion wrapping the message to send.
*/
func (n *Node) wrapABigOnion(msg []byte, ids []string) []byte {
	o := msg
	for i:=0; i<len(ids)-1; i++ {
		pe := n.getPubRoutingInfo(ids[i])
		c := n.Vault.Withdraw(ids[i])
		o = ocrypto.WrapOnion(pe.Pk, ids[i+1], c.Bytes(), o)
	}
	return o
}