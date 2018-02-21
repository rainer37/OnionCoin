package bank

import (
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
)
const BANK_PREFIX = "[BANK]"

type Bank struct {
	sk *rsa.PrivateKey
}

func print(str ...interface{}) {
	fmt.Print(BANK_PREFIX+" ")
	fmt.Println(str...)
}

func InitBank(sk *rsa.PrivateKey) *Bank{
	print("i'm a bank!")
	bank := new(Bank)
	bank.sk = sk
	return bank
}

/*
	Blindly sign the RawCoin received.
 */
func (bank *Bank) SignRawCoin(coinSeg []byte) []byte {
	return ocrypto.BlindSign(bank.sk, coinSeg)
}

func (bank *Bank) VerifyCoin(c *coin.Coin) bool { return false }
func (bank *Bank) MakeCoin() {}

func GetBankIDSet() []string {
	// TODO: generate set of bank based on cur time.
	return []string{"FAKEID1339", "FAKEID1338"}
}