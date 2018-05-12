package node

import(
	"fmt"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"time"
	"github.com/rainer37/OnionCoin/coin"
	bc "github.com/rainer37/OnionCoin/blockChain"
	"encoding/json"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/records"
	"strings"
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
	currentBanks = []string{"FAKEID1338", "FAKEID1339"} // TODO: take these superpower out.
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

/*
	periodically check if the node is the bank, and update the status.
 */
func (n *Node) bankStatusDetection() {
	ticker := time.NewTicker(time.Second * 2)
	for range ticker.C {
		currentBanks = n.chain.GetCurBankIDSet()
		if n.iamBank() {
			print("		My Turn To be Bank!")
			n.bankProxy.SetStatus(true)
		} else {
			// print("		Damn, not a bank!")
			n.bankProxy.SetStatus(false)
		}
	}
}

/*
	timer to check epoch change, update banksets, and start proposing timer.
 */
func (n *Node) epochTimer() {
	epochLen := int64(util.EPOCHLEN)

	defer func() {
		print("BOOM!\n\n\n\n")
	}()

	if time.Now().Unix() % epochLen != 0 {
		nextEpoch := (time.Now().Unix()/epochLen + 1) * epochLen
		diff := nextEpoch - time.Now().Unix()
		print(diff)
		timer1 := time.NewTimer(time.Duration(diff) * time.Second)
		<-timer1.C
	}

	ticker := time.NewTicker(time.Duration(epochLen) * time.Second)
	// n.syncOnce()
	for t := range ticker.C {
		percent := float64(ocrypto.RSATime) / float64(time.Since(ela).Nanoseconds() / 1000000)
		fmt.Println(t.Unix() / epochLen, msgSendCount - bcCount , omsgCount, pathLength, ocrypto.RSAStep, ocrypto.AESStep, ocrypto.RSATime, ocrypto.AESTime, percent * 100,"%")
		fmt.Println(currentBanks)
		currentBanks = n.chain.GetCurBankIDSet()
		go func() {
			if n.iamBank() {
				// start proposing timer
				n.bankProxy.SetStatus(true)

				propTimer := time.NewTimer(util.PROPOSINGTIME * time.Second)
				go func() {
					<-propTimer.C
					HashCmpMap = make(map[string]int)
					print("Time to propose my txns", t.Unix())
					n.syncOnce()
					go func() {
						for _, b := range currentBanks {
							if b == n.ID { continue }
							bpe := n.getPubRoutingInfo(b)
							if bpe == nil { continue }
							txnsBytes := n.getTxnsInBuffer()
							if string(txnsBytes) != "null" {
								n.sendOMsg(TXNAGGRE, txnsBytes, bpe)
							}
						}
					}()
				}()

				// start pushing timer
				pushTimer := time.NewTimer(util.PUSHTIME * time.Second)
				go func() {
					<-pushTimer.C
					print("Time to push my block", t.Unix())
					go func() {
						// n.bankProxy.GenerateNewBlock()
						nb := n.bankProxy.GenNewBlock()
						if nb != nil {
							HashCmpMap[string(nb.CurHash)] = 1
							for _, b := range currentBanks {
								if b == n.ID { continue }
								bpe := n.getPubRoutingInfo(b)
								if bpe == nil { continue }
								n.sendOMsg(HASHCMP, nb.CurHash, bpe)
							}
							cmpTimer := time.NewTimer(2 * time.Second)
							<-cmpTimer.C
							if n.bankProxy.IsMajorityHash(string(nb.CurHash)) {
								n.chain.StoreBlock(nb)
							} else {
								print("i has minor hash, wait for sync ***************************")
							}
							n.bankProxy.CleanBuffer()
						}
					}()
				}()
			} else if n.iamNextBank() {
				n.bankProxy.SetStatus(false)
				print("!!! i'm one of the next gen banks, so? !!!")
				n.syncOnce()
				//pullTimer := time.NewTimer(bc.PUSHTIME * time.Second)
				//go func() {
				//	<-pullTimer.C
				//	fmt.Println("Time to push my block", t.Unix())
				//	go func() {
				//		n.bankProxy.GenerateNewBlock()
				//	}()
				//}()
			} else {
				n.bankProxy.SetStatus(false)
				n.syncOnce()
			}
		}()
	}
}

func (n *Node) random_exchg() {
	ticker := time.NewTicker(time.Second * 1)
	for range ticker.C {
		//fmt.Println("Tick at", t.Unix(), "SEND:", msgSendCount, "RECEIVED:", msgReceived, "OPS:", opCount)
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
		//print(t.Unix(), "SEND:", msgSendCount, "OPS:", opCount, "MSGCOUNT:", omsgCount, "PATHLEN:", pathLength)
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

func (n *Node) getTxnsInBuffer() []byte {
	txns, err := json.Marshal(n.bankProxy.GetTxnBuffer())
	util.CheckErr(err)
	return txns
}

func (n *Node) isSlientHours() bool {
	nextEpoch := (time.Now().Unix()/util.EPOCHLEN + 1) * util.EPOCHLEN
	t := time.Now().Unix()
	if t > nextEpoch - util.PROPOSINGDELAY {
		return true
	}
	return false
}

/*
	check if n.ID is one of current bank ids.
 */
func (n *Node) iamBank() bool {
	return n.checkBankStatus(n.ID)
}

func (n *Node) isBank(id string) bool {
	return n.checkBankStatus(id)
}

func (n *Node) iamNextBank() bool {
	return util.Contains(n.chain.GetNextBankIDSet(), n.ID)
}

/*
	Check if the id given is a current bank.
 */
func (n* Node) checkBankStatus(id string) bool {
	return util.Contains(currentBanks, id)
}

func print(str ...interface{}) {
	if silent {
		return
	}
	fmt.Print(NODEPREFIX+" ")
	fmt.Println(str...)
}