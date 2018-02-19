package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/coin"
	"crypto/sha256"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/node"
)

func TestRawCoinToBytes(t *testing.T) {

	rwcn := coin.NewRawCoin("rainer")
	hashid := sha256.Sum256([]byte("rainer"))

	if rwcn.GetRIDHash() != hashid {
		t.Error("wrong hash of id")
	}

	bytes := rwcn.ToBytes()

	if string(bytes[:32]) != string(hashid[:]) {
		t.Error("wrong bytes format")
	}
}

func TestRawCoinBlindbyBank(t *testing.T) {
	rwcn := coin.NewRawCoin("rainer")
	bankSK := ocrypto.RSAKeyGen()

	brc, bfid := node.BlindRawCoin(rwcn, &bankSK.PublicKey)

	signedBRC := ocrypto.BlindSign(bankSK, brc)

	coin := node.UnBlindSignedRawCoin(signedBRC, bfid, &bankSK.PublicKey)

	if !ocrypto.VerifyBlindSig(&bankSK.PublicKey,rwcn.ToBytes(),coin) {
		t.Error("wrong raw coin ex")
	}

	if !node.ValidateCoinByKey(coin, "rainer", &bankSK.PublicKey) {
		t.Error("invalid coin")
	}
}
