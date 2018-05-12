package coin

import(
	"os"
	"io/ioutil"
	"encoding/json"
	"sync"
	"github.com/rainer37/OnionCoin/util"
	"strconv"
	"math/rand"
)

const COINDIR = "coin/"

type Vault struct {
	Coins map[string][]*Coin
	me sync.RWMutex
	balance int64
}

func (vault *Vault) Len() int {
	return len(vault.Coins)
}

func InitVault() *Vault {
	vault := new(Vault)
	vault.Coins = make(map[string][]*Coin)
	vault.me = sync.RWMutex{}
	vault.balance = 1

	if ok, _ := util.Exists(COINDIR); !ok {
		os.Mkdir(COINDIR, 0777)
	}

	files, err := ioutil.ReadDir(COINDIR)
	util.CheckErr(err)

	for _, f := range files {
		// print("loading coin for", RID)
		coinBytes, err := ioutil.ReadFile(COINDIR + f.Name())
		util.CheckErr(err)

		ncoin := new(Coin)
		err = json.Unmarshal(coinBytes, ncoin)
		util.CheckErr(err)

		vault.insertCoin(ncoin)
	}
	return vault
}

func (vault *Vault) Contains(rid string) bool {
	if coins, ok := vault.Coins[rid]; ok {
		if len(coins) > 0 { return true }
	}
	return false
}

/*
	insert coin into vault and store it on disk.
 */
func (vault *Vault) Deposit(coin *Coin) {
	vault.insertCoin(coin)
	vault.writeCoinToDisk(coin)
}


func (vault *Vault) insertCoin(ncoin *Coin) {
	rid := ncoin.GetRID()
	_, ok := vault.Coins[rid]
	if !ok {
		vault.Coins[rid] = make([]*Coin, 1)
		vault.Coins[rid][0] = ncoin
	} else {
		vault.Coins[rid] = append(vault.Coins[rid], ncoin)
	}
	vault.balance++
}

func (vault *Vault) writeCoinToDisk(c *Coin) {
	e := strconv.FormatUint(c.Epoch, 10)
	r := strconv.FormatUint(rand.Uint64(), 10)
	coinPath := COINDIR + c.RID + "_" + e + "_" + r
	ioutil.WriteFile(coinPath, c.Bytes(), 0644)
}

/*
	withdraw a coin from vault, and delete the file on disk.
 */
func (vault *Vault) Withdraw(rid string) *Coin {
	if !vault.Contains(rid) { return nil }

	vault.me.Lock()
	c := vault.Coins[rid][0]
	if len(vault.Coins[rid]) > 1 {
		vault.Coins[rid] = vault.Coins[rid][1:]
	} else {
		vault.Coins[rid] = []*Coin{}
	}
	vault.me.Unlock()
	files, err := ioutil.ReadDir(COINDIR)
	util.CheckErr(err)
	for _, f := range files {
		if rid == f.Name()[:len(rid)] {
			os.Remove(COINDIR + f.Name())
			break
		}
	}
	vault.balance--
	return c
}

func (vault *Vault) GetBalance() int64 { return vault.balance }
func (vault *Vault) GetNumCoins(rid string) int {
	if !vault.Contains(rid) { return 0 }
	return len(vault.Coins[rid])
}

func (vault *Vault) String() string {
	s := ""
	for _, cs := range vault.Coins {
		for _, c := range cs {
			s += c.String() + "\n"
		}
	}
	return s
}
