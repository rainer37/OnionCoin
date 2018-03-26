package coin

import(
	"os"
	"io/ioutil"
	"strings"
)

var balance = 1

type Vault struct {
	Coins map[string][]*Coin
}

func (vault *Vault) Len() int {
	return len(vault.Coins)
}

func (vault *Vault) InitVault() {
	vault.Coins = make(map[string][]*Coin)
	if ok, _ := exists("coin"); !ok {
		os.Mkdir(COINDIR, 0777)
	}
	files, err := ioutil.ReadDir(COINDIR)
	checkErr(err)
	for _, f := range files {
		rid := strings.Split(f.Name(), "_")[0]
		// print("loading coin for", rid)
		coinBytes, err := ioutil.ReadFile(COINDIR+f.Name())
		checkErr(err)
		coins, ok := vault.Coins[rid]
		ncoin := NewCoin(rid,coinBytes)
		if !ok {
			vault.Coins[rid] = []*Coin{ncoin}
		} else {
			coins = append(coins, ncoin)
		}
	}
}

func (vault *Vault) Contains(coin *Coin) bool {
	if coins, ok := vault.Coins[coin.RID]; ok {
		if len(coins) > 0 {
			return true
		}
	}
	return false
}

func (vault *Vault) Deposit(coin *Coin) {
	if !vault.Contains(coin) {
		vault.Coins[coin.RID] = []*Coin{coin}
	} else {
		vault.Coins[coin.RID] = append(vault.Coins[coin.RID], coin)
	}
	coin.Store() // store the coin on disk
	// balance++
}

/*
	withdraw a coin from vault
 */
func (vault *Vault) Withdraw(rid string) *Coin {
	// print("Withdrawing Coin :"+rid)
	if vault.Coins[rid] == nil {
		files, err := ioutil.ReadDir(COINDIR)
		checkErr(err)
		for _, f := range files {
			if rid == f.Name()[:len(rid)] {
				coinData, err := ioutil.ReadFile(f.Name())
				checkErr(err)
				ncoin := NewCoin(rid, coinData)
				vault.Coins[rid] = []*Coin{ncoin}
				return ncoin
			}
		}
		// print("No coin for this dude", rid)
		return nil
	}
	c := vault.Coins[rid][0]
	//if len(vault.Coins[rid]) > 1 {
	//	vault.Coins[rid] = vault.Coins[rid][1:]
	//} else {
	//	vault.Coins[rid] = vault.Coins[rid][:0]
	//}
	// balance--
	return c
}

func GetBalance() int {
	return balance
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
