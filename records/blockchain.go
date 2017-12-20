package records

/*
	
	Blockchain core api for external usage.

	New_Chain()
	Mine_block()
	Verify_block()

*/

import(
	"fmt"
)

type Txn struct {
	time string
	signature string
	reward_pt float64
	id string
}

func New_Chain() {
	fmt.Println("New BlockChain created by you...")
}

func Mine_block() {

}

func Verify_block() {

}

func get_block_from_disk() {

}

