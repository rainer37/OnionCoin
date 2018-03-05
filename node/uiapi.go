package node

import (
	"github.com/rainer37/OnionCoin/records"
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/bank"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"crypto/sha256"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/blockChain"
)

const COSIGNTIMEOUT = 2
const NUMCOSIGNER = 2
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
		payload = append(payload, blindrwcn...) // TODO: append a real COIN

		fo := n.prepareOMsg(RAWCOINEXCHANGE,payload,bpe.Pk)

		exMap[bfid] = make(chan []byte)

		n.sendActive(fo, bpe.Port)

		var realCoin []byte

		select{
		case reply := <-exMap[bfid]:
			realCoin = reply
			close(exMap[bfid])
		case <-time.After(COSIGNTIMEOUT * time.Second):
			print(b, "no response, try next bank")
			counter++
			continue
		}

		print("waiting for response from", b)

		revealedCoin := UnBlindBytes(realCoin, bfid, &bpe.Pk)

		counter++

		expected := ocrypto.EncryptBig(&bpe.Pk, revealedCoin)

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
		n.Deposit(coin.NewCoin(dstID, rc))
	} else {
		print("Not Enough Banks To Forge a Coin, Try Next Epoch")
	}
}

/*
	Upon received a valid coin, the bank signs the coin and pass it to other banks
	Till enough signatures gained, then publish it as a transaction.
	Does the last CoSigner solves the puzzle of blind signers?
 */
func (n *Node) coSignValidCoin(c []byte, counter uint16) {

	hash := sha256.Sum256(c[:128]) // get the hash(32) of coin
	signedHash := n.blindSign(hash[:]) // sign the coin(128)

	signedHash = append(c, signedHash...)
	newCounter := make([]byte, 2)
	binary.BigEndian.PutUint16(newCounter, counter+1)
	signedHash = append(newCounter, signedHash...) // add updated counter to the head.

	if counter == NUMCOSIGNER {
		print("Enough verifiers got, publish it")
		txn := new(blockChain.CNEXTxn)
		n.publicTxn(txn)
		return
	}

	i := 0
	if time.Now().Unix() % 2 == 0{
		i = 1
	}
	bid := bank.GetBankIDSet()[i] // TODO: randomly pick another bank.
	tpk := records.GetKeyByID(bid)

	if tpk == nil {
		print("Cannot find the key by id")
		return
	}

	payload := n.prepareOMsg(COSIGN, signedHash, tpk.Pk)

	print("sending aggregated signed coin and cosign counter", newCounter)
	n.sendActive(payload, tpk.Port)
}

func (n *Node) publicTxn(txn blockChain.Txn) {

}