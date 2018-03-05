package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/sha256"
)

func GenTestBlockWithTwoTxn() *blockChain.Block {
	id_1 := "rainer"
	sk_1 := ocrypto.RSAKeyGen()
	pk_1 := sk_1.PublicKey
	pkHash_1 := sha256.Sum256(ocrypto.EncodePK(pk_1))

	id_2 := "Ella"
	sk_2 := ocrypto.RSAKeyGen()
	pk_2 := sk_2.PublicKey
	pkHash_2 := sha256.Sum256(ocrypto.EncodePK(pk_2))

	bk_1 := "bank1"
	bsk_1 := ocrypto.RSAKeyGen()

	bk_2 := "bank2"
	bsk_2 := ocrypto.RSAKeyGen()

	hash11 := ocrypto.BlindSign(bsk_1, pkHash_1[:])
	hash12 := ocrypto.BlindSign(bsk_2, pkHash_1[:])
	hash21 := ocrypto.BlindSign(bsk_1, pkHash_2[:])
	hash22 := ocrypto.BlindSign(bsk_2, pkHash_2[:])

	verifiers := []string{bk_1, bk_2}

	txn1 := blockChain.NewPKRTxn(id_1, pk_1, append(hash11, hash12...), verifiers)
	txn2 := blockChain.NewPKRTxn(id_2, pk_2, append(hash21, hash22...), verifiers)

	txns := []blockChain.Txn{txn1, txn2}
	block := blockChain.NewBlock(txns)
	return block
}

func TestBlockGen(t *testing.T) {
	// chain := blockChain.InitBlockChain()

	block := GenTestBlockWithTwoTxn()

	if block.NumTxn != 2 {
		t.Error("wrong number of txns")
	}
}