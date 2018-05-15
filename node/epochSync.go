package node

import (
	"time"
	"github.com/rainer37/OnionCoin/util"
	"github.com/rainer37/OnionCoin/ocrypto"
	"fmt"
	"encoding/json"
)

/*
	timer to check epoch change, update banksets, and start proposing timer.
 */
func (n *Node) epochTimer() {
	epochLen := int64(util.EPOCHLEN)

	defer func() {
		print("BOOM!\n\n\n\n")
	}()

	// wait until beginning of next epoch upon boot.
	now := time.Now().Unix()
	if now % epochLen != 0 {
		nextEpoch := (now / epochLen + 1) * epochLen
		diff := nextEpoch - now
		//print(diff)
		<-time.NewTimer(time.Duration(diff) * time.Second).C
	}

	epochTicker := time.NewTicker(time.Duration(epochLen) * time.Second)

	/*
		During beginning each epoch, log the stats.
		Then the banks will propose the block in the middle of epoch.
		Upon consensus achieved on major block, add it to chain.
		The banks for next epoch pull the new block from the chain.
	 */
	for t := range epochTicker.C {

		n.logStats(t)
		currentBanks = n.chain.GetCurBankIDSet()

		go func() {
			if n.iamBank() {
				n.bankProxy.SetStatus(true)

				// start proposing timer
				go n.proposeBlock()
				// start pushing timer
				go n.pushBlock()

			} else {
				n.bankProxy.SetStatus(false)
				n.syncOnce()

				if n.iamNextBank() {
					print("!!! i'm one of the next gen banks, so? !!!")
					go n.publishBlock()
				}
			}
		}()
	}
}

/*
	aggregate the txns from the local buffer,
	then propose to other banks, if receive new txns add them.
 */
func (n *Node) proposeBlock() {
	<-time.NewTimer(util.PROPOSINGTIME * time.Second).C

	HashCmpMap = make(map[string]int)
	n.syncOnce()

	for _, b := range currentBanks {
		if b == n.ID { continue }
		txnsBytes := n.getTxnsInBuffer()
		if string(txnsBytes) == "null" { continue }
		print("Time to propose my txns to", b)
		n.sendOMsgWithID(TXNAGGRE, txnsBytes, b)
	}
}

/*
	Send out the hash of local block to other.
	compare the local blocks with other hashes received,
	if the local is not the major block, discared.
	if it is, add it to the chain.
 */
func (n *Node) pushBlock() {
	<-time.NewTimer(util.PUSHTIME * time.Second).C

	nb := n.bankProxy.GenNewBlock()

	if nb == nil { return }

	HashCmpMap[string(nb.CurHash)] = 1

	go func() {
		for _, b := range currentBanks {
			if b == n.ID { continue }
			print("Time to push my block hash to", b)
			n.sendOMsgWithID(HASHCMP, nb.CurHash, b)
		}
	}()

	<-time.NewTimer(util.DECISIONDELAY * time.Second).C

	if n.bankProxy.IsMajorityHash(string(nb.CurHash)) {
		print("I have major block[" + string(nb.CurHash[:8]) + "]")
		n.chain.StoreBlock(nb)
	} else {
		print("I have minor block[" + string(nb.CurHash[:8]) + "]" +
			" wait for sync ***************************")
	}

	n.bankProxy.CleanBuffer()
}

/*
	print/log the stats for experiments.
 */
func (n *Node) logStats(t time.Time) {
	percent := float64(ocrypto.RSATime) /
		float64(time.Since(ela).Nanoseconds() / 1000000)
	fmt.Println(t.Unix() / util.EPOCHLEN, msgSendCount - bcCount ,
		omsgCount, pathLength, ocrypto.RSAStep, ocrypto.AESStep,
		ocrypto.RSATime, ocrypto.AESTime, percent * 100,"%")
	fmt.Println(currentBanks)
}

func (n *Node) pullBlock() {

}

func (n *Node) getTxnsInBuffer() []byte {
	txns, err := json.Marshal(n.bankProxy.GetTxnBuffer())
	util.CheckErr(err)
	return txns
}

func (n *Node) isSlientHours() bool {
	now := time.Now().Unix()
	nextEpoch := (now / util.EPOCHLEN + 1) * util.EPOCHLEN
	if now > nextEpoch - util.PROPOSINGDELAY {
		return true
	}
	return false
}

/*
	Check if the id given is a current bank.
 */
func (n* Node) checkBankStatus(id string) bool { return util.Contains(currentBanks, id) }
func (n *Node) iamBank() bool { return n.checkBankStatus(n.ID) }
func (n *Node) isBank(id string) bool { return n.checkBankStatus(id) }
func (n *Node) iamNextBank() bool { return util.Contains(n.chain.GetNextBankSet(), n.ID) }