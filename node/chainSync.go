package node

import (
	"encoding/binary"
	"time"
	"fmt"
	"github.com/rainer37/OnionCoin/bank"
	"math/rand"
	"github.com/rainer37/OnionCoin/blockChain"
	"encoding/json"
)

type DepthHashPair struct {
	Depth int64
	Hash []byte
}

/*
	Try sync block chain with peers periodically.
	randomly picks a bank to send CHAINSYNC req with my current depth.
 */
func (n *Node) syncBlockChain() {
	// go n.blockChainEventQueue()
	if n.ID == "FAKEID1338" {
		return
	}
	ticker := time.NewTicker(time.Millisecond * 5000)
	for t := range ticker.C {
		fmt.Println("Tick at", t.Unix())
		go func() {
			banks := bank.GetBankIDSet()
			bid := banks[rand.Int() % len(banks)] // picks random bank but not me
			for bid == n.ID {
				bid = banks[rand.Int() % len(banks)]
			}
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(n.chain.Size()))
			bpk := n.getPubRoutingInfo(bid)
			p := n.prepareOMsg(CHAINSYNC, buf, bpk.Pk)
			n.sendActive(p, bpk.Port)
		}()
	}
}

/*
	Upon received a request for blockChain Sync, reply with blocks.
	request : peer current depth
	reply : my current depth and the missing blocks starting with peer's current depth
 */
func (n *Node) chainSyncRequested(payload []byte, senderID string) {
	peerDepth := int64(binary.BigEndian.Uint64(payload))
	n.chainSyncHelper(peerDepth, senderID)
}

func (n *Node) chainSyncHelper(peerDepth int64, senderID string) {
	myDepth := make([]byte, 8)
	binary.BigEndian.PutUint64(myDepth, uint64(n.chain.Size()))

	for i := peerDepth; i < n.chain.Size(); i++ {
		spk := n.getPubRoutingInfo(senderID)
		blocks := n.chain.GenBlockBytes(i)
		p := n.prepareOMsg(CHAINSYNCACK, append(myDepth, blocks...), spk.Pk)
		n.sendActive(p, spk.Port)
		// break
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
	// print(oneBlock.Depth, n.chain.Size())

	if oneBlock.Depth == n.chain.Size() && string(oneBlock.PrevHash) == string(n.chain.GetLastBlock().CurHash) {
		print("smooth chain, continue")
		n.chain.StoreBlock(oneBlock)
	} else {
		print("broken chain")
		n.handleBranching(senderID)
	}

}

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
	print(len(arr))

	if len(arr) == 0 { return }

	var start int64 = 1
	for _, v := range arr {
		if string(n.chain.Blocks[v.Depth].CurHash) == string(v.Hash) {
			start = v.Depth
			break
		}
	}

	startByte := make([]byte, 8)
	binary.BigEndian.PutUint64(startByte, uint64(start))
	spe := n.getPubRoutingInfo(senderID)
	p := n.prepareOMsg(CHAINREPAIRREPLY, startByte, spe.Pk)
	n.sendActive(p, spe.Port)

	// n.chainSyncHelper(start, senderID)
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

	banks := bank.GetBankIDSet()

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