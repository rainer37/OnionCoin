package blockChain

type Txns []*Txn

type Txn interface {
	txn()
}

type PKRegTxn struct {}
type CNEXTxn struct {}
type BCNRDMTxn struct {}

func (pkr *PKRegTxn) txn() {}

/*
	translate []Txn into bytes
 */
func (t Txns) toBytes() []byte {
	return []byte{}
}

/*
	Translate bytes into []Txn
 */
func produceTxns(data []byte) Txns {
	return Txns{}
}