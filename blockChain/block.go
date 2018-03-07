package blockChain

import (
	"strconv"
	"bytes"
	"crypto/sha256"
	"time"
)

const MAXNUMTXN = 2

type Block struct {
	PrevHash  []byte
	CurHash   []byte
	Depth     int64
	Ts        int64
	NumTxn    int
	Txns      []Txn
	TxnHashes [][]byte
}

func NewBlock(txns []Txn) *Block {
	b := new(Block)
	b.Txns = txns
	b.TxnHashes = make([][]byte, len(txns))
	for i, t := range b.Txns {
		h := sha256.Sum256(t.ToBytes())
		b.TxnHashes[i] = h[:]
	}
	b.NumTxn = len(txns)
	b.Depth = -1 // default to -1
	return b
}

func (b *Block) GetCurHash() []byte {
	timestamp := []byte(strconv.FormatInt(b.Ts, 10))
	content := TxnsToBytes(b.Txns)
	data := bytes.Join([][]byte{b.PrevHash, content, timestamp}, []byte{})
	hash := sha256.Sum256(data)
	return hash[:]
}

func (b *Block) GetTS() int64 { return b.Ts }
func (b *Block) GetPrevHash() []byte { return b.PrevHash }
func (b *Block) GetNumTxn() int { return b.NumTxn }
func (b *Block) GetTxn(index int) Txn { return b.Txns[index]}
func (b *Block) GetContent() []byte { return TxnsToBytes(b.Txns) }
func (b *Block) AddTxn(txn *Txn) {}

type User struct {
	name string
}


func (b *Block) String() string {
	s := "\nCurHash: " + string(b.CurHash) + "\n"
	s += "PreHash: " + string(b.PrevHash) + "\n"
	s += "TimeStamp: " + time.Unix(b.Ts, b.Ts).String() + "\n"
	s += "Number of Txns: " + strconv.Itoa(b.NumTxn) + "\n"
	return s
}

func BytesToBlock(blockBytes []byte) *Block { return nil }