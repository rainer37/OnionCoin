package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
	"github.com/rainer37/OnionCoin/blockChain"
	"time"
	"strings"
)

const BCOINSIZE = 128
const COSIGNTIMEOUT = 2

var exMap = map[string]chan []byte{} // channels for coin exchanging

/*
	bank processing coin exchange request.
	RAWCOIN(128) | BFID(8) | COINREWARD(128) |
	If valid coin received, sign the rawCoin and send it back.
	meanwhile starting coSign protocol to get coin published.
 */
func (n *Node) receiveRawCoin(payload []byte, senderID string) {
	//print("Make a wish")
	if len(payload) != BCOINSIZE * 2 + 16 {
		print("Wrong coin exchange len", len(payload))
		return }

	c := payload[BCOINSIZE+16:]

	rwcn := make([]byte, BCOINSIZE)
	copy(rwcn, payload[8:BCOINSIZE+8])
	bfid := make([]byte, 8)
	copy(bfid, payload[BCOINSIZE+8:BCOINSIZE+16])


	// check validity, if not, abort
	if !n.ValidateCoin(c, senderID) {
		print("invalid coin refuse signing it")
		return
	}

	//print("valid coin, continue")

	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, 0)
	tsBytes := payload[:8]

	pb := append(tsBytes, c...)
	pb = append(counterBytes, pb...)
	// start CoSign protocol with counter 0.

	n.coSignValidCoin(pb)

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
	Generate the genesis coin with my signed pk.
 */
func (n *Node) GetGenesisCoin() *coin.Coin {
	pkHash := sha256.Sum256(ocrypto.EncodePK(n.sk.PublicKey))
	gcoin := n.blindSign(pkHash[:])
	return coin.NewCoin(n.ID, gcoin)
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

	// TODO: use more than just the genesis coin
	gcoin := n.Vault.Withdraw(n.ID).Bytes()
	// print(len(gcoin))
	// banks := n.chain.GetCurBankIDSet()
	banks := currentBanks
	// print(banks)
	banksPk := []rsa.PublicKey{} // records which banks are helping

	counter := 0
	layers := 0
	rc := rwcn.ToBytes()

	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	for layers < blockChain.NUMCOSIGNER && counter < len(banks) {
		bid := banks[counter]
		bpe := n.getPubRoutingInfo(bid)

		if bpe == nil {
			return
		}

		//print("Requesting", bid, "for signing rawCoin")

		blindrwcn, bfid := BlindBytes(rc, &bpe.Pk)
		payload := append(blindrwcn, []byte(bfid)...)
		payload = append(payload, gcoin...) // TODO: append a real COINREWARD
		payload = append(tsBytes, payload...)

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

		//print("waiting for response from", bid)

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
		// print("New Coin Forged, Thanks Fellas!", len(rc))
		n.Deposit(coin.NewCoin(dstID, rc))
		// print(n.Vault.Coins)
	} else {
		//print("Not Enough Banks To Forge a Coin, Try Next Epoch")
	}
}

/*
	Upon received a valid coin, the bank signs the coin and pass it to other banks
	Till enough signatures gained, then publish it as a transaction.
	Does the last CoSigner solves the puzzle of blind signers?
 */
func (n *Node) coSignValidCoin(c []byte) {

	counter := binary.BigEndian.Uint16(c[:2]) // get cosign counter first 2 bytes

	cc := c[2:]

	hashAndIds := sha256.Sum256(cc[8:136]) // get the hash(32) of coin

	signedHash := n.blindSign(hashAndIds[:]) // sign the coin(128)
	signedHash = append(cc, signedHash...)

	newCounter := make([]byte, 2)
	binary.BigEndian.PutUint16(newCounter, counter+1)

	idBytes := make([]byte, IDLEN)
	copy(idBytes, n.ID)

	signedHash = append(signedHash, idBytes[:]...) // append verifier to it

	// when there is enough sigs gathered, try publish the txn.
	if counter+1 == blockChain.NUMCOSIGNER {
		//print("Enough verifiers got, publish it")
		t, cnum, cbytes, sigs, verifiers := decodeCNCosign(signedHash, counter+1)
		txn := blockChain.NewCNEXTxn(cnum, cbytes, t, sigs, verifiers)
		// TODO: go n.broadcastTxn(txn)
		ok := n.bankProxy.AddTxn(txn)
		if ok {
			//print("time to publish this block")
			// n.publishBlock()
		}
		return
	}


	signedHash = append(newCounter, signedHash...) // add updated counter to the head.cvx

	// randomly picks banks other than me
	bid := n.pickOneRandomBank()
	tpk := n.getPubRoutingInfo(bid)
	payload := n.prepareOMsg(COINCOSIGN, signedHash, tpk.Pk)

	//print("sending aggregated signed coin and cosign counter:", newCounter)
	n.sendActive(payload, tpk.Port)
}

/*
	Decode the bytes from CoSign protocol into correspoding info.
 */
func decodeCNCosign(content []byte, counter uint16) (ts int64, cnum uint64, cbytes []byte, sigs []byte, verifiers []string) {
	ts = int64(binary.BigEndian.Uint64(content[:10]))
	cbytes = content[8:136]
	cnum = binary.BigEndian.Uint64(cbytes) // TODO: get real cnum

	sigs_vrfers := content[136:]

	for i:=0; i<int(counter); i++ {
		b := sigs_vrfers[i*144:(i+1)*144-1]
		ver,sig := b[128:], b[:128]
		sigs = append(sigs, sig...)
		verifiers = append(verifiers, strings.Trim(string(ver), "\x00"))
	}

	return
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
