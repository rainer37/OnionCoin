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
	fmt.Println("OnionCoin Client v1.0.0 Started...")
	/*
	1. 	check if there is local ledger/registry/coins:
			yes: read it from disk
			no: create a new one
		in either case update with peers:

	2.	start the crypto manager
	3.	start the p2p net server listening on [ip:port]

	5.	UI initiated
	*/
	coin.Test_coin()
	p2p.Init_p2p_net()
	ocrypto.Crypto_test()
	vault.Init_vault()

	for {
		// receiving user commands
	}
}