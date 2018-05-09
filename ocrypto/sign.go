package ocrypto
/*
	Signing primitives derived from crypto/ras's signing and verifying
 */
import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto"
	"time"
	"github.com/rainer37/OnionCoin/util"
)

func RSASign(sk *rsa.PrivateKey, msg []byte) []byte {
	start := time.Now()
	hashed := sha256.Sum256(msg)
	signature, err := rsa.SignPKCS1v15(rng, sk, crypto.SHA256, hashed[:])
	util.CheckErr(err)
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++

	return signature
}

func RSAVerify(pk *rsa.PublicKey, sig []byte, msg []byte) bool {
	start := time.Now()
	hashed := sha256.Sum256(msg)
	err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashed[:], sig)
	ela := time.Since(start)
	RSATime += ela.Nanoseconds()/nano
	RSAStep++

	if err != nil {
		return false
	}
	return true
}