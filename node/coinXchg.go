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
	"bytes"
	"encoding/json"
	"sync"
)

const BCOINSIZE = 128 // raw
const COSIGNTIMEOUT = 2

var exMap = map[string]chan []byte{} // channels for coin exchanging
var m = sync.RWMutex{}

/*
	bank processing coin exchange request.
	RAWCOIN(128) | BFID(8) | COINREWARD(128) |
	If valid coin received, sign the rawCoin and send it back.
	meanwhile starting coSign protocol to get coin published.
 */
func (n *Node) receiveRawCoin(payload []byte, senderID string) {
	//print("Make a wish")
	if len(payload) <= BCOINSIZE+ 16 {
		print("Wrong coin exchange len", len(payload))
		return }

	c := payload[BCOINSIZE+16:]
	cLen := len(payload) - BCOINSIZE - 16

	rwcn := make([]byte, BCOINSIZE)
	copy(rwcn, payload[8:BCOINSIZE+8])

	bfid := make([]byte, 8)
	copy(bfid, payload[BCOINSIZE+8: BCOINSIZE+16])


	// check validity, if not, abort
	if !n.ValidateCoin(c, senderID) {
		print("invalid coin refuse signing it")
		return
	}

	//print("valid coin, continue")

	counterBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(counterBytes, 0)
	tsBytes := payload[:8]

	cLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(cLenBytes, uint32(cLen))

	pb := append(tsBytes, cLenBytes...)
	pb = append(pb, c...)
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
	return coin.NewCoin(n.ID, gcoin, []string{n.ID})
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
	// print(rwcn.GetCoinNum(), string(gcoin))
	// print(len(gcoin))
	banks := currentBanks

	signerBanks := []string{} // records which banks are helping

	counter := 0
	layers := 0
	rc := rwcn.ToBytes()

	tsBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(tsBytes, uint64(time.Now().Unix()))

	for layers < blockChain.NUMCOSIGNER && counter < len(banks) {
		banks = currentBanks
		bid := banks[counter]
		bpe := n.getPubRoutingInfo(bid)

		if bpe == nil {
			return
		}

		// print("Requesting", bid, "for signing rawCoin")

		blindrwcn, bfid := BlindBytes(rc, &bpe.Pk)

		payload := bytes.Join([][]byte{tsBytes, blindrwcn, []byte(bfid), gcoin}, []byte{})

		fo := n.prepareOMsg(RAWCOINEXCHANGE, payload, bpe.Pk)

		m.Lock()
		exMap[bfid] = make(chan []byte)
		m.Unlock()

		n.sendActive(fo, bpe.Port)

		var realCoin []byte

		m.Lock()
		select{
		case reply := <-exMap[bfid]:
			realCoin = reply
			close(exMap[bfid])
			m.Unlock()
		case <-time.After(COSIGNTIMEOUT * time.Second):
			print(bid, "no response, try next bank")
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
			print("not equal after blindSign, bad bank!", bid, len(expected), len(rc))
			continue
		}

		rc = revealedCoin[:32] // reduce the size to 32 to avoid long asym enc.
		signerBanks = append(signerBanks, bid)
		layers++
	}

	if layers == blockChain.NUMCOSIGNER {
		// print("New Coin Forged, Thanks Fellas!", len(rc))
		n.Deposit(coin.NewCoin(dstID, rc, signerBanks))
		// print(n.Vault.Coins)
	} else {
		print("Not Enough Banks To Forge a Coin, Try Next Epoch")
	}
}

/*
	Upon received a valid coin, the bank signs the coin and pass it to other banks
	Till enough signatures gained, then publish it as a transaction.
	Does the last CoSigner solves the puzzle of blind signers?
	| counter(2) | ts(8) | cLen(4) | coin(cLen) | ... sigs
 */
func (n *Node) coSignValidCoin(c []byte) {

	counter := binary.BigEndian.Uint16(c[:2]) // get cosign counter first 2 bytes

	cc := c[2:]

	cLen := binary.BigEndian.Uint32(cc[8:12])
	hashAndIds := sha256.Sum256(cc[8 + 4: 8 + 4 + cLen]) // get the hash(32) of coin

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
		t, cnum, cbytes, sigs, verifiers := n.decodeCNCosign(signedHash, counter+1)
		var c coin.Coin
		json.Unmarshal(cbytes, &c)
		txn := blockChain.NewCNEXTxn(cnum, cbytes, t, sigs, verifiers)

		// go n.broadcastTxn(txn, blockChain.MSG)
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
func (n *Node) decodeCNCosign(content []byte, counter uint16) (ts int64, cnum uint64, cbytes []byte, sigs []byte, verifiers []string) {
	ts = int64(binary.BigEndian.Uint64(content[:8]))
	cLen := binary.BigEndian.Uint32(content[8:12])
	// cbytes = content[12 : 12 + cLen]
	cbytes = make([]byte, cLen)
	copy(cbytes, content[12 : 12 + cLen])

	var c coin.Coin
	json.Unmarshal(cbytes, &c)
	cnum, _ = n.getCoinNumAndIDHash(c)
	// print(string(c.String()))
	sigs_vrfers := content[12 + cLen:]

	for i:=0; i<int(counter); i++ {
		b := sigs_vrfers[i*(128 + 16):(i+1)*(128 + 16)-1]
		ver,sig := b[128:], b[:128]
		sigs = append(sigs, sig...)
		verifiers = append(verifiers, strings.Trim(string(ver), "\x00"))
	}

	return
}

func (n *Node) getCoinNumAndIDHash(c coin.Coin) (uint64, []byte) {
	content := c.Content
	signers := c.Signers

	for i:=len(signers)-1; i>=0; i-- {
		b := signers[i]
		bpe := n.getPubRoutingInfo(b)
		content = ocrypto.EncryptBig(&bpe.Pk, content)
	}

	// print(len(content), string(content))
	if len(content) != 32 + 8 {
		return 0, nil
	}
	return binary.BigEndian.Uint64(content[32:]), content[:32]
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
func (n *Node) ValidateCoin(coinBytes []byte, senderID string) bool {

	var ncoin coin.Coin
	json.Unmarshal(coinBytes, &ncoin)
	// print(string(ncoin.Bytes()))

	// first check if it is a genesis coin.
	spe := n.getPubRoutingInfo(senderID)
	encSPK := sha256.Sum256(ocrypto.EncodePK(spe.Pk))
	targetHash := ocrypto.EncryptBig(&spe.Pk, ncoin.Content)

	if string(encSPK[:]) == string(targetHash) {
		// print(senderID, "GCoin received")
		return true
	}

	// TODO: remove this
	return true
	// if not gcoin, check if the signers are in the same epoch, then check the signatures.
	whoWasBanks := n.chain.GetBankSetWhen(int64(ncoin.Epoch) * blockChain.EPOCHLEN)
	// print(ncoin.Signers)
	// print(whoWasBanks)

	for _, s := range(ncoin.Signers) {
		if !contains(whoWasBanks, s) {
			print("One of the signers is not supposed to be bank at that moment!")
			return false
		}
	}

	coinNum, idHash := n.getCoinNumAndIDHash(ncoin)
	print("CoinNum:", coinNum)
	h := sha256.Sum256([]byte(senderID))

	if string(idHash) != string(h[:]) {
		print("The coin id is not the senderID, sorry you have to use your own coin!")
		return false
	}

	// then check if the coinNum has been recorded in previous txns.
	if !blockChain.IsFreeCoinNum(coinNum) {
		print("Same Coin Num Found!")
		return false
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
