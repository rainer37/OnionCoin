package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
)

func TestPKEncrypt(t *testing.T) {
	msg := "hello world"
	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	cipher := ocrypto.PKEncrypt(pk, []byte(msg))
	plain := ocrypto.PKDecrypt(sk, cipher)

	if string(plain) != msg {
		t.Error("Wrong Encryption")
	}
}