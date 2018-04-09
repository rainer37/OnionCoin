package node

import (
	"encoding/binary"
	"time"
	"math/rand"
	"github.com/rainer37/OnionCoin/blockChain"
	"encoding/json"
)

const SYNCPEROID = 5

type DepthHashPair struct {
	Depth int64
	Hash []byte
}

/*
	Try sync block chain with peers periodically.
	randomly picks a bank to send CHAINSYNC req with my current depth, and the hash of last block.
 */
func (n *Node) syncBlockChain() {

	ticker := time.NewTicker(time.Second * SYNCPEROID)
	for t := range ticker.C {
		print("TS:", t.Unix(), "CHAINLEN:", n.chain.Size(), "LASTHASH:", string(n.chain.GetLastBlock().GetCurHash()[:8]))

		// if !n.iamBank() { continue }

		go n.syncOnce()

	}
}

func (n *Node) syncOnce() {
	bid := n.pickOneRandomBank()
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(n.chain.Size()))
	b := n.chain.GetLastBlock()
	bhash := b.CurHash
	bpk := n.getPubRoutingInfo(bid)
	if bpk == nil { return }
	p := n.prepareOMsg(CHAINSYNC, append(buf, bhash[:]...) , bpk.Pk)
	n.sendActive(p, bpk.Port)
}

/*
	Select a random bank from bank set, but not me as bank.
 */
func (n *Node) pickOneRandomBank() string {
	banks := currentBanks
	bid := banks[rand.Int() % len(banks)] // picks random bank but not me
	for bid == n.ID {
		bid = banks[rand.Int() % len(banks)]
	}
	return bid
}

/*
	Upon received a request for blockChain Sync, reply with blocks.
	request : [peer current depth(8) | hash of last block(32)]
	reply : my current depth and the missing blocks starting with peer's current depth
 */
func (n *Node) chainSyncRequested(payload []byte, senderID string) {
	peerDepth := int64(binary.BigEndian.Uint64(payload[:8]))
	bHash := payload[8:]
	
	// if the peer has longer chain, ignore it.
	if peerDepth >= n.chain.Size() {
		// print(senderID, "has longer chain up to depth", peerDepth - 1)
		return
	}

	if string(n.chain.Blocks[peerDepth-1].CurHash) == string(bHash) {
		// if the chain is simply shorten

		print(senderID, "has short chain up to depth", peerDepth - 1)

		for i := peerDepth; i < n.chain.Size(); i++ {
			spk := n.getPubRoutingInfo(senderID)
			blocks := n.chain.GenBlockBytes(i)
			// print("sending block", i)

			myDepth := make([]byte, 8)
			binary.BigEndian.PutUint64(myDepth, uint64(i))

			p := n.prepareOMsg(CHAINSYNCACK, append(myDepth, blocks...), spk.Pk)
			n.sendActive(p, spk.Port)
		}
	} else {
		n.handleBranching(senderID)
		print("!!! found", senderID, "has minor branch")
	}
}

/*
	Upon received a sync ack from a bank.
	compare a the hash of the block to see if it could connect.
	if not, ignore it, otherwise store it.
	// TODO: the blocks may not came in orders.
 */
func (n *Node) chainSyncAckReceived(payload []byte, senderID string) {
	bDepth := binary.BigEndian.Uint64(payload[:8])
	print("received block with depth:", bDepth, "from", senderID)
	oneBlock := blockChain.DeMuxBlock(payload[8:])
	n.chain.StoreBlock(oneBlock)
}

/*
	upon detecting the peer has broken chain
	send some of partial local block hashes for the peer to trim the chain.
	current use first 8 bytes out of 32 bytes sha256.
 */
func (n *Node) handleBranching(senderID string) {
	a := []DepthHashPair{}
	i := n.chain.Size() - 100
	if i < 1 {
		i = 1
	}
	for ; i < n.chain.Size() ; i++ {
		b := n.chain.Blocks[i]
		bHash := b.CurHash
		dhp := DepthHashPair{i, bHash[:8]}
		a = append(a, dhp)
	}
	arr, err := json.Marshal(a)
	checkErr(err)
	spe := n.getPubRoutingInfo(senderID)
	p := n.prepareOMsg(CHAINREPAIR, arr, spe.Pk)
	n.sendActive(p, spe.Port)
}

/*
	receive a list of partial hashes from the peer with longer blockChain.
	Compare the hashes until finding the broken park, and then trim it.
 */
func (n *Node) chainRepairReceived(payload []byte, senderID string) {
	var arr []DepthHashPair
	json.Unmarshal(payload, &arr)

	if len(arr) == 0 { return }

	var start int64 = -1
	for _, v := range arr {
		if v.Depth < n.chain.Size() && string(n.chain.Blocks[v.Depth].CurHash[:8]) == string(v.Hash[:8]) {
			start = v.Depth
			// break
		}
	}
	if start != -1 {
		print("!!! i am broken at", start)
		n.chain.TrimChain(int64(start))
		n.syncOnce()
	}
}

/*
	Repair the blockChain by trimming at specific point
	and sync with others later.
 */
func (n *Node) repairChain(payload []byte) {
	trimmingStart := binary.BigEndian.Uint64(payload[:8])
	print("trimming everything starting at", trimmingStart + 1)
	n.chain.TrimChain(int64(trimmingStart + 1))
}

/*
	broadcast the txn to other banks with best effort.
 */
func (n *Node) broadcastTxn(txn blockChain.Txn, txnType rune) {
	banks := currentBanks
	for _, b := range banks {
		if b != n.ID{
			bpe := n.getPubRoutingInfo(b)
			p := n.prepareOMsg(TXNRECEIVE, append([]byte{byte(txnType)}, txn.ToBytes()...), bpe.Pk)
			n.sendActive(p, bpe.Port)
		}
	}
}

/*
	publish the new blocks to all other banks.
 */
func (n *Node) publishBlock() {

	// banks := n.chain.GetCurBankIDSet()
	banks := currentBanks
	for _, b := range banks {
		if b == n.ID { continue }
		bpe := n.getPubRoutingInfo(b)
		bbytes := n.chain.GenBlockBytes(n.chain.Size() - 1)
		depthByte := make([]byte, 8)
		binary.BigEndian.PutUint64(depthByte, uint64(n.chain.Size()-1))
		p := n.prepareOMsg(PUBLISHINGBLOCK, append(depthByte, bbytes...), bpe.Pk)
		n.sendActive(p, bpe.Port)
	}
}