package ocrypto

import(
	"fmt"
)

const CRYPTO_PREFIX string = "[CRYP]"

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", CRYPTO_PREFIX, str)
	case string:
		fmt.Println(CRYPTO_PREFIX, str.(string))
	default:

	}
}

func Crypto_test() {
	print("Crypto ToolKit.")
}