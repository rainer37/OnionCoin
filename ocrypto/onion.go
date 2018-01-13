package ocrypto

import (
	"encoding/binary"
	"fmt"
	"crypto/rsa"
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

func WrapOnion(pk rsa.PublicKey, nextID string, coin []byte, content []byte) []byte {

	b := make([]byte, 0)

	nlen := make([]byte, 4)
	binary.BigEndian.PutUint32(nlen, uint32(len(nextID)))
	b = append(b, nlen...)

	clen := make([]byte, 4)
	binary.BigEndian.PutUint32(clen, uint32(len(coin)))
	b = append(b, clen...)

	ilen := make([]byte, 8)
	binary.BigEndian.PutUint64(ilen, uint64(len(content)))
	b = append(b, ilen...)

	b = append(b, []byte(nextID)...)
	b = append(b, coin...)
	b = append(b, content...)
	b = append(b, []byte{'c','h','a','o','s'}...)

	return PKEncrypt(pk, b)
}

func PeelOnion(sk *rsa.PrivateKey, onion []byte) *Onion {
	return FormatOnion(DecryptOnion(sk, onion))
}

func DecryptOnion(sk *rsa.PrivateKey, onion []byte) []byte {
	return PKDecrypt(sk, onion)
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

func (o *Onion) String() string {
	return fmt.Sprintf("nextID: %s coin: %v inner: %v chaos: %v",
		o.NextID, o.Coin, o.InnerOnion, o.Chaos)
}