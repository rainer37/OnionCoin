package coin

import(
	"fmt"
)

const COIN_PREFIX = "[COIN]"
const COIN_LEN = 32

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

func (c *Coin) Bytes() []byte {
	b := make([]byte, COIN_LEN)
	cmsg := []byte("ThisIsNotACoin")
	for i:=0;i<len(cmsg);i++ {
		b[i] = cmsg[i]
	}
	return b
}