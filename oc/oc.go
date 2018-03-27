package main

import(
	"fmt"
	"os"
	"github.com/rainer37/OnionCoin/node"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ui"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/x509"
	"io/ioutil"
	"log"
)

const LOCALHOST = "127.0.0.1"
const SELFSKEYPATH = "self.sk"
const USAGE = "[MAIN] Usage:" + "\n\toc i [myport]" + "\n\toc j [myport] [joinport]"
/*
	Load saved local states.
	0. check if there is local SK, if not generate one and write it to file.
 */
func loadSavedStates(addr string) (status int){
	status = 0

	if yes, _ := exists(addr); !yes {
		os.Mkdir(addr, 0777)
		status = 1
	}

	os.Chdir(addr) // go into oc info dir

	if yes, _ := exists(SELFSKEYPATH); !yes {
		file, err := os.Create(SELFSKEYPATH)
		defer file.Close()
		checkErr(err)
		sk := ocrypto.RSAKeyGen()
		skBytes := x509.MarshalPKCS1PrivateKey(sk)
		ioutil.WriteFile(SELFSKEYPATH, skBytes, 0644)
		status = 1
	}

	return
}

func main() {

	if len(os.Args) < 2 || len(os.Args) > 5{
		fmt.Println(USAGE)
		os.Exit(1)
	}

	cmd := os.Args[1]

	if cmd != "i" && cmd != "j" {
		fmt.Println(USAGE)
		os.Exit(1)
	}

	fmt.Println("[MAIN] OnionCoin v1.0.0 Started...")

	defer func() {
		fmt.Println("[MAIN] OnionCoin shudown.")
	}()

	port := os.Args[2]

	status := loadSavedStates(port)

	n := node.NewNode(port)
	n.IP = LOCALHOST
	n.ID = node.FAKEID+n.Port

	gcoin := n.GetGenesisCoin()
	n.Vault.Deposit(gcoin)

	records.GenerateKeyRepo()

	if cmd == "j" {
		joinAddr := os.Args[3]
		go n.IniJoin(joinAddr, status)
	} else if cmd == "i" {
		go n.SelfInit()
	}

	go ui.Listen(n)

	select {}
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}
