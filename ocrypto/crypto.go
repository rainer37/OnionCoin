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

func rsaTimeInc(start time.Time) {
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++
}

func aesTimeInc(start time.Time) {
	ela := time.Since(start)
	AESTime += ela.Nanoseconds()/1000
	AESStep++
}

// generate a pub/private key pair.
// public key is inside of PrivateKey by invoking .PublicKey
func RSAKeyGen() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rng, util.RSAKEYLEN)
	util.CheckErr(err)
	return key
}

func PKEncrypt(pk rsa.PublicKey, payload []byte) []byte {
	defer rsaTimeInc(time.Now())
	c, err := rsa.EncryptOAEP(sha256.New(), rng, &pk, payload, LABEL)
	util.CheckErr(err)
	return c
}

func PKDecrypt(sk *rsa.PrivateKey, payload []byte) []byte {
	defer rsaTimeInc(time.Now())
	plain, err := rsa.DecryptOAEP(sha256.New(), rng, sk, payload, LABEL)
	util.CheckErr(err)
	return plain
}

func RSASign(sk *rsa.PrivateKey, msg []byte) []byte {
	defer rsaTimeInc(time.Now())
	hashed := util.Sha(msg)
	signature, err := rsa.SignPKCS1v15(rng, sk, crypto.SHA256, hashed[:])
	util.CheckErr(err)
	return signature
}

func RSAVerify(pk *rsa.PublicKey, sig []byte, msg []byte) bool {
	defer rsaTimeInc(time.Now())
	hashed := util.Sha(msg)
	err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashed[:], sig)
	if err != nil { return false }
	return true
}

func AESEncrypt(key []byte, payload []byte) ([]byte, error) {
	defer aesTimeInc(time.Now())
	c, err := aes.NewCipher(key)
	if err != nil { return nil, err}
	gcm, err := cipher.NewGCM(c)
	if err != nil { return nil, err}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rng, nonce); err != nil {
		return nil, err
	}
	ncipher := gcm.Seal(nonce, nonce, payload, nil)
	return ncipher, nil
}

func AESDecrypt(key []byte, cipherText []byte) ([]byte, error){
	defer aesTimeInc(time.Now())
	c, err := aes.NewCipher(key)
	if err != nil { return nil, err }
	gcm, err := cipher.NewGCM(c)
	if err != nil { return nil, err}
	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("cipherText too short")
	}
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	plain, err := gcm.Open(nil, nonce, cipherText, nil)
	return plain, err
}

/*
	Return cipher, encrypted symkey, and err.
 */
func BlockEncrypt(msg []byte, pk rsa.PublicKey) ([]byte, []byte , error) {
	symkey := make([]byte, 32)
	rand.Read(symkey) // generate random bytes for encryption
	ncipher, err := AESEncrypt(symkey, msg)
	if err != nil { return nil, nil, err}
	cipherKey := PKEncrypt(pk, symkey)
	return ncipher, cipherKey, nil
}

func BlockDecrypt(cipher []byte, cipherKey []byte, sk *rsa.PrivateKey) ([]byte, error) {
	key := PKDecrypt(sk, cipherKey)
	plain, err := AESDecrypt(key, cipher)
	return plain, err
}

func EncodePK(pubkey rsa.PublicKey) []byte {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, uint32(pubkey.E))
	return append(pubkey.N.Bytes(), bs...)
}

func DecodePK(enkey []byte) rsa.PublicKey {
	NLen := util.RSAKEYLEN / 8
	if len(enkey) != NLen + 4 {
		panic("wrong length of encoded pk")
	}
	i := new(big.Int)
	i.SetBytes(enkey[:NLen])
	e := int(binary.BigEndian.Uint32(enkey[NLen:]))
	return rsa.PublicKey{i, e}
}