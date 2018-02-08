package main

import(
	"fmt"
	"os"
	"github.com/rainer37/OnionCoin/node"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
)

const LOCALHOST = "127.0.0.1"

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

	n := node.NewNode()
	n.IP = LOCALHOST
	n.Port = os.Args[2]
	n.ID = node.FAKE_ID+n.Port

	records.GenerateKeyRepo("")

	// for testing
	now := time.Now().Unix()
	if n.Port == "1338" {
		records.InsertEntry("ID1", ocrypto.RSAKeyGen().PublicKey, now, LOCALHOST, "port1")
		records.InsertEntry("MyID2", ocrypto.RSAKeyGen().PublicKey, now, LOCALHOST, "pt2")
		records.InsertEntry("HIsID3", ocrypto.RSAKeyGen().PublicKey, now, LOCALHOST, "p3")
	}
	//b := records.GetKeyByID("ID1")
	//fmt.Println(b.Pk, b.IP, b.Port, b.Time)
	//ab := b.Bytes()
	//p := records.BytesToPKEntry(ab)
	//fmt.Println(p.Pk, p.IP, p.Port, p.Time, len(ab))

	//for i,v := range(records.KeyRepo) {
	//	fmt.Println(i)
	//	fmt.Println(v.Pk)
	//}

	if cmd == "j" {
		joinPort := os.Args[3]
		n.IniJoin(joinPort)
	} else if cmd == "i" {
		n.SelfInit()
	}
}