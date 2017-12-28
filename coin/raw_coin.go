package coin

type RawCoin struct {
	RID string // receiver's ID
	Content []byte
}

func NewRawCoin() *RawCoin {
	coin := new(RawCoin)
	coin.RID = "1338"
	coin.Content = []byte("hello world")
	print("Coin is an Onion : " + string(coin.RID))
	return coin
}

func (c *RawCoin) GetContent() []byte {
	return c.Content
}

func (c *RawCoin) GetRID() string {
	return c.RID
}