package coin

import(
	"fmt"
	"os"
	"crypto/sha256"
	"strconv"
	"log"
	"io/ioutil"
	"time"
)

const COINPREFIX = "[COIN]"
const COINLEN = 128
const COINDIR = "coin/"

type Coin struct {
	RID string // receiver's ID
	epoch uint64
	Content []byte
}

func print(str ...interface{}) {
	fmt.Print(COINPREFIX +" ")
	fmt.Println(str...)
}

func NewCoin(rid string, content []byte) *Coin {
	coin := new(Coin)
	coin.RID = rid
	coin.Content = content
	coin.epoch = uint64(time.Now().Unix())
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
	return c.GetContent()
}

func (c *Coin) String() string {
	hash := sha256.Sum256(c.Content)
	return c.RID + " : " + string(hash[:])
}

func (c *Coin) Store() {
	e := strconv.FormatUint(c.epoch, 10)
	coinPath := COINDIR+c.RID+"_"+e
	if ok, _ := exists(coinPath); !ok {
		file, err := os.Create(coinPath)
		defer file.Close()
		checkErr(err)
	}
	ioutil.WriteFile(coinPath, c.Content, 0644)
	print("successfully save a coin on disk", coinPath)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}