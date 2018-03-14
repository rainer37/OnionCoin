package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/bank"
	"crypto/rsa"
	"github.com/rainer37/OnionCoin/ocrypto"
	"time"
	"crypto/sha256"
	"encoding/binary"
	"github.com/rainer37/OnionCoin/blockChain"
	"strings"
	"math/rand"
)

const COSIGNTIMEOUT = 2

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

	// print(rwcn.GetCoinNum())
	//c := n.Vault.Withdraw(n.ID) // get a coin from the vault.
	//
	//if c == nil {
	//	print("no coin for exchange a new one")
	//	return
	//}

	banks := bank.GetBankIDSet()
	print(banks)
	banksPk := []rsa.PublicKey{} // records which banks are helping

	counter := 0
	layers := 0
	rc := rwcn.ToBytes()

	for layers < blockChain.NUMCOSIGNER && counter < len(banks) {
		bid := banks[counter]
		bpe := n.getPubRoutingInfo(bid)

		print("Requesting", bid, "for signing rawCoin")

		blindrwcn, bfid := BlindBytes(rc, &bpe.Pk)

		payload := append(blindrwcn, []byte(bfid)...)

		payload = append(payload, blindrwcn...) // TODO: append a real COIN

		fo := n.prepareOMsg(RAWCOINEXCHANGE, payload, bpe.Pk)

		exMap[bfid] = make(chan []byte)

		n.sendActive(fo, bpe.Port)

		var realCoin []byte

		select{
		case reply := <-exMap[bfid]:
			realCoin = reply
			close(exMap[bfid])
		case <-time.After(COSIGNTIMEOUT * time.Second):
			print(bid, "no response, try next bank")
			counter++
			continue
		}

		print("waiting for response from", bid)

		revealedCoin := UnBlindBytes(realCoin, bfid, &bpe.Pk)

		counter++

		expected := ocrypto.EncryptBig(&bpe.Pk, revealedCoin)

		if string(expected) != string(rc) {
			print("not equal after blindSign, bad bank!", bid)
			continue
		}

		rc = revealedCoin
		banksPk = append(banksPk, bpe.Pk)
		layers++
	}

	if layers == blockChain.NUMCOSIGNER {
		print("New Coin Forged, Thanks Fellas!")
		n.Deposit(coin.NewCoin(dstID, rc))
		// print(n.Vault.Coins)
	} else {
		print("Not Enough Banks To Forge a Coin, Try Next Epoch")
	}
}

/*
	Upon received a valid coin, the bank signs the coin and pass it to other banks
	Till enough signatures gained, then publish it as a transaction.
	Does the last CoSigner solves the puzzle of blind signers?
 */
func (n *Node) coSignValidCoin(c []byte) {

	counter := binary.BigEndian.Uint16(c[:2]) // get cosign counter first 2 bytes

	c = c[2:]

	hashAndIds := sha256.Sum256(c[:128]) // get the hash(32) of coin

	signedHash := n.blindSign(hashAndIds[:]) // sign the coin(128)
	signedHash = append(c, signedHash...)

	newCounter := make([]byte, 2)
	binary.BigEndian.PutUint16(newCounter, counter+1)

	idBytes := make([]byte, 16)
	copy(idBytes, n.ID)

	signedHash = append(signedHash, idBytes[:]...) // append verifier to it

	// when there is enough sigs gathered, try publish the txn.
	if counter+1 == blockChain.NUMCOSIGNER {
		print("Enough verifiers got, publish it")
		print(len(signedHash), counter+1, "verifiers")
		cnum, cbytes, sigs, verifiers := decodeCNCosign(signedHash, counter+1)
		print(cnum, verifiers, len(sigs))
		txn := blockChain.NewCNEXTxn(cnum, cbytes, sigs, verifiers)
		// start broadcasting the new Txn.
		// TODO: go n.broadcastTxn(txn)
		ok := n.bankProxy.AddTxn(txn)
		if ok {
			print("time to publish this block")
			// n.publishBlock()
		}
		return
	}


	signedHash = append(newCounter, signedHash...) // add updated counter to the head.cvx


	// randomly picks banks other than me
	otherBanks := bank.GetBankIDSet()
	bid := otherBanks[0]

	for bid == n.ID {
		index := rand.Int() % len(otherBanks)
		bid = otherBanks[index]
	}

	tpk := n.getPubRoutingInfo(bid)

	payload := n.prepareOMsg(COSIGN, signedHash, tpk.Pk)

	print("sending aggregated signed coin and cosign counter:", newCounter)
	n.sendActive(payload, tpk.Port)
}

/*
	Decode the bytes from CoSign protocol into correspoding info.
 */
func decodeCNCosign(content []byte, counter uint16) (cnum uint64, cbytes []byte, sigs []byte, verifiers []string) {
	cbytes = content[:128]
	cnum = binary.BigEndian.Uint64(cbytes) // TODO: get real cnum

	sigs_vrfrs := content[128:]

	for i:=0; i<int(counter); i++ {
		b := sigs_vrfrs[i*144:(i+1)*144-1]
		ver,sig := b[128:], b[:128]
		sigs = append(sigs, sig...)
		verifiers = append(verifiers, strings.Trim(string(ver), "\x00"))
	}

	return
}

/*
	broadcast the txn to other banks with best effort.
 */
func (n *Node) broadcastTxn(txn blockChain.Txn, txnType rune) {
	for _, b := range bank.GetBankIDSet() {
		if b != n.ID{
			bpe := n.getPubRoutingInfo(b)
			p := n.prepareOMsg(TXNRECEIVE, append([]byte{byte(txnType)}, txn.ToBytes()...), bpe.Pk)
			go n.sendActive(p, bpe.Port)
		}
	}
}