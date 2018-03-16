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
)

const BKCHPREFIX = "[BKCH] "
const CHAINDIR = "chainData/"
const TINDEXDIR = CHAINDIR + "TIndex"
const NUMCOSIGNER = 2

var slient = false
var GENESISBLOCK = Block{[]byte("ONCE UPON A TIME IN OLD ERA"), []byte("GENESIS_HASH_ON_MAR_2018"), 0, 0, 0, nil, nil}

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

	return chain
}

/*
	Load all the blocks from disk, and add it to blockchain struct
	update TIndex with all info loaded
 */
func (chain *BlockChain) loadChainAndIndex() {
	// read TIndex from disk if there is one.
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
				print("broken chain detected")
			}
		}
	}
}

/*
	Add a new block to the blockChain.
	Update the prehash, txnHashes and depth before add it.
 */
func (chain *BlockChain) AddNewBlock(block *Block) bool {
	print("Adding a new block")

	// should try pull the block again from the network first before publish it.
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	block.PrevHash = prevBlock.CurHash
	block.Depth = prevBlock.Depth + 1
	block.Ts = time.Now().Unix()

	// hashes := bytes.Join(block.TxnHashes, []byte{})
	// depth := make([]byte, 8)
	// binary.BigEndian.PutUint64(depth, uint64(block.Depth))

	// payload := bytes.Join([][]byte{depth, block.PrevHash, hashes}, []byte{})
	// print(len(payload))

	block.CurHash = block.GetCurHash()
	chain.StoreBlock(block)

	return true
}

/*
	store the block bytes on disk in json, and update TIndex
 */
func (chain *BlockChain) StoreBlock(b *Block) {
	print("writing block to disk, depth:", b.Depth)

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

	print("new block written, depth:", b.Depth)

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
	// print("Index updated")
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
	generate a block from the given bytes, dynamically determine the actual type of the block
 */
func DeMuxBlock (blockBytes []byte) *Block {
	var block *Block
	json.Unmarshal(blockBytes, &block)
	block.Txns = block.Txns[:0]
	var b interface{}
	json.Unmarshal(blockBytes, &b)
	itemsMap := b.(map[string]interface{})

	for i , v := range itemsMap {
		if i == "Txns" {
			for _ , vv := range v.([]interface{}) {
				tflag := 0
				for iii := range vv.(map[string]interface{}) {
					if iii == "Pk" {
						tflag = 1
						break
					} else if iii == "CoinNum" {
						tflag = 2
						break
					}
				}
				if tflag == 1 {
					vvBytes, _ := json.Marshal(vv)
					var ptxn PKRegTxn
					json.Unmarshal(vvBytes, &ptxn)
					block.Txns = append(block.Txns, ptxn)
				} else if tflag == 2 {
					vvBytes, _ := json.Marshal(vv)
					var ctxn CNEXTxn
					json.Unmarshal(vvBytes, &ctxn)
					block.Txns = append(block.Txns, ctxn)
				}
			}
		}
	}
	return block
}

/*
	generate one block bytes in json
 */
func (chain *BlockChain) GenBlockBytes(start int64) []byte {
	blocks := chain.Blocks[start]
	b, err := json.Marshal(blocks)
	checkErr(err)
	return b
}

func (chain *BlockChain) GetAllPeerIDs() []string {
	peers := []string{}
	for i,v := range chain.TIndex.PKIndex {
		if i == "FAKEID1338" || i == "FAKEID1339" {
			continue
		}

		remainder := (time.Now().Unix() + int64(v)) % 3
		//print("TS:", time.Now().Unix(), "ID:", int64(v), "REM:", remainder)
		if  remainder == 1 {
			peers = append(peers, i)
			//print(i, "is one of the bank")
		}

	}
	return peers
}

func (chain *BlockChain) GetBankIDSet() []string {
	return chain.GetBankSetWhen(time.Now().Unix())
}

func (chain *BlockChain) GetBankSetWhen(t int64) []string {
	superBank := []string{"FAKEID1339", "FAKEID1338"}
	// superBank = append(superBank, chain.GetAllPeerIDs()...)
	return superBank
}

// TODO: check if current chain is almost syncd
func (chain *BlockChain) IsAlmostSyncd() bool {
	print("Current chainLength:", chain.Size())
	if chain.Size() > 1 {
		return true
	}
	return false
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
