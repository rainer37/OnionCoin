package blockChain

/*
	Blockchain core api for external usage.
*/

import(
	"fmt"
	"crypto/rsa"
	"os"
	"log"
	"time"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"github.com/rainer37/OnionCoin/ocrypto"
	"sync"
	"sort"
)

const BKCHPREFIX = "[BKCH] "
const CHAINDIR = "chainData/"
const TINDEXDIR = CHAINDIR + "TIndex"
const NUMCOSIGNER = 2
const EPOCHLEN = 10
const PROPOSINGDELAY = 4
const PUSHINGDELAY = 2
const PROPOSINGTIME = EPOCHLEN - PROPOSINGDELAY
const PUSHTIME = EPOCHLEN - PUSHINGDELAY
const MAXNUMTXN = 500
const MATUREDIFF = 2

var slient = false
var GENESISBLOCK = Block{[]byte("ONCE UPON A TIME IN OLD ERA"), []byte("_OC_GENESIS_HASH_ON_18_MAR_2018_"), 0, 0, 0, nil, nil}

func print(str ...interface{}) {
	if slient {
		return
	}
	fmt.Print(BKCHPREFIX+" ")
	fmt.Println(str...)
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


type BlockChain struct {
	Blocks []*Block
	TIndex ChainIndex
}

var TI *ChainIndex

func InitBlockChain() *BlockChain {

	chain := new(BlockChain)
	chain.Blocks = []*Block{&GENESISBLOCK}

	if ok, _ := exists(TINDEXDIR); !ok {
		os.Create(TINDEXDIR)
	}

	chain.TIndex.mutex = &sync.Mutex{}

	if ok, _ := exists(CHAINDIR); !ok {
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
func (chain *BlockChain) loadChainAndIndex() {
	chain.Blocks = []*Block{&GENESISBLOCK}

	chain.TIndex.PKIndex = make(map[string]int64)
	chain.TIndex.CNIndex = make(map[string]int64)

	files, err := ioutil.ReadDir(CHAINDIR)
	checkErr(err)

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
func (chain *BlockChain) AddNewBlock(block *Block) bool {
	// print("Adding a new block")

	// should try pull the block again from the network first before publish it.
	prevBlock := chain.Blocks[chain.Size()-1]
	block.PrevHash = prevBlock.CurHash
	block.Depth = prevBlock.Depth + 1
	block.Ts = time.Now().Unix()
	block.CurHash = block.GetCurHash()
	chain.StoreBlock(block)

	return true
}

/*
	store the block bytes on disk in json, and update TIndex
 */
func (chain *BlockChain) StoreBlock(b *Block) {
	// print("writing block to disk, depth:", b.Depth)

	blockData, err := json.Marshal(b)
	checkErr(err)

	blockFileName := strconv.FormatInt(b.Depth, 10)

	if ok, _ := exists(CHAINDIR + blockFileName); ok {
		print("duplicate block depth detected!")
		return
	}

	f, err := os.Create(CHAINDIR + blockFileName)
	checkErr(err)
	f.Write(blockData)
	f.Close()

	chain.updateIndex(b)
	chain.Blocks = append(chain.Blocks, b)

	print("new block written, depth:", b.Depth, "Epoch:", b.Ts / EPOCHLEN)
}

/*
	return the pub-key associated with Id from blockchain.
 */
func (chain *BlockChain) GetPKFromChain(id string) *rsa.PublicKey {
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
	TODO: save the valid txns that are not in the new chain for future use.

 */
func (chain *BlockChain) TrimChain(start int64) {
	n := chain.Size()
	for i := start; i < n; i++ {
		os.Remove(CHAINDIR + strconv.FormatUint(uint64(i), 10))
	}
	chain.loadChainAndIndex()
}

func (chain *BlockChain) Size() int64 {
	return int64(len(chain.Blocks))
}

func (chain *BlockChain) GetBlock(index int64) *Block {
	return chain.Blocks[index]
}

func (chain *BlockChain) GetLastBlock() *Block {
	return chain.GetBlock(chain.Size()-1)
}

/*
	generate one block bytes in json
 */
func (chain *BlockChain) GenBlockBytes(index int64) []byte {
	blo := chain.GetBlock(index)
	b, err := json.Marshal(blo)
	checkErr(err)
	return b
}

/*
	update the TIndex for quick search for coinNum and pk.
 */
func (chain *BlockChain) updateIndex(b *Block) {
	for _, t := range b.Txns {
		//chain.TIndex.mutex.Lock()
		switch v := t.(type) {
		case PKRegTxn:
			chain.TIndex.PKIndex[v.Id] = b.Depth
		case CNEXTxn:
			chain.TIndex.CNIndex[strconv.FormatUint(v.CoinNum, 10)] = b.Depth
		case BCNRDMTxn:
			print("duo nothing now")
		default:
			print("what the fuck is this txn")
		}
		//chain.TIndex.mutex.Lock()
	}

	chain.TIndex.ChainLen = chain.Size()
	chain.TIndex.LastUpdate = time.Now().Unix()

	indexData, err := json.Marshal(chain.TIndex)
	checkErr(err)

	f, err := os.Create(TINDEXDIR)
	checkErr(err)
	f.Write(indexData)
	f.Close()
	print("Index updated")
}

/*
	read a block from disk and turn it into Block by given name(block index)
 */
func readABlock(name string) *Block {
	blockPath := CHAINDIR + name
	dat, err := ioutil.ReadFile(blockPath)
	checkErr(err)
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

func (chain *BlockChain) GetAllPeerIDs(max int64) []string {
	peers := []string{}
	for i,v := range chain.TIndex.PKIndex {
		if i == "FAKEID1338" || i == "FAKEID1339" {
			continue
		}

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

func (chain *BlockChain) GetNextBankIDSet() []string {
	nbanks := chain.GetBankSetWhen(time.Now().Unix() + EPOCHLEN)
	// print(nbanks)
	return nbanks
}

func (chain *BlockChain) GetBankSetWhen(t int64) []string {
	superBank := []string{"FAKEID1339", "FAKEID1338"} // TODO: super banks for now, remove them.

	curEpoch := t / EPOCHLEN

	matureLen := chain.GetMatureBlockLen(t)
	allPeers := chain.GetAllPeerIDs(matureLen)
	numBanks := len(allPeers) / 2

	//print("Everyone:",allPeers, "Mature", matureLen, "NumBanks", numBanks)

	theChosen := []string{}
	counter := 0
	i := 0
	for counter < numBanks {
		newChosen := allPeers[(curEpoch * int64(i+1)) % int64(len(allPeers))]
		if !contains(theChosen, newChosen) {
			theChosen = append(theChosen, newChosen)
			counter++
		}
		i++
	}
	// print("Epoch:", curEpoch, "Everyone:",allPeers, "Mature", matureLen, "Chosen:", theChosen)
	//print("Epoch:", curEpoch, "Chosen:", theChosen)
	return append(superBank, theChosen...)
}

func (chain *BlockChain) GetPrevBanks() []string {
	return chain.GetBankSetWhen(time.Now().Unix() - EPOCHLEN)
}

func (chain* BlockChain) GetMatureBlockLen(t int64) int64 {
	curEpoch := t / EPOCHLEN
	mLen := 0
	for i, b := range chain.Blocks {
		// print("$$", b.Ts / EPOCHLEN, curEpoch - 2)
		if b.Ts / EPOCHLEN < curEpoch - MATUREDIFF {
			mLen = i
		} else {
			break
		}
	}
	return int64(mLen)
}

func contains(arr []string, t string) bool {
	for _,v := range arr {if v == t {return true}}
	return false
}

/*
	check if the current chain length is matrue.
 */
func (chain *BlockChain) IsAlmostSyncd() bool {
	if chain.Size() > chain.GetMatureBlockLen(time.Now().Unix()) {
		return true
	}
	return false
}

/*
	Check if the new coin num has been used before.
 */
func IsFreeCoinNum(coinNum uint64) bool {
	if _, ok := TI.CNIndex[strconv.FormatUint(coinNum, 10)]; !ok {
		return true
	}
	return false
}

func (chain *BlockChain) String() (s string) {
	for _,v := range chain.Blocks {
		b, err := json.Marshal(v)
		checkErr(err)
		s += string(b) + "\n"
	}
	return
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func checkErr(err error){
	if err != nil { log.Fatal(err) }
}
