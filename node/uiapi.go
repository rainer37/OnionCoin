package node

import (
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/bank"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"math/big"
	"time"
)

/*
	Exchanging an existing coin to a newCoin with dstID, and random coinNum.
	1. generate a rawCoin with dstID.
	2. Lookup for banks and their address.
	3. iteratively blind the rawCoin and send it to one of the Bank with a valid coin.
	4. Unblind the signed rawCoin, go to 3 if not enough banks sign the rawCoin.
	5. deposit the newCoin.
*/

func (n *Node) CoinExchange(dstID string) {
	dstID = "FAKEID" + dstID
	rwcn := coin.NewRawCoin(dstID)

	print(rwcn.GetCoinNum())

	banks := bank.GetBankIDSet()
	print(banks)
	banksPk := []rsa.PublicKey{}

	counter := 0
	layers := 0
	rc := rwcn.ToBytes()

	for layers < NUMSIGNINGBANK && counter < len(banks) {
		b := banks[counter]
		bpe := records.GetKeyByID(b)

		if bpe == nil {
			print("ERR finding bank id provided", b)
			continue
		}

		print("Requesting", b, "for signing rawCoin")

		blindrwcn, bfid := BlindBytes(rc, &bpe.Pk)

		payload := append(blindrwcn, []byte(bfid)...)
		payload = append(payload, blindrwcn...)

		fo := records.MarshalOMsg(RAWCOINEXCHANGE,payload,n.ID,n.sk,bpe.Pk)

		exMap[bfid] = make(chan []byte)

		n.sendActive(fo, bpe.Port)

		// TODO: start a timer for network timeout

		var realCoin []byte

		select{
		case reply := <-exMap[bfid]:
			realCoin = reply
		case <-time.After(3 * time.Second):
			print(b, "no response, try next bank")
			counter++
			continue
		}

		print("waiting for response")

		revealedCoin := UnBlindSignedRawCoin(realCoin, bfid, &bpe.Pk)

		counter++

		expected := ocrypto.Encrypt(new(big.Int), &bpe.Pk, new(big.Int).SetBytes(revealedCoin)).Bytes()

		if string(expected) != string(rc) {
			print("not equal, bad bank!", b)
			continue
		}

		rc = revealedCoin
		banksPk = append(banksPk, bpe.Pk)
		layers++
	}

	if layers == NUMSIGNINGBANK {
		print("New Coin Forged, Thanks Fellas!")
	} else {
		print("Not Enough Banks To Forge a Coin, Try Next Epoch")
	}
}
