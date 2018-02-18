package main

import(
	"fmt"
	"os"
	"github.com/rainer37/OnionCoin/node"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"github.com/rainer37/OnionCoin/ui"
)

const LOCALHOST = "127.0.0.1"

func loadSavedStates(n *node.Node) {

}

func main() {

	if len(os.Args) < 2 || len(os.Args) > 5{
		fmt.Println("[MAIN] Usage:" +
			"\n\toc i [myport]" +
			"\n\toc j [myport] [joinport]")
		os.Exit(1)
	}

	fmt.Println("[MAIN] OnionCoin v1.0.0 Started...")

	defer func() {
		fmt.Println("[MAIN] OnionCoin shudown.")
	}()

	cmd := os.Args[1]

	port := os.Args[2]
	n := node.NewNode(port)
	n.IP = LOCALHOST
	n.ID = node.FAKEID +n.Port

	loadSavedStates(n)
	records.GenerateKeyRepo()

	// for testing
	now := time.Now().Unix()
	if n.Port == "1337" {
		records.InsertEntry("ID1", ocrypto.RSAKeyGen().PublicKey, now, LOCALHOST, "port1")
		records.InsertEntry("MyID2", ocrypto.RSAKeyGen().PublicKey, now, LOCALHOST, "pt2")
		records.InsertEntry("HIsID3", ocrypto.RSAKeyGen().PublicKey, now, LOCALHOST, "p3")
	}

	if cmd == "j" {
		joinPort := os.Args[3]
		go n.IniJoin(joinPort)
	} else if cmd == "i" {
		go n.SelfInit()
	}

	go ui.Listen(n)

	select {}
}