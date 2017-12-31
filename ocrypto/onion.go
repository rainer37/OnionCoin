package ocrypto

import (
	"encoding/binary"
	"fmt"
)

/*
	Onion Bytes Format(bytes):
	| nextID_len(4) | coin_len(4) | innerOnion_len(8) | next_ID | coin | innerOnion | chaos |
 */

type Onion struct {
	NextID string
	Coin []byte
	InnerOnion []byte
	Chaos []byte
}

type OnionMaker struct {}

func decryptOnion(sk []byte, onion []byte) []byte {
	return nil
}

/*
	Transforms a DRCRYPTED onion in bytes into Onion struct
 */
func FormatOnion(onion []byte) *Onion {
	totol_len := uint64(len(onion))
	nextID_len := uint64(binary.BigEndian.Uint32(onion[:4]))
	coin_len := uint64(binary.BigEndian.Uint32(onion[4:8]))
	innerOnion_len := binary.BigEndian.Uint64((onion[8:16]))
	chaos_len := totol_len - nextID_len - coin_len - innerOnion_len - 16

	//fmt.Printf("Total Len: %d nextLen: %d coinLen: %d, innerLen: %d chaosLen: %d\n", totol_len, nextID_len, coin_len, innerOnion_len, chaos_len)

	cur := uint64(16)

	o := new(Onion)
	o.NextID = string(onion[cur:cur+nextID_len])
	cur = cur+nextID_len
	o.Coin = onion[cur:cur+coin_len]
	cur = cur+coin_len
	o.InnerOnion = onion[cur: totol_len - chaos_len]
	o.Chaos = onion[totol_len - chaos_len:]
	return o
}

func CookOnion(sk []byte, onion []byte) *Onion {
	return FormatOnion(decryptOnion(sk, onion))
}

func (o *OnionMaker) MakeOnionHeart() *Onion { return nil }

func (o *OnionMaker) wrap(pk []byte, nextID string, len int, onionByte []byte) []byte {
	return nil
}

func (o *OnionMaker) peel (sk []byte, onionBytes []byte) (oret *Onion) {
	return
}

func (o *Onion) String() string {
	return fmt.Sprintf("nextID: %s coin: %v inner: %v chaos: %v", o.NextID, o.Coin, o.InnerOnion, o.Chaos)
}