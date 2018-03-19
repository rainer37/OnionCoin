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
	"github.com/rainer37/OnionCoin/bank"
	bc "github.com/rainer37/OnionCoin/blockChain"
	"strings"
)

const NODEPREFIX = "[NODE]"
const FAKEID = "FAKEID"
const SELFSKEYPATH = "self.sk"
const IDLEN = 16

var slient = false
var opCount = 0
var pathLength = 0

type Node struct {
	ID string
	IP string
	Port string
	*coin.Vault
	sk *rsa.PrivateKey
	pkChan chan []byte // for pk lookup await when joining
	bankProxy *bank.Bank
	regChan chan []byte
	iplookup chan string
	blockBuffer chan int
	chain *bc.BlockChain
}

func NewNode(port string) *Node {
	n := new(Node)
	n.Vault = new(coin.Vault)
	n.Port = port
	n.pkChan = make(chan []byte)
	n.regChan = make(chan []byte)
	n.iplookup = make(chan string)
	n.blockBuffer = make(chan int)
	n.sk = produceSK()
	n.InitVault()
	n.chain = bc.InitBlockChain()
	return n
}

func (n *Node) GetBalance() int {
	return n.Vault.Len()
}

func (n *Node) addr() string {
	return n.IP + ":" + n.Port
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

	if pe == nil {
		print("no such dude in system")
		return nil
	}

	if pe.Port != "" {
		return pe
	}

	n.LookUpIP(id)
	targetAddr := <-n.iplookup

	ip, port := strings.Split(targetAddr, "@")[0],  strings.Split(targetAddr, "@")[1]

	records.InsertEntry(id, pe.Pk, time.Now().Unix(), ip, port)

	return records.GetKeyByID(id)
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

func (n *Node) bankStatusDetection() {
	ticker := time.NewTicker(time.Second * 2)
	for range ticker.C {
		n.chain.GetBankIDSet()
	}
}

func (n *Node) random_exchg() {
	ticker := time.NewTicker(time.Second * 2)
	for range ticker.C {
		//fmt.Println("Tick at", t.Unix(), "SEND:", msgSendCount, "RECEIVED:", msgReceived, "OPS:", opCount)
		if !n.iamBank() {

			go func() {
				if coin.GetBalance() > 0 {

					n.CoinExchange(n.ID)
					opCount++

				} else {
					fmt.Println("Not enough balance")
				}
			}()
		}

	}
}

func (n *Node) random_msg() {
	ticker := time.NewTicker(time.Second * 3)
	for t := range ticker.C {
		fmt.Println(t.Unix(), "SEND:", msgSendCount, "OPS:", opCount, "MSGCOUNT:", omsgCount, "PATHLEN:", pathLength)
		if !n.iamBank() {
			go func() {
				if coin.GetBalance() > 0 {

					path := records.RandomPath()
					pathLength += len(path)
					msg := "hello, i am " + n.ID

					// fmt.Println("Path:", path)

					n.SendOninoMsg(path, msg)

					opCount++

				} else {
					fmt.Println("Not enough balance")
				}
			}()
		}

	}
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	if slient {
		return
	}
	fmt.Print(NODEPREFIX+" ")
	fmt.Println(str...)
}