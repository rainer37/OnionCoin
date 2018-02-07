package main

import(
	"fmt"
	"os"
	"github.com/rainer37/OnionCoin/node"
	"github.com/rainer37/OnionCoin/records"
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

	if cmd == "j" {
		joinPort := os.Args[3]
		n.Join(joinPort)
	} else if cmd == "i" {
		n.SelfInit()
	}
}