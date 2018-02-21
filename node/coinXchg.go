package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"crypto/sha256"
	"math/big"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/records"
	"math/rand"
)

const BCOINSIZE = 128
const NUMSIGNINGBANK = 2
var exMap = map[string]chan []byte{}
/*
	bank processing coin exchange request.
	RAWCOIN(128) | BFID(8) | COIN(128) |
	If valid coin received, sign the rawCoin and send it back.
 */
func (n *Node) receiveRawCoin(payload []byte, senderID string) {
	print("Make a wish")
	if len(payload) != BCOINSIZE * 2 + 8 { return }

	c := payload[BCOINSIZE+8:]
	if !n.ValidateCoin(c, senderID) { return }

	print("valid coin, continue")
	rwcn := payload[:BCOINSIZE]
	bfid := payload[BCOINSIZE:BCOINSIZE+8]

	newCoin := n.blindSign(rwcn)
	spk := records.GetKeyByID(senderID)
	p := records.MarshalOMsg(RAWCOINSIGNED,append(newCoin, bfid...),n.ID,n.sk,spk.Pk)
	n.sendActive(p, spk.Port)
}

/*
	received newSignedCoin from Bank.
	| newCoin(128) | bfid(8) |
 */
func (n *Node) receiveNewCoin(payload []byte, senderID string) {
	if len(payload) != BCOINSIZE+8 { return }
	newCoin := payload[:BCOINSIZE]
	bfid := payload[BCOINSIZE:]
	bpk := records.GetKeyByID(senderID)
	realCoin := UnBlindSignedRawCoin(newCoin, string(bfid), &bpk.Pk)
	print(len(realCoin))

	if ValidateCoinByKey(realCoin, senderID, &bpk.Pk) {
		print("Man i got a coin")
		exMap[string(bfid)] <- realCoin
	} else {
		print("Bad bank sucks")
		exMap[string(bfid)] <- []byte("BADBANK")
	}
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
func UnBlindSignedRawCoin(signedRC []byte, bfID string, bankPK *rsa.PublicKey) []byte {
	bf := coin.GetBF(bfID)
	if bf == nil {
		return nil
	}
	coin := ocrypto.Unblind(bankPK,signedRC, bf)
	return coin
}

func (n *Node) blindSign(rawCoin []byte) []byte {
	return ocrypto.BlindSign(n.sk, rawCoin)
}

/*
	validate the coin received by decrypting the coin multiple times then check against coinNum and senderID.
 */
func (n *Node) ValidateCoin(coin []byte, senderID string) bool {
	// TODO: validate coin that is signed by other banks
	return true
	// return ValidateCoinByKey(coin, senderID, &n.sk.PublicKey)
}

/*
	Validate a received coin by checking if the rid matches senderID, and if the coinNum is free.
 */
func ValidateCoinByKey(coinBytes []byte, senderID string, pk *rsa.PublicKey) bool {
	c := ocrypto.Encrypt(new(big.Int), pk, new(big.Int).SetBytes(coinBytes)).Bytes()

	if len(c) != 40 {
		return false
	}

	idHash := sha256.Sum256([]byte(senderID))
	targetHash := c[:32]
	coinNum := c[32:]

	if string(idHash[:]) != string(targetHash) {
		return false
	}

	if !coin.IsFreeCoinNum(binary.BigEndian.Uint64(coinNum)) {
		return false
	}

	return true
}
