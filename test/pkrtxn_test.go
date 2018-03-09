package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"crypto/sha256"
	"github.com/rainer37/OnionCoin/blockChain"
	"strings"
	"encoding/binary"
)

func TestPKRTxnToBytes(t *testing.T) {
	newBieID := "Rainer"
	newBiePK := ocrypto.RSAKeyGen().PublicKey
	ts := time.Now().Unix()
	signers := []string{"FAKEID1338", "FAKEID1339"}

	s1SK := ocrypto.RSAKeyGen()
	s2SK := ocrypto.RSAKeyGen()

	encodedPK := ocrypto.EncodePK(newBiePK)
	encodedPKHash := sha256.Sum256(encodedPK)

	s1SigOnHash := ocrypto.BlindSign(s1SK, append(encodedPKHash[:], []byte(newBieID)...))
	s2SigOnHash := ocrypto.BlindSign(s2SK, append(encodedPKHash[:], []byte(newBieID)...))

	content := append(s1SigOnHash, s2SigOnHash...)

	ptxn := blockChain.NewPKRTxn(newBieID, newBiePK, content, signers)
	txnBytes := ptxn.ToBytes()

	if strings.Trim(string(txnBytes[:16]), "\x00") != newBieID {
		t.Error("ID not equal")
	}

	if string(ocrypto.EncodePK(newBiePK)) != string(txnBytes[16:148]) {
		t.Error("PK not equal")
	}

	if binary.BigEndian.Uint64(txnBytes[148:156]) != uint64(ts) {
		t.Error("Ts not equal")
	}

	if string(s1SigOnHash) != string(txnBytes[156:156+128]) {
		t.Error("first signed hash not equal")
	}

	if string(s2SigOnHash) != string(txnBytes[156+128:156+256]) {
		t.Error("second signed hash not equal")
	}

	if strings.Trim(string(txnBytes[156+256:156+256+16]), "\x00") != "FAKEID1338" {
		t.Error("first signer id not equal")
	}

	if strings.Trim(string(txnBytes[156+256+16:156+256+32]), "\x00") != "FAKEID1339" {
		t.Error("second signer id not equal")
	}

	ntxn := blockChain.ProduceTxn(txnBytes, '0').(blockChain.PKRegTxn)

	if ntxn.Id != newBieID {
		t.Error("ID not equal")
	}

	if string(ntxn.Pk) != string(encodedPK) {
		t.Error("PK not equal")
	}

	if ntxn.Ts != ts {
		t.Error("Ts not equal")
	}

	if string(s1SigOnHash) != string(ntxn.Sigs[:128]) {
		t.Error("first signed hash not equal")
	}

	if string(s2SigOnHash) != string(ntxn.Sigs[128:]) {
		t.Error("second signed hash not equal")
	}

	if ntxn.Verifiers[0] != "FAKEID1338" {
		t.Error("first signer id not equal")
	}

	if ntxn.Verifiers[1] != "FAKEID1339" {
		t.Error("second signer id not equal")
	}

}

func TestCNEXTxnToBytes(t *testing.T) {
	newBieID := "Rainer"

	ts := time.Now().Unix()
	signers := []string{"FAKEID1338", "FAKEID1339"}

	s1SK := ocrypto.RSAKeyGen()
	s2SK := ocrypto.RSAKeyGen()

	cBytes := make([]byte, 128)

	coinHash := sha256.Sum256(cBytes)

	s1SigOnHash := ocrypto.BlindSign(s1SK, append(coinHash[:], []byte(newBieID)...))
	s2SigOnHash := ocrypto.BlindSign(s2SK, append(coinHash[:], []byte(newBieID)...))

	content := append(s1SigOnHash, s2SigOnHash...)

	ctxn := blockChain.NewCNEXTxn(1234567, cBytes, content, signers)
	txnBytes := ctxn.ToBytes()

	if binary.BigEndian.Uint64(txnBytes[:8]) != 1234567 {
		t.Error("coinNum not equal")
	}

	if string(cBytes) != string(txnBytes[8:136]) {
		t.Error("coinBytes not equal")
	}

	if binary.BigEndian.Uint64(txnBytes[136:144]) != uint64(ts) {
		t.Error("Ts not equal")
	}

	if string(s1SigOnHash) != string(txnBytes[144:144+128]) {
		t.Error("first signed hash not equal")
	}

	if string(s2SigOnHash) != string(txnBytes[144+128:144+256]) {
		t.Error("second signed hash not equal")
	}

	if strings.Trim(string(txnBytes[144+256:144+256+16]), "\x00") != "FAKEID1338" {
		t.Error("first signer id not equal")
	}

	if strings.Trim(string(txnBytes[144+256+16:144+256+32]), "\x00") != "FAKEID1339" {
		t.Error("second signer id not equal")
	}

	ntxn := blockChain.ProduceTxn(txnBytes, '1').(blockChain.CNEXTxn)

	if ntxn.CoinNum != 1234567 {
		t.Error("coinNum not equal")
	}

	if string(ntxn.CoinBytes) != string(cBytes) {
		t.Error("coinBytes not equal")
	}

	if ntxn.Ts != ts {
		t.Error("Ts not equal")
	}

	if string(s1SigOnHash) != string(ntxn.Sigs[:128]) {
		t.Error("first signed hash not equal")
	}

	if string(s2SigOnHash) != string(ntxn.Sigs[128:]) {
		t.Error("second signed hash not equal")
	}

	if ntxn.Verifiers[0] != "FAKEID1338" {
		t.Error("first signer id not equal")
	}

	if ntxn.Verifiers[1] != "FAKEID1339" {
		t.Error("second signer id not equal")
	}
}