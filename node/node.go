package node

import(
	"fmt"
	"github.com/rainer37/OnionCoin/vault"
	"github.com/rainer37/OnionCoin/coin"
	"log"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/bank"
	"crypto/x509"
	"os"
	"io/ioutil"
)

const NODEPREFIX = "[NODE]"
const FAKEID = "FAKEID"
const NEWBIE = "N"
const OLDBIE = "O"
const SELFSKEYPATH = "self.sk"

type Node struct {
	ID string
	IP string
	Port string
	*vault.Vault
	sk *rsa.PrivateKey
	pkChan chan []byte // for pk lookup await
	bankProxy *bank.Bank
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
	n.Vault = new(vault.Vault)
	n.InitVault()
	n.Port = port
	n.pkChan = make(chan []byte)
	n.sk = produceSK(port)
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
	ckey := ocrypto.PKDecrypt(n.sk, incoming[:ocrypto.SYMKEYLEN])
	omsg := new(records.OMsg)
	b, err := ocrypto.AESDecrypt(ckey, incoming[ocrypto.SYMKEYLEN:])
	if err == nil { return nil }
	omsg.B = b
	return omsg
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

/*
	Check if there are sk stored locally, if not create one.
 */
func produceSK(port string) *rsa.PrivateKey {
	if yes,_ := exists(port);!yes {
		os.Mkdir(port, 0777)
	}

	os.Chdir(port) // go into oc info dir

	if yes, _ := exists(SELFSKEYPATH); yes {
		dat, err := ioutil.ReadFile(SELFSKEYPATH)
		checkErr(err)
		sk, err := x509.ParsePKCS1PrivateKey(dat)
		checkErr(err)
		return sk
	}

	fmt.Println(os.Getwd())
	file, err := os.Create(SELFSKEYPATH)
	defer file.Close()
	checkErr(err)
	sk := ocrypto.RSAKeyGen()
	skBytes := x509.MarshalPKCS1PrivateKey(sk)
	go ioutil.WriteFile(SELFSKEYPATH, skBytes, 0644)
	return sk
}