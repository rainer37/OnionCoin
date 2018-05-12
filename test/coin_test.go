package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/coin"
	"crypto/sha256"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/node"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/blockChain"
	"os"
)

func TestRawCoinToBytes(t *testing.T) {
	blockChain.InitBlockChain()
	defer os.RemoveAll("chainData/")
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

	brc, bfid := node.BlindBytes(rwcn.ToBytes(), &bankSK.PublicKey)

	signedBRC := ocrypto.BlindSign(bankSK, brc)

	coin := node.UnBlindBytes(signedBRC, bfid, &bankSK.PublicKey)

	if !ocrypto.VerifyBlindSig(&bankSK.PublicKey,rwcn.ToBytes(),coin) {
		t.Error("wrong raw coin ex")
	}

	if !node.ValidateCoinByKey(coin, "rainer", &bankSK.PublicKey) {
		t.Error("invalid coin")
	}
}

func TestRawCoinBlindbyBanks(t *testing.T) {
	rwcn := coin.NewRawCoin("rainer")
	bankSK := ocrypto.RSAKeyGen()
	bankSK1 := ocrypto.RSAKeyGen()

	brc, bfid := node.BlindBytes(rwcn.ToBytes(), &bankSK.PublicKey)

	signedBRC := ocrypto.BlindSign(bankSK, brc)

	coin := node.UnBlindBytes(signedBRC, bfid, &bankSK.PublicKey)

	if !ocrypto.VerifyBlindSig(&bankSK.PublicKey,rwcn.ToBytes(),coin) {
		t.Error("wrong raw coin ex")
	}

	if !node.ValidateCoinByKey(coin, "rainer", &bankSK.PublicKey) {
		t.Error("invalid coin")
	}

	nbrc, nbfid := ocrypto.Blind(&bankSK1.PublicKey, coin)

	nsignedBRC := ocrypto.BlindSign(bankSK1, nbrc)

	ncoin := ocrypto.Unblind(&bankSK1.PublicKey, nsignedBRC, nbfid)

	if !ocrypto.VerifyBlindSig(&bankSK1.PublicKey,coin,ncoin) {
		t.Error("wrong raw coin ex")
	}

	c := ocrypto.EncryptBig(&bankSK1.PublicKey, ncoin)

	cc := ocrypto.EncryptBig(&bankSK.PublicKey, c)

	if rwcn.GetCoinNum() != binary.BigEndian.Uint64(cc[32:]) {
		t.Error("Cannot decrypt second layer")
	}
}
