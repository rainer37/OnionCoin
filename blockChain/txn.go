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

func (pkr *PKRegTxn) toBytes() []byte { return []byte{} }
func (pkr *PKRegTxn) getVerifiers() []string { return []string{} }

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