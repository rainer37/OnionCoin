package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
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

	oc := ocrypto.WrapOnion(key.PublicKey,nextID, coin, inner)

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
}

/*
	testing single wrap and peel
 */
func TestOnionSingleWrap(t *testing.T) {
	msg := []byte("Hello RainEr, this is you from future")

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	coin := coin.NewCoin()
	coinByte := coin.Bytes()

	nextHopID := "Ella"

	encryptedOnion := ocrypto.WrapOnion(pk, nextHopID, coinByte, msg)
	eID, eCoin, eMsg := ocrypto.PeelOnion(sk, encryptedOnion)

	if eID != nextHopID {
		t.Error("Wrong ID")
	}

	if string(eCoin) != string(coinByte) {
		t.Error("Wrong ID")
	}

	if string(eMsg) != string(eMsg) {
		t.Error("Wrong ID")
	}

}

func TestOnionTwoWrap(t *testing.T) {
	msg := []byte("Hello RainEr, this is you from future")

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey
	c := coin.NewCoin()
	coinByte := c.Bytes()
	nextHopID := "Ella"

	layerOne := ocrypto.WrapOnion(pk, nextHopID, coinByte, msg)

	sk2 := ocrypto.RSAKeyGen()
	pk2 := sk2.PublicKey
	coin2 := coin.NewCoin()
	coinByte2 := coin2.Bytes()
	nextHopID2 := "Alle"

	layerTwo := ocrypto.WrapOnion(pk2, nextHopID2, coinByte2, layerOne)

	eID, eCoin, eMsg := ocrypto.PeelOnion(sk2, layerTwo)

	if eID != nextHopID2 {
		t.Error("Wrong ID")
	}

	if string(eCoin) != string(coinByte2) {
		t.Error("Wrong Coin")
	}

	if string(layerOne) != string(eMsg) {
		t.Error("Wrong InnerOnion")
	}

	eID2, eCoin2, eMsg2 := ocrypto.PeelOnion(sk, layerOne)

	if eID2 != nextHopID {
		t.Error("Wrong ID2")
	}

	if string(eCoin2) != string(coinByte) {
		t.Error("Wrong Coin2")
	}

	if string(eMsg2) != string(msg) {
		t.Error("Wrong Msg")
	}

}
