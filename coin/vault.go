package coin

import(
	"fmt"
	"os"
)

var debugged = false

type Vault struct {
	Coins map[string]*Coin
}

func (vault *Vault) Len() int {
	return len(vault.Coins)
}

func (vault *Vault) InitVault() {
	vault.Coins = make(map[string]*Coin)
	if ok, _ := exists("coin"); !ok {
		os.Mkdir(COINDIR, 0777)
	}
	print("Vault Created.")
}

func (vault *Vault) Contains(coin *Coin) bool {
	if _, ok := vault.Coins[coin.GetRID()]; ok {
		return true
	}
	return false
}

func (vault *Vault) Deposit(coin *Coin) error {
	print("Depositing Coin :"+coin.GetRID())
	if !vault.Contains(coin) {
		vault.Coins[coin.GetRID()] = coin
		coin.Store() // store the coin on disk
		return nil
	}
	return fmt.Errorf("error: %s is in the vault", coin.GetRID())
}

func (vault *Vault) Withdraw(id string) *Coin {
	print("Withdrawing Coin :"+id)
	defer func(){
		delete(vault.Coins, id)
	}()
	return vault.Coins[id]
} 