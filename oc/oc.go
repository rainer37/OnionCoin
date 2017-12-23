package main

/*
	Client for testing.
*/

import(
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/vault"	
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/p2p"
	"github.com/rainer37/OnionCoin/node"
)

func main() {
	fmt.Println("[MAIN] OnionCoin v1.0.0 Started...")

	p2p.P2PInit()
	ocrypto.Crypto_test()
	n := node.NewNode()
	fmt.Println(n.GetBalance())

	var vault vault.Vault
	coin := coin.New_Coin()
	vault.InitVault()

	err := vault.Deposit(coin)
	if err != nil {
		println(err.Error())
	}

	vault.Withdraw("1338")

	//for {
		// receiving user commands
	//}
}