package ocrypto

import(
	"fmt"
)

const CRYPTO_PREFIX = "[CRYP]"

type CryptoTK struct {
	Ver Verifier
	Bsig BlindSig
	Sig Signer
	OMaker OnionMaker
}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", CRYPTO_PREFIX, str)
	case string:
		fmt.Println(CRYPTO_PREFIX, str.(string))
	default:

	}
}

func NewCryptoTK() *CryptoTK {
	print("Crypto ToolKit.")
	return new(CryptoTK)
}