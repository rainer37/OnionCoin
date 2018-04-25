package coin

import (
	"math/rand"
	"crypto/sha256"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/blockChain"
	"sync"
)

type RawCoin struct {
	rid string // receiver's ID
	ridHash [32]byte
	coinNum uint64
}

var met = sync.RWMutex{}
var rawCoinBFs = make(map[string][]byte)

func NewRawCoin(rid string) *RawCoin {
	rwcoin := new(RawCoin)
	rwcoin.rid = rid
	rwcoin.ridHash = sha256.Sum256([]byte(rid))
	rwcoin.coinNum = genFreeCN()
	return rwcoin
}

func (c *RawCoin) GetCoinNum() uint64 {
	return c.coinNum
}

func (c *RawCoin) GetRID() string {
	return c.rid
}

func (c *RawCoin) GetRIDHash() [32]byte {
	return c.ridHash
}

/*
	Record blind factor of the exchanging RawCoins.
 */
func RecordBF(rwid string, bf []byte) {
	met.Lock()
	rawCoinBFs[rwid] = bf
	met.Unlock()
}

func GetBF(bfid string) []byte {
	return rawCoinBFs[bfid]
}

/*
	raw coin to bytes of size 40.
 */
func (c *RawCoin) ToBytes() []byte {
	b := [8]byte{}
	binary.BigEndian.PutUint64(b[:], c.coinNum)
	bytes := append(c.ridHash[:], b[:]...)
	return bytes
}

/*
	generate a random free coin num.
 */
func genFreeCN() uint64 {
	for {
		cn := uint64(rand.Uint32())
		if blockChain.IsFreeCoinNum(cn) {
			return cn
		}
	}
}
