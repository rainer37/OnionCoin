package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"crypto/sha256"
	"github.com/rainer37/OnionCoin/blockChain"
	"fmt"
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

	fmt.Println(len(encodedPK), len(s1SigOnHash), len(s2SigOnHash))

	content := append(s1SigOnHash, s2SigOnHash...)

	ptxn := blockChain.NewPKRTxn(newBieID, newBiePK, content, signers)
	txnBytes := ptxn.ToBytes()
	fmt.Println(len(txnBytes))

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
}