package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/blockChain"
	"crypto/sha256"
	"strings"
	"encoding/binary"
)


func TestAddBlockToChain(t *testing.T) {
	chain := blockChain.InitBlockChain()

	size := chain.Size()

	block := GenTestBlockWithTwoTxn()

	gblock := chain.GetLastBlock()

	chain.AddNewBlock(block)

	if chain.Size() != size + 1 {
		t.Error("Wrong size")
	}

	depth := chain.Size()
	lb := chain.GetLastBlock()

	if lb.Depth != depth - 1 {
		t.Error("Wrong Depth")
	}

	if strings.Trim(string(lb.PrevHash), "\x00") != string(gblock.CurHash) {
		t.Error("Wrong prevHash")
	}

	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(lb.Ts))

	expectedHashes := append(lb.PrevHash, lb.GetContent()...)
	expectedHashes = append(expectedHashes, timestamp...)
	expectedSHA := sha256.Sum256(expectedHashes)

	if string(expectedSHA[:]) != string(lb.GetCurHash()) {
		t.Error("wrong current hash")
	}

}

func TestGetBlock(t *testing.T) {
	chain := blockChain.InitBlockChain()
	size := chain.Size()
	block_1 := GenTestBlockWithTwoTxn()
	block_2 := GenTestBlockWithTwoTxn()
	block_3 := GenTestBlockWithTwoTxn()

	chain.AddNewBlock(block_1)
	chain.AddNewBlock(block_2)
	chain.AddNewBlock(block_3)

	// fmt.Println(chain.Size())
	if chain.Size() != size + 3 {
		t.Error("Wrong Size")
	}

	b_1 := chain.GetBlock(1)
	b_2 := chain.GetBlock(2)
	b_3 := chain.GetBlock(3)

	phash1 := b_1.PrevHash
	phash2 := b_2.PrevHash
	phash3 := b_3.PrevHash

	if strings.Trim(string(phash1), "\x00") != "_OC_GENESIS_HASH_ON_18_MAR_2018_" {
		t.Error("Wrong prevHash")
	}

	if string(phash2) != string(b_1.CurHash) {
		t.Error("Wrong phash2")
	}

	if string(phash3) != string(b_2.CurHash) {
		t.Error("Wrong phash3")
	}

}