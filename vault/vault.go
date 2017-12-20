package vault

import(
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
)

const VAULT_PREFIX string = "[VULT]"

type Vault struct {
	Coins map[string]*coin.Coin
}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", VAULT_PREFIX, str)
	case string:
		println(VAULT_PREFIX, str.(string))
	default:

	}
}

func (vault *Vault) Len() int {
	return len(vault.Coins)
}

func (vault *Vault) Init_vault() {
	vault.Coins = make(map[string]*coin.Coin)
	print("Vault Created.")
}

func (vault *Vault) Deposit(coin *coin.Coin) {
	print("Depositing Coin :"+coin.Get_RID())
	vault.Coins[coin.Get_RID()] = coin
	print(vault.Len())
}

func (vault *Vault) Withdraw(id string) *coin.Coin {
	print("Withdrawing Coin :"+id)
	defer func(){
		delete(vault.Coins, id)
		print(vault.Len())
	}()
	return vault.Coins[id]
} 