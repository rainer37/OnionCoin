package coin

import (
	"math/rand"
	"encoding/binary"
	"sync"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/blockChain"
)

type RawCoin struct {
	rid string // receiver's ID
	ridHash [32]byte
	coinNum uint64
}

var met = sync.RWMutex{}
var rawCoinBFs = make(map[string][]byte) // map of blind factors

func NewRawCoin(rid string) *RawCoin {
	rwcoin := new(RawCoin)
	rwcoin.rid = rid
	rwcoin.ridHash = util.ShaHash([]byte(rid))
	rwcoin.coinNum = genFreeCN()
	return rwcoin
}

func (c *RawCoin) GetCoinNum() uint64 { return c.coinNum }
func (c *RawCoin) GetRID() string { return c.rid }
func (c *RawCoin) GetRIDHash() [32]byte { return c.ridHash }

/*
	Record blind factor of the exchanging RawCoins.
 */
func RecordBF(rwid string, bf []byte) {
	met.Lock()
	rawCoinBFs[rwid] = bf
	met.Unlock()
}

func GetBF(bfid string) []byte {
	met.Lock()
	defer met.Unlock()
	return rawCoinBFs[bfid]
}

/*
	raw coin to bytes of size 40.
 */
func (c *RawCoin) ToBytes() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, c.coinNum)
	return append(c.ridHash[:], b...)
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