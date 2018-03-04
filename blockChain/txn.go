package blockChain

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"encoding/binary"
	"time"
	"strings"
)

const (
	PK = '0'
	MSG = '1'
	UPDATE = '2'
)

type Txn interface {
	ToBytes() []byte
	GetVerifiers() []string
}

/*
	PublicKey Register Transactions
 */
type PKRegTxn struct {
	Id        string
	Pk        rsa.PublicKey
	Ts        int64
	Content   []byte // containing signatures of the hash of Pk and Id.
	Verifiers []string
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

func NewPKRTxn(id string, pk rsa.PublicKey, content []byte, verifiers []string) PKRegTxn {
	return PKRegTxn{id, pk, time.Now().Unix(), content, verifiers}
}

/*
	ID | PK | Ts | signedHashes | SignerIDs
 */
func (pkr PKRegTxn) ToBytes() []byte {
	pkrBytes := []byte{}

	nextIDBytes := make([]byte, 16)

	for i:=0;i<len(pkr.Id);i++ {
		nextIDBytes[i] = pkr.Id[i]
	}

	pkrBytes = append(pkrBytes, nextIDBytes...)
	pkrBytes = append(pkrBytes, ocrypto.EncodePK(pkr.Pk)...)

	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(pkr.Ts))

	pkrBytes = append(pkrBytes, timeBytes...)
	pkrBytes = append(pkrBytes, pkr.Content...)

	for _, signer := range pkr.Verifiers {
		sidBytes := make([]byte, 16)

		for i:=0;i<len(signer);i++ {
			sidBytes[i] = signer[i]
		}
		pkrBytes = append(pkrBytes, []byte(sidBytes)...)

	}
	
	return pkrBytes
}

func (pkr PKRegTxn) GetVerifiers() []string { return pkr.Verifiers }



/*
	CNEX format: coinNum(8) : signedCoin(128) : [S0, S1, S2...] : [V0, V1, V2...] : [VHash0, VHash1, VHash2...]
	Si(16): coin signer i;
	Vi(16): coin Verifiers(cosigners);
	VHashi(128) : cosigned hash of the signedCoin
 */
func (cnex CNEXTxn) ToBytes() []byte { return []byte{} }
func (cnex CNEXTxn) GetVerifiers() []string { return cnex.verifiers }

func (bcnrd BCNRDMTxn) ToBytes() []byte { return []byte{} }
func (bcnrd BCNRDMTxn) GetVerifiers() []string { return bcnrd.verifiers }

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
		txn := new(PKRegTxn)
		txn.Id = strings.Trim(string(data[:16]), "\x00")
		txn.Pk = ocrypto.DecodePK(data[16:148])
		txn.Ts = int64(binary.BigEndian.Uint64(data[148:156]))
		txn.Content = data[156:156 + 256]
		txn.Verifiers = append(txn.Verifiers, strings.Trim(string(data[156+256:156+256+16]), "\x00"))
		txn.Verifiers = append(txn.Verifiers, strings.Trim(string(data[156+256+16:156+256+32]), "\x00"))
	case MSG:
		txn := new(CNEXTxn)
		return txn
	case UPDATE:
		txn := new(BCNRDMTxn)
		return txn
	}
	return new(PKRegTxn)
}

/*
	Translate bytes into multiple Txns.
 */
func produceTxns(data []byte) []Txn {
	return []Txn{}
}