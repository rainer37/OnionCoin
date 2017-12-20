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
)

func main() {
	fmt.Println("OnionCoin v1.0.0 Started...")

	coin.New_Coin()
	p2p.Init_p2p_net()
	ocrypto.Crypto_test()
	
	var vault vault.Vault
	vault.Init_vault()

	//for {
		// receiving user commands
	//}
}