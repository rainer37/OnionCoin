package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
)

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

	key := ocrypto.RSAKeyGen()

	oc := ocrypto.WrapOnion(key.PublicKey, nextID, coin, inner)

	o := ocrypto.DecryptOnion(key, oc)

	//fmt.Println(o)

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
}