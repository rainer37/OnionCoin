package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/coin"
	"crypto/sha256"
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
