package coin

import(
	"fmt"
	"os"
	"strconv"
	"io/ioutil"
	"time"
	"encoding/json"
	"github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/util"
)

const COINPREFIX = "[COIN]"
const COINLEN = 128
const COINDIR = "coin/"

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
	coin.Epoch = uint64(time.Now().Unix()) / blockChain.EPOCHLEN
	coin.Signers = signers
	return coin
}

func (c *Coin) GetContent() []byte {
	b := make([]byte, COINLEN)
	copy(b, c.Content)
	return b
}

func (c *Coin) GetRID() string {
	return c.RID
}

func (c *Coin) Bytes() []byte {
	coinBytes, err :=json.Marshal(c)
	util.CheckErr(err)
	return coinBytes
}

func (c *Coin) String() string {
	return string(c.Bytes())
}

func (c *Coin) Store() {
	e := strconv.FormatUint(c.Epoch, 10)
	coinPath := COINDIR+c.RID+"_"+e
	if ok, _ := util.Exists(coinPath); !ok {
		file, err := os.Create(coinPath)
		defer file.Close()
		util.CheckErr(err)
	}
	ioutil.WriteFile(coinPath, c.Bytes(), 0644)
	// print("successfully save a coin on disk", coinPath)
}
