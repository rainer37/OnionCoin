package blockChain

import "crypto/rsa"

type Txns []*Txn

type Txn interface {
	toBytes() []byte
	getVerifiers() []string
}

/*
	PublicKey Register Transactions
 */
type PKRegTxn struct {
	id string
	pk *rsa.PublicKey
	ts int64
	verifiers []string
}

/*
	Coin Exchange Transactions
 */
type CNEXTxn struct {
	coinNum uint64
	verifiers []string
}

/*
	Bank Coin Redeem Transactions
 */
type BCNRDMTxn struct {
	txnID []byte
	casherID string
	verifiers []string
}

func (pkr PKRegTxn) toBytes() []byte { return []byte{} }
func (pkr PKRegTxn) getVerifiers() []string { return []string{} }

/*
	CNEX format: coinNum(8) : signedCoin(128) : [S0, S1, S2...] : [V0, V1, V2...] : [VHash0, VHash1, VHash2...]
	Si(16): coin signer i;
	Vi(16): coin verifiers(cosigners);
	VHashi(128) : cosigned hash of the signedCoin
 */
func (cnex CNEXTxn) toBytes() []byte { return []byte{} }
func (cnex CNEXTxn) getVerifiers() []string { return cnex.verifiers }

func (bcnrd BCNRDMTxn) toBytes() []byte { return []byte{} }
func (bcnrd BCNRDMTxn) getVerifiers() []string { return []string{} }
/*
	translate []Txn into bytes
 */
func (t Txns) txnToBytes() []byte {
	aggre := []byte{}
	for _,txn := range t {
		aggre = append(aggre, Txn(*txn).toBytes()...)
	}
	return aggre
}

/*
	Translate bytes into []Txn
 */
func produceTxns(data []byte) Txns {
	return Txns{}
}