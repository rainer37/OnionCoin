package ocrypto
/*
	Signing primitives derived from crypto/ras's signing and verifying
 */
import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto"
)

func RSASign(sk *rsa.PrivateKey, msg []byte) []byte {
	hashed := sha256.Sum256(msg)
	signature, err := rsa.SignPKCS1v15(rng, sk, crypto.SHA256, hashed[:])
	checkErr(err)
	return signature
}

func RSAVerify(pk *rsa.PublicKey, sig []byte, msg []byte) bool {
	hashed := sha256.Sum256(msg)
	err := rsa.VerifyPKCS1v15(pk, crypto.SHA256, hashed[:], sig)
	if err != nil {
		return false
	}
	return true
}