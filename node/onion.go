package node

import (
	"fmt"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"github.com/rainer37/OnionCoin/util"
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

	nextIDBytes := make([]byte, util.IDLEN)
	copy(nextIDBytes, nextID)

	cbytes := make([]byte, 256) // TODO: adjust the size by real coin size.
	copy(cbytes, coin)

	b := util.JoinBytes([][]byte{nextIDBytes, cbytes, content})
	cipher, cKey, err := ocrypto.BlockEncrypt(b, pk)
	util.CheckErr(err)

	return append(cKey, cipher...)
}
/*
	remove one layer of onion and return nexthopID, CoinBytes, and InnerOnion
 */
func PeelOnion(sk *rsa.PrivateKey, fullOnion []byte) (string, []byte, []byte) {
	// First SYMKEYLEN == 32 is the symmetric key.
	cKey, onion := fullOnion[:util.SYMKEYLEN], fullOnion[util.SYMKEYLEN:]
	decryptedOnion, err := ocrypto.BlockDecrypt(onion, cKey, sk)
	util.CheckErr(err)
	// print(len(decryptedOnion))
	return util.Strip(decryptedOnion[:util.IDLEN]),
	[]byte(util.Strip(decryptedOnion[util.IDLEN:util.IDLEN+256])),
	decryptedOnion[util.IDLEN+256:]
	//
	//return string(bytes.Trim(decryptedOnion[:util.IDLEN], "\x00")),
	//bytes.Trim(decryptedOnion[util.IDLEN:util.IDLEN+256], "\x00"),
	//decryptedOnion[util.IDLEN+256:]
}

func (o *Onion) String() string {
	return fmt.Sprintf("nextID: %s coin: %v inner: %v chaos: %v",
		o.NextID, o.Coin, o.InnerOnion, o.Chaos)
}
