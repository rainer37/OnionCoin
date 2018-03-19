package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/blockChain"
)
func TestBlindAndUnBlind(t *testing.T) {

	msg := []byte("hello world")
	sk := ocrypto.RSAKeyGen()
	pk := &sk.PublicKey
	bsig, bfactor := ocrypto.Blind(pk, msg)
	bsign := ocrypto.BlindSign(sk, bsig)
	nmsg := ocrypto.Unblind(pk, bsign, bfactor)

	if !ocrypto.VerifyBlindSig(pk, msg, nmsg) {
		t.Error("wrong blind signing")
	}
}

func TestBlindRawCoin(t *testing.T) {
	blockChain.InitBlockChain()
	msg := coin.NewRawCoin("rainer").ToBytes()
	sk := ocrypto.RSAKeyGen()
	pk := &sk.PublicKey
	bsig, bfactor := ocrypto.Blind(pk, msg)
	bsign := ocrypto.BlindSign(sk, bsig)
	nmsg := ocrypto.Unblind(pk, bsign, bfactor)

	if !ocrypto.VerifyBlindSig(pk, msg, nmsg) {
		t.Error("wrong blind signing")
	}
}
