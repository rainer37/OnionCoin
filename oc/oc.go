package main

/*
	Client for testing.
*/

import(
	"fmt"
	chain "github.com/rainer37/OnionCoin/blockchain_oc"
	"github.com/rainer37/OnionCoin/crypto_oc/onion"
	"github.com/rainer37/OnionCoin/peernet"
	"github.com/rainer37/OnionCoin/crypto_oc/pkpair"
)

func main() {
	fmt.Println("Onicrypto_oc/pkpairClient v1.0.0 Started...")
	chain.New_Chain()
	onion.Onion()
	peernet.Peernet_Init()
	peernet.Serve()

	pub, prv := pkpair.KeyGen()
	
	cipher := pkpair.Encrypt("Bring them down", pub)
	plain := pkpair.Decrypt(cipher, prv)

	fmt.Println(plain)

	for {

	}
}