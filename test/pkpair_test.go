package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
)

func TestPKEncrypt(t *testing.T) {
	msg := "the-key-has-to-be-32-bytes-long!"

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	cipher := ocrypto.PKEncrypt(pk, []byte(msg))
	plain := ocrypto.PKDecrypt(sk, cipher)

	if string(plain) != msg {
		t.Error("Wrong RSA Encryption Result")
	}
}

func TestAESEncryption(t *testing.T) {
	msg := []byte("hello world")
	key := []byte("the-key-has-to-be-32-bytes-long!")

	cipherText, err := ocrypto.AESEncrypt(key, msg)
	if err != nil { t.Error("Error when encryption") }
	plainText, err := ocrypto.AESDecrypt(key, cipherText)
	if err != nil { t.Error("Error when decryption") }

	if string(plainText) != string(msg) {
		t.Error("Wrong AES Encryption Result")
	}
}

func TestBlockEncryption(t *testing.T) {
	msg := []byte("Hello World")
	key := []byte("the-key-has-to-be-32-bytes-long!")

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	cipher, cKey, err := ocrypto.BlockEncrypt(msg, key, pk)
	if err != nil { t.Error("Error when block encryption") }

	plainText, err := ocrypto.BlockDecrypt(cipher, cKey, sk)
	if err != nil { t.Error("Error when block decryption") }

	if string(plainText) != string(msg) {
		t.Error("Wrong Block Encrytion Result")
	}
}