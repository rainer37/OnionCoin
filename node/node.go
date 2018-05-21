package node

import(
	"fmt"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"time"
	"github.com/rainer37/OnionCoin/coin"
	bc "github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/records"
	"strings"
	"github.com/rainer37/OnionCoin/ocrypto"
)

const NODEPREFIX = "[NODE]"
const FAKEID = "FAKEID"
const SELFSKEYPATH = "self.sk"

var silent = false
var opCount = 0
var pathLength = 0
var currentBanks []string
var ela time.Time

type Node struct {
	ID string
	IP string
	Port string
	*coin.Vault
	sk *rsa.PrivateKey
	pkChan chan []byte // for pk lookup await when joining
	bankProxy *Bank
	regChan chan []byte
	iplookup chan string
	feedbackChan chan rune
	chain *bc.BlockChain
}

func NewNode(port string) *Node {
	n := new(Node)
	n.Port = port
	n.pkChan = make(chan []byte)
	n.regChan = make(chan []byte)
	n.iplookup = make(chan string)
	n.feedbackChan = make(chan rune)
	n.sk = produceSK()
	n.Vault = coin.InitVault()
	n.chain = bc.InitBlockChain()
	currentBanks = []string{"FAKEID1338", "FAKEID1339"}
	// TODO: take these superpower out.
	ela = time.Now()
	return n
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

func (n *Node) GetBalance() int64 {
	return n.Vault.GetBalance()
}

func (n *Node) blindSign(rawCoin []byte) []byte {
	return ocrypto.BlindSign(n.sk, rawCoin)
}

/*
	Try retrieving the pub-key and routing information with given id.
	The PK must be retrieved from blockchain.
	The current net addr of associated node may need
	to be looked up if the routing info is outdated.
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

	ip, port := strings.Split(targetAddr, "@")[0],
	strings.Split(targetAddr, "@")[1]

	n.recordPE(id, pe.Pk, ip, port)

	return records.GetKeyByID(id)
}

/*
	Check if there are sk stored locally, if not create one.
	AND change dir into personal dir.
 */
func produceSK() *rsa.PrivateKey {
	dat, err := ioutil.ReadFile(SELFSKEYPATH)
	util.CheckErr(err)
	sk, err := x509.ParsePKCS1PrivateKey(dat)
	util.CheckErr(err)
	return sk
}

func (n *Node) random_exchg() {
	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {
		//fmt.Println("Tick at", t.Unix(), "SEND:", msgSendCount,
		// "RECEIVED:", msgReceived, "OPS:", opCount)
		if !n.iamBank() {

			go func() {
				if n.GetBalance() > 0 {

					if n.isSlientHours() { return }
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
	ticker := time.NewTicker(time.Second * 5)
	for range ticker.C {
		//print(t.Unix(), "SEND:", msgSendCount, "OPS:",
		// opCount, "MSGCOUNT:", omsgCount, "PATHLEN:", pathLength)
		if !n.iamBank() {
			go func() {
				if n.GetBalance() > 0 {

					if n.isSlientHours() { return }

					path := records.RandomPath()
					for _, b := range path {
						n.CoinExchange(b)
					}
					pathLength += len(path)
					msg := "hello, i am " + n.ID
					//print("COINS READY")
					// fmt.Println("Path:", path)

					n.SendOninoMsg(path, msg)
					//print("ONION FIRED")

					opCount++

				} else {
					fmt.Println("Not enough balance")
				}
			}()
		}

	}
}

func print(str ...interface{}) {
	if silent {
		return
	}
	fmt.Print(NODEPREFIX+" ")
	fmt.Println(str...)
}