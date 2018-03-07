package blockChain

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"encoding/binary"
	"time"
	"strings"
	"crypto/sha256"
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
	CoinNum   uint64
	CoinBytes []byte
	Ts        int64
	Sigs      []byte
	Verifiers []string
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
	return PKRegTxn{id, ocrypto.EncodePK(pk), time.Now().Unix(), sigs, verifiers}
}

func NewCNEXTxn(coinNum uint64, coinBytes []byte, sigs []byte, verifiers []string) CNEXTxn {
	return CNEXTxn{coinNum, coinBytes, time.Now().Unix(), sigs, verifiers}
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
	pkrBytes = append(pkrBytes, pkr.Pk...)

	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(pkr.Ts))

	pkrBytes = append(pkrBytes, timeBytes...)
	pkrBytes = append(pkrBytes, pkr.Sigs...)

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
func (pkr PKRegTxn) GetSigs() []byte { return pkr.Sigs }
func (pkr PKRegTxn) GetContent() []byte {
	pkHash := sha256.Sum256(pkr.Pk)
	return append(pkHash[:], []byte(pkr.Id)...)
}

/*
	CNEX format: coinNum(8) : signedCoin(128) : [S0, S1, S2...] : [V0, V1, V2...] : [VHash0, VHash1, VHash2...]
	Si(16): coin signer i;
	Vi(16): coin Verifiers(cosigners);
	VHashi(128) : cosigned hash of the signedCoin
 */
func (cnex CNEXTxn) ToBytes() []byte { return []byte{} }
func (cnex CNEXTxn) GetCoinNum() uint64 { return cnex.CoinNum }
func (cnex CNEXTxn) GetSigs() []byte { return cnex.Sigs }
func (cnex CNEXTxn) GetVerifiers() []string { return cnex.Verifiers }
func (cnex CNEXTxn) GetContent() []byte {
	cnHash := sha256.Sum256(cnex.CoinBytes)
	return cnHash[:]
}

func (bcnrd BCNRDMTxn) ToBytes() []byte { return []byte{} }
func (bcnrd BCNRDMTxn) GetSigs() []byte { return bcnrd.Sigs }
func (bcnrd BCNRDMTxn) GetVerifiers() []string { return bcnrd.Verifiers }
func (bcnrd BCNRDMTxn) GetContent() []byte { return []byte{} }

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
		txn.Id = strings.Trim(string(data[:16]), "\x00")
		txn.Pk = data[16:148]
		txn.Ts = int64(binary.BigEndian.Uint64(data[148:156]))
		txn.Sigs = data[156:156 + 256]
		txn.Verifiers = append(txn.Verifiers, strings.Trim(string(data[156+256:156+256+16]), "\x00"))
		txn.Verifiers = append(txn.Verifiers, strings.Trim(string(data[156+256+16:156+256+32]), "\x00"))
		return txn
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