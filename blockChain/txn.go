package blockChain

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"github.com/rainer37/OnionCoin/util"
	"encoding/json"
)

const (
	PK = '0'
	MSG = '1'
	UPDATE = '2'
)

type Txn interface {
	ToBytes() []byte
	GetVerifiers() []string
	GetContent() []byte
	GetSigs() []byte
	GetTS() int64
}

type TxnBase struct {
	Ts        	int64
	Sigs      	[]byte
	Verifiers 	[]string
}

/*
	PublicKey Register Transactions
 */
type PKRegTxn struct {
	Id        	string
	Pk        	[]byte
	TxnBase
}

/*
	Coin Exchange Transactions
 */
type CNEXTxn struct {
	CoinNum   	uint64
	CoinBytes 	[]byte
	TxnBase
}

/*
	Bank Coin Redeem Transactions
 */
type BCNRDMTxn struct {
	TxnID 		[]byte
	CasherID 	string
	TxnBase
}

func NewPKRTxn(id string, pk rsa.PublicKey,
	sigs []byte, vs []string) PKRegTxn {
	util.SortSigs(sigs, vs)
	return PKRegTxn{id, ocrypto.EncodePK(pk),
	TxnBase{time.Now().Unix(), sigs, vs}}
}

func NewCNEXTxn(coinNum uint64, coinBytes []byte, ts int64,
	sigs []byte, verifiers []string) CNEXTxn {
	util.SortSigs(sigs, verifiers)
	return CNEXTxn{coinNum, coinBytes,
	TxnBase{ts, sigs, verifiers}}
}

func NewBCNRDMTxn(TxnID []byte, casherID string, ts int64,
	sigs []byte, verifiers []string) BCNRDMTxn {
	util.SortSigs(sigs, verifiers)
	return BCNRDMTxn{TxnID, casherID,
	TxnBase{ts, sigs, verifiers}}
}
/*
	ID(16) | PK(132) | Ts(8) | signedHashes | SignerIDs
 */
func (pkr PKRegTxn) ToBytes() []byte {
	txnBytes, err := json.Marshal(pkr)
	util.CheckErr(err)
	return txnBytes
}

func (pkr PKRegTxn) GetVerifiers() []string { return pkr.Verifiers }
func (pkr PKRegTxn) GetSigs() []byte { return pkr.Sigs }
func (pkr PKRegTxn) GetTS() int64 { return pkr.Ts }

/*
	PK register txn content: pk + id
 */
func (pkr PKRegTxn) GetContent() []byte {
	pkHash := util.Sha(pkr.Pk)
	return append(pkHash[:], []byte(pkr.Id)...)
}

/*
	CNEX format:
	coinNum(8) : signedCoin(128) : [S0, S1, S2...]
	: [V0, V1, V2...] : [VHash0, VHash1, VHash2...]
	Si(16): coin signer i;
	Vi(16): coin Verifiers(cosigners);
	VHashi(128) : cosigned hash of the signedCoin
 */
func (cnex CNEXTxn) ToBytes() []byte {
	txnBytes, err := json.Marshal(cnex)
	util.CheckErr(err)
	return txnBytes
}

func (cnex CNEXTxn) GetCoinNum() uint64 { return cnex.CoinNum }
func (cnex CNEXTxn) GetSigs() []byte { return cnex.Sigs }
func (cnex CNEXTxn) GetVerifiers() []string { return cnex.Verifiers }
func (cnex CNEXTxn) GetTS() int64 { return cnex.Ts }

/*
	coin exchange content: coin bytes
 */
func (cnex CNEXTxn) GetContent() []byte {
	cnHash := util.Sha(cnex.CoinBytes)
	return cnHash[:]
}

func (bcnrd BCNRDMTxn) ToBytes() []byte {
	txnBytes, err := json.Marshal(bcnrd)
	util.CheckErr(err)
	return txnBytes
}

func (bcnrd BCNRDMTxn) GetSigs() []byte { return bcnrd.Sigs }
func (bcnrd BCNRDMTxn) GetVerifiers() []string { return bcnrd.Verifiers }
func (bcnrd BCNRDMTxn) GetTS() int64 { return bcnrd.Ts }
func (bcnrd BCNRDMTxn) GetContent() []byte {
	idByets := make([]byte, util.IDLEN)
	copy(idByets, bcnrd.CasherID)
	return append(bcnrd.TxnID, idByets...)
}
/*
	translate []Txn into bytes
 */
func TxnsToBytes(t []Txn) (aggre []byte) {
	for _,txn := range t {
		aggre = append(aggre, txn.ToBytes()...)
	}
	return
}

/*
	Translate bytes into one Txn
 */
func ProduceTxn(data []byte, txnType rune) Txn {
	switch txnType {
	case PK:
		txn := PKRegTxn{}
		json.Unmarshal(data, &txn)
		return txn
	case MSG:
		txn := CNEXTxn{}
		json.Unmarshal(data, &txn)
		return txn
	case UPDATE:
		txn := BCNRDMTxn{}
		json.Unmarshal(data, &txn)
		return txn
	}
	return nil
}