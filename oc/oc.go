package main

import(
	"fmt"
	"os"
	"github.com/rainer37/OnionCoin/node"
)

func main() {

	if len(os.Args) < 2 || len(os.Args) > 5{
		fmt.Println("[MAIN] Usage:\n\toc [port]")
		os.Exit(1)
	}

	fmt.Println("[MAIN] OnionCoin v1.0.0 Started...")

	defer func() {
		fmt.Println("[MAIN] OnionCoin shudown.")
	}()

	cmd := os.Args[1]

	n := node.NewNode()
	n.IP = "127.0.0.1"
	n.Port = os.Args[2]

	if cmd == "j" {
		n.Join(os.Args[3])
	} else if cmd == "i" {
		n.SelfInit()
	}
}