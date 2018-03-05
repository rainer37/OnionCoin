package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"fmt"
)

func TestPKEncrypt(t *testing.T) {
	msg := "the-key-has-to-be-32-bytes-lon!"

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	cipher := ocrypto.PKEncrypt(pk, []byte(msg))
	plain := ocrypto.PKDecrypt(sk, cipher)

	if string(plain) != msg {
		t.Error("Wrong RSA Encryption Result")
	}
}

func TestPKende(t *testing.T) {
	msg := []byte("hello World")

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	cipher := ocrypto.BlindSign(sk, msg)
	plain := ocrypto.EncryptBig(&pk, cipher)

	fmt.Println(string(plain), len(plain), len(cipher))
	if string(plain) != string(msg) {
		t.Error("Wrong pk en de")
	}
}

func TestAESEncryption(t *testing.T) {
	msg := []byte("hello world hello world")
	key := []byte("the-key-has-to-be-32-bytes-long!")

	cipherText, err := ocrypto.AESEncrypt(key, msg)
	//fmt.Println(len(cipherText))
	if err != nil { t.Error("Error when encryption") }
	plainText, err := ocrypto.AESDecrypt(key, cipherText)
	if err != nil { t.Error("Error when decryption") }

	if string(plainText) != string(msg) {
		t.Error("Wrong AES Encryption Result")
	}
}

func TestBlockEncryption(t *testing.T) {
	msg := []byte("Hello World")

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	cipher, cKey, err := ocrypto.BlockEncrypt(msg, pk)
	if err != nil { t.Error("Error when block encryption") }

	plainText, err := ocrypto.BlockDecrypt(cipher, cKey, sk)
	if err != nil { t.Error("Error when block decryption") }

	if string(plainText) != string(msg) {
		t.Error("Wrong Block Encrytion Result")
	}
}

func TestSignANDVerify(t *testing.T) {
	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	msg := []byte("Hello World")

	sig := ocrypto.RSASign(sk, msg)
	fmt.Println(len(sig))
	b := ocrypto.RSAVerify(&pk, sig, msg)
	if !b {
		t.Error("Signature verification failed")
	}
}

func TestPKEncodeDecode(t *testing.T) {
	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	b := ocrypto.EncodePK(pk)
	newPK := ocrypto.DecodePK(b)

	//fmt.Println(pk.N, pk.E)
	//fmt.Println(newPK.N, newPK.E)
	if newPK.N.Cmp(pk.N) != 0 || newPK.E != pk.E {
		t.Error("Error on pk encoding")
	}
}