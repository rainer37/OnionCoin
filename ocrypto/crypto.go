package ocrypto

import(
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"log"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"errors"
	"fmt"
)

const CRYPTO_PREFIX = "[CRYP]"
const KEY_LEN = 1024
const SYM_KEY_LEN = 128

var LABEL = []byte("orders")
var rng = rand.Reader

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

// generate a pub/private key pair.
// public key is inside of PrivateKey by invoking .PublicKey
func RSAKeyGen() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rng, KEY_LEN)
	checkErr(err)
	return key
}

func PKEncrypt(pk rsa.PublicKey, payload []byte) []byte {
	cipher, err := rsa.EncryptOAEP(sha256.New(), rng, &pk, payload, LABEL)
	checkErr(err)
	return cipher
}

func PKDecrypt(sk *rsa.PrivateKey, payload []byte) []byte {
	plain, err := rsa.DecryptOAEP(sha256.New(), rng, sk, payload, LABEL)
	checkErr(err)
	return plain
}

func AESEncrypt(key []byte, payload []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil { return nil, err}

	gcm, err := cipher.NewGCM(c)
	if err != nil { return nil, err}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, payload, nil), nil
}

func AESDecrypt(key []byte, cipherText []byte) ([]byte, error){
	c, err := aes.NewCipher(key)
	if err != nil { return nil, err }

	gcm, err := cipher.NewGCM(c)
	if err != nil { return nil, err}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	return gcm.Open(nil, nonce, cipherText, nil)
}

func BlockEncrypt(msg []byte, key []byte, pk rsa.PublicKey) ([]byte, []byte , error) {
	cipher, err := AESEncrypt(key, msg)
	if err != nil { return nil, nil, err}
	cipherKey := PKEncrypt(pk, key)
	return cipher, cipherKey, nil
}

func BlockDecrypt(cipher []byte, cipherKey []byte, sk *rsa.PrivateKey) ([]byte, error) {
	key := PKDecrypt(sk, cipherKey)
	return AESDecrypt(key, cipher)
}