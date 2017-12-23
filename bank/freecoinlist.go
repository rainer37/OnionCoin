package bank

//TODO: new data structure. ex. tree
type FreeCoinList struct {
	FreeCoins []uint64
}

func NewFCL() *FreeCoinList {
	fcl := new(FreeCoinList)
	return fcl
}

func (fcl *FreeCoinList) insert(cid uint64) {
	fcl.FreeCoins = append(fcl.FreeCoins, cid)
	println("Inserting Coin :",cid)
}

func (fcl *FreeCoinList) remove(cid uint64) {
	println("Removing Coin :",cid)
	//TODO: remove by id
}