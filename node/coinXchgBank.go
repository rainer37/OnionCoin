package node

import (
	"github.com/rainer37/OnionCoin/coin"
	"github.com/rainer37/OnionCoin/ocrypto"
	"encoding/binary"
	"encoding/json"
	"sync"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/blockChain"
)

const BCOINSIZE = 128 // raw

var exMap = map[string]chan []byte{} // channels for coin exchanging
var m = sync.RWMutex{}

/*
	bank processing coin exchange request.
	RAWCOIN(128) | BFID(8) | COINREWARD(128) |
	If valid coin received, sign the rawCoin and send it back.
	meanwhile starting coSign protocol to get coin published.
 */
func (n *Node) receiveRawCoin(payload []byte, senderID string) {

	c := payload[BCOINSIZE + 16:]
	cLen := len(payload) - BCOINSIZE - 16

	rwcn := util.NewBytes(BCOINSIZE, payload[8:BCOINSIZE+8])

	bfid := util.NewBytes(8, payload[BCOINSIZE+8: BCOINSIZE+16])

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

	pb := util.JoinBytes([][]byte{counterBytes, tsBytes, cLenBytes, c})

	// start CoSign protocol with counter 0.
	n.coSignValidCoin(pb)

	newCoin := n.blindSign(rwcn)

	n.sendOMsgWithID(RAWCOINSIGNED, append(newCoin, bfid...), senderID)
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

	cLen := binary.BigEndian.Uint32(cc[8 : 12])
	hashAndIds := util.Sha(cc[8 + 4: 8 + 4 + cLen]) // get the hash(32) of coin

	signedHash := n.blindSign(hashAndIds[:]) // sign the coin(128)
	signedHash = append(cc, signedHash...)

	newCounter := make([]byte, 2)
	binary.BigEndian.PutUint16(newCounter, counter+1)

	idBytes := util.NewBytes(util.IDLEN, []byte(n.ID))

	signedHash = append(signedHash, idBytes[:]...) // append verifier to it

	// when there is enough sigs gathered, try publish the txn.
	if counter+1 == util.NUMCOSIGNER {
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

	n.sendOMsgWithID(COINCOSIGN, signedHash, bid)
}

/*
	Decode the bytes from CoSign protocol into correspoding info.
 */
func (n *Node) decodeCNCosign(content []byte, counter uint16) (ts int64, cnum uint64,
	cbytes []byte, sigs []byte, verifiers []string) {
	ts = int64(binary.BigEndian.Uint64(content[:8]))
	cLen := binary.BigEndian.Uint32(content[8:12])
	// cbytes = content[12 : 12 + cLen]

	cbytes = util.NewBytes(int(cLen), content[12 : 12 + cLen])

	var c coin.Coin
	json.Unmarshal(cbytes, &c)
	cnum, _ = n.getCoinNumAndIDHash(c)
	// print(string(c.String()))
	sigsVrfers := content[12 + cLen:]

	for i:=0; i<int(counter); i++ {
		b := sigsVrfers[i*(128 + 16):(i+1)*(128 + 16)-1]
		ver,sig := b[128:], b[:128]
		sigs = append(sigs, sig...)
		verifiers = append(verifiers, util.Strip(ver))
	}

	return
}

func (n *Node) getCoinNumAndIDHash(c coin.Coin) (uint64, []byte) {
	content := c.Content
	signers := c.Signers
	// print(len(content))
	buf := []byte{}

	for i:=0;i<len(signers);i++ {
		b := signers[i]
		bpe := n.getPubRoutingInfo(b)
		buf = append(buf, ocrypto.EncryptBig(&bpe.Pk, content[i*128:(i+1)*128])...)
	}

	if len(buf) != 32 + 8 {
		print("The size of coin should be 40", len(buf))
		return 0, nil
	}
	return binary.BigEndian.Uint64(buf[32:]), buf[:32]
}

func (n *Node) validcoinwrap(coinBytes []byte, senderID string) bool {
	n.ValidateCoin(coinBytes, senderID)
	return true
}

func (n *Node) isGCoin(ncoin coin.Coin, coinNum uint64, senderID string) bool {
	if coinNum != 0 { return false }
	for _, s := range ncoin.Signers {
		if s != senderID { return false }
	}
	print(senderID, "'s GCoin received")
	return true
}

/*
	validate the coin received by decrypting the coin multiple times
	then join the bytes and check against coinNum and senderID.
 */
func (n *Node) ValidateCoin(coinBytes []byte, senderID string) bool {

	var ncoin coin.Coin
	json.Unmarshal(coinBytes, &ncoin)

	// TODO: remove this
	// return true
	// if not gcoin, check if the signers are in the same epoch,
	// then check the signatures.
	whoWasBanks := n.chain.GetBankSetWhen(int64(ncoin.Epoch) * util.EPOCHLEN)
	// print(ncoin.Signers)
	// print(whoWasBanks)

	coinNum, idHash := n.getCoinNumAndIDHash(ncoin)
	print("CoinNum:", coinNum, senderID)

	h := util.Sha([]byte(senderID))

	if string(idHash) != string(h[:]) {
		print("The coin id is not the senderID," +
			" sorry you have to use your own coin!")
		return false
	}

	// first check if it is a genesis coin.
	if n.isGCoin(ncoin, coinNum, senderID) { return true }

	for _, s := range(ncoin.Signers) {
		if !util.Contains(whoWasBanks, s) {
			print("One of the signers is not" +
				" supposed to be bank at that moment!")
			return false
		}
	}

	// then check if the coinNum has been recorded in previous txns.
	if !blockChain.IsFreeCoinNum(coinNum) {
		print("Same Coin Num Found!")
		return false
	}

	return true
}