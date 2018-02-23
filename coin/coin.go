package coin

import(
	"fmt"
	"os"
)

const COIN_PREFIX = "[COIN]"
const COIN_LEN = 32
const COINDIR = "coin"

type Coin struct {
	RID string // receiver's ID
	Content []byte
}

func print(str ...interface{}) {
	fmt.Print(COIN_PREFIX+" ")
	fmt.Println(str...)
}

func NewCoin(rid string, content []byte) *Coin {
	coin := new(Coin)
	coin.RID = rid
	coin.Content = content
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

func (c *Coin) String() string {
	return c.RID + " : " + string(c.Content)
}

func (c *Coin) Store() {
	if ok, _ := exists(c.RID); !ok {
		os.Mkdir(COINDIR, 0777)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}