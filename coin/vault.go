package coin

import(
	"os"
)

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
	print("Vault Created.")
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
	print("Depositing Coin :"+coin.RID)
	if !vault.Contains(coin) {
		vault.Coins[coin.RID] = []*Coin{coin}
	} else {
		vault.Coins[coin.RID] = append(vault.Coins[coin.RID], coin)
	}
	coin.Store() // store the coin on disk
}

/*
	withdraw a coin from vault
 */
func (vault *Vault) Withdraw(rid string) *Coin {
	print("Withdrawing Coin :"+rid)
	if vault.Coins[rid] == nil {
		return nil
	}
	c := vault.Coins[rid][0]
	// vault.Coins[id] = vault.Coins[id][1:]
	return c
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