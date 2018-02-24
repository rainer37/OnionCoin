package blockChain

/*
	Blockchain core api for external usage.
*/

import(
	"fmt"
	"crypto/rsa"
	"os"
)

const BKCHPREFIX = "BKCH"
const CHAINDIR = "chainData"
var GENESISBLOCK = NewBlock([]byte("0000000000000000"), []byte("WHERE OC BEGINS"))

func print(str ...interface{}) {
	fmt.Print(BKCHPREFIX+" ")
	fmt.Println(str...)
}

type BlockChain struct {
	blocks []*Block
}

func InitBlockChain() *BlockChain {
	chain := new(BlockChain)
	chain.blocks = append(chain.blocks, GENESISBLOCK)
	if ok, _ := exists(CHAINDIR); !ok {
		os.Mkdir(CHAINDIR, 0777)
	}
	return chain
}

func (chain *BlockChain) AddBlock(data []byte) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	nb := NewBlock(prevBlock.curHash, data)
	nb.Store() // write to disk
	chain.blocks = append(chain.blocks, nb)
}

/*
	return the pub-key associated with id from blockchain.
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