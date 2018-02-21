package coin

import(
	"fmt"
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
		return nil
	}
	return fmt.Errorf("error: %s is in the vault", coin.GetRID())
}

func (vault *Vault) Withdraw(id string) *Coin {
	print("Withdrawing Coin :"+id)
	defer func(){
		delete(vault.Coins, id)
		print(vault.Len())
	}()
	return vault.Coins[id]
} 