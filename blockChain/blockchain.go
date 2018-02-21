package blockChain

/*
	Blockchain core api for external usage.
*/

import(
	"fmt"
	"crypto/rsa"
)

const BKCHPREFIX = "BKCH"
var GENESISBLOCK = NewBlock([]byte{}, []byte("WHERE OC BEGINS"))

func print(str ...interface{}) {
	fmt.Print(BKCHPREFIX+" ")
	fmt.Println(str...)
}

type BlockChain struct {
	blocks []*Block
}

func (chain *BlockChain) AddBlock(data []byte) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	nb := NewBlock(prevBlock.curHash, data)
	chain.blocks = append(chain.blocks, nb)
}

func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{GENESISBLOCK}}
}

/*
	return the pub-key associated with id from blockchain.
 */
func GetPKFromChain(id string) *rsa.PublicKey {
	return nil
}