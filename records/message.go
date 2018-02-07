package records

import (
	"encoding/binary"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"time"
	"fmt"
	"bytes"
)

/*
	OnionCoin Msg Standard v1.0 as of 2018 Jan
	Raw Message:
	[0-127]		:	encrypted symmetric key
	[128-end] 	:   symmetrically encrypted payload

	OnionMsg:
	[128]	: 1 byte OpCode
	[??] 	: 16 byte node id
	[??]	: 4 byte unix time stamp
	[??] 	: 32 byte signed sha256 sum of [payload:ts]
	[??] 	: 4 byte length of payload
	[??]	: payload and chaos


*/

const (
	CIPHERKEYLEN = 128
	NODEIDLEN = 16
	TSLEN = 8
	HASHLEN = 128
	PAYLOADLENLEN = 4
)

type OnionMsg interface {
	GetOPCode() rune
	GetSenderID() string
	GetLenPayload() int
	GetPayload(len int) []byte
	GetPayloadTsHash() []byte
	GetTS() uint32
	VerifySig(pk *rsa.PublicKey) bool
}

type OMsg struct {
	B []byte
}

func (omsg *OMsg) GetOPCode() rune {
	return rune(omsg.B[0])
}

func (omsg *OMsg) GetSenderID() string {

	return string(bytes.Trim(omsg.B[1:1 + NODEIDLEN], "\x00"))
}

func (omsg *OMsg) GetTS() uint32 {
	return uint32(binary.BigEndian.Uint32(omsg.B[1 + NODEIDLEN : 1 + NODEIDLEN + TSLEN]))
}

func (omsg *OMsg) GetLenPayload() int {
	return int(binary.BigEndian.Uint32(omsg.B[1 + NODEIDLEN + TSLEN + HASHLEN : 1 + NODEIDLEN + TSLEN + HASHLEN + PAYLOADLENLEN]))
}

func (omsg *OMsg) GetPayload(len int) []byte {
	return omsg.B[1 + NODEIDLEN + TSLEN + HASHLEN + PAYLOADLENLEN :1 + NODEIDLEN + TSLEN + HASHLEN + PAYLOADLENLEN + len]
}

func (omsg *OMsg) GetPayloadTsHash() []byte {
	return omsg.B[1 + NODEIDLEN + TSLEN : 1 + NODEIDLEN + TSLEN + HASHLEN]
}

func (omsg *OMsg) VerifySig(pk *rsa.PublicKey) bool {
	payload := omsg.GetPayload(omsg.GetLenPayload())
	hash := omsg.GetPayloadTsHash()
	return ocrypto.RSAVerify(pk, hash, payload)
}

func UnmarshalOMsg(msg []byte, sk *rsa.PrivateKey) (*OMsg, bool) {
	omsg := new(OMsg)
	b, err := ocrypto.BlockDecrypt(msg[CIPHERKEYLEN:], msg[:CIPHERKEYLEN], sk)
	if err != nil {
		fmt.Println(1,err.Error())
		return nil, false
	}
	omsg.B = b
	return omsg, true
}

/*
	sk : own secret key
	pk : target public key
 */
func MarshalOMsg(opCode rune, payload []byte, nodeID string, sk *rsa.PrivateKey, pk rsa.PublicKey) []byte {
	buffer := make([]byte,1)
	buffer[0] = byte(opCode)

	buf := make([]byte, NODEIDLEN)
	for i:=0;i<NODEIDLEN;i++ {
		if i < len(nodeID) {
			buf[i] = nodeID[i]
		}
	}
	buffer = append(buffer, buf...)

	buf = make([]byte, TSLEN)
	binary.BigEndian.PutUint64(buf, uint64(time.Now().Unix()))
	buffer = append(buffer, buf...)

	buf = make([]byte, HASHLEN)
	sig := ocrypto.RSASign(sk, payload)
	for i:=0;i<HASHLEN;i++ {
		buf[i] = sig[i]
	}
	buffer = append(buffer, buf...)

	buf = make([]byte, PAYLOADLENLEN)
	binary.BigEndian.PutUint32(buf, uint32(len(payload)))
	buffer = append(buffer, buf...)

	buffer = append(buffer, payload...)

	key := []byte("the-key-has-to-be-32-bytes-long!") // TODO: generate a random 32-byte key

	cipher, ckey, err := ocrypto.BlockEncrypt(buffer, key, pk)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	ckey = append(ckey, cipher...)

	return ckey
}


