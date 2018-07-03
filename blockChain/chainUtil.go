package blockChain

import (
	"encoding/json"
	"github.com/rainer37/OnionCoin/util"
	"io/ioutil"
)

/*
	read a block from disk and turn it into Block by given name(block index)
 */
func readABlock(name string) *Block {
	blockPath := CHAINDIR + name
	dat, err := ioutil.ReadFile(blockPath)
	util.CheckErr(err)
	return DeMuxBlock(dat)
}

/*
	generate a block from the given bytes, dynamically determine the actual type of the txns.
 */
func DeMuxBlock (blockBytes []byte) *Block {
	var block *Block
	json.Unmarshal(blockBytes, &block)
	block.Txns = []Txn{}
	var b interface{}
	json.Unmarshal(blockBytes, &b)
	itemsMap := b.(map[string]interface{})

	for i , v := range itemsMap {
		if i == "Txns" {
			block.Txns = DemuxTxnsHelper(v.([]interface{}))
		}
	}

	return block
}

/*
	Dynamically Demux the txns in json to correct Txn.
 */
func DemuxTxns(txnsBytes []byte) (buffer []Txn) {
	var b interface{}
	json.Unmarshal(txnsBytes, &b)
	itemsMap := b.([]interface{})
	buffer = DemuxTxnsHelper(itemsMap)
	return
}

func DemuxTxnsHelper(itemsMap []interface{}) (buffer []Txn){
	for _, v := range itemsMap {
		tflag := 0
		for iii := range v.(map[string]interface{}) {
			if iii == "Pk" {
				tflag = 1
				break
			} else if iii == "CoinNum" {
				tflag = 2
				break
			}
		}
		if tflag == 1 {
			vvBytes, _ := json.Marshal(v)
			var ptxn PKRegTxn
			json.Unmarshal(vvBytes, &ptxn)
			buffer = append(buffer, ptxn)
		} else if tflag == 2 {
			vvBytes, _ := json.Marshal(v)
			var ctxn CNEXTxn
			json.Unmarshal(vvBytes, &ctxn)
			buffer = append(buffer, ctxn)
		}
	}
	return
}

func (chain *Chain) String() (s string) {
	for _,v := range chain.Blocks {
		b, err := json.Marshal(v)
		util.CheckErr(err)
		s += string(b) + "\n"
	}
	return
}