package records

/*
	
	Blockchain core api for external usage.

*/

import(
	"fmt"
	"time"
)

const BKCH_PREFIX = "BKCH"

func print(str ...interface{}) {
	fmt.Print(BKCH_PREFIX+" ")
	fmt.Println(str...)
}

type Txn struct {
	coinNum uint64
	redeemer string
	signers []string // ids of signing bank nodes
	ts time.Time
}

type CoinBlock struct {
	curHash []byte
	preHash []byte
	txns [32]Txn
	ts time.Time
}

func NewCoinBlock() {
	print("creating new CoinBlock")
}

func NewTxn() {
	print("creating new Transaction")
}

func get_block_from_disk() {

}

