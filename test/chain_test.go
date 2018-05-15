package test

import (
	"testing"
	"github.com/rainer37/OnionCoin/blockChain"
	"crypto/sha256"
	"strings"
	"encoding/binary"
	"bytes"
	"os"
)


func TestAddBlockToChain(t *testing.T) {
	chain := blockChain.InitBlockChain()
	defer os.RemoveAll("chainData/")

	size := chain.Size()
	block := GenTestBlockWithTwoTxn(chain)
	gblock := chain.GetLastBlock()

	chain.AddNewBlock(block)

	if chain.Size() != size + 1 {
		t.Error("Wrong size")
	}

	depth := chain.Size()
	lb := chain.GetLastBlock()

	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(lb.Ts))

	txnsBytes := make([]byte, 32)
	for _, h := range lb.TxnHashes {
		for i, c := range h {
			txnsBytes[i] += c
		}
	}

	data := bytes.Join([][]byte{lb.PrevHash, txnsBytes, timestamp}, []byte{})
	expectedSHA := sha256.Sum256(data)

	if string(expectedSHA[:]) != string(lb.CurHash) {
		t.Error("wrong current hash")
	}

	if lb.Depth != depth - 1 {
		t.Error("Wrong Depth")
	}

	if strings.Trim(string(lb.PrevHash), "\x00") != string(gblock.CurHash) {
		t.Error("Wrong prevHash")
	}

	if blockChain.TI.PKIndex["rainer"] != 1 || blockChain.TI.PKIndex["Ella"] != 1{
		t.Error("Wrong Index Update")
	}
}

func TestGetBlock(t *testing.T) {
	chain := blockChain.InitBlockChain()
	defer os.RemoveAll("chainData/")

	size := chain.Size()

	block_1 := GenTestBlockWithTwoTxn(chain)
	chain.AddNewBlock(block_1)
	block_2 := GenTestBlockWithTwoTxn(chain)
	chain.AddNewBlock(block_2)
	block_3 := GenTestBlockWithTwoTxn(chain)
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

	if strings.Trim(string(phash1), "\x00") != blockChain.GENSIS_HASH {
		t.Error("Wrong prevHash")
	}

	if string(phash2) != string(b_1.CurHash) {
		t.Error("Wrong phash2")
	}

	if string(phash3) != string(b_2.CurHash) {
		t.Error("Wrong phash3")
	}

	b_last := chain.GetLastBlock()

	if string(b_last.CurHash) != string(b_3.CurHash) {
		t.Error("wrong getting last block")
	}

	if blockChain.TI.PKIndex["rainer"] != 3 || blockChain.TI.PKIndex["Ella"] != 3 {
		t.Error("Wrong Index Update")
	}
}