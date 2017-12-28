package coin

import(
	"fmt"
)

const COIN_PREFIX = "[COIN]"

type Coin struct {
	RID string // receiver's ID
	Content []byte
}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", COIN_PREFIX, str)
	case string:
		fmt.Println(COIN_PREFIX, str.(string))
	default:

	}
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