package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/coin"
)

/*
	testing single wrap and peel
 */
func TestOnionSingleWrap(t *testing.T) {
	msg := []byte("Hello RainEr, this is you from future")

	sk := ocrypto.RSAKeyGen()
	pk := sk.PublicKey

	coin := coin.NewCoin("1", []byte("rainer"))
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
	c := coin.NewCoin("1", []byte("rainer"))
	coinByte := c.Bytes()
	nextHopID := "Ella"

	layerOne := ocrypto.WrapOnion(pk, nextHopID, coinByte, msg)

	sk2 := ocrypto.RSAKeyGen()
	pk2 := sk2.PublicKey
	coin2 := coin.NewCoin("1", []byte("tianrigu"))
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
