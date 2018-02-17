package vault

import(
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
)

const VAULTPREFIX = "[VULT]"

var debugged = false

type Vault struct {
	Coins map[string]*coin.Coin
}

func print(str ...interface{}) {
	fmt.Print(VAULTPREFIX+" ")
	fmt.Println(str...)
}

func (vault *Vault) Len() int {
	return len(vault.Coins)
}

func (vault *Vault) InitVault() {
	vault.Coins = make(map[string]*coin.Coin)
	print("Vault Created.")
}

func (vault *Vault) Contains(coin *coin.Coin) bool {
	if _, ok := vault.Coins[coin.GetRID()]; ok {
		return true
	}
	return false
}

func (vault *Vault) Deposit(coin *coin.Coin) error {
	print("Depositing Coin :"+coin.GetRID())
	if !vault.Contains(coin) {
		vault.Coins[coin.GetRID()] = coin
		//print(vault.Len())
		return nil
	}
	return fmt.Errorf("error: %s is in the vault", coin.GetRID())
}

func (vault *Vault) Withdraw(id string) *coin.Coin {
	print("Withdrawing Coin :"+id)
	defer func(){
		delete(vault.Coins, id)
		print(vault.Len())
	}()
	return vault.Coins[id]
} 