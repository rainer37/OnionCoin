package ocrypto

import (
	"encoding/binary"
	"fmt"
	"crypto/rsa"
)

/*
	Onion Bytes Format(bytes):

	| cipherKey(32)	| nextID_len(4) | coin_len(4) | innerOnion_len(8) | next_ID | coin | innerOnion | chaos |
 */

type Onion struct {
	NextID string
	Coin []byte
	InnerOnion []byte
	Chaos []byte
}

func WrapOnion(pk rsa.PublicKey, symKey []byte, nextID string, coin []byte, content []byte) []byte {

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
	b = append(b, []byte{'c','h','a','o','s'}...) // TODO: replace this with random length of bytes

	cipher, cKey, err := BlockEncrypt(b, symKey, pk)
	checkErr(err)

	cipher = append(cKey, cipher...)

	return cipher
}

func PeelOnion(sk *rsa.PrivateKey, cKey []byte, fullOnion []byte) *Onion {
	// First SYM_KEY_LEN == 32 is the symmetric key.
	cKey, onion := fullOnion[:SYM_KEY_LEN], fullOnion[SYM_KEY_LEN:]
	return FormatOnion(DecryptOnion(sk, cKey, onion))
}

func DecryptOnion(sk *rsa.PrivateKey, cKey []byte, onion []byte) []byte {
	msg, err := BlockDecrypt(onion, cKey, sk)
	checkErr(err)
	return msg
}

/*
	Transforms a DRCRYPTED onion in bytes into Onion struct
 */
func FormatOnion(onion []byte) *Onion {
	totolLen := uint64(len(onion))
	nextIDLen := uint64(binary.BigEndian.Uint32(onion[:4]))
	coinLen := uint64(binary.BigEndian.Uint32(onion[4:8]))
	innerOnionLen := binary.BigEndian.Uint64(onion[8:16])
	chaosLen := totolLen - nextIDLen - coinLen - innerOnionLen - 16

	//fmt.Printf("Total Len: %d nextLen: %d coinLen: %d, innerLen: %d chaosLen: %d\n",
	// totol_len, nextID_len, coin_len, innerOnion_len, chaos_len)

	cur := uint64(16)

	o := new(Onion)
	o.NextID = string(onion[cur:cur+nextIDLen])
	cur = cur+ nextIDLen
	o.Coin = onion[cur:cur+coinLen]
	cur = cur+ coinLen
	o.InnerOnion = onion[cur: totolLen -chaosLen]
	o.Chaos = onion[totolLen - chaosLen:]
	return o
}

func (o *Onion) String() string {
	return fmt.Sprintf("nextID: %s coin: %v inner: %v chaos: %v",
		o.NextID, o.Coin, o.InnerOnion, o.Chaos)
}

func (o *Onion) GenForwardOMsg() {

}