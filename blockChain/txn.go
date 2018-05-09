package blockChain

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"crypto/sha256"
	"encoding/json"
	"github.com/rainer37/OnionCoin/util"
)

const (
	PK = '0'
	MSG = '1'
	UPDATE = '2'
)
const SIGL = 128

type Txn interface {
	ToBytes() []byte
	GetVerifiers() []string
	GetContent() []byte // getting the signing content
	GetSigs() []byte
	GetTS() int64
}

/*
	PublicKey Register Transactions
 */
type PKRegTxn struct {
	Id        	string
	Pk        	[]byte
	Ts        	int64
	Sigs      	[]byte // containing signatures of the hash of Pk and Id.
	Verifiers 	[]string
}

/*
	Coin Exchange Transactions
 */
type CNEXTxn struct {
	CoinNum   	uint64
	CoinBytes 	[]byte
	Ts        	int64
	Sigs      	[]byte
	Verifiers 	[]string
}

/*
	Bank Coin Redeem Transactions
 */
type BCNRDMTxn struct {
	TxnID 		[]byte
	CasherID 	string
	Ts        	int64
	Sigs   		[]byte // containing signatures of the hash of Pk and Id.
	Verifiers 	[]string
}

func NewPKRTxn(id string, pk rsa.PublicKey, sigs []byte, verifiers []string) PKRegTxn {
	sortSigs(sigs, verifiers)
	return PKRegTxn{id, ocrypto.EncodePK(pk), time.Now().Unix(), sigs, verifiers}
}

func NewCNEXTxn(coinNum uint64, coinBytes []byte, ts int64, sigs []byte, verifiers []string) CNEXTxn {
	sortSigs(sigs, verifiers)
	return CNEXTxn{coinNum, coinBytes, ts, sigs, verifiers}
}

func NewBCNRDMTxn(TxnID []byte, casherID string, ts int64, sigs []byte, verifiers []string) BCNRDMTxn {
	sortSigs(sigs, verifiers)
	return BCNRDMTxn{TxnID, casherID, ts, sigs, verifiers}
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
/*
	PK register txn content: pk + id
 */
func (pkr PKRegTxn) GetContent() []byte {
	pkHash := sha256.Sum256(pkr.Pk)
	return append(pkHash[:], []byte(pkr.Id)...)
}

func (pkr PKRegTxn) GetTS() int64 { return pkr.Ts }

/*
	CNEX format: coinNum(8) : signedCoin(128) : [S0, S1, S2...] : [V0, V1, V2...] : [VHash0, VHash1, VHash2...]
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
	cnHash := sha256.Sum256(cnex.CoinBytes)
	return cnHash[:]
}

func (bcnrd BCNRDMTxn) ToBytes() []byte {
	txnBytes, err := json.Marshal(bcnrd)
	util.CheckErr(err)
	return txnBytes
}

func (bcnrd BCNRDMTxn) GetSigs() []byte { return bcnrd.Sigs }
func (bcnrd BCNRDMTxn) GetVerifiers() []string { return bcnrd.Verifiers }
func (bcnrd BCNRDMTxn) GetContent() []byte {
	idByets := make([]byte, 16)
	copy(idByets, bcnrd.CasherID)
	return append(bcnrd.TxnID, idByets...)
}
func (bcnrd BCNRDMTxn) GetTS() int64 { return bcnrd.Ts }

/*
	translate []Txn into bytes
 */
func TxnsToBytes(t []Txn) []byte {
	aggre := []byte{}
	for _,txn := range t {
		aggre = append(aggre, txn.ToBytes()...)
	}
	return aggre
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

func sortSigs(sigs []byte, verifiers []string) {
	for i:=0; i<len(verifiers) - 1; i++ {
		for j:=0; j<len(verifiers) -i - 1; j++ {
			if verifiers[j] > verifiers[j+1] {
				verifiers[j+1], verifiers[j] = verifiers[j], verifiers[j+1]
				temp := make([]byte, SIGL)
				copy(temp, sigs[(j+1) * SIGL:(j+2) * SIGL])
				copy(sigs[(j+1) * SIGL:(j+2) * SIGL],sigs[j*SIGL:(j+1)*SIGL])
				copy(sigs[j*SIGL:(j+1)*SIGL], temp)
			}
		}
	}
}
