package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"strconv"
	"crypto/rsa"
	"crypto/sha256"
	"math/big"
	"encoding/binary"
)

/*
	bank processing coin exchange request.

 */
func coinExProtocol(payload []byte) {
	print("Make a wish")
}

/*
	Blind a RawCoin's bytes with random blind factor, and record bf.
	return blinded RawCoin and blindFactor id for future unblinding.
 */
func BlindRawCoin(rwcn *coin.RawCoin, bankPK *rsa.PublicKey) ([]byte, string) {
	brwcn, bfac := ocrypto.Blind(bankPK, rwcn.ToBytes())
	bfid := strconv.FormatUint(rwcn.GetCoinNum(), 10)
	coin.RecordBF(bfid ,bfac)
	return brwcn, bfid
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

func (n *Node) ValidateCoin(coin []byte, senderID string) bool {
	// TODO: validate coin that is signed by other banks
	return ValidateCoinByKey(coin, senderID, &n.sk.PublicKey)
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
