package vault

import(
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
)

const VAULT_PREFIX = "[VULT]"

var debugged bool = false

type Vault struct {
	Coins map[string]*coin.Coin
}

func print(str interface{}) {

	if !debugged {
		return
	}

	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", VAULT_PREFIX, str)
	case string:
		fmt.Println(VAULT_PREFIX, str.(string))
	default:

	}
}

func (vault *Vault) Len() int {
	return len(vault.Coins)
}

func (vault *Vault) InitVault() {
	vault.Coins = make(map[string]*coin.Coin)
	print("Vault Created.")
}

func (vault *Vault) Contains(coin *coin.Coin) bool {
	if _, ok := vault.Coins[coin.Get_RID()]; ok {
		return true
	}
	return false
}

func (vault *Vault) Deposit(coin *coin.Coin) error {
	print("Depositing Coin :"+coin.Get_RID())
	if !vault.Contains(coin) {
		vault.Coins[coin.Get_RID()] = coin
		print(vault.Len())
		return nil
	}
	return fmt.Errorf("Error: %s is in the Vault", coin.Get_RID())
}

func (vault *Vault) Withdraw(id string) *coin.Coin {
	print("Withdrawing Coin :"+id)
	defer func(){
		delete(vault.Coins, id)
		print(vault.Len())
	}()
	return vault.Coins[id]
} 