package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"github.com/rainer37/OnionCoin/blockChain"
)

const BCOINSIZE = 128
var exMap = map[string]chan []byte{} // channels for coin exchanging

/*
	bank processing coin exchange request.
	RAWCOIN(128) | BFID(8) | COINREWARD(128) |
	If valid coin received, sign the rawCoin and send it back.
	meanwhile starting coSign protocol to get coin published.
 */
func (n *Node) receiveRawCoin(payload []byte, senderID string) {
	//print("Make a wish")
	if len(payload) != BCOINSIZE * 2 + 8 {
		print("Wrong coin exchange len", len(payload))
		return }

	c := payload[BCOINSIZE+8:]

	// check validity, if not, abort
	if !n.ValidateCoin(c, senderID) {
		print("invalid coin refuse signing it")
		return
	}

	//print("valid coin, continue")

	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, 0)
	// start CoSign protocol with counter 0.
	n.coSignValidCoin(append(counterBytes, c...))

	rwcn := payload[:BCOINSIZE]
	bfid := payload[BCOINSIZE:BCOINSIZE+8]

	newCoin := n.blindSign(rwcn)
	spk := n.getPubRoutingInfo(senderID)

	if spk == nil {
		print("Cannot find the key with senderID")
		return
	}

	//print("reply with partial newCoin")

	p := n.prepareOMsg(RAWCOINSIGNED,append(newCoin, bfid...),spk.Pk)
	n.sendActive(p, spk.Port)
}

/*
	received newSignedCoin from Bank.
	| newCoin(128) | bfid(8) |
 */
func (n *Node) receiveNewCoin(payload []byte, senderID string) {

	if len(payload) != BCOINSIZE+8 {
		return
	}

	newCoin := payload[:BCOINSIZE]
	bfid := payload[BCOINSIZE:]
	exMap[string(bfid)] <- newCoin
}

/*
	Blind a RawCoin's bytes with random blind factor, and record bf.
	return blinded RawCoin and blindFactor id for future unblinding.
 */
func BlindBytes(b []byte, bankPK *rsa.PublicKey) ([]byte, string) {
	brwcn, bfac := ocrypto.Blind(bankPK, b)
	bfid := make([]byte, 8)
	binary.BigEndian.PutUint64(bfid,rand.Uint64())
	coin.RecordBF(string(bfid) ,bfac)
	return brwcn, string(bfid)
}

/*
	Unblind the SignedRawCoin received, using saved blind factor.
 */
func UnBlindBytes(signedRC []byte, bfID string, bankPK *rsa.PublicKey) []byte {
	bf := coin.GetBF(bfID)
	if bf == nil {
		return nil
	}
	c := ocrypto.Unblind(bankPK,signedRC, bf)
	return c
}

func (n *Node) blindSign(rawCoin []byte) []byte {
	return ocrypto.BlindSign(n.sk, rawCoin)
}

/*
	validate the coin received by decrypting the coin multiple times then check against coinNum and senderID.
 */
func (n *Node) ValidateCoin(coin []byte, senderID string) bool {
	// TODO: validate coin that is signed by other banks

	// first check if it is a genesis coin.
	spe := n.getPubRoutingInfo(senderID)
	encSPK := sha256.Sum256(ocrypto.EncodePK(spe.Pk))
	targetHash := ocrypto.EncryptBig(&spe.Pk, coin)

	if string(encSPK[:]) == string(targetHash) {
		print(senderID, "GCoin received")
		return true
	}

	return true
	// return ValidateCoinByKey(coin, senderID, &n.sk.PublicKey)
}

/*
	Validate a received coin by checking if the rid matches senderID, and if the coinNum is free.
 */
func ValidateCoinByKey(coinBytes []byte, senderID string, pk *rsa.PublicKey) bool {
	c := ocrypto.EncryptBig(pk, coinBytes)
	if len(c) != 40 {
		return false
	}

	idHash := sha256.Sum256([]byte(senderID))
	targetHash := c[:32]
	coinNum := c[32:]

	if string(idHash[:]) != string(targetHash) {
		return false
	}

	if !blockChain.IsFreeCoinNum(binary.BigEndian.Uint64(coinNum)) {
		return false
	}

	return true
}
