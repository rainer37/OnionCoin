package blockChain

import (
	"strconv"
	"bytes"
	"crypto/sha256"
	"time"
	"encoding/binary"
)

type Block struct {
	PrevHash  []byte
	CurHash   []byte
	Depth     int64
	Ts        int64
	NumTxn    int
	TxnHashes [][]byte
	Txns      []Txn
}

/*
	create a new block from txns, without depth, ts, and hashes set.
 */
func NewBlock(txns []Txn) *Block {
	b := new(Block)
	b.Txns = txns
	b.TxnHashes = make([][]byte, len(txns))
	for i, t := range b.Txns {
		h := sha256.Sum256(t.GetContent())
		b.TxnHashes[i] = h[:]
	}
	b.NumTxn = len(txns)
	b.Depth = -1 // default to -1
	return b
}

/*
	compute the hash of the block from {prevhash, content, and ts}.
 */
func (b *Block) GetCurHash() []byte {
	timestamp := make([]byte, 8)
	binary.BigEndian.PutUint64(timestamp, uint64(b.Ts))
	txnsBytes := TxnsToBytes(b.Txns)
	data := bytes.Join([][]byte{b.PrevHash, txnsBytes, timestamp}, []byte{})
	print(len(data))
	hash := sha256.Sum256(data)
	return hash[:]
}

func (b *Block) GetTS() int64 { return b.Ts }
func (b *Block) GetPrevHash() []byte { return b.PrevHash }
func (b *Block) GetNumTxn() int { return b.NumTxn }
func (b *Block) GetTxn(index int) Txn { return b.Txns[index]}
func (b *Block) GetContent() []byte { return TxnsToBytes(b.Txns) }

func (b *Block) String() string {
	s := "\nCurHash: " + string(b.CurHash) + "\n"
	s += "PreHash: " + string(b.PrevHash) + "\n"
	s += "TimeStamp: " + time.Unix(b.Ts, b.Ts).String() + "\n"
	s += "Depth: " + strconv.FormatInt(b.Depth, 10)
	s += "Number of Txns: " + strconv.Itoa(b.NumTxn) + "\n"
	s += "Txn Hashes: "
	for _, v := range b.TxnHashes {
		s += "["+string(v)+"]\n"
	}
	return s
}