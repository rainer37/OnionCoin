package vault

import(
	"github.com/rainer37/OnionCoin/coin"
)

const VAULT_PREFIX string = "[VULT]"

type Coin coin.Coin

type Vault struct {
	Coins []Coin
}

func print(str string) {
	println(VAULT_PREFIX, str)
}

func (vault Vault) Init_vault() {
	vault.Coins = make([]Coin, 10)
	print("Vault Created.")
}