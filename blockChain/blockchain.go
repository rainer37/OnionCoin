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
)

const BKCHPREFIX = "[BKCH] "
const CHAINDIR = "chainData/"

var GENESISBLOCK = NewBlock([]Txn{})

func print(str ...interface{}) {
	fmt.Print(BKCHPREFIX+" ")
	fmt.Println(str...)
}

type BlockChain struct {
	Blocks []*Block
}

func InitBlockChain() *BlockChain {
	chain := new(BlockChain)
	GENESISBLOCK.CurHash = []byte("GENESIS_HASH_ON_MAR_2018")
	chain.Blocks = append(chain.Blocks, GENESISBLOCK)
	if ok, _ := exists(CHAINDIR); !ok {
		os.Mkdir(CHAINDIR, 0777)
	}
	return chain
}

func (chain *BlockChain) AddBlock(block *Block) bool {
	print("Adding a block")
	// should try pull the block again from the network first before publish it.
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	block.PrevHash = prevBlock.CurHash
	block.Depth = chain.getNextDepth()
	block.Ts = time.Now().Unix()
	block.CurHash = block.GetCurHash()

	chain.Store(block) // write to disk
	chain.Blocks = append(chain.Blocks, block)
	return true
}

func (chain *BlockChain) getNextDepth() int64 {
	return int64(len(chain.Blocks))
}

/*
	return the pub-key associated with Id from blockchain.
 */
func GetPKFromChain(id string) *rsa.PublicKey {
	return nil
}

func (chain *BlockChain) Size() int64 {
	return int64(len(chain.Blocks))
}

func (chain *BlockChain) GetBlock(index int64) *Block {
	return chain.Blocks[index]
}

func (chain *BlockChain) GetLastBlock() *Block {
	return chain.GetBlock(int64(len(chain.Blocks)-1))
}

/*
	Store blockData to Disk.
 */
func (chain *BlockChain) Store(b *Block) {
	print("writing block to disk")
	// print(b)
	blockData, err := json.Marshal(b)
	checkErr(err)

	blockFileName := strconv.FormatInt(b.Depth, 10)

	if ok, _ := exists(CHAINDIR+blockFileName); ok {
		print("duplicate block depth detected!")
		return
	}

	f, err := os.Create(CHAINDIR+blockFileName)
	checkErr(err)

	f.Write(blockData)
	f.Close()
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
