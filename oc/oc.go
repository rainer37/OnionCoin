package main

/*
	Client for testing.
*/

import(
	"fmt"
	chain "github.com/rainer37/OnionCoin/blockchain_oc"
	"github.com/rainer37/OnionCoin/crypto_oc/onion"
)

func main() {
	fmt.Println("OnionCoin Client v1.0.0 Started...")
	chain.New_Chain()
	onion.Onion()
}