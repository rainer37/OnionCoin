package blockChain

import (
	"time"
	"strconv"
	"bytes"
	"crypto/sha256"
)

const MAXNUMTXN = 10

type Block struct {
	prevHash []byte
	curHash []byte
	ts int64
	numTxn int
	txns Txns
}

func NewBlock(prevHash []byte, payload []byte) *Block {
	b := new(Block)
	b.prevHash = prevHash
	b.txns = produceTxns(payload)
	b.ts = time.Now().Unix()
	b.curHash = b.GetCurHash()
	return b
}

func (b *Block) GetCurHash() []byte {
	timestamp := []byte(strconv.FormatInt(b.ts, 10))
	data := bytes.Join([][]byte{b.prevHash, b.txns.txnToBytes(), timestamp}, []byte{})
	hash := sha256.Sum256(data)
	return hash[:]
}

func (b *Block) GetTS() int64 { return b.ts}
func (b *Block) GetPrevHash() []byte { return b.prevHash }
func (b *Block) GetNumTxn() int { return b.numTxn }
func (b *Block) GetTxn(index int) *Txn { return b.txns[index]}
func (b *Block) AddTxn(txn *Txn) {}
/*
	Store blockData to Disk.
 */
func (b *Block) Store() {}
func BytesToBlock(blockBytes []byte) *Block { return nil }