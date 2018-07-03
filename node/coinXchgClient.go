package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"encoding/binary"
	"crypto/rsa"
	"math/rand"
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
//func (n *Node) GetGenesisCoin() *coin.Coin {
//	pkHash := util.Sha(ocrypto.EncodePK(n.sk.PublicKey))
//	gcoin := n.blindSign(pkHash[:])
//	return coin.NewCoin(n.ID, gcoin, []string{n.ID})
//}

func (n *Node) GetGenesisCoin() *coin.Coin {
	myIDHash := util.Sha([]byte(n.ID))
	//myIDHash := []byte("rainerrainerrainerrainerrainerer")
	gbytes := util.NewBytes(40, myIDHash[:])
	gbs := util.SplitBytes(gbytes, util.NUMCOSIGNER)
	gcoin := []byte{}
	for _, gb := range gbs {
		gcoin = append(gcoin, n.blindSign(gb)...)
	}
	signers := []string{n.ID}
	for i:=0;i<util.NUMCOSIGNER-1;i++ {
		signers = append(signers, n.ID)
	}
	return coin.NewCoin(n.ID, gcoin, signers)
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
//	gcoin := n.GetGenesisCoin().Bytes()
	gcoin := n.Withdraw(dstID).Bytes()

	if gcoin == nil {
		print("No More Coins To exchange")
		return
	}

	banks := shuffleBanks() // get banks with random orders.

	signerBanks := []string{} // records which banks are helping

	counter := 0
	layers := 0
	rc := rwcn.ToBytes()
	tsBytes := util.CurTSBytes()
	newCoin := []byte{}

	rawCoinSegs := util.SplitBytes(rc, util.NUMCOSIGNER)
	print("coin segments num:", len(rawCoinSegs),
		"coin len:", len(gcoin))

	for layers < util.NUMCOSIGNER && counter < len(banks) {
		bid := banks[counter]
		bpe := n.getPubRoutingInfo(bid)

		blindRawSeg, bfid := BlindBytes(rawCoinSegs[layers], &bpe.Pk)
		print("blindRawSeg size:", len(blindRawSeg))

		payload := util.JoinBytes([][]byte{tsBytes, blindRawSeg, []byte(bfid), gcoin})
		n.sendOMsg(RAWCOINEXCHANGE, payload, bpe)

		var signedRawSeg []byte
		exMap[bfid] = make(chan []byte)

		m.Lock()
		select {
			case signedRawSeg = <-exMap[bfid]:
			case <-time.After(util.COSIGNTIMEOUT * time.Second):
				signedRawSeg = nil
		}

		close(exMap[bfid])
		m.Unlock()

		counter++

		if signedRawSeg == nil { continue }

		revealedCoin := UnBlindBytes(signedRawSeg, bfid, &bpe.Pk)

		expected := ocrypto.EncryptBig(&bpe.Pk, revealedCoin)

		if string(expected) != string(rawCoinSegs[layers]) {
			print("not equal after blindSign, bad bank!",
				bid, len(expected), len(rawCoinSegs[layers]))
			continue
		}

		newCoin = append(newCoin, revealedCoin...)
		signerBanks = append(signerBanks, bid)
		layers++
	}

	if layers != util.NUMCOSIGNER {
		print("Not Enough Banks To Forge a Coin, Try Next Epoch")
		return
	}
	print("successfully get a coin for", dstID)
	n.Deposit(coin.NewCoin(dstID, newCoin, signerBanks))
}