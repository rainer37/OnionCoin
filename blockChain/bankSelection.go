package blockChain

import (
	"sort"
	"time"
	"github.com/rainer37/OnionCoin/util"
)

func (chain *BlockChain) GetAllPeerIDs(max int64) []string {
	peers := []string{}
	for i,v := range chain.TIndex.PKIndex {
		//if i == "FAKEID1338" || i == "FAKEID1339" {
		//	continue
		//}

		if v <= max {
			peers = append(peers, i)
		}
	}
	sort.Strings(peers)
	return peers
}

func (chain *BlockChain) GetCurBankIDSet() []string {
	return chain.GetBankSetWhen(time.Now().Unix())
}

func (chain *BlockChain) GetNextBankSet() []string {
	nbanks := chain.GetBankSetWhen(time.Now().Unix() + util.EPOCHLEN)
	// print(nbanks)
	return nbanks
}

func (chain *BlockChain) GetBankSetWhen(t int64) []string {
	superBank := []string{"FAKEID1339", "FAKEID1338"}
	// TODO: super banks for now, remove them.
	// return superBank
	curEpoch := t / util.EPOCHLEN

	matureLen := chain.GetMatureBlockLen(t)
	allPeers := chain.GetAllPeerIDs(matureLen)
	// fmt.Println(allPeers, matureLen)
	numBanks := len(allPeers) / 2
	// return append(superBank, allPeers...)
	theChosen := []string{}
	counter := 0
	i := 0
	for counter < numBanks {
		newChosen := allPeers[(curEpoch * int64(i+1)) % int64(len(allPeers))]
		if !util.Contains(theChosen, newChosen) {
			theChosen = append(theChosen, newChosen)
			counter++
		}
		i++
	}
	if matureLen < 2 {
		return superBank
	}
	// print("Epoch:", curEpoch, "Everyone:",
	// allPeers, "Mature", matureLen, "Chosen:", theChosen)
	return theChosen
	//	return append(superBank, theChosen...)
}

func (chain *BlockChain) GetPrevBanks() []string {
	return chain.GetBankSetWhen(time.Now().Unix() - util.EPOCHLEN)
}

func (chain* BlockChain) GetMatureBlockLen(t int64) int64 {
	curEpoch := t / util.EPOCHLEN
	mLen := 0
	for i, b := range chain.Blocks {
		// print("$$", b.Ts / EPOCHLEN, curEpoch - 2)
		if b.Ts / util.EPOCHLEN < curEpoch - util.MATUREDIFF {
			mLen = i
		} else {
			break
		}
	}
	return int64(mLen)
}