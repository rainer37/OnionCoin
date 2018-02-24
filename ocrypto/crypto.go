package ocrypto

import(
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"log"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"errors"
	"fmt"
	"encoding/binary"
	"math/big"
	"math/rand"
)

const CRYPTOPREFIX = "[CRYP]"
const RSAKEYLEN = 1024
const SYMKEYLEN = 128

var LABEL = []byte("orders")
var rng = crand.Reader

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}

func print(str ...interface{}) {
	fmt.Print(CRYPTOPREFIX+" ")
	fmt.Println(str...)
}

// generate a pub/private key pair.
// public key is inside of PrivateKey by invoking .PublicKey
func RSAKeyGen() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rng, RSAKEYLEN)
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
	if _, err = io.ReadFull(rng, nonce); err != nil {
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

func BlockEncrypt(msg []byte, pk rsa.PublicKey) ([]byte, []byte , error) {
	buf := make([]byte, 32) // generate random bytes
	rand.Read(buf)
	symkey := buf
	cipher, err := AESEncrypt(symkey, msg)
	if err != nil { return nil, nil, err}
	cipherKey := PKEncrypt(pk, symkey)
	return cipher, cipherKey, nil
}

func BlockDecrypt(cipher []byte, cipherKey []byte, sk *rsa.PrivateKey) ([]byte, error) {
	key := PKDecrypt(sk, cipherKey)
	return AESDecrypt(key, cipher)
}

func EncodePK(pubkey rsa.PublicKey) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(pubkey.E))
	return append(pubkey.N.Bytes(), bs...)
}

func DecodePK(enkey []byte) rsa.PublicKey {
	if len(enkey) != 132 {
		panic(nil)
	}
	i := new(big.Int)
	i.SetBytes(enkey[:128])
	e := int(binary.LittleEndian.Uint32(enkey[128:]))
	key := rsa.PublicKey{i, e}
	return key
}