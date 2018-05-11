package ocrypto

import(
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"errors"
	"fmt"
	"encoding/binary"
	"math/big"
	"math/rand"
	"time"
	"github.com/rainer37/OnionCoin/util"
	"crypto"
)

const CRYPTOPREFIX = "[CRYP]"

var LABEL = []byte("orders")
var rng = crand.Reader

var RSATime int64 = 0
var AESTime int64 = 0
var nano int64 = 1000000
var RSAStep = 0
var AESStep = 0

func print(str ...interface{}) {
	fmt.Print(CRYPTOPREFIX+" ")
	fmt.Println(str...)
}

// generate a pub/private key pair.
// public key is inside of PrivateKey by invoking .PublicKey
func RSAKeyGen() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rng, util.RSAKEYLEN)
	util.CheckErr(err)
	return key
}

func PKEncrypt(pk rsa.PublicKey, payload []byte) []byte {
	start := time.Now()
	cipher, err := rsa.EncryptOAEP(sha256.New(), rng, &pk, payload, LABEL)
	util.CheckErr(err)
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++
	return cipher
}

func PKDecrypt(sk *rsa.PrivateKey, payload []byte) []byte {
	start := time.Now()
	plain, err := rsa.DecryptOAEP(sha256.New(), rng, sk, payload, LABEL)
	util.CheckErr(err)
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++

	return plain
}

func RSASign(sk *rsa.PrivateKey, msg []byte) []byte {
	start := time.Now()
	hashed := util.ShaHash(msg)
	signature, err := rsa.SignPKCS1v15(rng, sk, crypto.SHA256, hashed[:])
	util.CheckErr(err)
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++
	return signature
}

func RSAVerify(pk *rsa.PublicKey, sig []byte, msg []byte) bool {
	start := time.Now()
	hashed := util.ShaHash(msg)
	err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashed[:], sig)
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++
	if err != nil { return false }
	return true
}

func AESEncrypt(key []byte, payload []byte) ([]byte, error) {

	start := time.Now()

	c, err := aes.NewCipher(key)
	if err != nil { return nil, err}

	gcm, err := cipher.NewGCM(c)
	if err != nil { return nil, err}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rng, nonce); err != nil {
		return nil, err
	}

	cipher := gcm.Seal(nonce, nonce, payload, nil)

	ela := time.Since(start)
	AESTime += ela.Nanoseconds()/1000
	AESStep++
	return cipher, nil
}

func AESDecrypt(key []byte, cipherText []byte) ([]byte, error){
	start := time.Now()

	c, err := aes.NewCipher(key)
	if err != nil { return nil, err }

	gcm, err := cipher.NewGCM(c)
	if err != nil { return nil, err}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	plain, err := gcm.Open(nil, nonce, cipherText, nil)

	ela := time.Since(start)
	AESTime += ela.Nanoseconds()/1000
	AESStep++
	return plain, err
}

/*
	Return cipher, encrypted symkey, and err.
 */
func BlockEncrypt(msg []byte, pk rsa.PublicKey) ([]byte, []byte , error) {
	buf := make([]byte, 32)
	rand.Read(buf) // generate random bytes for encryption
	symkey := buf
	cipher, err := AESEncrypt(symkey, msg)
	if err != nil { return nil, nil, err}
	cipherKey := PKEncrypt(pk, symkey)
	return cipher, cipherKey, nil
}

func BlockDecrypt(cipher []byte, cipherKey []byte, sk *rsa.PrivateKey) ([]byte, error) {
	key := PKDecrypt(sk, cipherKey)
	plain, err := AESDecrypt(key, cipher)
	return plain, err
}

func EncodePK(pubkey rsa.PublicKey) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(pubkey.E))
	epk := append(pubkey.N.Bytes(), bs...)
	return epk
}

func DecodePK(enkey []byte) rsa.PublicKey {
	NLen := util.RSAKEYLEN / 8
	if len(enkey) != NLen + 4 {
		panic("wrong length of encoded pk")
	}
	i := new(big.Int)
	i.SetBytes(enkey[:NLen])
	e := int(binary.LittleEndian.Uint32(enkey[NLen:]))
	key := rsa.PublicKey{i, e}
	return key
}