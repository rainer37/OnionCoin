package node

import (
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/bank"
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

	counter := 0
	rc := rwcn.ToBytes()
	for counter < NUMSIGNINGBANK {
		b := banks[counter]
		bpe := records.GetKeyByID(b)

		if bpe == nil {
			print("ERR finding bank id provided", b)
			continue
		}

		print("Requesting", b, "for signing rawCoin")

		blindrwcn, bfid := BlindBytes(rc, &bpe.Pk)
		print(blindrwcn, bfid)

		payload := append(blindrwcn, []byte(bfid)...)
		payload = append(payload, blindrwcn...)
		print(len(payload))

		fo := records.MarshalOMsg(RAWCOINEXCHANGE,payload,n.ID,n.sk,bpe.Pk)
		n.sendActive(fo, bpe.Port)

		print("waiting for response")

		exMap[bfid] = make(chan []byte)
		realCoin := <-exMap[bfid]

		counter++

		if string(realCoin) == "BADBANK" {
			print("C'mon man, sign it properly!")
		}

		rc = realCoin

	}
	print("new coin got", len(rc))
}
