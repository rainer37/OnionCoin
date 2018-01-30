package bank

import (
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
)
const BANK_PREFIX = "[BANK]"

type Bank struct {
	FCL *FreeCoinList
}

func print(str ...interface{}) {
	fmt.Print(BANK_PREFIX+" ")
	fmt.Println(str...)
}

func InitBank() *Bank{
	print("i'm a bank!")
	bank := new(Bank)
	bank.FCL = NewFCL()
	return bank
}

func (bank *Bank) Sign(coinSeg []byte) []byte { return nil }
func (bank *Bank) VerifyCoin(c *coin.Coin) bool { return false }
func (bank *Bank) MakeCoin() {}
func (bank *Bank) GenFreeList() {}
func (bank *Bank) send() {}
func (bank *Bank) receive() {}