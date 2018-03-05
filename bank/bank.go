package bank

import (
	"fmt"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/records"
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
	// TODO: validate the txn received, if not proofable, discards it.
	ok := bank.validateTxn(txn)
	if !ok {
		print("Invalid Cheating Txn, discard it")
		return
	}
	bank.txnBuffer = append(bank.txnBuffer, txn)
	print("Txn added")

	if len(bank.txnBuffer) == blockChain.MAXNUMTXN {
		bank.publishBlock()
	}
}

/*
	Check if the Txn is valid by checking the sigs against the claims banks.
 */
func (bank *Bank) validateTxn(txn blockChain.Txn) bool {
	verifiers := txn.GetVerifiers()
	sigs := txn.GetSigs()

	if len(verifiers) != len(sigs) / 128 {
		print("number of sigs does not match number of banks")
		return false
	}

	bankSetWhenSigning := getBankSetWhen(1234)

	matchCounter := 0
	for _, s := range verifiers {
		for _, v := range bankSetWhenSigning {
			if s == v {
				matchCounter++
			}
		}
	}
	if matchCounter != len(verifiers) {
		print("Some signer was not a bank at that time")
		return false
	}

	// verify every signature is proper by checking against the original signed message.

	content := txn.GetContent()

	for i:=0;i<len(verifiers);i++ {
		pk := records.GetKeyByID(verifiers[i]).Pk
		expectedContent := ocrypto.EncryptBig(&pk, sigs[i * 128 : (i+1) * 128])

		if string(expectedContent) != string(content) {
			print("Wrong sig by", verifiers[i])
			return false
		}

	}

	return true
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

func getBankSetWhen(t int64) []string {
	// TODO: generate set of bank based on cur time.
	return []string{"FAKEID1339", "FAKEID1338"}
}