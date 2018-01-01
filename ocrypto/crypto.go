package ocrypto

import(
	"fmt"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"log"
)

const CRYPTO_PREFIX = "[CRYP]"

type CryptoTK struct {
	Ver Verifier
	Bsig BlindSig
	Sig Signer
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	fmt.Print(CRYPTO_PREFIX+" ")
	fmt.Println(str...)
}

func RSAKeyGen() *rsa.PrivateKey {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	checkErr(err)

	return key
}

func PKEncrypt(pk rsa.PublicKey, payload []byte) []byte {
	label := []byte("orders")
	rng := rand.Reader

	cipher, err := rsa.EncryptOAEP(sha256.New(), rng, &pk, payload, label)

	checkErr(err)

	return cipher
}

func PKDecrypt(sk *rsa.PrivateKey, payload []byte) []byte {

	label := []byte("orders")
	rng := rand.Reader

	plain, err := rsa.DecryptOAEP(sha256.New(), rng, sk, payload, label)
	checkErr(err)

	return plain
}

func NewCryptoTK() *CryptoTK {
	print("New Crypto ToolKit.")
	return new(CryptoTK)
}