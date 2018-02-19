package coin

import (
	"math/rand"
	"crypto/sha256"
	"encoding/binary"
)

type RawCoin struct {
	rid string // receiver's ID
	ridHash [32]byte
	coinNum uint64
}

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

func RecordBF(rwid string, bf []byte) {
	rawCoinBFs[rwid] = bf
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
		cn := rand.Uint64()
		if IsFreeCoinNum(cn) {
			return cn
		}
	}
}

/*
	check against the free coin list on blockchain to see if this num is free to use.
	TODO: check against blockchain
 */
func IsFreeCoinNum(coinNum uint64) bool {
	return true
}