package records

import (
	"encoding/binary"
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


type OnionMsg interface {
	GetOPCode() rune
	GetSenderID() string
	GetLenPayload() int
	GetPayload(len int) []byte
	GenPayloadTsHash() []byte
	VerifySig() bool
}

type OMsg struct {
	B []byte
}

func (omsg *OMsg) GetOPCode() rune {
	return rune(omsg.B[0])
}

func (omsg *OMsg) GetSenderID() string {
	return string(omsg.B[1:17])
}

func (omsg *OMsg) GetLenPayload() int {
	return int(binary.BigEndian.Uint32(omsg.B[52:56]))
}

func (omsg *OMsg) GetPayload(len int) []byte {
	return omsg.B[56:56+len]
}

func (omsg *OMsg) GenPayloadTsHash() []byte {
	return omsg.B[21:53]
}

func (omsg *OMsg) VerifySig() bool {
	return true
}
