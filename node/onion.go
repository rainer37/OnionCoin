package node

import (
	"fmt"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/coin"
	"bytes"
	"github.com/rainer37/OnionCoin/ocrypto"
)

/*
	Onion Bytes Format(bytes):

	| cipherKey(32)	| next_ID(16) | coin(COINLEN) | innerOnion | chaos |
 */

type Onion struct {
	NextID string
	Coin []byte
	InnerOnion []byte
	Chaos []byte
}

/*
	wrap nextID, Coin, and content with given pub-key to form a layer of Onion.
 */
func WrapOnion(pk rsa.PublicKey, nextID string, coin []byte, content []byte) []byte {

	nextIDBytes := make([]byte, IDLEN)
	copy(nextIDBytes, nextID)

	b := append(nextIDBytes, coin...)
	b = append(b, content...)

	cipher, cKey, err := ocrypto.BlockEncrypt(b, pk)
	checkErr(err)

	cipher = append(cKey, cipher...)
	return cipher
}
/*
	remove one layer of onion and return nexthopID, CoinBytes, and InnerOnion
 */
func PeelOnion(sk *rsa.PrivateKey, fullOnion []byte) (string, []byte, []byte) {
	// First SYMKEYLEN == 32 is the symmetric key.
	cKey, onion := fullOnion[:ocrypto.SYMKEYLEN], fullOnion[ocrypto.SYMKEYLEN:]
	decryptedOnion, err := ocrypto.BlockDecrypt(onion, cKey, sk)
	checkErr(err)
	return string(bytes.Trim(decryptedOnion[:IDLEN], "\x00")),
	decryptedOnion[IDLEN:IDLEN+coin.COINLEN],
	decryptedOnion[IDLEN+coin.COINLEN:]
}

func (o *Onion) String() string {
	return fmt.Sprintf("nextID: %s coin: %v inner: %v chaos: %v",
		o.NextID, o.Coin, o.InnerOnion, o.Chaos)
}
