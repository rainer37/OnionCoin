package bank

import "fmt"

const BANK_PREFIX string = "[BANK]"

type Bank struct {

}

func print(str interface{}) {
	switch str.(type) {
	case int, uint, uint64:
		fmt.Printf("%s %d\n", BANK_PREFIX, str)
	case string:
		println(BANK_PREFIX, str.(string))
	default:

	}
}


func New_Bank() *Bank{
	print("i'm a bank!")
	bank := new(Bank)
	return bank
	}