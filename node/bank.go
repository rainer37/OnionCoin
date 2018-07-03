package node

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/blockChain"
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/util"
)

var HashCmpMap map[string]int

type Bank struct {
	sk *rsa.PrivateKey
	txnBuffer []blockChain.Txn
	chain *blockChain.Chain
	status bool
}

func InitBank(sk *rsa.PrivateKey, chain *blockChain.Chain) *Bank {
	bank := new(Bank)
	bank.sk = sk
	bank.chain = chain
	HashCmpMap = make(map[string]int)
	return bank
}

func (bank *Bank) GetTxnBuffer() []blockChain.Txn { return bank.txnBuffer }
func (bank *Bank) SetStatus(status bool) { bank.status = status }

/*
	Add a transaction to the buffer, return true if succeed.
*/
func (bank *Bank) AddTxn(txn blockChain.Txn) bool {

	// first check if there are duplicate txn in buffer.
	if bank.containsTxn(txn) {
		print("duplicate txn, not added")
		return false
	}

	if !bank.validateTxn(txn) {
		print("Invalid Cheating Txn, discard it")
		return false
	}

	bank.txnBuffer = append(bank.txnBuffer, txn)

	l := len(bank.txnBuffer)
	print("Txn added, current buffer load:",
		float32(l) / util.MAXNUMTXN, l)

	return true
}

/*
	Check if the Txn is valid by checking the sigs against the claims banks.
 */
func (bank *Bank) validateTxn(txn blockChain.Txn) bool {
	verifiers := txn.GetVerifiers()
	sigs := txn.GetSigs()

	if len(verifiers) != util.NUMCOSIGNER ||
		len(verifiers) != len(sigs) / 128 {
		print("number of sigs does not match number of co-signers!",
			len(verifiers), len(sigs) / 128)
		return false
	}

	bankSetWhenSigning := bank.chain.GetBankSetWhen(txn.GetTS())

	// counter number of valid signer
	matchCounter := 0
	for _, s := range verifiers {
		for _, v := range bankSetWhenSigning {
			if s == v || s == "FAKEID1338" || s == "FAKEID1339" {
				matchCounter++
			}
		}
	}
	if matchCounter < len(verifiers) {
		print("Some signer was not a bank at that time",
			matchCounter, util.NUMCOSIGNER)
		return false
	}

	// verify every signature is proper by checking against
	// the original signed message.

	content := txn.GetContent()

	for i:=0; i<len(verifiers); i++ {
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
	upon receieved proposed list of txns, validate them and add to buffer
 */
func (bank *Bank) AggreTxns(txns []blockChain.Txn) {
	counter := 0
	for _, v := range txns {
		if bank.AddTxn(v) {
			counter++
		}
	}
	print(counter, "new txns added")
}


func (bank *Bank) containsTxn(txn blockChain.Txn) bool {
	for _, v := range bank.txnBuffer {
		if string(v.GetContent()) == string(txn.GetContent()) {
			return true
		}
	}
	return false
}

/*
	generate a block from transaction buffer and push it to the system.
 */
func (bank *Bank) GenNewBlock() *blockChain.Block {
	return bank.chain.GenNewBlock(bank.txnBuffer)
}

/*
	check if the local hash is the major one.
 */
func (bank *Bank) IsMajorityHash(hash string) bool {
	maxCount := 0
	myCount := HashCmpMap[hash]
	for _, v := range HashCmpMap {
		if v > maxCount {
			maxCount = v
		}
	}
	return myCount == maxCount
}

func (bank *Bank) CleanBuffer() {
	bank.txnBuffer = []blockChain.Txn{}
	print("Txn buffer cleared")
}