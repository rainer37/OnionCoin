package node

import(
	"fmt"
	"log"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"time"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/bank"
	bc "github.com/rainer37/OnionCoin/blockChain"
	"math/rand"
	"encoding/binary"
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
	events chan func(rune, []byte)
}

func NewNode(port string) *Node {
	print("Create a new node.")
	n := new(Node)
	n.Vault = new(coin.Vault)
	n.Port = port
	n.pkChan = make(chan []byte)
	n.regChan = make(chan []byte)
	n.events = make(chan func(rune, []byte))
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
	go n.blockChainEventQueue()
	ticker := time.NewTicker(time.Millisecond * 5000)
	for t := range ticker.C {
		go func() {
			fmt.Println("Tick at", t.Unix())
			banks := bank.GetBankIDSet()
			bid := banks[rand.Int() % len(banks)]
			bpk := n.getPubRoutingInfo(bid)
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(n.chain.Size()))
			p := n.prepareOMsg(CHAINSYNC, buf, bpk.Pk)
			n.sendActive(p, bpk.Port)
		}()
	}
}

func (n *Node) blockChainEventQueue() {
	for {
		select {
		case f := <-n.events:
			f('1', nil)
			print("event got")
		default:
		}
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

func (n* Node) AddEvent(f func(rune, []byte)) {
	n.events <- f
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	fmt.Print(NODEPREFIX+" ")
	fmt.Println(str...)
}