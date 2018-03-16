package node

import (
	"encoding/binary"
	"time"
	"fmt"
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
		fmt.Println("$$$", t.Unix(), "CHAINLEN:", n.chain.Size())

		go func() {
			bid := n.pickOneRandomBank()
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(n.chain.Size()))
			b := n.chain.GetLastBlock()
			bhash := b.CurHash
			bpk := n.getPubRoutingInfo(bid)
			if bpk == nil { return }
			p := n.prepareOMsg(CHAINSYNC, append(buf, bhash[:]...) , bpk.Pk)
			n.sendActive(p, bpk.Port)
		}()

	}
}

/*
	Select a random bank from bank set, but not me as bank.
 */
func (n *Node) pickOneRandomBank() string {
	banks := n.chain.GetBankIDSet()
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

	myDepth := make([]byte, 8)
	binary.BigEndian.PutUint64(myDepth, uint64(n.chain.Size()))

	// if the peer has longer chain, ignore it.
	if peerDepth >= n.chain.Size() {
		print(senderID, "has longer chain", peerDepth)
		return
	}

	if string(n.chain.Blocks[peerDepth-1].CurHash) == string(bHash) {
		// if the chain is simply shorten

		print(senderID, "has short chain up to depth", peerDepth)

		for i := peerDepth; i < n.chain.Size(); i++ {
			spk := n.getPubRoutingInfo(senderID)
			blocks := n.chain.GenBlockBytes(i)
			p := n.prepareOMsg(CHAINSYNCACK, append(myDepth, blocks...), spk.Pk)
			n.sendActive(p, spk.Port)
		}
	} else {

		print("!!! found", senderID, "has minor branch")

		n.handleBranching(senderID)
		// or the chain is a minor branch, send REPAIR
	}


}

/*
	Upon received a sync ack from a bank.
	compare a the hash of the block to see if it could connect.
	if not, ignore it, otherwise store it.
 */
func (n *Node) chainSyncAckReceived(payload []byte, senderID string) {
	peerMaxDepth := binary.BigEndian.Uint64(payload[:8])
	print("This peer has chain up to depth", peerMaxDepth)
	oneBlock := blockChain.DeMuxBlock(payload[8:])
	n.chain.StoreBlock(oneBlock)
}

// TODO: append more partial hashes of blocks.
func (n *Node) handleBranching(senderID string) {
	a := []DepthHashPair{}
	i := n.chain.Size() - 3
	if i < 1 {
		i = 1
	}
	for ; i < n.chain.Size() ; i++ {
		b := n.chain.Blocks[i]
		bHash := b.CurHash
		dhp := DepthHashPair{i, bHash}
		a = append(a, dhp)
	}
	arr, err := json.Marshal(a)
	checkErr(err)
	spe := n.getPubRoutingInfo(senderID)
	p := n.prepareOMsg(CHAINREPAIR, arr, spe.Pk)
	n.sendActive(p, spe.Port)
}


func (n *Node) chainRepairReceived(payload []byte, senderID string) {
	var arr []DepthHashPair
	json.Unmarshal(payload, &arr)

	if len(arr) == 0 { return }

	var start int64 = 1
	for _, v := range arr {
		if v.Depth < n.chain.Size() && string(n.chain.Blocks[v.Depth].CurHash) == string(v.Hash) {
			start = v.Depth
			break
		}
	}

	print("!!! i am broken at", start)
	n.chain.TrimChain(int64(start + 1))
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
	publish the new blocks to all other banks.
 */
func (n *Node) publishBlock() {

	banks := n.chain.GetBankIDSet()

	for _, b := range banks {
		if b == n.ID {
			continue
		}

		bpe := n.getPubRoutingInfo(b)
		bbytes := n.chain.GenBlockBytes(n.chain.Size() - 1)
		depthByte := make([]byte, 8)
		binary.BigEndian.PutUint64(depthByte, uint64(n.chain.Size()-1))
		p := n.prepareOMsg(PUBLISHINGBLOCK, append(depthByte, bbytes...), bpe.Pk)
		n.sendActive(p, bpe.Port)

	}
}