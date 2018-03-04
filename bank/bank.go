package bank

import (
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/blockChain"
)
const BANK_PREFIX = "[BANK]"

type Bank struct {
	sk *rsa.PrivateKey
	txnBuffer []blockChain.Txn
	chain *blockChain.BlockChain
}

func print(str ...interface{}) {
	fmt.Print(BANK_PREFIX+" ")
	fmt.Println(str...)
}

func InitBank(sk *rsa.PrivateKey, chain *blockChain.BlockChain) *Bank{
	print("i'm a bank!")
	bank := new(Bank)
	bank.sk = sk
	bank.chain = chain
	return bank
}

/*
	Blindly sign the RawCoin received.
 */
func (bank *Bank) SignRawCoin(coinSeg []byte) []byte {
	return ocrypto.BlindSign(bank.sk, coinSeg)
}

/*
	Add a transaction to the buffer
*/
func (bank *Bank) AddTxn(txn blockChain.Txn) {
	bank.txnBuffer = append(bank.txnBuffer, txn)
	print("Txn added")

	if len(bank.txnBuffer) == blockChain.MAXNUMTXN {
		bank.publishBlock()
	}
}

/*
	generate a block from transaction buffer and push it to the system.
 */
func (bank *Bank) publishBlock() {
	print("Fresh Block!", len(bank.txnBuffer))
	newBlock := blockChain.NewBlock(bank.txnBuffer)
	bank.chain.AddBlock(newBlock)
}

func (bank *Bank) VerifyCoin(c *coin.Coin) bool { return false }
func (bank *Bank) MakeCoin() {}

func GetBankIDSet() []string {
	// TODO: generate set of bank based on cur time.
	return []string{"FAKEID1339", "FAKEID1338"}
}