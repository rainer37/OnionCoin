package main

import(
	"fmt"
	"os"
	//"github.com/rainer37/OnionCoin/coin"
	//"github.com/rainer37/OnionCoin/vault"
	//"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/node"
)

func main() {

	if len(os.Args) < 2 || len(os.Args) > 5{
		fmt.Println("[MAIN] Usage:\n\toc [port]")
		os.Exit(1)
	}

	port := os.Args[1]

	fmt.Println("[MAIN] OnionCoin v1.0.0 Started...")

	/*
	ocrypto.NewCryptoTK()
	n := node.NewNode()
	fmt.Println("[MAIN] Balance:", n.GetBalance())

	var vault vault.Vault
	coin := coin.New_Coin()
	vault.InitVault()

	err := vault.Deposit(coin)
	if err != nil {
		println(err.Error())
	}

	vault.Withdraw("1338")
	*/

	new(node.Node).PeerNetInit(port)
}