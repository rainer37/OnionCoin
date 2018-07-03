package coin

import(
	"fmt"
	"time"
	"encoding/json"
	"github.com/rainer37/OnionCoin/util"
)

const COINPREFIX = "[COIN]"

type Coin struct {
	RID     string // receiver's ID
	Epoch   uint64
	Content []byte
	Signers []string
}

func print(str ...interface{}) {
	fmt.Print(COINPREFIX +" ")
	fmt.Println(str...)
}

func NewCoin(rid string, content []byte, signers []string) *Coin {
	coin := new(Coin)
	coin.RID = rid
	coin.Content = content
	coin.Epoch = uint64(time.Now().Unix()) / util.EPOCHLEN
	coin.Signers = signers
	return coin
}

func (c *Coin) Bytes() []byte {
	coinBytes, err := json.Marshal(c)
	util.CheckErr(err)
	return coinBytes
}

func (c *Coin) String() string {
	return string(c.Bytes())
}