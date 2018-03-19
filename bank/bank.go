package bank

import (
	"crypto/rsa"
	"fmt"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/records"
	"crypto/sha256"
)

const BANK_PREFIX = "[BANK]"
var slient = true

type Bank struct {
	sk *rsa.PrivateKey
	txnBuffer []blockChain.Txn
	chain *blockChain.BlockChain
}

func print(str ...interface{}) {
	if slient {
		return
	}
	fmt.Print(BANK_PREFIX+" ")
	fmt.Println(str...)
}

func InitBank(sk *rsa.PrivateKey, chain *blockChain.BlockChain) *Bank {
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
func (b *Bank) AddTxn(txn blockChain.Txn) bool {

	// first check if there are duplicate txn in buffer.
	for _, t := range b.txnBuffer {
		h := sha256.Sum256(t.GetContent())
		h2 := sha256.Sum256(txn.GetContent())
		if string(h[:]) == string(h2[:]) {
			print("duplicate txn, ignore")
			return false
		}
	}

	ok := b.validateTxn(txn)
	if !ok {
		print("Invalid Cheating Txn, discard it")
		return false
	}
	b.txnBuffer = append(b.txnBuffer, txn)
	print("Txn added, current buffer load:", float32(len(b.txnBuffer)) / blockChain.MAXNUMTXN)

	for len(b.txnBuffer) >= blockChain.MAXNUMTXN {
		b.generateNewBlock()
	}

	return false
}

/*
	Check if the Txn is valid by checking the sigs against the claims banks.
 */
func (bank *Bank) validateTxn(txn blockChain.Txn) bool {
	verifiers := txn.GetVerifiers()
	sigs := txn.GetSigs()

	if len(verifiers) != blockChain.NUMCOSIGNER || len(verifiers) != len(sigs) / 128 {
		print("number of sigs does not match number of banks", len(verifiers), len(sigs) / 128)
		return false
	}

	bankSetWhenSigning := bank.chain.GetBankSetWhen(1234)

	// counter number of valid signer
	matchCounter := 0
	for _, s := range verifiers {
		for _, v := range bankSetWhenSigning {
			if s == v {
				matchCounter++
			}
		}
	}
	if matchCounter != len(verifiers) {
		print("Some signer was not a bank at that time", matchCounter, blockChain.NUMCOSIGNER)
		return false
	}

	// verify every signature is proper by checking against the original signed message.

	content := txn.GetContent()

	for i:=0;i<len(verifiers);i++ {
		pk := records.GetKeyByID(verifiers[i]).Pk
		expectedContent := ocrypto.EncryptBig(&pk, sigs[i * 128 : (i+1) * 128])

		if string(expectedContent) != string(content) {
			print("Wrong sig obtained from", verifiers[i])
			return false
		}

	}

	return true
}

/*
	generate a block from transaction buffer and push it to the system.
 */
func (bank *Bank) generateNewBlock() bool {
	print("Fresh Block!", len(bank.txnBuffer))
	newBlock := blockChain.NewBlock(bank.txnBuffer[:blockChain.MAXNUMTXN])
	ok := bank.chain.AddNewBlock(newBlock)
	if ok {
		bank.cleanBuffer()
		return true
	}
	return false
}

func (bank *Bank) cleanBuffer() {
	if len(bank.txnBuffer) > blockChain.MAXNUMTXN {
		bank.txnBuffer = bank.txnBuffer[blockChain.MAXNUMTXN:]
	} else {
		bank.txnBuffer = []blockChain.Txn{}
	}
	print("Txn buffer cleared")
}
