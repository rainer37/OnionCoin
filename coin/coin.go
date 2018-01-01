package coin

import(
	"fmt"
)

const COIN_PREFIX = "[COIN]"

type Coin struct {
	RID string // receiver's ID
	Content []byte
}

func print(str ...interface{}) {
	fmt.Print(COIN_PREFIX+" ")
	fmt.Println(str...)
}

func NewCoin() *Coin {
	coin := new(Coin)
	coin.RID = "1338"
	coin.Content = []byte("hello world")
	print("Coin is an Onion : " + string(coin.RID))
	return coin
}

func (c *Coin) GetContent() []byte {
	return c.Content
}

func (c *Coin) GetRID() string {
	return c.RID
}

func (c *Coin) toByte() {}