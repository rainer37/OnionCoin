package records

import (
	"encoding/binary"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"time"
	"fmt"
	"github.com/rainer37/OnionCoin/util"
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
	TSLEN = 8
	HASHLEN = util.RSAKEYLEN / 8 //128
	PAYLOADLENLEN = 4
	TS_OFFSET = util.IDLEN + 1
	PLD_LEN_OFFSET = TS_OFFSET + TSLEN
	PLD_OFFSET = PLD_LEN_OFFSET + HASHLEN + PAYLOADLENLEN
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
	return util.GetID(omsg.B[1:1 + util.IDLEN]) // trim trailing NULL
}

func (omsg *OMsg) GetTS() uint32 {
	return uint32(binary.BigEndian.Uint32(omsg.B[TS_OFFSET: PLD_LEN_OFFSET]))
}

func (omsg *OMsg) GetLenPayload() int {
	return int(binary.BigEndian.Uint32(omsg.B[PLD_LEN_OFFSET + HASHLEN : PLD_OFFSET]))
}

func (omsg *OMsg) GetPayload() []byte {
	return omsg.B[PLD_OFFSET : PLD_OFFSET + omsg.GetLenPayload()]
}

func (omsg *OMsg) GetPayloadTsHash() []byte {
	return omsg.B[PLD_LEN_OFFSET : PLD_LEN_OFFSET + HASHLEN]
}

func (omsg *OMsg) VerifySig(pk *rsa.PublicKey) bool {
	payload := omsg.GetPayload()
	hash := omsg.GetPayloadTsHash()
	return ocrypto.RSAVerify(pk, hash, payload)
}

/*
	First Decrypt the encrypted symkey with my private key,
	then decrypt the rest of msg with sym key, and return omsg.
 */
func UnmarshalOMsg(msg []byte, sk *rsa.PrivateKey) (*OMsg, bool) {
	omsg := new(OMsg)
	if len(msg) < util.CIPHERKEYLEN {
		return nil, false
	}
	b, err := ocrypto.BlockDecrypt(msg[util.CIPHERKEYLEN:], msg[:util.CIPHERKEYLEN], sk)
	if err != nil {
		fmt.Println("block Decrypt",err.Error())
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
	op_buf := make([]byte,1)
	op_buf[0] = byte(opCode)

	if len(nodeID) > util.IDLEN { return nil }

	id_buf := make([]byte, util.IDLEN)
	copy(id_buf[:], nodeID)

	ts_buf := make([]byte, TSLEN)
	binary.BigEndian.PutUint64(ts_buf, uint64(time.Now().Unix()))

	hash_buf := make([]byte, HASHLEN)
	sig := ocrypto.RSASign(sk, payload)
	copy(hash_buf, sig)

	pld_len_buf := make([]byte, PAYLOADLENLEN)
	binary.BigEndian.PutUint32(pld_len_buf, uint32(len(payload)))

	buffer := util.JoinBytes([][]byte{op_buf, id_buf, ts_buf, hash_buf, pld_len_buf, payload})
	cipher, ckey, err := ocrypto.BlockEncrypt(buffer, pk)

	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return append(ckey, cipher...)
}