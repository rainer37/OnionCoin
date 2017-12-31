package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"fmt"
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

	fmt.Print(o)
}

func TestDecryptOnion(t *testing.T) {
}

func TestCookOnion(t *testing.T) {
}
