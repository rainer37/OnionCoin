package pkpair

/*
	RSA key pair encrypt/decrypt.
*/

import(
	"crypto/rsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
)

var pub_key *rsa.PublicKey
var prv_key *rsa.PrivateKey
var label []byte

func Encrypt(msg string, pub *rsa.PublicKey) []byte{

	secretMessage := []byte(msg)

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, secretMessage, label)
	
	if err != nil { fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err); return nil }

	//fmt.Printf("Ciphertext:\n%x\n", ciphertext)

	return ciphertext
}

func Decrypt(ciphertext []byte, prv *rsa.PrivateKey) string {
	
	plainText, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, prv, ciphertext, label)

	if err != nil { fmt.Println(err); os.Exit(1) }

	//fmt.Printf("OAEP decrypted [%x] to \n[%s]\n", ciphertext, plainText)

	return string(plainText)
}

func KeyGen() (*rsa.PublicKey, *rsa.PrivateKey){
	prv_key,_ = rsa.GenerateKey(rand.Reader, 2048)
	pub_key = &prv_key.PublicKey
	label = []byte("orders")
	return pub_key, prv_key
}

func GetKeys() (*rsa.PublicKey, *rsa.PrivateKey){
	return pub_key, prv_key
}