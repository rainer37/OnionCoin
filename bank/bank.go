package bank

import (
	"fmt"
	)
const BANK_PREFIX = "[BANK]"

type Bank struct {
	FCL *FreeCoinList
}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", BANK_PREFIX, str)
	case string:
		fmt.Println(BANK_PREFIX, str.(string))
	default:

	}
}

func InitBank() *Bank{
	print("i'm a bank!")
	bank := new(Bank)
	bank.FCL = NewFCL()
	return bank
}

func (bank *Bank) Sign() {}
func (bank *Bank) VerifyCoin() {}
func (bank *Bank) MakeCoin() {}
func (bank *Bank) GenFreeList() {}