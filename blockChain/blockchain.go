package blockChain

/*
	Blockchain core api for external usage.
*/

import(
	"crypto/rsa"
	"os"
	"time"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"github.com/rainer37/OnionCoin/ocrypto"
	"sync"
	"github.com/rainer37/OnionCoin/util"
	"fmt"
	"sort"
)

const BKCHPREFIX = "[BKCH] "
const CHAINDIR = "chainData/"
const TINDEXDIR = CHAINDIR + "TIndex"
const GENSIS_HASH = "_OC_GENESIS_HASH_ON_18_MAR_2018_"

const silent = false
var GENESISBLOCK = Block{
	[]byte("ONCE UPON A TIME IN OLD ERA"),
[]byte(GENSIS_HASH), 0, 0, 0,
nil, nil }

type TxnSorter []Txn
func (a TxnSorter) Len() int           { return len(a) }
func (a TxnSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TxnSorter) Less(i, j int) bool {
	earlier := a[i].GetTS() < a[j].GetTS()
	sameTime := a[i].GetTS() == a[j].GetTS()
	smallContent := string(a[i].GetContent()) < string(a[j].GetContent())
	return  earlier || (sameTime && smallContent)
}

/*
	Fast Index from all blocks stored on disk.
	PKIndex : "id" 		: blockIndex
	CNIndex : "coinNum" : blockIndex
 */
type ChainIndex struct {
	LastUpdate int64
	ChainLen int64
	PKIndex map[string]int64
	CNIndex map[string]int64
	mutex *sync.Mutex
}

type Chain struct {
	Blocks []*Block
	TIndex ChainIndex
}

var TI *ChainIndex

func InitBlockChain() *Chain {

	chain := new(Chain)
	chain.Blocks = []*Block{&GENESISBLOCK}

	if ok, _ := util.Exists(TINDEXDIR); !ok {
		os.Create(TINDEXDIR)
	}

	chain.TIndex.mutex = &sync.Mutex{}

	if ok, _ := util.Exists(CHAINDIR); !ok {
		os.Mkdir(CHAINDIR, 0777)
		chain.TIndex.PKIndex = make(map[string]int64)
		chain.TIndex.CNIndex = make(map[string]int64)
	} else {
		chain.loadChainAndIndex()
		print("Current Stored Chain Len", chain.Size())
	}

	TI = &chain.TIndex
	return chain
}

/*
	Load all the blocks from disk, and add it to blockchain struct
	update TIndex with all info loaded
 */
func (chain *Chain) loadChainAndIndex() {
	chain.Blocks = []*Block{&GENESISBLOCK}

	chain.TIndex.PKIndex = make(map[string]int64)
	chain.TIndex.CNIndex = make(map[string]int64)

	files, err := ioutil.ReadDir(CHAINDIR)
	util.CheckErr(err)

	for _, f := range files {
		if f.Name() != "TIndex" && f.Name() != ".DS_Store"{
			b := readABlock(f.Name())
			if string(chain.GetLastBlock().CurHash) == string(b.PrevHash) {
				chain.Blocks = append(chain.Blocks, b)
				chain.updateIndex(b)
			} else {
				print("broken chain detected", f.Name())
			}
		}
	}
}

/*
	Add a new block to the blockChain.
	Update the prehash, txnHashes and depth before add it.
 */
func (chain *Chain) AddNewBlock(block *Block) {
	// chain.Blocks = append(chain.Blocks, block)
	chain.StoreBlock(block)
}

/*
	generate a block from transaction buffer and push it to the system.
 */
func (chain *Chain) GenNewBlock(txnBuffer []Txn) *Block {
	if len(txnBuffer) == 0 { return nil }
	print("Fresh Block with", len(txnBuffer), "txns")

	sort.Sort(TxnSorter(txnBuffer))
	newBlock := NewBlock(txnBuffer)

	prevBlock := chain.Blocks[chain.Size()-1]

	newBlock.PrevHash = prevBlock.CurHash
	newBlock.Depth = prevBlock.Depth + 1
	newBlock.Ts = time.Now().Unix()
	newBlock.CurHash = newBlock.GetCurHash()

	print("NewBlock Hash: [", string(newBlock.CurHash[:8]), "]")
	return newBlock
}

/*
	store the block bytes on disk in json, and update TIndex
 */
func (chain *Chain) StoreBlock(b *Block) {
	// print("writing block to disk, depth:", b.Depth)

	blockData, err := json.Marshal(b)
	util.CheckErr(err)

	blockFileName := strconv.FormatInt(b.Depth, 10)

	if ok, _ := util.Exists(CHAINDIR + blockFileName); ok {
		print("duplicate block depth detected!")
		return
	}

	ioutil.WriteFile(CHAINDIR + blockFileName, blockData, 0644)
	chain.updateIndex(b)
	chain.Blocks = append(chain.Blocks, b)

	print("new block written, depth:",
		b.Depth, "Epoch:", b.Ts / util.EPOCHLEN)
}

/*
	return the pub-key associated with Id from blockchain.
 */
func (chain *Chain) GetPKFromChain(id string) *rsa.PublicKey {
	tpk, ok := chain.TIndex.PKIndex[id]
	if !ok { return nil }

	blockPath := strconv.FormatInt(tpk, 10)
	b := readABlock(blockPath)

	for _, t := range b.Txns {
		switch v := t.(type) {
		case PKRegTxn:
			if v.Id == id {
				pk := ocrypto.DecodePK(v.Pk)
				return &pk
			}
		}
	}
	return nil
}

/*
	Trim the blockchain starting at some point, and delete them from disk
	then update chain struct and TIndex.
 */
func (chain *Chain) TrimChain(start int64) {
	n := chain.Size()
	for i := start; i < n; i++ {
		os.Remove(CHAINDIR + strconv.FormatUint(uint64(i), 10))
	}
	chain.loadChainAndIndex()
}

func (chain *Chain) Size() int64 { return int64(len(chain.Blocks)) }
func (chain *Chain) GetBlock(index int64) *Block { return chain.Blocks[index] }
func (chain *Chain) GetLastBlock() *Block { return chain.GetBlock(chain.Size()-1) }

/*
	generate one block bytes in json
 */
func (chain *Chain) GenBlockBytes(index int64) []byte {
	blo := chain.GetBlock(index)
	b, err := json.Marshal(blo)
	util.CheckErr(err)
	return b
}

/*
	update the TIndex for quick search for coinNum and pk.
 */
func (chain *Chain) updateIndex(b *Block) {
	for _, t := range b.Txns {
		switch v := t.(type) {
		case PKRegTxn:
			chain.TIndex.PKIndex[v.Id] = b.Depth
		case CNEXTxn:
			chain.TIndex.CNIndex[strconv.FormatUint(v.CoinNum, 10)] = b.Depth
		case BCNRDMTxn:
			print("duo nothing now")
		default:
			print("what the hell is this txn?")
		}
	}

	chain.TIndex.ChainLen = chain.Size()
	chain.TIndex.LastUpdate = time.Now().Unix()

	indexData, err := json.Marshal(chain.TIndex)
	util.CheckErr(err)

	f, err := os.Create(TINDEXDIR)
	util.CheckErr(err)
	f.Write(indexData)
	f.Close()
	// print("Index updated")
}

/*
	Check if the new coin num has been used before.
 */
func IsFreeCoinNum(coinNum uint64) bool {
	_, ok := TI.CNIndex[strconv.FormatUint(coinNum, 10)]
	return !ok
}

func print(str ...interface{}) {
	if silent { return }
	fmt.Print(BKCHPREFIX+" ")
	fmt.Println(str...)
}