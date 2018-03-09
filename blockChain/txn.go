package blockChain

import (
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"encoding/binary"
	"time"
	"strings"
	"crypto/sha256"
	"bytes"
)

const (
	PK = '0'
	MSG = '1'
	UPDATE = '2'
)

type Txn interface {
	ToBytes() []byte
	GetVerifiers() []string
	GetContent() []byte // getting the signing content
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
	ID(16) | PK(132) | Ts(8) | signedHashes | SignerIDs
 */
func (pkr PKRegTxn) ToBytes() []byte {
	pkrBytes := []byte{}

	nextIDBytes := make([]byte, 16)
	copy(nextIDBytes, pkr.Id)

	pkrBytes = append(pkrBytes, nextIDBytes...)
	pkrBytes = append(pkrBytes, pkr.Pk...)

	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(pkr.Ts))

	pkrBytes = append(pkrBytes, timeBytes...)
	pkrBytes = append(pkrBytes, pkr.Sigs...)

	for _, signer := range pkr.Verifiers {
		sidBytes := make([]byte, 16)
		copy(sidBytes, signer)
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
func (cnex CNEXTxn) ToBytes() []byte {
	coinNumBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(coinNumBytes, cnex.CoinNum)

	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(cnex.Ts))

	verBytes := []byte{}
	for _, signer := range cnex.Verifiers {
		sidBytes := make([]byte, 16)
		copy(sidBytes, signer)
		verBytes = append(verBytes, []byte(sidBytes)...)

	}

	cnexBytes := bytes.Join([][]byte{coinNumBytes, cnex.CoinBytes, timeBytes, cnex.Sigs, verBytes}, []byte{})

	return cnexBytes
}
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
		counter := 156
		txn.Sigs = data[counter: counter + 128 * NUMCOSIGNER]
		counter = counter + 128 * NUMCOSIGNER
		txn.Verifiers = make([]string, NUMCOSIGNER)
		for i:=0;i<NUMCOSIGNER;i++ {
			txn.Verifiers[i] = strings.Trim(string(data[counter + i * 16: counter + (i+1) * 16]), "\x00")
		}
		return txn
	case MSG:
		txn := CNEXTxn{}
		txn.CoinNum = binary.BigEndian.Uint64(data[:8])
		txn.CoinBytes = data[8:136]
		txn.Ts = int64(binary.BigEndian.Uint64(data[136:144]))
		counter := 144
		txn.Sigs = data[counter: counter + 128 * NUMCOSIGNER]
		counter = counter + 128 * NUMCOSIGNER
		txn.Verifiers = make([]string, NUMCOSIGNER)
		for i:=0;i<NUMCOSIGNER;i++ {
			txn.Verifiers[i] = strings.Trim(string(data[counter + i * 16: counter + (i+1) * 16]), "\x00")
		}
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