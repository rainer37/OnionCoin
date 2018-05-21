package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"encoding/binary"
	"crypto/rsa"
	"math/rand"
	"github.com/rainer37/OnionCoin/blockChain"
)

/*
	received newSignedCoin from Bank.
	| newCoin(128) | bfid(8) |
 */
func (n *Node) receiveNewCoin(payload []byte, senderID string) {
	if len(payload) != BCOINSIZE + 8 { return }
	newCoin := payload[:BCOINSIZE]
	bfid := payload[BCOINSIZE:]
	exMap[string(bfid)] <- newCoin
}

/*
	Generate the genesis coin with my signed pk.
 */
func (n *Node) GetGenesisCoin() *coin.Coin {
	pkHash := util.Sha(ocrypto.EncodePK(n.sk.PublicKey))
	gcoin := n.blindSign(pkHash[:])
	return coin.NewCoin(n.ID, gcoin, []string{n.ID})
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
	if bf == nil { return nil }
	return ocrypto.Unblind(bankPK,signedRC, bf)
}

func shuffleBanks() []string {
	l := len(currentBanks)
	bs := make([]string, l)
	for i:=0; i<l; i++ {
		b := currentBanks[rand.Int() % l]
		for util.Contains(bs, b) {
			b = currentBanks[rand.Int() % l]
		}
		bs[i] = b
	}
	return bs
}

/*
	Exchanging an existing coin to a newCoin with dstID, and random coinNum.
	1. generate a rawCoin with dstID.
	2. Lookup for banks and their address.
	3. iteratively blind the rawCoin and send it to one of the Bank with a valid coin.
	4. Unblind the signed rawCoin, go to 3 if not enough banks sign the rawCoin.
	5. deposit the newCoin.
*/
func (n *Node) CoinExchange(dstID string) {
	rwcn := coin.NewRawCoin(dstID)

	// gcoin := n.Vault.Withdraw(n.ID).Bytes()
	gcoin := n.GetGenesisCoin().Bytes()
	if gcoin == nil {
		print("No More Coins To exchange")
		return
	}

	banks := shuffleBanks() // get banks with random orders.

	signerBanks := []string{} // records which banks are helping

	counter := 0
	layers := 0
	rc := rwcn.ToBytes()

	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	for layers < util.NUMCOSIGNER && counter < len(banks) {
		banks = currentBanks
		bid := banks[counter]

		bpe := n.getPubRoutingInfo(bid)
		if bpe == nil { continue }

		// print("Requesting", bid, "for signing rawCoin")

		blindrwcn, bfid := BlindBytes(rc, &bpe.Pk)

		payload := util.JoinBytes([][]byte{tsBytes, blindrwcn, []byte(bfid), gcoin})
		n.sendOMsg(RAWCOINEXCHANGE, payload, bpe)

		var realCoin []byte

		m.Lock()
		exMap[bfid] = make(chan []byte)

		select{
		case reply := <-exMap[bfid]:
			realCoin = reply
			close(exMap[bfid])
			m.Unlock()
		case <-time.After(util.COSIGNTIMEOUT * time.Second):
			// print(bid, "no response, try next bank")
			close(exMap[bfid])
			counter++
			m.Unlock()
			continue
		}
		//print("waiting for response from", bid)

		revealedCoin := UnBlindBytes(realCoin, bfid, &bpe.Pk)

		counter++

		expected := ocrypto.EncryptBig(&bpe.Pk, revealedCoin)

		if string(expected) != string(rc) {
			// print("not equal after blindSign, bad bank!", bid, len(expected), len(rc))
			continue
		}

		rc = revealedCoin
		signerBanks = append(signerBanks, bid)
		layers++
	}

	if layers != util.NUMCOSIGNER {
		// print("Not Enough Banks To Forge a Coin, Try Next Epoch")
		return
	}

	n.Deposit(coin.NewCoin(dstID, rc, signerBanks))
}

/*
	Validate a received coin by checking if the rid matches senderID, and if the coinNum is free.
 */
func ValidateCoinByKey(coinBytes []byte, senderID string, pk *rsa.PublicKey) bool {
	c := ocrypto.EncryptBig(pk, coinBytes)

	if len(c) != 32 + 8 { return false }

	idHash := util.Sha([]byte(senderID))
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