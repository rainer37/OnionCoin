package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"fmt"
)


func TestOnionToOMsg(t *testing.T) {
	key := ocrypto.RSAKeyGen()
	msg := []byte("the-key-has-to-be-32-bytes-long!")
	cipher := ocrypto.PKEncrypt(key.PublicKey, msg)
	fmt.Println(len(cipher))
	symcipher, _ := ocrypto.AESEncrypt(msg, msg)
	fmt.Println(len(symcipher))
}

func TestFormatOnion(t *testing.T) {
	b := []byte{0,0,0,3,0,0,0,4,0,0,0,0,0,0,0,7,'w','h','o',0,0,0,9,0,0,0,0,0,0,0xE,1,2,3}
	o := ocrypto.FormatOnion(b)

	if o.NextID != "who" {
		t.Error("nextID")
	}

	if string(o.Coin) != string([]byte{0,0,0,9}) {
		t.Error("coin")
	}

	if string(o.InnerOnion) != string([]byte{0,0,0,0,0,0,0xE}) {
		t.Error("inner")
	}

	if string(o.Chaos) != string([]byte{1,2,3}) {
		t.Error("chaos")
	}
}

func TestPeelOnion(t *testing.T) {
}

func TestWrapOnion(t *testing.T) {
	nextID := "rainer"
	coin := []byte{1,2,3,4}
	inner := []byte{6,7,8,9,0}

	sym_key := []byte("the-key-has-to-be-32-bytes-long!")

	key := ocrypto.RSAKeyGen()

	oc := ocrypto.WrapOnion(key.PublicKey, sym_key, nextID, coin, inner)

	o := ocrypto.DecryptOnion(key, ocrypto.PKEncrypt(key.PublicKey, sym_key), oc[256:])

	if len(o) != 36 {
		t.Error("Crying")
	}

	if string(o[:4]) != string([]byte{0,0,0,6}) {
		t.Error("Crying again on nlen")
	}

	if string(o[4:8]) != string([]byte{0,0,0,4}) {
		t.Error("Crying again on clen")
	}

	if string(o[8:16]) != string([]byte{0,0,0,0,0,0,0,5}) {
		t.Error("Crying again on ilen")
	}

	if string(o[16:22]) != string([]byte{'r','a','i','n','e','r'}) {
		t.Error("Crying again again on nextID")
	}

	if string(o[22:26]) != string([]byte{1,2,3,4}) {
		t.Error("Crying again again on coin")
	}

	if string(o[26:31]) != string([]byte{6,7,8,9,0}) {
		t.Error("Crying again again on inner")
	}

	if string(o[31:]) != "chaos" {
		t.Error("Crying again again on chaos")
	}

	op := ocrypto.PeelOnion(key, sym_key, oc)

	if string(op.Chaos) != "chaos" {
		t.Error("Peel Chaos")
	}

	if string(op.NextID) != "rainer" {
		t.Error("Peel nextID")
	}

	if string(op.Coin) != string([]byte{1,2,3,4}) {
		t.Error("Peel Coin")
	}
}