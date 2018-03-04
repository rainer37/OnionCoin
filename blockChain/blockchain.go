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
)

const BKCHPREFIX = "[BKCH] "
const CHAINDIR = "chainData"

var GENESISBLOCK = NewBlock([]Txn{})

func print(str ...interface{}) {
	fmt.Print(BKCHPREFIX+" ")
	fmt.Println(str...)
}

type BlockChain struct {
	blocks []*Block
}

func InitBlockChain() *BlockChain {
	chain := new(BlockChain)
	GENESISBLOCK.CurHash = []byte("GENESIS_HASH_ON_MAR_2018")
	chain.blocks = append(chain.blocks, GENESISBLOCK)
	if ok, _ := exists(CHAINDIR); !ok {
		os.Mkdir(CHAINDIR, 0777)
	}
	return chain
}

func (chain *BlockChain) AddBlock(block *Block) {
	print("Adding a block")
	prevBlock := chain.blocks[len(chain.blocks)-1]

	block.PrevHash = prevBlock.CurHash
	block.Depth = chain.getNextDepth()
	block.Ts = time.Now().Unix()
	block.CurHash = block.GetCurHash()
	block.Store() // write to disk
	chain.blocks = append(chain.blocks, block)
}
/*
func (chain *BlockChain) AddBlock(data []byte) {
	print("Add a block")
	prevBlock := chain.blocks[len(chain.blocks)-1]
	nb := NewBlock(prevBlock.CurHash, data)
	nb.Depth = chain.getNextDepth()
	nb.Store() // write to disk
	chain.blocks = append(chain.blocks, nb)
}
*/
func (chain *BlockChain) getNextDepth() int64 {
	return int64(len(chain.blocks))
}
/*
	return the pub-key associated with Id from blockchain.
 */
func GetPKFromChain(id string) *rsa.PublicKey {
	return nil
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
