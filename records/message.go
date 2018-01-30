package records

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

